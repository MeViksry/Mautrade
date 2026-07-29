use crate::adapters::common::{
    account_mode, binance_symbol, client_order_id, decimal_from_str, form_body, hmac_sha256_hex, millis, quantity,
    quote_value, report, response_text, side_title, status_from_fill, sum_decimal_strings,
};
use crate::credentials::InternalCredentialProvider;
use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus, OrderSide};
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::sleep;

const REAL_BASE_URL: &str = "https://api.bybit.com";
const DEMO_BASE_URL: &str = "https://api-demo.bybit.com";
const TESTNET_BASE_URL: &str = "https://api-testnet.bybit.com";
const RECV_WINDOW: &str = "5000";

#[derive(Debug, Clone)]
pub struct BybitExecutionClient {
    credentials: Arc<InternalCredentialProvider>,
    http: reqwest::Client,
}

impl BybitExecutionClient {
    pub fn new(credentials: Arc<InternalCredentialProvider>) -> Self {
        Self {
            credentials,
            http: reqwest::Client::new(),
        }
    }

    fn base_url(account_mode: &str) -> &'static str {
        match account_mode {
            "demo" => DEMO_BASE_URL,
            "testnet" => TESTNET_BASE_URL,
            _ => REAL_BASE_URL,
        }
    }
}

#[async_trait]
impl ExchangeExecutionClient for BybitExecutionClient {
    async fn place_order(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        let credentials = self.credentials.fetch_for_order(&req).await?;
        if credentials.exchange != "bybit" {
            return Err(ExecutionError::InvalidOrder(
                "credential exchange does not match bybit adapter".to_string(),
            ));
        }

        let account_mode = account_mode(&req, &credentials.account_mode);
        let base_url = Self::base_url(&account_mode);
        let symbol = binance_symbol(&req.symbol);
        let qty = match req.side {
            OrderSide::Buy => quote_value(&req)?.to_string(),
            OrderSide::Sell => quantity(&req)?.to_string(),
        };
        let market_unit = match req.side {
            OrderSide::Buy => "quoteCoin",
            OrderSide::Sell => "baseCoin",
        };
        let order_link_id = client_order_id(&req, 36);
        let body = BybitCreateOrderRequest {
            category: "spot",
            symbol: &symbol,
            side: side_title(&req.side),
            order_type: "Market",
            qty: &qty,
            time_in_force: "IOC",
            order_link_id: &order_link_id,
            is_leverage: 0,
            order_filter: "Order",
            market_unit,
        };
        let body_json = serde_json::to_string(&body)
            .map_err(|err| ExecutionError::Exchange(format!("bybit encode order request failed: {err}")))?;
        let timestamp = millis().to_string();
        let signature_payload = format!("{timestamp}{}{RECV_WINDOW}{body_json}", credentials.api_key);
        let signature = hmac_sha256_hex(&credentials.api_secret, &signature_payload)?;

        let response = self
            .http
            .post(format!("{base_url}/v5/order/create"))
            .header("X-BAPI-API-KEY", credentials.api_key.clone())
            .header("X-BAPI-TIMESTAMP", timestamp)
            .header("X-BAPI-RECV-WINDOW", RECV_WINDOW)
            .header("X-BAPI-SIGN-TYPE", "2")
            .header("X-BAPI-SIGN", signature)
            .header(reqwest::header::CONTENT_TYPE, "application/json")
            .body(body_json)
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("bybit place order request failed: {err}")))?;
        let body = response_text(response, "bybit place order").await?;
        let created: BybitCreateOrderResponse = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("bybit order response decode failed: {err}")))?;
        if created.ret_code != 0 {
            return Err(ExecutionError::Exchange(format!(
                "bybit rejected order: {}",
                created.ret_msg
            )));
        }
        let order_id = created.result.order_id;
        self.poll_order_report(base_url, &credentials, &req, &symbol, &order_id, &order_link_id)
            .await
    }
}

