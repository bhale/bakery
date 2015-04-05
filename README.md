## Bakery

Includes a set of microservices to expose smart sensor networks, 
or dumb sensors with an embedded gateway device to collectors that feed the
data to Cassandra (and Apache Spark?)

The Bakery is managed by head pasty chef Kubernetes.

## Running the Bakery

#### Creating a CoreOS / Kubernetes Cluster
https://github.com/pires/kubernetes-vagrant-coreos-cluster.git

#### Creating a Kubernetes controlled Cassandra cluster
https://github.com/GoogleCloudPlatform/kubernetes/tree/master/examples/cassandra

You will also need to generate containers from the projects in this repository,
and start a RabbitMQ container.