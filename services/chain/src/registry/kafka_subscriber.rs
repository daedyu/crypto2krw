use anyhow::Result;
use rdkafka::config::ClientConfig;
use rdkafka::consumer::{Consumer, StreamConsumer};
use rdkafka::message::Message;
use serde::Deserialize;
use tracing::{info, warn};

use super::redis_store::AddressRegistry;

const TOPIC_WALLET_CREATED: &str = "crypto2krw.core.wallet.created";

#[derive(Deserialize)]
struct WalletCreatedData {
    address: String,
    network: String,
    // user_id, currency 등은 이 컨텍스트에서 불필요
}

#[derive(Deserialize)]
struct CloudEvent {
    data: WalletCreatedData,
}

/// wallet.created 이벤트를 구독하여 감시 주소를 Redis에 등록.
/// main.rs에서 별도 tokio task로 실행되며 컨텍스트 취소 시 종료.
pub async fn run(brokers: &str, registry: AddressRegistry) -> Result<()> {
    let consumer: StreamConsumer = ClientConfig::new()
        .set("bootstrap.servers", brokers)
        .set("group.id", "chain-watcher-registry")
        .set("auto.offset.reset", "earliest")
        .set("enable.auto.commit", "true")
        .create()?;

    consumer.subscribe(&[TOPIC_WALLET_CREATED])?;
    info!("registry subscriber started, topic={TOPIC_WALLET_CREATED}");

    loop {
        match consumer.recv().await {
            Err(e) => {
                warn!("kafka recv error in registry: {e}");
            }
            Ok(msg) => {
                let Some(payload) = msg.payload() else { continue };

                let event: CloudEvent = match serde_json::from_slice(payload) {
                    Ok(e) => e,
                    Err(e) => {
                        warn!("malformed wallet.created message: {e}");
                        continue;
                    }
                };

                let data = &event.data;
                if let Err(e) = registry.add_address(&data.network, &data.address).await {
                    warn!("failed to register address {}: {e}", data.address);
                } else {
                    info!(
                        "registered watch address network={} address={}",
                        data.network, data.address
                    );
                }
            }
        }
    }
}
