mod config;
mod idempotency;
mod publisher;
mod registry;
mod watcher;

use anyhow::Result;
use tracing::info;
use tracing_subscriber::EnvFilter;

use registry::AddressRegistry;

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .json()
        .init();

    dotenvy::dotenv().ok();

    let cfg = config::Config::from_env()?;
    info!("chain-watcher starting");

    let redis_client = redis::Client::open(cfg.redis_url.as_str())?;
    let idempotency_store = idempotency::Store::new(&cfg.redis_url).await?;
    let kafka_publisher = publisher::KafkaPublisher::new(&cfg.kafka_brokers)?;
    let registry = AddressRegistry::new(redis_client);

    // 1. 주소 레지스트리 구독자 먼저 시작 (wallet.created 이벤트 → Redis set)
    {
        let reg = registry.clone();
        let brokers = cfg.kafka_brokers.clone();
        tokio::spawn(async move {
            if let Err(e) = registry::kafka_subscriber::run(&brokers, reg).await {
                tracing::error!("registry subscriber error: {e}");
            }
        });
    }

    // 2. 잠시 대기 — 레지스트리가 기존 주소를 Redis에 로드할 시간
    tokio::time::sleep(std::time::Duration::from_secs(2)).await;

    let mut handles = Vec::new();

    // Solana watcher (HTTP 폴링, 10초 간격)
    {
        let store = idempotency_store.clone();
        let publisher = kafka_publisher.clone();
        let reg = registry.clone();
        let solana_rpc = cfg.solana_rpc_url.clone();
        handles.push(tokio::spawn(async move {
            watcher::solana::run(solana_rpc, reg, store, publisher).await
        }));
    }

    // Ethereum watcher (HTTP JSON-RPC 폴링, 15초 간격)
    {
        let store = idempotency_store.clone();
        let publisher = kafka_publisher.clone();
        let reg = registry.clone();
        let eth_rpc = cfg.ethereum_rpc_url.clone();
        let usdt_contract = cfg.usdt_erc20_contract.clone();
        handles.push(tokio::spawn(async move {
            watcher::ethereum::run(eth_rpc, usdt_contract, reg, store, publisher).await
        }));
    }

    // Tron watcher (TronGrid REST 폴링, 10초 간격)
    {
        let store = idempotency_store.clone();
        let publisher = kafka_publisher.clone();
        let reg = registry.clone();
        let tron_api = cfg.tron_api_url.clone();
        let usdt_trc20 = cfg.usdt_trc20_contract.clone();
        handles.push(tokio::spawn(async move {
            watcher::tron::run(tron_api, usdt_trc20, reg, store, publisher).await
        }));
    }

    for handle in handles {
        handle.await??;
    }

    Ok(())
}