impl BybitExecutionClient {
    async fn poll_order_report(
        &self,
        base_url: &str,
        credentials: &crate::credentials::ExchangeCredentials,
        req: &ExecutionRequest,
        symbol: &str,
        order_id: &str,
        order_link_id: &str,
    ) -> Result<ExecutionReport, ExecutionError> {
        let mut last_status = String::new();
        for attempt in 0..12 {
            match self
                .fetch_order(base_url, credentials, symbol, order_id, order_link_id)
                .await
            {
                Ok(Some(order)) => {
                    last_status = order.order_status.clone();
                    let filled_quantity = decimal_from_str(&order.cum_exec_qty);
                    let fill_value_quote = decimal_from_str(&order.cum_exec_value);
                    if filled_quantity > rust_decimal::Decimal::ZERO && fill_value_quote > rust_decimal::Decimal::ZERO {
                        let fee_values = order
                            .cum_fee_detail
                            .as_ref()
                            .map(|values| values.values().map(String::as_str).collect::<Vec<_>>())
                            .unwrap_or_default();
                        let fee = if fee_values.is_empty() {
                            decimal_from_str(&order.cum_exec_fee)
                        } else {
                            sum_decimal_strings(fee_values)
                        };
                        return Ok(report(
                            req,
                            status_from_fill(filled_quantity, &order.order_status),
                            filled_quantity,
                            fill_value_quote,
                            fee,
                            Some(order.order_id),
                            Some(&order.order_status),
                        ));
                    }
                    if order.order_status.eq_ignore_ascii_case("Rejected")
                        || order.order_status.eq_ignore_ascii_case("Cancelled")
                    {
                        return Ok(report(
                            req,
                            ExecutionStatus::Failed,
                            rust_decimal::Decimal::ZERO,
                            rust_decimal::Decimal::ZERO,
                            rust_decimal::Decimal::ZERO,
                            Some(order.order_id),
                            Some(&order.order_status),
                        ));
                    }
                }
                Ok(None) => {}
                Err(err) => {
                    if attempt == 11 {
                        return Err(err);
                    }
                }
            }
            sleep(Duration::from_millis(500)).await;
        }

        Ok(report(
            req,
            ExecutionStatus::Failed,
            rust_decimal::Decimal::ZERO,
            rust_decimal::Decimal::ZERO,
            rust_decimal::Decimal::ZERO,
            Some(order_id.to_string()),
            Some(if last_status.is_empty() {
                "order status was not available after polling"
            } else {
                &last_status
            }),
        ))
    }

    async fn fetch_order(
        &self,
        base_url: &str,
        credentials: &crate::credentials::ExchangeCredentials,
        symbol: &str,
        order_id: &str,
        order_link_id: &str,
    ) -> Result<Option<BybitOrder>, ExecutionError> {
        let query = form_body(&[
            ("category", "spot".to_string()),
            ("symbol", symbol.to_string()),
            ("orderId", order_id.to_string()),
            ("orderLinkId", order_link_id.to_string()),
            ("openOnly", "1".to_string()),
            ("limit", "1".to_string()),
        ]);
        let timestamp = millis().to_string();
        let signature_payload = format!("{timestamp}{}{RECV_WINDOW}{query}", credentials.api_key);
        let signature = hmac_sha256_hex(&credentials.api_secret, &signature_payload)?;

        let response = self
            .http
            .get(format!("{base_url}/v5/order/realtime?{query}"))
            .header("X-BAPI-API-KEY", credentials.api_key.clone())
            .header("X-BAPI-TIMESTAMP", timestamp)
            .header("X-BAPI-RECV-WINDOW", RECV_WINDOW)
            .header("X-BAPI-SIGN-TYPE", "2")
            .header("X-BAPI-SIGN", signature)
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("bybit fetch order request failed: {err}")))?;
        let body = response_text(response, "bybit fetch order").await?;
        let fetched: BybitOrderRealtimeResponse = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("bybit fetch order decode failed: {err}")))?;
        if fetched.ret_code != 0 {
            return Err(ExecutionError::Exchange(format!(
                "bybit rejected order lookup: {}",
                fetched.ret_msg
            )));
        }
        Ok(fetched.result.list.into_iter().next())
    }
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct BybitCreateOrderRequest<'a> {
    category: &'a str,
    symbol: &'a str,
    side: &'a str,
    #[serde(rename = "orderType")]
    order_type: &'a str,
    qty: &'a str,
    #[serde(rename = "timeInForce")]
    time_in_force: &'a str,
    #[serde(rename = "orderLinkId")]
    order_link_id: &'a str,
    #[serde(rename = "isLeverage")]
    is_leverage: i32,
    #[serde(rename = "orderFilter")]
    order_filter: &'a str,
    #[serde(rename = "marketUnit")]
    market_unit: &'a str,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct BybitCreateOrderResponse {
    ret_code: i64,
    ret_msg: String,
    result: BybitCreateOrderResult,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct BybitCreateOrderResult {
    order_id: String,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct BybitOrderRealtimeResponse {
    ret_code: i64,
    ret_msg: String,
    result: BybitOrderRealtimeResult,
}

#[derive(Debug, Deserialize)]
struct BybitOrderRealtimeResult {
    #[serde(default)]
    list: Vec<BybitOrder>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct BybitOrder {
    order_id: String,
    order_status: String,
    #[serde(default)]
    cum_exec_qty: String,
    #[serde(default)]
    cum_exec_value: String,
    #[serde(default)]
    cum_exec_fee: String,
    #[serde(default)]
    cum_fee_detail: Option<HashMap<String, String>>,
}
