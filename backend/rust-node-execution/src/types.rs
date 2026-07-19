use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionRequest {
    pub id: String,
    pub idempotency_key: String,
    pub master_signal_id: String,
    pub user_id: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub layer_id: Option<String>,
    pub exchange: String,
    pub symbol: String,
    pub side: OrderSide,
    #[serde(default, skip_serializing_if = "Option::is_none", with = "rust_decimal::serde::str_option")]
    pub quantity: Option<Decimal>,
    #[serde(default, skip_serializing_if = "Option::is_none", with = "rust_decimal::serde::str_option")]
    pub quote_value: Option<Decimal>,
    pub created_at: String,
}

#[derive(Debug, Clone, Deserialize, Serialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum OrderSide {
    Buy,
    Sell,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionReport {
    pub request_id: String,
    pub idempotency_key: String,
    pub master_signal_id: String,
    pub user_id: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub layer_id: Option<String>,
    pub exchange: String,
    pub symbol: String,
    pub side: OrderSide,
    pub status: ExecutionStatus,
    #[serde(with = "rust_decimal::serde::str")]
    pub filled_quantity: Decimal,
    #[serde(with = "rust_decimal::serde::str")]
    pub fill_price: Decimal,
    #[serde(with = "rust_decimal::serde::str")]
    pub fill_value_quote: Decimal,
    #[serde(with = "rust_decimal::serde::str")]
    pub exchange_fee: Decimal,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub exchange_order_id: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub error_code: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub error_message: Option<String>,
    pub executed_at: String,
}

#[derive(Debug, Clone, Deserialize, Serialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum ExecutionStatus {
    Success,
    Failed,
    Skipped,
    Partial,
}
