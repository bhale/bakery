## Bakery

Includes a set of microservices to expose smart sensor networks, 
or dumb sensors with an embedded gateway device to collectors that feed the
data to Cassandra (and Apache Spark?)

The Bakery is managed by head pasty chef Kubernetes.

## TODO

Create Collection containers to support gathering sensor data from XMPP, ModBus, BacNet.
Create a hardware gateway to expose legacy sensors to the network (BeagleBone?)

## Running the Bakery

#### Creating a CoreOS / Kubernetes Cluster
https://github.com/pires/kubernetes-vagrant-coreos-cluster.git

#### Creating a Kubernetes controlled Cassandra cluster
https://github.com/GoogleCloudPlatform/kubernetes/tree/master/examples/cassandra

You will also need to generate containers from the projects in this repository,
and start a RabbitMQ container.

#### SQL setup

```
CREATE KEYSPACE bakery WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '3'}  AND durable_writes = true;

CREATE TABLE bakery.sensor_states (
    sensor_id text,
    event_time timestamp,
    state text,
    PRIMARY KEY (sensor_id, event_time)
) WITH CLUSTERING ORDER BY (event_time ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = '{"keys":"ALL", "rows_per_partition":"NONE"}'
    AND comment = ''
    AND compaction = {'min_threshold': '4', 'class': 'org.apache.cassandra.db.compaction.SizeTieredCompactionStrategy', 'max_threshold': '32'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND dclocal_read_repair_chance = 0.1
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND read_repair_chance = 0.0
    AND speculative_retry = '99.0PERCENTILE';
```
