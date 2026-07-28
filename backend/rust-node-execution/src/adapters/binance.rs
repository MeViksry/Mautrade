use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest};
use async_trait::async_trait;

#[derive(Debug, Default)]
pub struct BinanceExecutionClient;

#[async_trait]
impl ExchangeExecutionClient for BinanceExecutionClient {
    async fn place_order(&self, _req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        Err(ExecutionError::Exchange(
            "Binance adapter is scaffolded; live/testnet signed spot order execution is not implemented yet".to_string(),
        ))
    }
}
