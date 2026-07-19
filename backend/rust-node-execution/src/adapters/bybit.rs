use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest};
use async_trait::async_trait;

#[derive(Debug, Default)]
pub struct BybitExecutionClient;

#[async_trait]
impl ExchangeExecutionClient for BybitExecutionClient {
    async fn place_order(&self, _req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        Err(ExecutionError::Exchange(
            "Bybit adapter is scaffolded; live signing and exchange filters are not implemented yet".to_string(),
        ))
    }
}
