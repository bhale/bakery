package main

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1beta1"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"io/ioutil"
	//"log"
	"encoding/json"
	//"fmt"
	"net/http"
	"time"
)

// Get all active endpoints from Kubernetes API
func getEndpoints(pod string) []string {
	resp, _ := http.Get("http://172.17.8.101:8080/api/v1beta1/endpoints/" + pod + "?namespace=default")
	body, _ := ioutil.ReadAll(resp.Body)
	var endpoints v1beta1.Endpoints
	json.Unmarshal(body, &endpoints)
	return (endpoints.Endpoints)
}

var session *gocql.Session

func main() {
	// connect to the cluster
	endpoints := getEndpoints("cassandra")
	cluster := gocql.NewCluster(endpoints...)
	cluster.Keyspace = "bakery"
	cluster.Consistency = gocql.Quorum
	session, _ = cluster.CreateSession()
	defer session.Close()

	r := mux.NewRouter()
	r.HandleFunc("/sensors", SensorsHandler)
	http.ListenAndServe(":9090", r)
}


type Sensor struct {
	SensorId  string    `json:"sensor_id"`
	State     string    `json:"state"`
	EventTime time.Time `json:"event_time"`
}

type Sensors struct {
	Sensors []Sensor `json:"sensors"`
}

func SensorHandler(rw http.ResponseWriter, r *http.Request) {

	sensor := Sensor{}

	q := session.Query(`SELECT sensor_id, state, event_time FROM sensor_states LIMIT 1`)
	q.Scan(&sensor.SensorId, &sensor.State, &sensor.EventTime)

	js, _ := json.Marshal(sensor)
	rw.Write(js)
}

func SensorsHandler(rw http.ResponseWriter, r *http.Request) {

	sensors := make([]Sensor, 0)

	q := session.Query(`SELECT sensor_id, state, event_time FROM sensor_states LIMIT 100`).Iter()
	for {
		sensor := Sensor{}
		if q.Scan(&sensor.SensorId, &sensor.State, &sensor.EventTime) {
			sensors = append(sensors, sensor)
		} else {
			break 
		}
	}
	var sensorsList Sensors
	sensorsList.Sensors = sensors
	js, _ := json.Marshal(sensorsList)
	rw.Write(js)
}

