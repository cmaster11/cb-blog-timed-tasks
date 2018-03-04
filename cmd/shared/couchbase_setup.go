package shared

import (
	"log"
	"cb-blog-timed-tasks/internal"
)

func SetupCouchbaseController() *internal.CouchbaseController {
	controller, err := internal.NewCouchbaseController(
		ConfigCouchbaseHostName,
		ConfigCouchbaseBucketName,
		ConfigCouchbaseManagerUsername,
		ConfigCouchbaseManagerPassword,
	)

	if err != nil {
		log.Panicf("failed to initialize couchbase controller: %v", err)
	}

	return controller
}
