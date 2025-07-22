package rabbitmq

import (
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

func Init() error {
	var err error

	// if os.Getenv("ENV") != "production" {
	// 	err = godotenv.Load()
	// 	if err != nil {
	// 		log.Fatal("Error loading .env file")
	// 	}
	// }

	url := os.Getenv("RABBITMQ_URL") // amqps://username:password@host/vhost
	conn, err = amqp.Dial(url)
	if err != nil {
		return err
	}

	channel, err = conn.Channel()
	if err != nil {
		return err
	}

	// Ensure queue exists
	_, err = channel.QueueDeclare(
		"TestQueue", // queue name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	return err
}

func Publish(message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = channel.Publish(
		"",          // exchange
		"TestQueue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Failed to publish message: %v\n", err)
	}
	return err
}

func Close() {
	if channel != nil {
		channel.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
