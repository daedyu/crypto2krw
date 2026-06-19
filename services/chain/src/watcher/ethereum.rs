use anyhow::{Context, Result};
use reqwest::Client;
use serde::Deserialize;
use serde_json::{json, Value};
use std::time::Duration;
use tokio::time::sleep;
use tracing::{info, warn};

use crate::idempotency::Store;
use crate::publisher::{DepositDetectedData, KafkaPublisher};
use crate::registry::AddressRegistry;

const NETWORK: &str = "ETHEREUM";
const POLL_INTERVAL: Duration = Duration::from_secs(15);
const REQUIRED_CONFIRMATIONS: u64 = 12;
// Transfer(address indexed from, address indexed to, uint256 value)
const TRANSFER_TOPIC: &str =
    "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef";
// 1 ETH = 1e18 wei
const WEI_PER_ETH: u128 = 1_000_000_000_000_000_000;
// USDT-ERC20 has 6 decimals
const USDT_DECIMALS: u128 = 1_000_000;

pub async fn run(
    rpc_url: String,
    usdt_contract: String,
    registry: AddressRegistry,
    idempotency: Store,
    publisher: KafkaPublisher,
) -> Result<()> {
    info!("ethereum watcher starting, rpc={rpc_url}");
    let client = Client::new();
    let mut last_processed_block: Option<u64> = None;

    loop {
        let addresses = registry.get_addresses(NETWORK).await.unwrap_or_default();
        if addresses.is_empty() {
            sleep(POLL_INTERVAL).await;
            continue;
        }

        match get_latest_block_number(&client, &rpc_url).await {
            Ok(latest) => {
                let safe_block = latest.saturating_sub(REQUIRED_CONFIRMATIONS);
                let from_block = last_processed_block
                    .map(|b| b + 1)
                    .unwrap_or(safe_block);

                if from_block <= safe_block {
                    if let Err(e) = scan_range(
                        &client,
                        &rpc_url,
                        &usdt_contract,
                        &addresses,
                        &idempotency,
                        &publisher,
                        from_block,
                        safe_block,
                    )
                    .await
                    {
                        warn!("ethereum scan error blocks {from_block}..{safe_block}: {e}");
                    } else {
                        last_processed_block = Some(safe_block);
                    }
                }
            }
            Err(e) => warn!("eth_blockNumber failed: {e}"),
        }

        sleep(POLL_INTERVAL).await;
    }
}

async fn scan_range(
    client: &Client,
    rpc_url: &str,
    usdt_contract: &str,
    addresses: &[String],
    idempotency: &Store,
    publisher: &KafkaPublisher,
    from_block: u64,
    to_block: u64,
) -> Result<()> {
    // 1. USDT-ERC20 Transfer 이벤트 조회 (eth_getLogs)
    let address_topics: Vec<String> = addresses
        .iter()
        .map(|a| address_to_topic(a))
        .collect();

    let logs_res = eth_get_logs(
        client,
        rpc_url,
        usdt_contract,
        &[TRANSFER_TOPIC.to_string()],
        &address_topics,
        from_block,
        to_block,
    )
    .await
    .context("eth_getLogs for USDT")?;

    for log in logs_res {
        process_usdt_log(log, addresses, idempotency, publisher).await;
    }

    // 2. ETH native transfer: 블록별 tx.to 확인
    let addr_set: std::collections::HashSet<String> =
        addresses.iter().map(|a| a.to_lowercase()).collect();

    for block_num in from_block..=to_block {
        if let Err(e) =
            scan_eth_block(client, rpc_url, block_num, &addr_set, idempotency, publisher).await
        {
            warn!("scan block {block_num}: {e}");
        }
    }

    Ok(())
}

async fn scan_eth_block(
    client: &Client,
    rpc_url: &str,
    block_num: u64,
    addr_set: &std::collections::HashSet<String>,
    idempotency: &Store,
    publisher: &KafkaPublisher,
) -> Result<()> {
    let block_hex = format!("0x{block_num:x}");
    let body = json!({
        "jsonrpc": "2.0", "id": 1,
        "method": "eth_getBlockByNumber",
        "params": [block_hex, true]
    });

    let resp: Value = client
        .post(rpc_url)
        .json(&body)
        .send()
        .await?
        .json()
        .await?;

    let txs = match resp["result"]["transactions"].as_array() {
        Some(t) => t.clone(),
        None => return Ok(()),
    };

    for tx in txs {
        let to = tx["to"].as_str().unwrap_or("").to_lowercase();
        if to.is_empty() || !addr_set.contains(&to) {
            continue;
        }

        let value_hex = tx["value"].as_str().unwrap_or("0x0");
        let value_wei = hex_to_u128(value_hex);
        if value_wei == 0 {
            continue;
        }

        let tx_hash = tx["hash"].as_str().unwrap_or("").to_string();
        if tx_hash.is_empty() {
            continue;
        }

        let is_new = idempotency.try_mark_seen(NETWORK, &tx_hash).await?;
        if !is_new {
            continue;
        }

        let amount_str = wei_to_eth_string(value_wei);
        info!("ETH deposit: to={to} amount={amount_str} tx={tx_hash}");

        publisher
            .publish_deposit_detected(DepositDetectedData {
                chain_tx_hash: tx_hash,
                network: NETWORK.to_string(),
                to_address: to,
                currency: "ETH".to_string(),
                amount: amount_str,
                block_number: Some(block_num),
            })
            .await?;
    }

    Ok(())
}

