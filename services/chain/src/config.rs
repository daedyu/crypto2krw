use anyhow::{Context, Result};
use std::env;

pub struct Config {
    pub kafka_brokers:        String,
    pub redis_url:            String,
    pub solana_rpc_url:       String,
    pub ethereum_rpc_url:     String,
    pub usdt_erc20_contract:  String,
    pub tron_api_url:         String,
    pub usdt_trc20_contract:  String,
}

impl Config {
    pub fn from_env() -> Result<Self> {
        Ok(Self {
            kafka_brokers: env::var("KAFKA_BROKERS")
                .unwrap_or_else(|_| "localhost:9093".to_string()),
            redis_url: env::var("REDIS_URL")
                .unwrap_or_else(|_| "redis://localhost:6379".to_string()),
            solana_rpc_url: env::var("SOLANA_RPC_URL")
                .context("SOLANA_RPC_URL required")?,
            ethereum_rpc_url: env::var("ETHEREUM_RPC_URL")
                .context("ETHEREUM_RPC_URL required")?,
            usdt_erc20_contract: env::var("USDT_ERC20_CONTRACT")
                .unwrap_or_else(|_| "0xdAC17F958D2ee523a2206206994597C13D831ec7".to_string()),
            tron_api_url: env::var("TRON_API_URL")
                .unwrap_or_else(|_| "https://api.trongrid.io".to_string()),
            usdt_trc20_contract: env::var("USDT_TRC20_CONTRACT")
                .unwrap_or_else(|_| "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t".to_string()),
        })
    }
}
