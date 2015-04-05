package main

import (
	"encoding/json"
	"fmt"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1beta1"
	"github.com/gocql/gocql"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Sensor struct {
	Value int
	Id    int
	Unit  string
	Name  string
}

// Get all active endpoints from Kubernetes API
func getEndpoints(pod string) []string {
	resp, _ := http.Get("http://172.17.8.101:8080/api/v1beta1/endpoints/" + pod + "?namespace=default")
	body, _ := ioutil.ReadAll(resp.Body)
	var endpoints v1beta1.Endpoints
	json.Unmarshal(body, &endpoints)
	return (endpoints.Endpoints)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func parseSensors(body []byte) []Sensor {
	var sensors []Sensor
	json.Unmarshal(body, &sensors)
	return (sensors)
}

func main() {
	// connect to the cluster
	endpoints := getEndpoints("cassandra")
	cluster := gocql.NewCluster(endpoints...)
	cluster.Keyspace = "bakery"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@172.17.8.102:5672/")
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

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			//log.Printf(" [x] %s - %i - %i", d.Body)
			//readout := string(d.Body[:])

			sensors := parseSensors(d.Body)
			for _, sensor := range sensors {

				//log.Printf(" [x] %s - %i - %i", sensor.Name, sensor.Id, sensor.Value)
				if err := session.Query(`INSERT INTO sensor_states (sensor_id,event_time,state) VALUES (?, ?, ?)`,
					strconv.Itoa(sensor.Id), time.Now(), strconv.Itoa(sensor.Value)).Exec(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}