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
	Value         int
	Aggregator_id int `json:aggregator_id`
	Id            int
	Unit          string
	Name          string
}

// Get all active endpoints from Kubernetes API
func getEndpoints(pod string) []string {
	resp, _ := http.Get("http://10.1.1.2:8080/api/v1beta1/endpoints/" + pod + "?namespace=default")
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
	conn, err := amqp.Dial("amqp://guest:guest@10.100.49.105:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"ingestion_queue", // name
		true,              // durable
		false,             // delete when usused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)

	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		3,     //prefetch count
		0,     // prefetch size
		false, // global
	)

	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var count int

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			//log.Printf(" [x] %s - %i - %i", d.Body)
			//readout := string(d.Body[:])

			sensors := parseSensors(d.Body)
			for _, sensor := range sensors {

				log.Printf(" [x] %s - %s - %s - %s", sensor.Name, sensor.Id, sensor.Value, sensor.Aggregator_id)
				if err := session.Query(`INSERT INTO sensor_states (sensor_id,event_time,state) VALUES (?, ?, ?)`,
					strconv.Itoa(sensor.Id), time.Now(), strconv.Itoa(sensor.Value)).Exec(); err != nil {
					log.Fatal(err)
				}

			}
			count++
			log.Println(count)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for tasks. To exit press CTRL+C")
	<-forever
}
