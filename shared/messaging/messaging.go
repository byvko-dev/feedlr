package messaging

import (
	"context"
	"errors"
	"time"

	"github.com/byvko-dev/feedlr/shared/helpers"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
)

type client struct {
	conn *amqp.Connection
}

func NewClient() *client {
	return &client{}
}

func (c *client) channel() (*amqp.Channel, error) {
	if c.conn == nil || c.conn.IsClosed() {
		return nil, ErrConnectionClosed
	}
	return c.conn.Channel()
}

func (c *client) connect(queue string) error {
	var err error
	c.conn, err = amqp.Dial(helpers.MustGetEnv("RABBITMQ_URL"))
	if err != nil {
		return err
	}

	ch, err := c.channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) close() {
	c.conn.Close()
}

func (c *client) Publish(queue string, content ...[]byte) error {
	err := c.connect(queue)
	if err != nil {
		return err
	}
	defer c.close()

	var errors []error
	for _, body := range content {
		ch, err := c.channel()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		defer ch.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errors = append(errors, ch.PublishWithContext(ctx,
			"",    // exchange
			queue, // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			}))
	}

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) Subscribe(queue string, prefetch int, fn func(body []byte), cancel <-chan struct{}) error {
	if prefetch == 0 {
		prefetch = 1
	}

	err := c.connect(queue)
	if err != nil {
		return err
	}
	defer c.close()

	ch, err := c.channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Qos(
		prefetch, // prefetch count
		0,        // prefetch size
		false,    // global
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	// Listen for connection close
	closeNotify := c.conn.NotifyClose(make(chan *amqp.Error))

	for {
		select {
		case <-closeNotify:
			return ErrConnectionClosed
		case <-cancel:
			return nil
		case msg := <-msgs:
			fn(msg.Body)
			msg.Ack(false)
		}
	}
}
