package shared

// --- Couchbase ---
const ConfigCouchbaseHostName = "127.0.0.1"
const ConfigCouchbaseBucketName = "timed_tasks"
const ConfigCouchbaseManagerUsername = "timed_tasks"
const ConfigCouchbaseManagerPassword = "timed_tasks_pwd"

// --- Producer ---

// How much will a producer wait before inserting a new task (ms)
const ConfigProducerSleepDuration = 20
// How many tasks should a producer insert into the database
const ConfigProducerMaxTasksCount = 1000
// How much in future shall a task be run (ms)
const ConfigProducerTaskExecuteAtDelay = 500

// --- Consumer ---

// How many consumers to run in parallel
const ConfigConsumersCount = 10
// How much time will the consumer take to process a task (ms, minimum base value)
const ConfigConsumerProcessingTime = 200
// How much processing time will be affected by random amount (processing_time + (rand() * processing_time * random_multiplier)
const ConfigConsumerProcessingTimeRandomMultiplier = 0.25
// Should the consumer terminate automatically when no more tasks are available?
const ConfigConsumerTerminateOnNoTasksAvailable = true
