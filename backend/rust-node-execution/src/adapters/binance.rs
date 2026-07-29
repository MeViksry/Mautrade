use crate::adapters::common::{
    account_mode, binance_symbol, client_order_id, decimal_from_str, form_body, hmac_sha256_hex, millis, quantity,
    quote_value, report, response_text, side_upper, status_from_fill, sum_decimal_strings,
};
use crate::credentials::InternalCredentialProvider;
use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest, OrderSide};
use async_trait::async_trait;
use serde::Deserialize;
use std::sync::Arc;

const REAL_BASE_URL: &str = "https://api.binance.com";
const TESTNET_BASE_URL: &str = "https://testnet.binance.vision";

#[derive(Debug, Clone)]
pub struct BinanceExecutionClient {
    credentials: Arc<InternalCredentialProvider>,
    http: reqwest::Client,
}

impl BinanceExecutionClient {
    pub fn new(credentials: Arc<InternalCredentialProvider>) -> Self {
        Self {
            credentials,
            http: reqwest::Client::new(),
        }
    }

    fn base_url(account_mode: &str) -> &'static str {
        match account_mode {
            "demo" | "testnet" => TESTNET_BASE_URL,
            _ => REAL_BASE_URL,
        }
    }
}

#[async_trait]
impl ExchangeExecutionClient for BinanceExecutionClient {
    async fn place_order(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        let credentials = self.credentials.fetch(&req.exchange_binding_id).await?;
        if credentials.exchange != "binance" {
            return Err(ExecutionError::InvalidOrder(
                "credential exchange does not match binance adapter".to_string(),
            ));
        }

        let account_mode = account_mode(&req, &credentials.account_mode);
        let symbol = binance_symbol(&req.symbol);
        let mut pairs = vec![
            ("symbol", symbol),
            ("side", side_upper(&req.side).to_string()),
            ("type", "MARKET".to_string()),
            ("newClientOrderId", client_order_id(&req, 36)),
            ("newOrderRespType", "FULL".to_string()),
            ("recvWindow", "5000".to_string()),
            ("timestamp", millis().to_string()),
        ];
        match req.side {
            OrderSide::Buy => pairs.push(("quoteOrderQty", quote_value(&req)?.to_string())),
            OrderSide::Sell => pairs.push(("quantity", quantity(&req)?.to_string())),
        }

        let unsigned_body = form_body(&pairs);
        let signature = hmac_sha256_hex(&credentials.api_secret, &unsigned_body)?;
        let body = format!("{unsigned_body}&signature={signature}");
        let response = self
            .http
            .post(format!("{}/api/v3/order", Self::base_url(&account_mode)))
            .header("X-MBX-APIKEY", credentials.api_key)
            .header(reqwest::header::CONTENT_TYPE, "application/x-www-form-urlencoded")
            .body(body)
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("binance place order request failed: {err}")))?;

        let body = response_text(response, "binance place order").await?;
        let order: BinanceOrderResponse = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("binance order response decode failed: {err}")))?;
        Ok(order.into_report(&req))
    }
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct BinanceOrderResponse {
    #[serde(default)]
    order_id: serde_json::Value,
    #[serde(default)]
    status: String,
    #[serde(default, rename = "executedQty")]
    executed_qty: String,
    #[serde(default, rename = "cummulativeQuoteQty")]
    cumulative_quote_qty: String,
    #[serde(default)]
    fills: Vec<BinanceFill>,
}

#[derive(Debug, Deserialize)]
struct BinanceFill {
    #[serde(default)]
    commission: String,
}

impl BinanceOrderResponse {
    fn into_report(self, req: &ExecutionRequest) -> ExecutionReport {
        let filled_quantity = decimal_from_str(&self.executed_qty);
        let fill_value_quote = decimal_from_str(&self.cumulative_quote_qty);
        let exchange_fee = sum_decimal_strings(self.fills.iter().map(|fill| fill.commission.as_str()));
        let status = status_from_fill(filled_quantity, &self.status);
        report(
            req,
            status,
            filled_quantity,
            fill_value_quote,
            exchange_fee,
            json_value_to_string(&self.order_id),
            Some(&self.status),
        )
    }
}

fn json_value_to_string(value: &serde_json::Value) -> Option<String> {
    match value {
        serde_json::Value::String(value) => Some(value.clone()),
        serde_json::Value::Number(value) => Some(value.to_string()),
        _ => None,
    }
}
