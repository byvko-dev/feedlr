package messaging

import (
	"context"
	"time"

	"github.com/byvko-dev/feedlr/shared/helpers"
	amqp "github.com/rabbitmq/amqp091-go"
)

type client struct {
	ch   *amqp.Channel
	conn *amqp.Connection
}

var clientCache *client

func GetClient() *client {
	if clientCache == nil {
		clientCache = &client{}
	}
	return clientCache
}

func (c *client) Connect(queues ...string) error {
	var err error
	c.conn, err = amqp.Dial(helpers.MustGetEnv("RABBITMQ_URL"))
	if err != nil {
		return err
	}

	c.ch, err = c.conn.Channel()
	if err != nil {
		return err
	}
	defer c.ch.Close()

	// Declare queues
	for _, queue := range queues {
		_, err := c.ch.QueueDeclare(
			queue, // name
			false, // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) Close() {
	c.conn.Close()
}

// Ensures that the channel is valid
func (c *client) channel() (*amqp.Channel, error) {
	if c.ch == nil || c.ch.IsClosed() {
		var err error
		c.ch, err = c.conn.Channel()
		if err != nil {
			return nil, err
		}
	}
	return c.ch, nil
}

func (c *client) Publish(queue string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := c.channel()
	if err != nil {
		return err
	}
	return ch.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (c *client) Subscribe(queue string, fn func(body []byte), cancel <-chan struct{}) error {
	ch, err := c.channel()
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-cancel:
			return nil
		case msg := <-msgs:
			fn(msg.Body)
		}
	}
}
