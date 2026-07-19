use async_nats::jetstream;
use futures_util::StreamExt;
use mautrade_rust_node_execution::adapters::{
    BinanceExecutionClient, BybitExecutionClient, OkxExecutionClient, TokocryptoExecutionClient,
};
use mautrade_rust_node_execution::engine::{failed_report, ExecutionRouter, StaticRouter};
use mautrade_rust_node_execution::ExecutionRequest;
use std::sync::Arc;
use tracing::{error, info};

const EXECUTION_STREAM: &str = "EXECUTION";
const RESULT_SUBJECT: &str = "execution.result";
const DLQ_SUBJECT: &str = "execution.dlq";

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    tracing_subscriber::fmt().json().init();

    let nats_url = std::env::var("NATS_URL").unwrap_or_else(|_| "nats://localhost:4222".to_string());
    let durable_name = std::env::var("EXECUTOR_DURABLE").unwrap_or_else(|_| "mautrade-rust-executor".to_string());

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

    let router = StaticRouter::new()
        .with_client("binance", Arc::new(BinanceExecutionClient))
        .with_client("okx", Arc::new(OkxExecutionClient))
        .with_client("bybit", Arc::new(BybitExecutionClient))
        .with_client("tokocrypto", Arc::new(TokocryptoExecutionClient));

    info!(durable = durable_name, "mautrade rust execution worker started");
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

        let report = match router.execute(request.clone()).await {
            Ok(report) => report,
            Err(err) => failed_report(&request, "execution_error", err.to_string()),
        };
        let report_payload = serde_json::to_vec(&report)?;
        js.publish(RESULT_SUBJECT, report_payload.into()).await?.await?;
        message.double_ack().await?;
    }

    Ok(())
}
