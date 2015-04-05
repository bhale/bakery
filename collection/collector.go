package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	//"os"
	"io/ioutil"
	"net/http"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func getSensors(aggregator string) []byte {
	//var sensors []Sensor
	resp, _ := http.Get("http://" + aggregator + "/sensors")
	body, _ := ioutil.ReadAll(resp.Body)
	/*
		json.Unmarshal(body, &sensors)

		for _, sensor := range sensors {
			log.Println(sensor)
		}
	*/
	return (body)
}

func main() {

	// Aggregators to poll

	aggregators := []string{"172.17.8.101:9005", "172.17.8.101:9004", "172.17.8.101:9003", "172.17.8.101:9002", "172.17.8.101:9001", "172.17.8.101:9000"}

	rabbitHost := "amqp://guest:guest@172.17.8.102:5672/"

	conn, err := amqp.Dial(rabbitHost)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	for _, aggregator := range aggregators {

		body := getSensors(aggregator)
		log.Println("Starting on " + aggregator)

		err = ch.Publish(
			"logs", // exchange
			"",     // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

		//log.Printf(" [x] Sent %s", body)
	}
}

