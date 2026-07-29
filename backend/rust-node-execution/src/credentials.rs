use crate::engine::ExecutionError;
use crate::types::{ExecutionRequest, OrderSide};
use serde::Deserialize;
use std::sync::Arc;
use std::time::Duration;

const INTERNAL_TOKEN_HEADER: &str = "X-Mautrade-Internal-Token";

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ExchangeCredentials {
    pub id: String,
    pub exchange: String,
    pub account_mode: String,
    pub status: String,
    pub api_key: String,
    pub api_secret: String,
    #[serde(default)]
    pub api_passphrase: String,
    pub permission_scope: String,
}

#[derive(Debug, Clone)]
pub struct InternalCredentialProvider {
    client: reqwest::Client,
    base_url: Arc<str>,
    token: Arc<str>,
}

impl InternalCredentialProvider {
    pub fn from_env() -> Result<Self, ExecutionError> {
        let base_url = std::env::var("MAUTRADE_INTERNAL_API_URL")
            .unwrap_or_else(|_| "http://localhost:8080".to_string())
            .trim()
            .trim_end_matches('/')
            .to_string();
        if base_url.is_empty() {
            return Err(ExecutionError::InvalidOrder(
                "MAUTRADE_INTERNAL_API_URL is required".to_string(),
            ));
        }

        let token = internal_token_from_values(
            std::env::var("EXECUTION_INTERNAL_TOKEN").ok(),
            std::env::var("EXCHANGE_CREDENTIAL_KEY").ok(),
        );
        if token.is_empty() {
            return Err(ExecutionError::InvalidOrder(
                "EXECUTION_INTERNAL_TOKEN or EXCHANGE_CREDENTIAL_KEY is required".to_string(),
            ));
        }

        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(12))
            .build()
            .map_err(|err| ExecutionError::Exchange(format!("credential http client: {err}")))?;

        Ok(Self {
            client,
            base_url: Arc::from(base_url),
            token: Arc::from(token),
        })
    }

    pub async fn fetch_for_order(&self, req: &ExecutionRequest) -> Result<ExchangeCredentials, ExecutionError> {
        let credentials = self.fetch(&req.exchange_binding_id).await?;
        if credential_allows_side(&credentials.status, &req.side) {
            return Ok(credentials);
        }
        if credentials.status.trim().eq_ignore_ascii_case("closing_only") {
            return Err(ExecutionError::InvalidOrder(
                "exchange binding is closing-only and cannot accept new buy orders".to_string(),
            ));
        }
        Err(ExecutionError::InvalidOrder(
            "exchange binding is not active".to_string(),
        ))
    }

    async fn fetch(&self, binding_id: &str) -> Result<ExchangeCredentials, ExecutionError> {
        let binding_id = binding_id.trim();
        if binding_id.is_empty() {
            return Err(ExecutionError::InvalidOrder(
                "exchange_binding_id is required".to_string(),
            ));
        }

        let url = format!(
            "{}/api/v1/internal/exchange-bindings/{}/credentials",
            self.base_url, binding_id
        );
        let response = self
            .client
            .get(url)
            .header(INTERNAL_TOKEN_HEADER, self.token.as_ref())
            .send()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("fetch exchange credential failed: {err}")))?;

        let status = response.status();
        let body = response
            .text()
            .await
            .map_err(|err| ExecutionError::Exchange(format!("read exchange credential response failed: {err}")))?;
        if !status.is_success() {
            return Err(ExecutionError::Exchange(format!(
                "fetch exchange credential returned {status}: {}",
                compact_error_body(&body)
            )));
        }

        let credentials: ExchangeCredentials = serde_json::from_str(&body)
            .map_err(|err| ExecutionError::Exchange(format!("decode exchange credential failed: {err}")))?;
        if credentials.api_key.trim().is_empty() || credentials.api_secret.trim().is_empty() {
            return Err(ExecutionError::InvalidOrder(
                "exchange credential is missing api key or secret".to_string(),
            ));
        }
        Ok(credentials)
    }
}

fn internal_token_from_values(execution_token: Option<String>, exchange_key: Option<String>) -> String {
    execution_token
        .as_deref()
        .map(str::trim)
        .filter(|value| !value.is_empty())
        .or_else(|| exchange_key.as_deref().map(str::trim).filter(|value| !value.is_empty()))
        .unwrap_or_default()
        .to_string()
}

fn credential_allows_side(status: &str, side: &OrderSide) -> bool {
    if status.trim().eq_ignore_ascii_case("active") {
        return true;
    }
    status.trim().eq_ignore_ascii_case("closing_only") && matches!(side, OrderSide::Sell)
}

fn compact_error_body(body: &str) -> String {
    let compact = body.split_whitespace().collect::<Vec<_>>().join(" ");
    if compact.chars().count() > 180 {
        compact.chars().take(180).collect::<String>() + "..."
    } else {
        compact
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn active_credentials_allow_buy_and_sell() {
        assert!(credential_allows_side("active", &OrderSide::Buy));
        assert!(credential_allows_side("active", &OrderSide::Sell));
    }

    #[test]
    fn closing_only_credentials_allow_sell_only() {
        assert!(!credential_allows_side("closing_only", &OrderSide::Buy));
        assert!(credential_allows_side("closing_only", &OrderSide::Sell));
    }

    #[test]
    fn internal_token_prefers_execution_token() {
        let token = internal_token_from_values(
            Some(" execution-token ".to_string()),
            Some("exchange-key".to_string()),
        );

        assert_eq!(token, "execution-token");
    }

    #[test]
    fn internal_token_falls_back_to_exchange_credential_key() {
        let token = internal_token_from_values(Some(" ".to_string()), Some(" exchange-key ".to_string()));

        assert_eq!(token, "exchange-key");
    }
}
