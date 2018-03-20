# Timed tasks with Go and Couchbase

**NOTE:** This code's discussion and overview can be found in Couchbase Blog's post [Timed tasks using Couchbase and Go](https://blog.couchbase.com/timed-tasks-using-couchbase-go/).

The goal of the project is to demonstrate how to use specific Couchbase features 
([pessimistic-locking](https://blog.couchbase.com/optimistic-or-pessimistic-locking-which-one-should-you-pick/), indexing, N1QL)
to implement a distributed timed tasks system.

The project consists of the following programs:

* Producer - inserts tasks into Couchbase.
* Consumer - retrieves tasks and processes them.
* Consumer cluster - runs multiple consumers in parallel. 

## Setup

### Couchbase

1. `cd` into `docker` folder.
2. Run `docker compose up` to deploy a local Couchbase server.
3. Wait ~20 seconds for complete initialization.

### Timed tasks

1. Install `govendor`: `go get -u github.com/kardianos/govendor`.
2. Install dependencies with `govendor sync`.
3. Edit `cmd/shared/config.go` producer and consumer variables as desired.

Note: a big enough amount of samples (>= 1000) is suggested, to gather correct statistics.

## Run

1. Start the producer.
2. Start the consumer cluster.
3. Wait for both programs to automatically terminate (if configured), or press `Enter` and wait for them to exit.

Note: execution could take up to 1-2 minutes, depending on configuration. 

Run with Windows PowerShell:
```
go build cmd\producer\producer.go ; .\producer.exe ; rm producer.exe
go build cmd\consumer_cluster\consumer_cluster.go ; .\consumer_cluster.exe ; rm consumer_cluster.exe
```

Run with bash:
```
go build cmd/producer/producer.go && ./producer && rm producer
go build cmd/consumer_cluster/consumer_cluster.go && ./consumer_cluster && rm consumer_cluster
```

## Statistics

When the consumer cluster execution terminates gracefully, a statistics table will be printed.

### Low concurrency

```
consumerProcessingTime = 500
consumerProcessingTimeRandomMultiplier = 0.25
```

When consumers have a low concurrency factory, there will be less failed locking attempts.

```
  totFound|  totLockedFound|  totProcessed|  couldNotLock|  loopEfficiency (%)|  lockEfficiency (%)|
      1060|             108|           106|             0|          100.000000|           49.532711|
      1050|             121|           105|             0|          100.000000|           46.460175|
      1050|             110|           105|             0|          100.000000|           48.837208|
      1060|              94|           106|             0|          100.000000|           52.999996|
      1050|             105|           105|             0|          100.000000|           50.000000|
      1050|              96|           105|             0|          100.000000|           52.238804|
      1050|             113|           105|             0|          100.000000|           48.165138|
      1050|             106|           105|             0|          100.000000|           49.763031|
      1060|             101|           106|             0|          100.000000|           51.207726|
      1060|             103|           106|             0|          100.000000|           50.717705|
```

### High concurrency

```
consumerProcessingTime = 100
consumerProcessingTimeRandomMultiplier = 1.0
```

When consumers have a high concurrency factory, there will be more failed locking attempts.

```
  totFound|  totLockedFound|  totProcessed|  couldNotLock|  loopEfficiency (%)|  lockEfficiency (%)|
      2540|            1043|           253|             1|           99.606300|           19.521605|
      2600|            1056|           259|             1|           99.615387|           19.695816|
      2540|            1067|           254|             0|          100.000000|           19.227858|
      2540|            1080|           254|             0|          100.000000|           19.040480|
      2560|             985|           255|             1|           99.609375|           20.564516|
      2530|            1044|           252|             1|           99.604744|           19.444445|
      2570|            1029|           257|             0|          100.000000|           19.984447|
      2580|            1052|           258|             0|          100.000000|           19.694656|
      2550|            1009|           255|             0|          100.000000|           20.174049|
      2510|            1069|           251|             0|          100.000000|           19.015152|
```
