package main

import (
	"flag"
	"fmt"
	"github.com/YafimK/consul-101/Common"
	"github.com/streadway/amqp"
	"log"
	"net/http"
)

var consulDefaultAddress = "127.0.0.1:8500"

const serviceName = "messageConsumer"
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

	exchangeName := "log_messages"
	bindingKey := "log.*"

	msgs, err := setupQueueConsumer(ch, exchangeName, bindingKey)
	failOnError(err, "Failed setting up queue consumer")

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

func setupQueueConsumer(ch *amqp.Channel, exchangeName string, bindingKey string) (<-chan amqp.Delivery, error) {
	err := ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(
		"",    // name - empty means a random, unique name will be assigned
		true,  // durable
		false, // delete when the last consumer unsubscribe
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = ch.QueueBind(
		q.Name,       // queue name
		bindingKey,   // binding key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer id - empty means a random, unique id will be assigned
		false,  // auto acknowledgement of message delivery
		false,
		false,
		false,
		nil,
	)
	return msgs, err
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `%v service is good`, serviceName)
}

func main() {
	amqpHost := flag.String("amqp", "amqp://guest:guest@0.0.0.0:5672/", "enter amqp server")
	hostname := flag.String("host", "0.0.0.0", "service bind hostname")
	port := flag.Int("port", defaultServicePort, "service bind port")
	consulAddress := flag.String("consul", consulDefaultAddress, "service bind port")
	flag.Parse()

	go serviceApi(*amqpHost)
	consulClient, err := Common.NewClient(*consulAddress)
	failOnError(err, "failed connecting to consul")
	err = consulClient.RegisterService(serviceName, serviceName, *hostname, *port)
	failOnError(err, "failed registering with consul")
	http.HandleFunc("/healthcheck", healthCheck)
	err = http.ListenAndServe(fmt.Sprintf("%v:%v", *hostname, *port), nil)
	failOnError(err, "error during listening for incoming connections:")
}
