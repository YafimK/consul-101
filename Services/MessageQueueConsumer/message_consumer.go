package main

import (
	"flag"
	"fmt"
	"github.com/YafimK/consul-101/Common"
	"github.com/streadway/amqp"
	"log"
	"net/http"
)

var consulDefaultAddress = "consul:8500"

const serviceName = "message_consumer"
const defaultServicePort = 3001

func failOnError(err error, msg string) {
	if err != nil {
		// Exit the program.
		log.Fatalf(fmt.Sprintf("%s: %s", msg, err))
	}
}

func serviceApi(rabbitHostAddress string) {
	conn, err := amqp.Dial(rabbitHostAddress)
	failOnError(err, "Error connecting to the broker")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exchangeName := "user_updates"
	bindingKey := "user.profile.*"

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Error creating the exchange")

	q, err := ch.QueueDeclare(
		"",    // name - empty means a random, unique name will be assigned
		true,  // durable
		false, // delete when the last consumer unsubscribes
		false,
		false,
		nil,
	)
	failOnError(err, "Error creating the queue")

	err = ch.QueueBind(
		q.Name,       // queue name
		bindingKey,   // binding key
		exchangeName, // exchange
		false,
		nil,
	)
	failOnError(err, "Error binding the queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer id - empty means a random, unique id will be assigned
		false,  // auto acknowledgement of message delivery
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register as a consumer")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received message: %s", d.Body)
			d.Ack(false)
		}
	}()
	fmt.Println("Service listening for events...")
	<-forever
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `%v service is good`, serviceName)
}

func main() {
	amqpHost := flag.String("amqp", "amqp://guest:guest@0.0.0.0:5672/", "enter amqp server")
	hostname := flag.String("host", "0.0.0.0", "service bind hostname")
	port := flag.Int("port", defaultServicePort, "service bind port")
	flag.Parse()

	go serviceApi(*amqpHost)
	err := Common.RegisterServiceWithConsul(serviceName, serviceName, *hostname, *port, consulDefaultAddress)
	failOnError(err, "failed registering with consul")

	http.HandleFunc("/healthcheck", healthCheck)
	err = http.ListenAndServe(fmt.Sprintf("%v:%v", *hostname, *port), nil)
	failOnError(err, "error during listening for incoming connections:")
}
