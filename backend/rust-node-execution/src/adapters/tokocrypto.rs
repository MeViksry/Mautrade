use crate::adapters::common::{
    account_mode, client_order_id, decimal_from_str, form_body, hmac_sha256_hex, millis, quantity, quote_value,
    report, response_text, status_from_fill, sum_decimal_strings, tokocrypto_symbol,
};
use crate::credentials::InternalCredentialProvider;
use crate::engine::{ExecutionError, ExchangeExecutionClient};
use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus, OrderSide};
use async_trait::async_trait;
use rust_decimal::Decimal;
use serde::Deserialize;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::sleep;

const BASE_URL: &str = "https://www.tokocrypto.com";

#[derive(Debug, Clone)]
pub struct TokocryptoExecutionClient {
    credentials: Arc<InternalCredentialProvider>,
    http: reqwest::Client,
}

impl TokocryptoExecutionClient {
    pub fn new(credentials: Arc<InternalCredentialProvider>) -> Self {
        Self {
            credentials,
            http: reqwest::Client::new(),
        }
    }
}

#[async_trait]
impl ExchangeExecutionClient for TokocryptoExecutionClient {
    async fn place_order(&self, req: ExecutionRequest) -> Result<ExecutionReport, ExecutionError> {
        let credentials = self.credentials.fetch_for_order(&req).await?;
        if credentials.exchange != "tokocrypto" {
            return Err(ExecutionError::InvalidOrder(
                "credential exchange does not match tokocrypto adapter".to_string(),
            ));
        }
        let account_mode = account_mode(&req, &credentials.account_mode);
        if account_mode != "real" {
            return Err(ExecutionError::InvalidOrder(
                "tokocrypto demo/testnet spot execution is not supported".to_string(),
            ));
        }

        let symbol = tokocrypto_symbol(&req.symbol);
        let mut pairs = vec![
            ("symbol", symbol.clone()),
            ("side", tokocrypto_side(&req.side).to_string()),
            ("type", "2".to_string()),
            ("timeInForce", "2".to_string()),
            ("clientId", client_order_id(&req, 36)),
            ("recvWindow", "5000".to_string()),
            ("timestamp", millis().to_string()),
        ];
        match req.side {
            OrderSide::Buy => pairs.push(("quoteOrderQty", quote_value(&req)?.to_string())),
            OrderSide::Sell => {
                let mut q = quantity(&req)?;
                if let Ok(step_size) = self.fetch_step_size(&symbol).await {
                    if step_size > rust_decimal::Decimal::ZERO {
                        q = (q / step_size).floor() * step_size;
                    }
                }
                let qty_str = q.normalize().to_string();
                pairs.push(("quantity", qty_str));
            }
        }
        let unsigned_body = form_body(&pairs);
        let signature = hmac_sha256_hex(&credentials.api_secret, &unsigned_body)?;
        let body = format!("{unsigned_body}&signature={signature}");

        let response = self
            .http
            .post(format!("{BASE_URL}/open/v1/orders"))
            .header("X-MBX-APIKEY", credentials.api_key.clone())
            .header(reqwest::header::CONTENT_TYPE, "application/x-www-form-urlencoded")
            .body(body)
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("tokocrypto place order request failed: {err}")))?;
        let body = response_text(response, "tokocrypto place order").await?;
        let created: TokocryptoOrderEnvelope = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("tokocrypto order response decode failed: {err}")))?;
        if created.code != 0 {
            return Err(ExecutionError::Exchange(format!(
                "tokocrypto rejected order: {}",
                created.message
            )));
        }
        let order = created
            .data
            .ok_or_else(|| ExecutionError::Exchange("tokocrypto order response missing data".to_string()))?;
        if let Some(report) = tokocrypto_order_report(&req, order.clone()) {
            return Ok(report);
        }
        self.poll_order_report(&credentials, &req, &order.order_id_string()).await
    }
}

impl TokocryptoExecutionClient {
    async fn poll_order_report(
        &self,
        credentials: &crate::credentials::ExchangeCredentials,
        req: &ExecutionRequest,
        order_id: &str,
    ) -> Result<ExecutionReport, ExecutionError> {
        let mut last_status = String::new();
        for attempt in 0..12 {
            match self.fetch_order(credentials, order_id).await {
                Ok(Some(order)) => {
                    last_status = tokocrypto_status_name(order.status).to_string();
                    if let Some(report) = tokocrypto_order_report(req, order.clone()) {
                        return Ok(report);
                    }
                    if matches!(order.status, 3 | 5 | 6) {
                        return Ok(report(
                            req,
                            ExecutionStatus::Failed,
                            Decimal::ZERO,
                            Decimal::ZERO,
                            Decimal::ZERO,
                            Some(order.order_id_string()),
                            Some(tokocrypto_status_name(order.status)),
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
            Some(if last_status.is_empty() {
                "order status was not available after polling"
            } else {
                &last_status
            }),
        ))
    }

    async fn fetch_order(
        &self,
        credentials: &crate::credentials::ExchangeCredentials,
        order_id: &str,
    ) -> Result<Option<TokocryptoOrder>, ExecutionError> {
        let mut pairs = vec![
            ("orderId", order_id.to_string()),
            ("recvWindow", "5000".to_string()),
            ("timestamp", millis().to_string()),
        ];
        let unsigned_query = form_body(&pairs);
        let signature = hmac_sha256_hex(&credentials.api_secret, &unsigned_query)?;
        pairs.push(("signature", signature));
        let query = form_body(&pairs);

        let response = self
            .http
            .get(format!("{BASE_URL}/open/v1/orders/detail?{query}"))
            .header("X-MBX-APIKEY", credentials.api_key.clone())
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("tokocrypto fetch order request failed: {err}")))?;
        let body = response_text(response, "tokocrypto fetch order").await?;
        let fetched: TokocryptoOrderEnvelope = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("tokocrypto fetch order decode failed: {err}")))?;
        if fetched.code != 0 {
            return Err(ExecutionError::Exchange(format!(
                "tokocrypto rejected order lookup: {}",
                fetched.message
            )));
        }
        Ok(fetched.data)
    }

