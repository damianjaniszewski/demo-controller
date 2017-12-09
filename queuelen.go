package main

import (
    "log"
    "errors"
    "os"
    "github.com/streadway/amqp"
)

func getQueueLen(queueName string) (int, error) {
  connection, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer channel.Close()

  q, err := channel.QueueInspect("demo-queue")
  if err != nil {
		log.Println("inspect error", err)
    return 0, errors.New("Cannot get queue len for: " + q.Name)
	}
  return q.Messages, nil
}