async fn process_usdt_log(
    log: EthLog,
    addresses: &[String],
    idempotency: &Store,
    publisher: &KafkaPublisher,
) {
    // topics[2] = to address (padded to 32 bytes)
    let to_topic = match log.topics.get(2) {
        Some(t) => t.clone(),
        None => return,
    };
    // 마지막 40자 = hex 주소
    let to_addr = format!("0x{}", &to_topic[to_topic.len().saturating_sub(40)..]).to_lowercase();

    let matched = addresses
        .iter()
        .any(|a| a.to_lowercase() == to_addr);
    if !matched {
        return;
    }

    let tx_hash = log.transaction_hash.clone();
    let is_new = match idempotency.try_mark_seen(NETWORK, &format!("{tx_hash}:usdt")).await {
        Ok(v) => v,
        Err(e) => { warn!("idempotency error: {e}"); return; }
    };
    if !is_new {
        return;
    }

    // data = uint256 value (32 bytes hex)
    let value_wei = hex_to_u128(&log.data);
    let amount_str = usdt_raw_to_string(value_wei);

    let block_number = log
        .block_number
        .as_deref()
        .and_then(|s| u64::from_str_radix(s.trim_start_matches("0x"), 16).ok());

    info!("USDT-ERC20 deposit: to={to_addr} amount={amount_str} tx={tx_hash}");

    if let Err(e) = publisher
        .publish_deposit_detected(DepositDetectedData {
            chain_tx_hash: tx_hash,
            network: NETWORK.to_string(),
            to_address: to_addr,
            currency: "USDT_ERC20".to_string(),
            amount: amount_str,
            block_number,
        })
        .await
    {
        warn!("publish USDT deposit error: {e}");
    }
}

// ── JSON-RPC helpers ─────────────────────────────────────────────────────────

async fn get_latest_block_number(client: &Client, rpc_url: &str) -> Result<u64> {
    let body = json!({
        "jsonrpc": "2.0", "id": 1,
        "method": "eth_blockNumber",
        "params": []
    });
    let resp: Value = client
        .post(rpc_url)
        .json(&body)
        .send()
        .await?
        .json()
        .await?;
    let hex = resp["result"]
        .as_str()
        .context("eth_blockNumber: missing result")?;
    Ok(u64::from_str_radix(hex.trim_start_matches("0x"), 16)?)
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct EthLog {
    topics: Vec<String>,
    data: String,
    transaction_hash: String,
    block_number: Option<String>,
}

async fn eth_get_logs(
    client: &Client,
    rpc_url: &str,
    contract: &str,
    topics0: &[String],
    topics2: &[String],
    from_block: u64,
    to_block: u64,
) -> Result<Vec<EthLog>> {
    let from_hex = format!("0x{from_block:x}");
    let to_hex = format!("0x{to_block:x}");

    // topics 필터: [Transfer topic, null (any from), [to addresses]]
    let topics_filter: Value = json!([
        topics0,
        Value::Null,
        topics2,
    ]);

    let body = json!({
        "jsonrpc": "2.0", "id": 1,
        "method": "eth_getLogs",
        "params": [{
            "fromBlock": from_hex,
            "toBlock":   to_hex,
            "address":   contract,
            "topics":    topics_filter,
        }]
    });

    let resp: Value = client
        .post(rpc_url)
        .json(&body)
        .send()
        .await?
        .json()
        .await?;

    if let Some(err) = resp.get("error") {
        anyhow::bail!("eth_getLogs RPC error: {err}");
    }

    let logs: Vec<EthLog> = serde_json::from_value(resp["result"].clone())
        .unwrap_or_default();
    Ok(logs)
}

// ── 변환 유틸 ────────────────────────────────────────────────────────────────

/// Ethereum address → 32바이트 패딩 토픽 (소문자)
fn address_to_topic(addr: &str) -> String {
    let clean = addr.trim_start_matches("0x").to_lowercase();
    format!("0x{:0>64}", clean)
}

fn hex_to_u128(hex: &str) -> u128 {
    let s = hex.trim_start_matches("0x");
    u128::from_str_radix(s, 16).unwrap_or(0)
}

/// wei → ETH 소수점 문자열 (18자리 고정)
fn wei_to_eth_string(wei: u128) -> String {
    let whole = wei / WEI_PER_ETH;
    let frac = wei % WEI_PER_ETH;
    format!("{whole}.{frac:018}")
}

/// USDT raw (6 decimals) → 소수점 문자열
fn usdt_raw_to_string(raw: u128) -> String {
    let whole = raw / USDT_DECIMALS;
    let frac = raw % USDT_DECIMALS;
    format!("{whole}.{frac:06}")
}
