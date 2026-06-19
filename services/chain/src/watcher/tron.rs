use anyhow::Result;
use reqwest::Client;
use serde::Deserialize;
use std::time::Duration;
use tokio::time::sleep;
use tracing::{info, warn};

use crate::idempotency::Store;
use crate::publisher::{DepositDetectedData, KafkaPublisher};
use crate::registry::AddressRegistry;

const NETWORK: &str = "TRON";
const POLL_INTERVAL: Duration = Duration::from_secs(10);
// USDT-TRC20 has 6 decimals
const USDT_DECIMALS: u128 = 1_000_000;

#[derive(Deserialize)]
struct TronTrc20Response {
    data: Option<Vec<TronTrc20Tx>>,
}

#[derive(Deserialize)]
struct TronTrc20Tx {
    transaction_id: String,
    token_info: TronTokenInfo,
    to: String,
    value: String,
    block_timestamp: Option<u64>,
}

#[derive(Deserialize)]
struct TronTokenInfo {
    address: String,
    decimals: u32,
}

pub async fn run(
    api_url: String,
    usdt_contract: String,
    registry: AddressRegistry,
    idempotency: Store,
    publisher: KafkaPublisher,
) -> Result<()> {
    info!("tron watcher starting, api={api_url}");
    let client = Client::builder()
        .user_agent("crypto2krw-chain-watcher/1.0")
        .build()?;

    loop {
        let addresses = registry.get_addresses(NETWORK).await.unwrap_or_default();
        if addresses.is_empty() {
            sleep(POLL_INTERVAL).await;
            continue;
        }

        for address in &addresses {
            if let Err(e) = poll_address(
                &client,
                &api_url,
                &usdt_contract,
                address,
                &idempotency,
                &publisher,
            )
            .await
            {
                warn!("tron poll error for address={address}: {e}");
            }
        }

        sleep(POLL_INTERVAL).await;
    }
}

async fn poll_address(
    client: &Client,
    api_url: &str,
    usdt_contract: &str,
    address: &str,
    idempotency: &Store,
    publisher: &KafkaPublisher,
) -> Result<()> {
    // TronGrid: GET /v1/accounts/{address}/transactions/trc20
    // contract_address 필터로 USDT만 조회
    let url = format!(
        "{api_url}/v1/accounts/{address}/transactions/trc20\
         ?contract_address={usdt_contract}&limit=50&order_by=block_timestamp,desc"
    );

    let resp = client.get(&url).send().await?;
    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().await.unwrap_or_default();
        anyhow::bail!("TronGrid API error {status}: {body}");
    }

    let tron_resp: TronTrc20Response = resp.json().await?;
    let txs = tron_resp.data.unwrap_or_default();

    for tx in txs {
        // to 주소가 감시 대상인지 확인
        if tx.to.to_lowercase() != address.to_lowercase() {
            continue;
        }

        // USDT 컨트랙트 주소 확인
        if tx.token_info.address.to_lowercase() != usdt_contract.to_lowercase() {
            continue;
        }

        let is_new = idempotency.try_mark_seen(NETWORK, &tx.transaction_id).await?;
        if !is_new {
            continue;
        }

        let decimals = 10u128.pow(tx.token_info.decimals);
        let raw: u128 = tx.value.parse().unwrap_or(0);
        let amount_str = raw_to_decimal_string(raw, decimals);

        info!(
            "USDT-TRC20 deposit: to={address} amount={amount_str} tx={}",
            tx.transaction_id
        );

        publisher
            .publish_deposit_detected(DepositDetectedData {
                chain_tx_hash: tx.transaction_id,
                network: NETWORK.to_string(),
                to_address: address.to_string(),
                currency: "USDT_TRC20".to_string(),
                amount: amount_str,
                block_number: None,
            })
            .await?;
    }

    Ok(())
}

fn raw_to_decimal_string(raw: u128, decimals: u128) -> String {
    let whole = raw / decimals;
    let frac = raw % decimals;
    let frac_digits = if decimals == USDT_DECIMALS { 6 } else { 18 };
    format!("{whole}.{frac:0>frac_digits$}")
}
