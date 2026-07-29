use async_nats::jetstream;
use futures_util::StreamExt;
use mautrade_rust_node_execution::adapters::{
    BinanceExecutionClient, BybitExecutionClient, OkxExecutionClient, TokocryptoExecutionClient,
};
use mautrade_rust_node_execution::credentials::InternalCredentialProvider;
use mautrade_rust_node_execution::engine::{
    failed_report, stale_request_report, ExecutionRouter, StaticRouter,
};
use mautrade_rust_node_execution::ExecutionRequest;
use std::sync::Arc;
use std::time::Duration;
use tracing::{error, info};

const EXECUTION_STREAM: &str = "EXECUTION";
const RESULT_SUBJECT: &str = "execution.result";
const DLQ_SUBJECT: &str = "execution.dlq";

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    tracing_subscriber::fmt().json().init();

    let nats_url = std::env::var("NATS_URL").unwrap_or_else(|_| "nats://localhost:4222".to_string());
    let durable_name = std::env::var("EXECUTOR_DURABLE").unwrap_or_else(|_| "mautrade-rust-executor".to_string());
    let max_request_age = execution_max_request_age();

    let client = async_nats::connect(nats_url).await?;
    let js = jetstream::new(client);
    let stream = js.get_stream(EXECUTION_STREAM).await?;
    let consumer = stream
        .get_or_create_consumer(
            &durable_name,
            jetstream::consumer::pull::Config {
                durable_name: Some(durable_name.clone()),
                ..Default::default()
            },
        )
        .await?;

    let credential_provider = Arc::new(InternalCredentialProvider::from_env()?);
    let router = StaticRouter::new()
        .with_client("binance", Arc::new(BinanceExecutionClient::new(credential_provider.clone())))
        .with_client("okx", Arc::new(OkxExecutionClient::new(credential_provider.clone())))
        .with_client("bybit", Arc::new(BybitExecutionClient::new(credential_provider.clone())))
        .with_client(
            "tokocrypto",
            Arc::new(TokocryptoExecutionClient::new(credential_provider.clone())),
        );

    info!(
        durable = durable_name,
        max_request_age_seconds = max_request_age.as_secs(),
        "mautrade rust execution worker started"
    );
    let mut messages = consumer.messages().await?;
    while let Some(message) = messages.next().await {
        let message = match message {
            Ok(message) => message,
            Err(err) => {
                error!(error = %err, "failed to receive execution message");
                continue;
            }
        };

        let request = match serde_json::from_slice::<ExecutionRequest>(&message.payload) {
            Ok(request) => request,
            Err(err) => {
                error!(error = %err, "invalid execution payload");
                js.publish(DLQ_SUBJECT, message.payload.clone()).await?.await?;
                message.ack().await?;
                continue;
            }
        };

        let report = match stale_request_report(&request, max_request_age) {
            Some(report) => report,
            None => match router.execute(request.clone()).await {
                Ok(report) => report,
                Err(err) => failed_report(&request, "execution_error", err.to_string()),
            },
        };
        let report_payload = serde_json::to_vec(&report)?;
        js.publish(RESULT_SUBJECT, report_payload.into()).await?.await?;
        message.double_ack().await?;
    }

    Ok(())
}

fn execution_max_request_age() -> Duration {
    let seconds = std::env::var("EXECUTION_MAX_REQUEST_AGE_SECONDS")
        .ok()
        .and_then(|value| value.trim().parse::<u64>().ok())
        .unwrap_or(300);
    Duration::from_secs(seconds)
}
