use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus};
use async_trait::async_trait;
use rust_decimal::Decimal;

#[derive(Debug, Default)]
pub struct BinanceExecutionClient;

#[async_trait]
impl ExchangeExecutionClient for BinanceExecutionClient {
    async fn place_order(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        Ok(paper_report(req, "binance-paper-order"))
    }
}

fn paper_report(req: ExecutionRequest, order_id: &str) -> ExecutionReport {
    ExecutionReport {
        request_id: req.id,
        idempotency_key: req.idempotency_key,
        master_signal_id: req.master_signal_id,
        user_id: req.user_id,
        layer_id: req.layer_id,
        exchange: req.exchange,
        symbol: req.symbol,
        side: req.side,
        status: ExecutionStatus::Success,
        filled_quantity: req.quantity.unwrap_or(Decimal::ZERO),
        fill_price: Decimal::ZERO,
        fill_value_quote: req.quote_value.unwrap_or(Decimal::ZERO),
        exchange_fee: Decimal::ZERO,
        exchange_order_id: Some(order_id.to_string()),
        error_code: None,
        error_message: None,
        executed_at: "paper".to_string(),
    }
}
