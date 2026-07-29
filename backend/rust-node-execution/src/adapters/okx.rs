use crate::adapters::common::{
    account_mode, client_order_id, decimal_from_str, form_body, hmac_sha256_base64, okx_symbol, okx_timestamp,
    quantity, quote_value, report, response_text, side_lower, status_from_fill, sum_decimal_strings,
};
use crate::credentials::InternalCredentialProvider;
use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus, OrderSide};
use async_trait::async_trait;
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::Duration;
use tokio::time::sleep;

const BASE_URL: &str = "https://openapi.okx.com";

#[derive(Debug, Clone)]
pub struct OkxExecutionClient {
    credentials: Arc<InternalCredentialProvider>,
    http: reqwest::Client,
}

impl OkxExecutionClient {
    pub fn new(credentials: Arc<InternalCredentialProvider>) -> Self {
        Self {
            credentials,
            http: reqwest::Client::new(),
        }
    }
}

#[async_trait]
impl ExchangeExecutionClient for OkxExecutionClient {
    async fn place_order(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        let credentials = self.credentials.fetch_for_order(&req).await?;
        if credentials.exchange != "okx" {
            return Err(ExecutionError::InvalidOrder(
                "credential exchange does not match okx adapter".to_string(),
            ));
        }
        if credentials.api_passphrase.trim().is_empty() {
            return Err(ExecutionError::InvalidOrder(
                "okx api passphrase is required".to_string(),
            ));
        }

        let account_mode = account_mode(&req, &credentials.account_mode);
        let inst_id = okx_symbol(&req.symbol);
        let size = match req.side {
            OrderSide::Buy => quote_value(&req)?.to_string(),
            OrderSide::Sell => quantity(&req)?.to_string(),
        };
        let target_currency = match req.side {
            OrderSide::Buy => "quote_ccy",
            OrderSide::Sell => "base_ccy",
        };
        let client_id = client_order_id(&req, 32);
        let order = OkxPlaceOrderRequest {
            inst_id: &inst_id,
            td_mode: "cash",
            client_order_id: &client_id,
            side: side_lower(&req.side),
            order_type: "market",
            size: &size,
            target_currency,
            ban_amend: true,
        };
        let body_json = serde_json::to_string(&order)
            .map_err(|err| ExecutionError::Exchange(format!("okx encode order request failed: {err}")))?;
        let timestamp = okx_timestamp();
        let path = "/api/v5/trade/order";
        let signature = hmac_sha256_base64(&credentials.api_secret, &(timestamp.clone() + "POST" + path + &body_json))?;

        let mut request = self
            .http
            .post(format!("{BASE_URL}{path}"))
            .header("OK-ACCESS-KEY", credentials.api_key.clone())
            .header("OK-ACCESS-SIGN", signature)
            .header("OK-ACCESS-TIMESTAMP", timestamp)
            .header("OK-ACCESS-PASSPHRASE", credentials.api_passphrase.clone())
            .header(reqwest::header::CONTENT_TYPE, "application/json")
            .body(body_json);
        if account_mode == "demo" || account_mode == "testnet" {
            request = request.header("x-simulated-trading", "1");
        }

        let response = request
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("okx place order request failed: {err}")))?;
        let body = response_text(response, "okx place order").await?;
        let created: OkxPlaceOrderResponse = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("okx order response decode failed: {err}")))?;
        if created.code != "0" {
            return Err(ExecutionError::Exchange(format!("okx rejected order: {}", created.message)));
        }
        let accepted = created
            .data
            .into_iter()
            .next()
            .ok_or_else(|| ExecutionError::Exchange("okx order response missing data".to_string()))?;
        if accepted.status_code != "0" {
            return Err(ExecutionError::Exchange(format!(
                "okx rejected order: {}",
                accepted.status_message
            )));
        }

        self.poll_order_report(&credentials, &account_mode, &req, &inst_id, &accepted.order_id)
            .await
    }
}

