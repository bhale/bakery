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

	aggregators := []string{"192.168.1.101:9000", "192.168.1.101:9001", "192.168.1.101:9002", "192.168.1.101:9003", "192.168.1.101:9004", "192.168.1.101:9005", "192.168.1.101:9006", "192.168.1.101:9007", "192.168.1.101:9008", "192.168.1.101:9009"}

	rabbitHost := "amqp://guest:guest@10.100.49.105:5672/"

	conn, err := amqp.Dial(rabbitHost)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	for _, aggregator := range aggregators {

		body := getSensors(aggregator)
		log.Println("Starting on " + aggregator)

		err = ch.Publish(
			"",                // exchange
			"ingestion_queue", // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(body),
			})
		failOnError(err, "Failed to publish a message")

		//log.Printf(" [x] Sent %s", body)
	}
}
