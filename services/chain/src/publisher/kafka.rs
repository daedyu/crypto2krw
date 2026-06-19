use anyhow::Result;
use chrono::Utc;
use rdkafka::config::ClientConfig;
use rdkafka::producer::{FutureProducer, FutureRecord};
use serde::Serialize;
use std::time::Duration;
use uuid::Uuid;

const TOPIC_DEPOSIT_DETECTED: &str = "crypto2krw.core.deposit.detected";

#[derive(Serialize)]
struct CloudEvent<T: Serialize> {
    specversion:     &'static str,
    id:              String,
    source:          &'static str,
    #[serde(rename = "type")]
    event_type:      &'static str,
    time:            String,
    datacontenttype: &'static str,
    data:            T,
}

#[derive(Serialize)]
pub struct DepositDetectedData {
    pub chain_tx_hash: String,
    pub network:       String,  // "SOLANA" | "ETHEREUM" | "TRON"
    pub to_address:    String,
    pub currency:      String,  // "SOL" | "USDT_ERC20" | "USDT_TRC20" | "ETH"
    pub amount:        String,  // Decimal string
    pub block_number:  Option<u64>,
}

#[derive(Clone)]
pub struct KafkaPublisher {
    producer: FutureProducer,
}

impl KafkaPublisher {
    pub fn new(brokers: &str) -> Result<Self> {
        let producer: FutureProducer = ClientConfig::new()
            .set("bootstrap.servers", brokers)
            .set("acks", "all")
            .set("retries", "5")
            .set("enable.idempotence", "true")
            .create()?;

        Ok(Self { producer })
    }

    pub async fn publish_deposit_detected(&self, data: DepositDetectedData) -> Result<()> {
        let event = CloudEvent {
            specversion:     "1.0",
            id:              Uuid::new_v4().to_string(),
            source:          "chain-watcher",
            event_type:      "crypto2krw.core.deposit.detected",
            time:            Utc::now().to_rfc3339(),
            datacontenttype: "application/json",
            data,
        };

        let payload = serde_json::to_string(&event)?;
        // 파티션 키: to_address (같은 주소의 이벤트 순서 보장)
        let key = event.data.to_address.clone();

        self.producer
            .send(
                FutureRecord::to(TOPIC_DEPOSIT_DETECTED)
                    .key(&key)
                    .payload(&payload),
                Duration::from_secs(10),
            )
            .await
            .map_err(|(err, _)| anyhow::anyhow!("kafka send error: {err}"))?;

        Ok(())
    }
}
