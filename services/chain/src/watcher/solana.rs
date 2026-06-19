use anyhow::{Context, Result};
use std::str::FromStr;
use std::time::Duration;
use tokio::time::sleep;
use tracing::{info, warn};

use solana_client::nonblocking::rpc_client::RpcClient;
use solana_client::rpc_client::GetConfirmedSignaturesForAddress2Config;
use solana_client::rpc_config::RpcTransactionConfig;
use solana_sdk::commitment_config::CommitmentConfig;
use solana_sdk::pubkey::Pubkey;
use solana_sdk::signature::Signature;
use solana_transaction_status::{EncodedTransaction, UiMessage, UiTransactionEncoding};

use crate::idempotency::Store;
use crate::publisher::{DepositDetectedData, KafkaPublisher};
use crate::registry::AddressRegistry;

const NETWORK: &str = "SOLANA";
const POLL_INTERVAL: Duration = Duration::from_secs(10);
/// 요청당 최대 조회 시그니처 수
const MAX_SIGS: usize = 100;
/// 1 SOL = 1_000_000_000 lamports
const LAMPORTS_PER_SOL: u64 = 1_000_000_000;

pub async fn run(
    rpc_url: String,
    registry: AddressRegistry,
    idempotency: Store,
    publisher: KafkaPublisher,
) -> Result<()> {
    info!("solana watcher starting, rpc={rpc_url}");
    let client = RpcClient::new_with_commitment(rpc_url, CommitmentConfig::finalized());

    loop {
        let addresses = registry.get_addresses(NETWORK).await.unwrap_or_default();

        if addresses.is_empty() {
            sleep(POLL_INTERVAL).await;
            continue;
        }

        for address in &addresses {
            if let Err(e) =
                poll_address(&client, address, &registry, &idempotency, &publisher).await
            {
                warn!("poll error for address={address}: {e}");
            }
        }

        sleep(POLL_INTERVAL).await;
    }
}

async fn poll_address(
    client: &RpcClient,
    address: &str,
    registry: &AddressRegistry,
    idempotency: &Store,
    publisher: &KafkaPublisher,
) -> Result<()> {
    let pubkey =
        Pubkey::from_str(address).with_context(|| format!("invalid pubkey: {address}"))?;

    // Redis에 저장된 마지막 처리 시그니처 이후의 새 tx만 조회
    let last_sig_str = registry.get_last_sig(NETWORK, address).await?;
    let until_sig = last_sig_str
        .as_deref()
        .and_then(|s| Signature::from_str(s).ok());

    let config = GetConfirmedSignaturesForAddress2Config {
        before: None,
        until: until_sig,
        limit: Some(MAX_SIGS),
        commitment: Some(CommitmentConfig::finalized()),
    };

    let sigs = client
        .get_signatures_for_address_with_config(&pubkey, config)
        .await
        .with_context(|| format!("get_signatures_for_address {address}"))?;

    if sigs.is_empty() {
        return Ok(());
    }

    // 재시작 시 중복 방지를 위해 루프 진입 전에 최신 시그니처 저장
    let newest_sig = &sigs[0].signature;
    registry.set_last_sig(NETWORK, address, newest_sig).await?;

    // 오래된 것부터 처리 (역순 — sigs는 newest-first)
    for sig_info in sigs.iter().rev() {
        // 실패 tx 스킵
        if sig_info.err.is_some() {
            continue;
        }

        let sig_str = &sig_info.signature;

        // 1차 방어선: Redis SET NX 멱등성 체크
        let is_new = idempotency.try_mark_seen(NETWORK, sig_str).await?;
        if !is_new {
            continue;
        }

        let sig = Signature::from_str(sig_str)
            .with_context(|| format!("parse signature {sig_str}"))?;

        let tx_config = RpcTransactionConfig {
            encoding: Some(UiTransactionEncoding::Json),
            commitment: Some(CommitmentConfig::finalized()),
            max_supported_transaction_version: Some(0),
        };

        let tx = match client.get_transaction_with_config(&sig, tx_config).await {
            Ok(tx) => tx,
            Err(e) => {
                warn!("get_transaction failed sig={sig_str}: {e}");
                continue;
            }
        };

        let Some(lamports) = parse_sol_received(&tx, address) else {
            continue; // SOL 수신 없는 tx (outgoing, SPL-only 등)
        };

        if lamports == 0 {
            continue;
        }

        let amount_str = lamports_to_sol(lamports);

        info!("SOL deposit detected: address={address} amount={amount_str} tx={sig_str}");

        publisher
            .publish_deposit_detected(DepositDetectedData {
                chain_tx_hash: sig_str.clone(),
                network:       NETWORK.to_string(),
                to_address:    address.to_string(),
                currency:      "SOL".to_string(),
                amount:        amount_str,
                block_number:  Some(sig_info.slot),
            })
            .await
            .with_context(|| format!("publish deposit for tx={sig_str}"))?;
    }

    Ok(())
}

/// target_address의 SOL 순증가량(lamports)을 반환한다.
/// 수신이 없거나 잔액이 감소한 경우 None.
fn parse_sol_received(
    tx: &solana_transaction_status::EncodedConfirmedTransactionWithStatusMeta,
    target_address: &str,
) -> Option<u64> {
    let meta = tx.transaction.meta.as_ref()?;

    // JSON 인코딩 기준으로 account_keys 추출
    let account_keys: Vec<String> = match &tx.transaction.transaction {
        EncodedTransaction::Json(ui_tx) => match &ui_tx.message {
            UiMessage::Raw(raw) => raw.account_keys.clone(),
            UiMessage::Parsed(parsed) => parsed
                .account_keys
                .iter()
                .map(|k| k.pubkey.clone())
                .collect(),
        },
        _ => return None,
    };

    let idx = account_keys.iter().position(|k| k == target_address)?;

    let pre = *meta.pre_balances.get(idx)?;
    let post = *meta.post_balances.get(idx)?;

    if post > pre {
        Some(post - pre)
    } else {
        None
    }
}

/// lamports를 SOL 소수점 문자열로 변환 ("1.500000000" 형식, 9자리 고정).
fn lamports_to_sol(lamports: u64) -> String {
    let whole = lamports / LAMPORTS_PER_SOL;
    let frac = lamports % LAMPORTS_PER_SOL;
    format!("{whole}.{frac:09}")
}
