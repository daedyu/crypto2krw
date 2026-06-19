use anyhow::Result;
use redis::AsyncCommands;

/// chain-watcher가 감시할 주소 목록과 마지막 처리 시그니처를 Redis에 관리.
///
/// SADD  chain-watcher:watch:{NETWORK}     {address}   — 감시 주소 등록
/// SMEMBERS chain-watcher:watch:{NETWORK}              — 감시 주소 조회
/// SET   chain-watcher:last-sig:{NETWORK}:{address}    — 마지막 처리 시그니처
/// GET   chain-watcher:last-sig:{NETWORK}:{address}    — 마지막 처리 시그니처 조회
#[derive(Clone)]
pub struct AddressRegistry {
    client: redis::Client,
}

impl AddressRegistry {
    pub fn new(client: redis::Client) -> Self {
        Self { client }
    }

    pub async fn add_address(&self, network: &str, address: &str) -> Result<()> {
        let key = format!("chain-watcher:watch:{network}");
        let mut conn = self.client.get_multiplexed_async_connection().await?;
        conn.sadd::<_, _, ()>(&key, address).await?;
        Ok(())
    }

    pub async fn get_addresses(&self, network: &str) -> Result<Vec<String>> {
        let key = format!("chain-watcher:watch:{network}");
        let mut conn = self.client.get_multiplexed_async_connection().await?;
        let members: Vec<String> = conn.smembers(&key).await.unwrap_or_default();
        Ok(members)
    }

    pub async fn get_last_sig(&self, network: &str, address: &str) -> Result<Option<String>> {
        let key = format!("chain-watcher:last-sig:{network}:{address}");
        let mut conn = self.client.get_multiplexed_async_connection().await?;
        let val: Option<String> = conn.get(&key).await?;
        Ok(val)
    }

    pub async fn set_last_sig(&self, network: &str, address: &str, sig: &str) -> Result<()> {
        let key = format!("chain-watcher:last-sig:{network}:{address}");
        let mut conn = self.client.get_multiplexed_async_connection().await?;
        conn.set::<_, _, ()>(&key, sig).await?;
        Ok(())
    }
}