    async fn fetch_step_size(&self, symbol: &str) -> Result<rust_decimal::Decimal, ExecutionError> {
        // Tokocrypto doesn't need to filter by symbol here, as their endpoint doesn't support the symbol query properly.
        // We fetch all symbols (cached or fast enough) and search for the symbol.
        let response = self
            .http
            .get(format!("{BASE_URL}/open/v1/common/symbols"))
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("tokocrypto common/symbols request failed: {err}")))?;
        let body = response_text(response, "tokocrypto common/symbols").await?;
        let fetched: TokocryptoExchangeInfoResponse = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("tokocrypto common/symbols decode failed: {err}")))?;
        
        // The API returns all symbols, we must find ours
        for symbol_info in fetched.data.list {
            if symbol_info.symbol == symbol {
                for filter in symbol_info.filters {
                    if filter.filter_type == "LOT_SIZE" {
                        return Ok(crate::adapters::common::decimal_from_str(&filter.step_size));
                    }
                }
            }
        }
        Ok(rust_decimal::Decimal::ZERO)
    }
}

fn tokocrypto_side(side: &OrderSide) -> i32 {
    match side {
        OrderSide::Buy => 0,
        OrderSide::Sell => 1,
    }
}

fn tokocrypto_status_name(status: i32) -> &'static str {
    match status {
        1 => "PARTIALLY_FILLED",
        2 => "FILLED",
        3 => "CANCELED",
        5 => "REJECTED",
        6 => "EXPIRED",
        -2 => "SYSTEM_PROCESSING",
        _ => "NEW",
    }
}

fn tokocrypto_order_report(req: &ExecutionRequest, order: TokocryptoOrder) -> Option<ExecutionReport> {
    let filled_quantity = decimal_from_str(&order.executed_qty);
    let fill_value_quote = decimal_from_str(&order.executed_quote_qty);
    if filled_quantity <= Decimal::ZERO || fill_value_quote <= Decimal::ZERO {
        return None;
    }
    let status_name = tokocrypto_status_name(order.status);
    let fee = sum_decimal_strings([order.tax_fee.as_str()]);
    Some(report(
        req,
        status_from_fill(filled_quantity, status_name),
        filled_quantity,
        fill_value_quote,
        fee,
        Some(order.order_id_string()),
        Some(status_name),
    ))
}

#[derive(Debug, Deserialize)]
struct TokocryptoOrderEnvelope {
    code: i32,
    #[serde(default, alias = "msg", alias = "message")]
    message: String,
    #[serde(default)]
    data: Option<TokocryptoOrder>,
}

#[derive(Debug, Clone, Deserialize)]
struct TokocryptoOrder {
    #[serde(default, rename = "orderId")]
    order_id: serde_json::Value,
    #[serde(default)]
    status: i32,
    #[serde(default, rename = "executedQty")]
    executed_qty: String,
    #[serde(default, rename = "executedQuoteQty")]
    executed_quote_qty: String,
    #[serde(default, rename = "taxFee")]
    tax_fee: String,
}

impl TokocryptoOrder {
    fn order_id_string(&self) -> String {
        match &self.order_id {
            serde_json::Value::String(value) => value.clone(),
            serde_json::Value::Number(value) => value.to_string(),
            _ => String::new(),
        }
    }
}

#[derive(Debug, Deserialize, Default)]
struct TokocryptoExchangeInfoResponse {
    #[serde(default)]
    data: TokocryptoExchangeInfoData,
}

#[derive(Debug, Deserialize, Default)]
struct TokocryptoExchangeInfoData {
    #[serde(default)]
    list: Vec<TokocryptoSymbolInfo>,
}

#[derive(Debug, Deserialize, Default)]
struct TokocryptoSymbolInfo {
    #[serde(default)]
    symbol: String,
    #[serde(default)]
    filters: Vec<TokocryptoFilter>,
}

#[derive(Debug, Deserialize, Default)]
#[serde(rename_all = "camelCase")]
struct TokocryptoFilter {
    #[serde(default)]
    filter_type: String,
    #[serde(default)]
    step_size: String,
}

