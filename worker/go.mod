module github.com/byvko-dev/feedlr/worker

go 1.20

replace github.com/byvko-dev/feedlr/prisma => ../prisma

replace github.com/byvko-dev/feedlr/shared => ../shared

require github.com/byvko-dev/feedlr/shared v0.0.0-00010101000000-000000000000

require github.com/rabbitmq/amqp091-go v1.8.0 // indirect