impl OkxExecutionClient {
    async fn poll_order_report(
        &self,
        credentials: &crate::credentials::ExchangeCredentials,
        account_mode: &str,
        req: &ExecutionRequest,
        inst_id: &str,
        order_id: &str,
    ) -> Result<ExecutionReport, ExecutionError> {
        let mut last_state = String::new();
        for attempt in 0..12 {
            match self.fetch_order(credentials, account_mode, inst_id, order_id).await {
                Ok(Some(order)) => {
                    last_state = order.state.clone();
                    let filled_quantity = decimal_from_str(&order.acc_fill_size);
                    let avg_price = decimal_from_str(&order.avg_price);
                    if filled_quantity > Decimal::ZERO && avg_price > Decimal::ZERO {
                        let fill_value_quote = filled_quantity * avg_price;
                        let fee = sum_decimal_strings([order.fee.as_str()]);
                        return Ok(report(
                            req,
                            status_from_fill(filled_quantity, &order.state),
                            filled_quantity,
                            fill_value_quote,
                            fee,
                            Some(order.order_id),
                            Some(&order.state),
                        ));
                    }
                    if order.state == "canceled" || order.state == "order_failed" {
                        return Ok(report(
                            req,
                            ExecutionStatus::Failed,
                            Decimal::ZERO,
                            Decimal::ZERO,
                            Decimal::ZERO,
                            Some(order.order_id),
                            Some(&order.state),
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
            Decimal::ZERO,
            Decimal::ZERO,
            Decimal::ZERO,
            Some(order_id.to_string()),
            Some(if last_state.is_empty() {
                "order status was not available after polling"
            } else {
                &last_state
            }),
        ))
    }

    async fn fetch_order(
        &self,
        credentials: &crate::credentials::ExchangeCredentials,
        account_mode: &str,
        inst_id: &str,
        order_id: &str,
    ) -> Result<Option<OkxOrder>, ExecutionError> {
        let query = form_body(&[("instId", inst_id.to_string()), ("ordId", order_id.to_string())]);
        let path_with_query = format!("/api/v5/trade/order?{query}");
        let timestamp = okx_timestamp();
        let signature = hmac_sha256_base64(&credentials.api_secret, &(timestamp.clone() + "GET" + &path_with_query))?;

        let mut request = self
            .http
            .get(format!("{BASE_URL}{path_with_query}"))
            .header("OK-ACCESS-KEY", credentials.api_key.clone())
            .header("OK-ACCESS-SIGN", signature)
            .header("OK-ACCESS-TIMESTAMP", timestamp)
            .header("OK-ACCESS-PASSPHRASE", credentials.api_passphrase.clone());
        if account_mode == "demo" || account_mode == "testnet" {
            request = request.header("x-simulated-trading", "1");
        }

        let response = request
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("okx fetch order request failed: {err}")))?;
        let body = response_text(response, "okx fetch order").await?;
        let fetched: OkxOrderResponse = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("okx fetch order decode failed: {err}")))?;
        if fetched.code != "0" {
            return Err(ExecutionError::Exchange(format!(
                "okx rejected order lookup: {}",
                fetched.message
            )));
        }
        Ok(fetched.data.into_iter().next())
    }
}

#[derive(Debug, Serialize)]
struct OkxPlaceOrderRequest<'a> {
    #[serde(rename = "instId")]
    inst_id: &'a str,
    #[serde(rename = "tdMode")]
    td_mode: &'a str,
    #[serde(rename = "clOrdId")]
    client_order_id: &'a str,
    side: &'a str,
    #[serde(rename = "ordType")]
    order_type: &'a str,
    #[serde(rename = "sz")]
    size: &'a str,
    #[serde(rename = "tgtCcy")]
    target_currency: &'a str,
    #[serde(rename = "banAmend")]
    ban_amend: bool,
}

#[derive(Debug, Deserialize)]
struct OkxPlaceOrderResponse {
    code: String,
    #[serde(default, rename = "msg")]
    message: String,
    #[serde(default)]
    data: Vec<OkxPlaceOrderData>,
}

#[derive(Debug, Deserialize)]
struct OkxPlaceOrderData {
    #[serde(default, rename = "ordId")]
    order_id: String,
    #[serde(default, rename = "sCode")]
    status_code: String,
    #[serde(default, rename = "sMsg")]
    status_message: String,
}

#[derive(Debug, Deserialize)]
struct OkxOrderResponse {
    code: String,
    #[serde(default, rename = "msg")]
    message: String,
    #[serde(default)]
    data: Vec<OkxOrder>,
}

#[derive(Debug, Deserialize)]
struct OkxOrder {
    #[serde(default, rename = "ordId")]
    order_id: String,
    #[serde(default)]
    state: String,
    #[serde(default, rename = "accFillSz")]
    acc_fill_size: String,
    #[serde(default, rename = "avgPx")]
    avg_price: String,
    #[serde(default)]
    fee: String,
}
