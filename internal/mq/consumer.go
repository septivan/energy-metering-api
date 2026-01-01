package mq

// RabbitMQ Consumer - DISABLED FOR NOW
// Uncomment code below when RabbitMQ is needed

/*
import (
	"context"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"energy-metering-api/internal/config"
	"energy-metering-api/internal/websocket"
)

type Consumer struct {
	amqpURL       string
	hub           *websocket.Hub
	exchange      string
	queue         string
	routingKey    string
	retryInterval time.Duration
	logger        *zap.Logger
}

func NewConsumer(cfg *config.Config, hub *websocket.Hub, logger *zap.Logger) *Consumer {
	return &Consumer{
		amqpURL:       cfg.RabbitMQURL,
		hub:           hub,
		exchange:      cfg.RabbitMQExchange,
		queue:         cfg.RabbitMQQueue,
		routingKey:    cfg.RabbitMQRoutingKey,
		retryInterval: cfg.RabbitMQRetryInterval,
		logger:        logger,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("starting RabbitMQ consumer",
		zap.String("amqp_url", c.amqpURL),
		zap.String("exchange", c.exchange),
		zap.String("queue", c.queue),
		zap.String("routing_key", c.routingKey),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer stopped due to context cancellation")
			return
		default:
		}

		conn, err := amqp.Dial(c.amqpURL)
		if err != nil {
			c.logger.Error("failed to dial RabbitMQ, retrying...",
				zap.Error(err),
				zap.Duration("retry_in", c.retryInterval),
			)
			time.Sleep(c.retryInterval)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			c.logger.Error("failed to open channel, retrying...", zap.Error(err))
			conn.Close()
			time.Sleep(c.retryInterval)
			continue
		}

		exch := c.exchange
		if err := ch.ExchangeDeclare(exch, "topic", true, false, false, false, nil); err != nil {
			c.logger.Error("failed to declare exchange, retrying...",
				zap.Error(err),
				zap.String("exchange", exch),
			)
			ch.Close()
			conn.Close()
			time.Sleep(c.retryInterval)
			continue
		}

		q, err := ch.QueueDeclare(c.queue, true, false, false, false, nil)
		if err != nil {
			c.logger.Error("failed to declare queue, retrying...",
				zap.Error(err),
				zap.String("queue", c.queue),
			)
			ch.Close()
			conn.Close()
			time.Sleep(c.retryInterval)
			continue
		}

		if err := ch.QueueBind(q.Name, c.routingKey, exch, false, nil); err != nil {
			c.logger.Error("failed to bind queue, retrying...",
				zap.Error(err),
				zap.String("queue", q.Name),
				zap.String("routing_key", c.routingKey),
				zap.String("exchange", exch),
			)
			ch.Close()
			conn.Close()
			time.Sleep(c.retryInterval)
			continue
		}

		c.logger.Info("RabbitMQ consumer connected successfully",
			zap.String("queue", q.Name),
			zap.String("exchange", exch),
			zap.String("routing_key", c.routingKey),
		)

		msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
		if err != nil {
			c.logger.Error("failed to start consuming, retrying...", zap.Error(err))
			ch.Close()
			conn.Close()
			time.Sleep(c.retryInterval)
			continue
		}

		done := make(chan bool)
		go func() {
			messageCount := 0
			for d := range msgs {
				messageCount++
				c.logger.Debug("received message from RabbitMQ",
					zap.Int("message_count", messageCount),
					zap.String("routing_key", d.RoutingKey),
					zap.Int("body_size", len(d.Body)),
				)
				c.hub.Broadcast(d.Body)
			}
			c.logger.Warn("message channel closed")
			done <- true
		}()

		select {
		case <-ctx.Done():
			c.logger.Info("shutting down consumer...")
			ch.Close()
			conn.Close()
			return
		case <-done:
			c.logger.Warn("connection lost, reconnecting...")
			ch.Close()
			conn.Close()
			time.Sleep(c.retryInterval)
			continue
		}
	}
}
*/
