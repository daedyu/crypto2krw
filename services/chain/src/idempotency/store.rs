use anyhow::Result;
use redis::AsyncCommands;
use std::time::Duration;

const TTL_SECS: u64 = 60 * 60 * 24 * 30; // 30일

/// Redis SET NX를 이용한 tx_hash 중복 처리 방지.
/// chain-watcher의 1차 방어선 — 계정계 DB UNIQUE 제약이 최후 보루.
#[derive(Clone)]
pub struct Store {
    client: redis::Client,
}

impl Store {
    pub async fn new(redis_url: &str) -> Result<Self> {
        let client = redis::Client::open(redis_url)?;
        // 연결 확인
        let mut conn = client.get_multiplexed_async_connection().await?;
        redis::cmd("PING").exec_async(&mut conn).await?;
        Ok(Self { client })
    }

    /// 이미 본 tx_hash이면 false 반환, 새로운 tx_hash이면 true 반환 후 마킹.
    pub async fn try_mark_seen(&self, network: &str, tx_hash: &str) -> Result<bool> {
        let key = format!("chain-watcher:seen:{network}:{tx_hash}");
        let mut conn = self.client.get_multiplexed_async_connection().await?;

        // SET key 1 NX EX ttl
        let result: Option<String> = conn
            .set_options(
                &key,
                "1",
                redis::SetOptions::default()
                    .conditional_set(redis::ExistenceCheck::NX)
                    .with_expiration(redis::SetExpiry::EX(TTL_SECS)),
            )
            .await?;

        // SET NX 성공 시 Some("OK"), 키 존재 시 None
        Ok(result.is_some())
    }
}
