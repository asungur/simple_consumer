package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go"
	"github.com/streadway/amqp"
)

func sendSimpleMessage() {
	api_key := os.Getenv("MAILGUN_API_KEY")
	sandbox_domain := os.Getenv("MAILGUN_DOMAIN")

	mg := mailgun.NewMailgun(sandbox_domain, api_key)
	sender := "testingsomething@example.com"
	subject := "security warning"
	body := "Your fallback api is exposed"
	recipient := "sunguralican@gmail.com"
	message := mg.NewMessage(sender, subject, body, recipient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %s Resp: %s\n", id, resp)
}

func processMessages() {
	amqpServerURL := os.Getenv("AMQP_SERVER_URL")
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}
	defer connectRabbitMQ.Close()
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()
	messages, err := channelRabbitMQ.Consume(
		"FallbackAPIQueue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(err)
	}
	log.Println("You've connected to RabbitMQ")
	log.Println("Waiting for messages")

	if len(messages) == 0 {
		log.Println("You don't have any messages in the queue")
	}
	for message := range messages {
		if string(message.Body) == "A request has been sent via fallback API" {
			sendSimpleMessage()
		}
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	gocron.Every(2).Minutes().Do(processMessages)
	<-gocron.Start()
}
