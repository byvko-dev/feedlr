package main

import (
	"encoding/json"

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

func Subscribe(fn func(tasks.Task), cancel <-chan struct{}) error {
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-cancel:
			return nil
		case msg := <-msgs:
			var task tasks.Task
			err := json.Unmarshal(msg.Body, &task)
			if err != nil {
				return err
			}

			fn(task)
		}
	}
}
