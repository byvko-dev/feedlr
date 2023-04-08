package messaging

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/byvko-dev/feedlr/shared/helpers"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
)

type client struct {
	ch   *amqp.Channel
	conn *amqp.Connection
}

func NewClient() *client {
	return &client{}
}

func (c *client) connect(queue string) error {
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

	_, err = c.ch.QueueDeclare(
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

	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(content))
	for _, body := range content {
		wg.Add(1)
		go func(body []byte) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			errors <- ch.PublishWithContext(ctx,
				"",    // exchange
				queue, // routing key
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
		}(body)
	}
	wg.Done()
	close(errors)

	for err := range errors {
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

	closeNotify := c.conn.NotifyClose(make(chan *amqp.Error))

	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

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
