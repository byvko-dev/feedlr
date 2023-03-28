package main

import (
	prisma "github.com/byvko-dev/feedlr/prisma/client"
)

func main() {
	client := prisma.NewClient()
	if err := client.Connect(); err != nil {
		panic(err)
	}
	defer client.Disconnect()
	//
}
