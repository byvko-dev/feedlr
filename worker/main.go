package main

import "github.com/byvko-dev/feedlr/shared/tasks"

func main() {
	defer Disconnect() // Close the connection to RabbitMQ

	cancel := make(chan struct{})
	Subscribe(func(task tasks.Task) {
		// Post the webhook to Discord as Embed
		//

	}, cancel)
}
