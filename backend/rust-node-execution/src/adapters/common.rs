use crate::engine::ExecutionError;
use crate::types::{ExecutionReport, ExecutionRequest, ExecutionStatus, OrderSide};
use base64::Engine;
use hmac::{Hmac, Mac};
use rust_decimal::Decimal;
use sha2::Sha256;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use time::format_description::well_known::Rfc3339;
use time::OffsetDateTime;
use url::form_urlencoded;

type HmacSha256 = Hmac<Sha256>;

pub fn account_mode(req: &ExecutionRequest, credential_mode: &str) -> String {
    req.account_mode
        .as_deref()
        .map(str::trim)
        .filter(|value| !value.is_empty())
        .unwrap_or(credential_mode)
        .to_ascii_lowercase()
}

pub fn binance_symbol(symbol: &str) -> String {
    symbol
        .chars()
        .filter(|ch| *ch != '/' && *ch != '-' && *ch != '_')
        .collect::<String>()
        .to_ascii_uppercase()
}

pub fn okx_symbol(symbol: &str) -> String {
    symbol.replace(['/', '_'], "-").to_ascii_uppercase()
}

pub fn tokocrypto_symbol(symbol: &str) -> String {
    symbol.replace(['/', '-'], "_").to_ascii_uppercase()
}

pub fn client_order_id(req: &ExecutionRequest, max_len: usize) -> String {
    let mut id = String::from("mt");
    for ch in req.idempotency_key.chars().chain(req.id.chars()) {
        if ch.is_ascii_alphanumeric() {
            id.push(ch);
        }
        if id.len() >= max_len {
            break;
        }
    }
    if id.len() <= 2 {
        id.push_str(&millis().to_string());
    }
    id.truncate(max_len);
    id
}

pub fn quote_value(req: &ExecutionRequest) -> Result<Decimal, ExecutionError> {
    positive_decimal(req.quote_value, "quote_value")
}

pub fn quantity(req: &ExecutionRequest) -> Result<Decimal, ExecutionError> {
    positive_decimal(req.quantity, "quantity")
}

fn positive_decimal(value: Option<Decimal>, field: &str) -> Result<Decimal, ExecutionError> {
    let value = value.ok_or_else(|| ExecutionError::InvalidOrder(format!("{field} is required")))?;
    if value <= Decimal::ZERO {
        return Err(ExecutionError::InvalidOrder(format!("{field} must be greater than zero")));
    }
    Ok(value)
}

pub fn decimal_from_str(value: &str) -> Decimal {
    value.trim().parse::<Decimal>().unwrap_or(Decimal::ZERO)
}

pub fn sum_decimal_strings<'a>(values: impl IntoIterator<Item = &'a str>) -> Decimal {
    values
        .into_iter()
        .fold(Decimal::ZERO, |sum, value| sum + decimal_from_str(value).abs())
}

pub fn average_price(quantity: Decimal, value_quote: Decimal) -> Decimal {
    if quantity <= Decimal::ZERO {
        Decimal::ZERO
    } else {
        value_quote / quantity
    }
}

pub fn status_from_fill(quantity: Decimal, exchange_status: &str) -> ExecutionStatus {
    let status = exchange_status.trim().to_ascii_uppercase();
    if quantity <= Decimal::ZERO {
        ExecutionStatus::Failed
    } else if status == "PARTIALLY_FILLED" || status == "PARTIAL" || status == "PARTIALLYFILLED" {
        ExecutionStatus::Partial
    } else {
        ExecutionStatus::Success
    }
}

pub fn report(
    req: &ExecutionRequest,
    status: ExecutionStatus,
    filled_quantity: Decimal,
    fill_value_quote: Decimal,
    exchange_fee: Decimal,
    exchange_order_id: Option<String>,
    exchange_status: Option<&str>,
) -> ExecutionReport {
    let mut error_code = None;
    let mut error_message = None;
    if status == ExecutionStatus::Failed {
        error_code = Some("order_not_filled".to_string());
        error_message = Some(
            exchange_status
                .filter(|value| !value.trim().is_empty())
                .map(|value| format!("exchange order was not filled: {value}"))
                .unwrap_or_else(|| "exchange order was not filled".to_string()),
        );
    }

    ExecutionReport {
        request_id: req.id.clone(),
        idempotency_key: req.idempotency_key.clone(),
        master_signal_id: req.master_signal_id.clone(),
        user_id: req.user_id.clone(),
        layer_id: req.layer_id.clone(),
        exchange: req.exchange.clone(),
        symbol: req.symbol.clone(),
        side: req.side.clone(),
        status,
        filled_quantity,
        fill_price: average_price(filled_quantity, fill_value_quote),
        fill_value_quote,
        exchange_fee,
        exchange_order_id,
        error_code,
        error_message,
        executed_at: now_rfc3339(),
    }
}

pub fn form_body(pairs: &[(&str, String)]) -> String {
    let mut serializer = form_urlencoded::Serializer::new(String::new());
    for (key, value) in pairs {
        serializer.append_pair(key, value);
    }
    serializer.finish()
}

pub fn hmac_sha256_hex(secret: &str, payload: &str) -> Result<String, ExecutionError> {
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes())
        .map_err(|err| ExecutionError::Exchange(format!("hmac init failed: {err}")))?;
    mac.update(payload.as_bytes());
    Ok(hex::encode(mac.finalize().into_bytes()))
}

pub fn hmac_sha256_base64(secret: &str, payload: &str) -> Result<String, ExecutionError> {
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes())
        .map_err(|err| ExecutionError::Exchange(format!("hmac init failed: {err}")))?;
    mac.update(payload.as_bytes());
    Ok(base64::engine::general_purpose::STANDARD.encode(mac.finalize().into_bytes()))
}

pub async fn response_text(response: reqwest::Response, label: &str) -> Result<String, ExecutionError> {
    let status = response.status();
    let body = response
        .text()
        .await
        .map_err(|err| ExecutionError::Exchange(format!("{label}: read response failed: {err}")))?;
    if !status.is_success() {
        return Err(ExecutionError::Exchange(format!(
            "{label}: exchange returned {status}: {}",
            compact_body(&body)
        )));
    }
    Ok(body)
}

pub fn compact_body(body: &str) -> String {
    let compact = body.split_whitespace().collect::<Vec<_>>().join(" ");
    if compact.chars().count() > 240 {
        compact.chars().take(240).collect::<String>() + "..."
    } else {
        compact
    }
}

pub fn millis() -> u128 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_else(|_| Duration::from_secs(0))
        .as_millis()
}

pub fn now_rfc3339() -> String {
    OffsetDateTime::now_utc()
        .format(&Rfc3339)
        .unwrap_or_else(|_| "1970-01-01T00:00:00Z".to_string())
}

pub fn okx_timestamp() -> String {
    OffsetDateTime::now_utc()
        .format(&Rfc3339)
        .unwrap_or_else(|_| "1970-01-01T00:00:00Z".to_string())
}

pub fn side_upper(side: &OrderSide) -> &'static str {
    match side {
        OrderSide::Buy => "BUY",
        OrderSide::Sell => "SELL",
    }
}

pub fn side_title(side: &OrderSide) -> &'static str {
    match side {
        OrderSide::Buy => "Buy",
        OrderSide::Sell => "Sell",
    }
}

pub fn side_lower(side: &OrderSide) -> &'static str {
    match side {
        OrderSide::Buy => "buy",
        OrderSide::Sell => "sell",
    }
}
