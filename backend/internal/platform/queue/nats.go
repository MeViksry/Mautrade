package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	ExecutionStreamName       = "EXECUTION"
	ExecutionResultStreamName = "EXECUTION_RESULTS"
	ExecutionResultConsumer   = "mautrade-go-api-results"
	BuyRequestSubject         = "execution.buy.request"
	SellRequestSubject        = "execution.sell.request"
	ResultSubject             = "execution.result"
	DeadLetterSubject         = "execution.dlq"
)

type Client struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

func Connect(ctx context.Context, url string) (*Client, error) {
	conn, err := nats.Connect(url, nats.Name("mautrade-go-api"), nats.Timeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("nats: connect: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("nats: jetstream: %w", err)
	}

	client := &Client{conn: conn, js: js}
	if err := client.EnsureExecutionStream(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	if err := client.EnsureExecutionResultStream(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) Close() {
	if c == nil || c.conn == nil {
		return
	}
	c.conn.Drain()
	c.conn.Close()
}

func (c *Client) EnsureExecutionStream(ctx context.Context) error {
	if c == nil {
		return nil
	}
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: ExecutionStreamName,
		Subjects: []string{
			BuyRequestSubject,
			SellRequestSubject,
		},
		Retention: jetstream.WorkQueuePolicy,
		Storage:   jetstream.FileStorage,
		Replicas:  1,
		MaxAge:    7 * 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("nats: ensure execution stream: %w", err)
	}
	return nil
}

func (c *Client) EnsureExecutionResultStream(ctx context.Context) error {
	if c == nil {
		return nil
	}
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: ExecutionResultStreamName,
		Subjects: []string{
			ResultSubject,
			DeadLetterSubject,
		},
		Retention: jetstream.LimitsPolicy,
		Storage:   jetstream.FileStorage,
		Replicas:  1,
		MaxAge:    14 * 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("nats: ensure execution result stream: %w", err)
	}
	return nil
}

type ExecutionRequest struct {
	ID             string `json:"id"`
	IdempotencyKey string `json:"idempotency_key"`
	MasterSignalID string `json:"master_signal_id"`
	UserID         string `json:"user_id"`
	LayerID        string `json:"layer_id,omitempty"`
	Exchange       string `json:"exchange"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	Quantity       string `json:"quantity,omitempty"`
	QuoteValue     string `json:"quote_value,omitempty"`
	CreatedAt      string `json:"created_at"`
}

func (c *Client) PublishExecutionRequest(ctx context.Context, req ExecutionRequest) error {
	return c.PublishExecutionRequestWithMsgID(ctx, req, req.IdempotencyKey)
}

func (c *Client) PublishExecutionRequestWithMsgID(ctx context.Context, req ExecutionRequest, msgID string) error {
	if c == nil {
		return nil
	}
	if msgID == "" {
		msgID = req.IdempotencyKey
	}
	subject := BuyRequestSubject
	if req.Side == "sell" {
		subject = SellRequestSubject
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("nats: marshal execution request: %w", err)
	}
	if _, err := c.js.Publish(ctx, subject, payload, jetstream.WithMsgID(msgID)); err != nil {
		return fmt.Errorf("nats: publish execution request: %w", err)
	}
	return nil
}

type ExecutionResultHandler func(ctx context.Context, data []byte) error

func (c *Client) ConsumeExecutionResults(ctx context.Context, handler ExecutionResultHandler) (func(), error) {
	if c == nil {
		return func() {}, nil
	}
	stream, err := c.js.Stream(ctx, ExecutionResultStreamName)
	if err != nil {
		return nil, fmt.Errorf("nats: open execution result stream: %w", err)
	}
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       ExecutionResultConsumer,
		FilterSubject: ResultSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    5,
	})
	if err != nil {
		return nil, fmt.Errorf("nats: create execution result consumer: %w", err)
	}
	consumeContext, err := consumer.Consume(func(msg jetstream.Msg) {
		handleCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		if err := handler(handleCtx, msg.Data()); err != nil {
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	if err != nil {
		return nil, fmt.Errorf("nats: consume execution results: %w", err)
	}
	return consumeContext.Stop, nil
}
