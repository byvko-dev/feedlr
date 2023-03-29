package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/tasks"
	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var queue amqp.Queue
var ch *amqp.Channel

func init() {
	var err error
	conn, err = amqp.Dial(helpers.MustGetEnv("RABBITMQ_URL"))
	if err != nil {
		panic(err)
	}

	ch, err = conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err = ch.QueueDeclare(
		helpers.MustGetEnv("TASKS_QUEUE"), // name
		false,                             // durable
		false,                             // delete when unused
		false,                             // exclusive
		false,                             // no-wait
		nil,                               // arguments
	)
	if err != nil {
		panic(err)
	}
}

func Disconnect() {
	ch.Close()
	conn.Close()
}

func Publish(body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(
		ctx,
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func NewTask(task tasks.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return Publish(payload)
}
