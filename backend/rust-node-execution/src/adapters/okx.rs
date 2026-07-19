use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest};
use async_trait::async_trait;

#[derive(Debug, Default)]
pub struct OkxExecutionClient;

#[async_trait]
impl ExchangeExecutionClient for OkxExecutionClient {
    async fn place_order(&self, _req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        Err(ExecutionError::Exchange(
            "OKX adapter is scaffolded; live signing and exchange filters are not implemented yet".to_string(),
        ))
    }
}
