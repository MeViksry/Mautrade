use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus};
use async_trait::async_trait;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::sync::Arc;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ExecutionError {
    #[error("unsupported exchange: {0}")]
    UnsupportedExchange(String),
    #[error("invalid order: {0}")]
    InvalidOrder(String),
    #[error("exchange error: {0}")]
    Exchange(String),
}

#[async_trait]
pub trait ExchangeExecutionClient: Send + Sync {
    async fn place_order(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError>;
}

#[async_trait]
pub trait ExecutionRouter: Send + Sync {
    async fn execute(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError>;
}

#[derive(Default)]
pub struct StaticRouter {
    clients: HashMap<String, Arc<dyn ExchangeExecutionClient>>,
}

impl StaticRouter {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_client(mut self, exchange: impl Into<String>, client: Arc<dyn ExchangeExecutionClient>) -> Self {
        self.clients.insert(exchange.into(), client);
        self
    }
}

#[async_trait]
impl ExecutionRouter for StaticRouter {
    async fn execute(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        validate_request(&req)?;
        let exchange = req.exchange.to_lowercase();
        let client = self
            .clients
            .get(&exchange)
            .ok_or_else(|| ExecutionError::UnsupportedExchange(exchange.clone()))?;
        client.place_order(req).await
    }
}

pub fn validate_request(req: &ExecutionRequest) -> Result<(), ExecutionError> {
    if req.idempotency_key.trim().is_empty() {
        return Err(ExecutionError::InvalidOrder("idempotency_key is required".to_string()));
    }
    if req.symbol.trim().is_empty() {
        return Err(ExecutionError::InvalidOrder("symbol is required".to_string()));
    }
    if req.quantity.is_none() && req.quote_value.is_none() {
        return Err(ExecutionError::InvalidOrder(
            "quantity or quote_value is required".to_string(),
        ));
    }
    Ok(())
}

pub fn failed_report(req: &ExecutionRequest, code: impl Into<String>, message: impl Into<String>) -> ExecutionReport {
    ExecutionReport {
        request_id: req.id.clone(),
        idempotency_key: req.idempotency_key.clone(),
        master_signal_id: req.master_signal_id.clone(),
        user_id: req.user_id.clone(),
        layer_id: req.layer_id.clone(),
        exchange: req.exchange.clone(),
        symbol: req.symbol.clone(),
        side: req.side.clone(),
        status: ExecutionStatus::Failed,
        filled_quantity: Decimal::ZERO,
        fill_price: Decimal::ZERO,
        fill_value_quote: Decimal::ZERO,
        exchange_fee: Decimal::ZERO,
        exchange_order_id: None,
        error_code: Some(code.into()),
        error_message: Some(message.into()),
        executed_at: chrono_like_utc_now(),
    }
}

fn chrono_like_utc_now() -> String {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|duration| format!("{}.{:09}Z", duration.as_secs(), duration.subsec_nanos()))
        .unwrap_or_else(|_| "0.000000000Z".to_string())
}
