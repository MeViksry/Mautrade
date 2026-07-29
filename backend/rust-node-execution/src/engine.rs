use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus};
use async_trait::async_trait;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use thiserror::Error;
use time::format_description::well_known::Rfc3339;
use time::OffsetDateTime;

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
    if req.exchange_binding_id.trim().is_empty() {
        return Err(ExecutionError::InvalidOrder(
            "exchange_binding_id is required".to_string(),
        ));
    }
    if let Some(account_mode) = &req.account_mode {
        match account_mode.trim().to_ascii_lowercase().as_str() {
            "" | "real" | "demo" | "testnet" => {}
            other => {
                return Err(ExecutionError::InvalidOrder(format!(
                    "account_mode must be real, demo, or testnet; got {other}"
                )));
            }
        }
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

pub fn stale_request_report(req: &ExecutionRequest, max_age: Duration) -> Option<ExecutionReport> {
    if max_age.is_zero() {
        return None;
    }
    let created_at = OffsetDateTime::parse(req.created_at.trim(), &Rfc3339).ok()?;
    let age = OffsetDateTime::now_utc() - created_at;
    if age.whole_seconds() < max_age.as_secs() as i64 {
        return None;
    }
    Some(failed_report(
        req,
        "stale_execution_request",
        format!(
            "execution request is older than {} seconds; rejected to prevent delayed market order",
            max_age.as_secs()
        ),
    ))
}

fn chrono_like_utc_now() -> String {
    OffsetDateTime::now_utc()
        .format(&Rfc3339)
        .unwrap_or_else(|_| "1970-01-01T00:00:00Z".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::OrderSide;

    fn request_with_account_mode(account_mode: Option<&str>) -> ExecutionRequest {
        ExecutionRequest {
            id: "job-1".to_string(),
            idempotency_key: "signal:user:binding:buy".to_string(),
            master_signal_id: "signal-1".to_string(),
            user_id: "user-1".to_string(),
            layer_id: None,
            exchange_binding_id: "binding-1".to_string(),
            exchange: "binance".to_string(),
            account_mode: account_mode.map(str::to_string),
            symbol: "BTCUSDT".to_string(),
            side: OrderSide::Buy,
            quantity: None,
            quote_value: Some(Decimal::new(10, 0)),
            created_at: "2026-07-28T00:00:00Z".to_string(),
        }
    }

    #[test]
    fn validate_request_accepts_testnet_account_mode() {
        let req = request_with_account_mode(Some("testnet"));

        assert!(validate_request(&req).is_ok());
    }

    #[test]
    fn validate_request_rejects_unknown_account_mode() {
        let req = request_with_account_mode(Some("practice"));

        assert!(validate_request(&req).is_err());
    }

    #[test]
    fn validate_request_requires_exchange_binding_id() {
        let mut req = request_with_account_mode(Some("real"));
        req.exchange_binding_id.clear();

        assert!(validate_request(&req).is_err());
    }

    #[test]
    fn stale_request_report_rejects_old_market_orders() {
        let mut req = request_with_account_mode(Some("real"));
        req.created_at = "2026-01-01T00:00:00Z".to_string();

        let report = stale_request_report(&req, Duration::from_secs(300)).expect("expected stale report");

        assert_eq!(report.status, ExecutionStatus::Failed);
        assert_eq!(report.error_code.as_deref(), Some("stale_execution_request"));
    }

    #[test]
    fn stale_request_report_allows_disabled_guard() {
        let mut req = request_with_account_mode(Some("real"));
        req.created_at = "2026-01-01T00:00:00Z".to_string();

        assert!(stale_request_report(&req, Duration::ZERO).is_none());
    }

    #[test]
    fn failed_report_uses_rfc3339_executed_at() {
        let req = request_with_account_mode(Some("real"));

        let report = failed_report(&req, "test_error", "failed for test");

        assert!(OffsetDateTime::parse(&report.executed_at, &Rfc3339).is_ok());
    }
}
