package internal

import (
	"github.com/couchbase/gocb"
	"fmt"
	"strings"
	"log"
	"errors"
)

const cbIndexNameTimedTask = "idx_timed_task"

// Keeps care of all Couchbase-related operations
type CouchbaseController struct {
	bucket *gocb.Bucket
}

// Instantiates a new CouchbaseController object
func NewCouchbaseController(
	hostName string,
	bucketName string,
	username string,
	password string,
) (*CouchbaseController, error) {

	authString := fmt.Sprintf("%s:%s@", username, password)
	connectionString := fmt.Sprintf("couchbase://%s%s", authString, hostName)

	log.Printf("Connecting to Couchbase %s", connectionString)

	cluster, err := gocb.Connect(connectionString)
	if err != nil {
		return nil, errors.New("error connecting to couchbase cluster: " + err.Error())
	}

	// Let's use user/password auth to facilitate the whole authentication process
	authenticator := gocb.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
	cluster.Authenticate(authenticator)

	// No password because we're using a couchbase user to access the bucket
	bucket, err := cluster.OpenBucket(bucketName, "")
	if err != nil {
		return nil, errors.New("error connecting to open couchbase bucket: " + err.Error())
	}

	controller := CouchbaseController{
		bucket: bucket,
	}

	err = controller.ensureIndices()
	if err != nil {
		return nil, errors.New("error ensuring couchbase indices: " + err.Error())
	}

	return &controller, nil
}

func (controller *CouchbaseController) Close() {
	err := controller.bucket.Close()
	if err != nil {
		panic(err)
	}
}

// Queries Couchbase to get the next available timed task ids
func (controller *CouchbaseController) QueryNextTaskIds(limit int) ([]string, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	/*
	Query requirements:

	- Query NON-locked documents
	- `ExecuteAt` must be in the past

	 */

	/*
	A locked document's CAS value will equal -1. We can use this property to detect unlocked documents.
	https://developer.couchbase.com/documentation/server/5.0/sdk/concurrent-mutations-cluster.html#story-h2-6
	 */
	whereClauseNonLocked := fmt.Sprintf("META().`cas` <> -1")

	/*
	We're storing timestamps using milliseconds fashion, so we just need to check that a document's ExecuteAt field
	is in the past.
	https://developer.couchbase.com/documentation/server/current/n1ql/n1ql-language-reference/datefun.html#datefun__fn-date-now-millis
	 */
	whereClauseExecuteAt := fmt.Sprintf("`%s` <= NOW_MILLIS()", DbFieldTaskExecuteAt)

	// Unique WHERE clauses string
	whereClauses := strings.Join([]string{
		whereClauseNonLocked,
		whereClauseExecuteAt,
	}, " AND ")

	/*
	We can also suggest Couchbase to use index we generated on initialization, to improve the performance of the
	query.
	https://developer.couchbase.com/documentation/server/current/n1ql/n1ql-language-reference/hints.html
	  */
	useIndexHint := fmt.Sprintf("USE INDEX (%s USING GSI)", cbIndexNameTimedTask)

	/*
	We want to get the results ordered by oldest-first ExecuteAt field.
	 */
	orderBy := fmt.Sprintf("ORDER BY `%s` ASC", DbFieldTaskExecuteAt)
	limitString := fmt.Sprintf("LIMIT %d", limit)

	// Full query generation
	queryString := fmt.Sprintf("SELECT `%s` FROM `%s` %s WHERE %s %s %s",
		DbFieldTaskId,
		controller.bucket.Name(),
		useIndexHint,
		whereClauses,
		orderBy,
		limitString,
	)

	// log.Printf("(QueryNextTask) Executing query %s", queryString)

	query := gocb.NewN1qlQuery(queryString)

	/*
	Now, the generated query can be optimized in the following ways:

	- Have the Couchbase SDK cache the query, by marking it NON-adhoc
		https://developer.couchbase.com/documentation/server/current/sdk/n1ql-query.html#toplevel__prepare-stmts
	- Wait for the generated index to be up-to-date before executing the query. This can be achieved
		by requiring a specific query consistency.
		In our case, we require the `request_plus` consistency level, which will force our index to be up-to-date
		with the latest changes in the bucket, before being able to execute the query.
		https://developer.couchbase.com/documentation/server/current/architecture/querying-data-with-n1ql.html#story-h2-2
	 */
	query.AdHoc(false)
	query.Consistency(gocb.RequestPlus)

	// Actual query execution
	rows, err := controller.bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return nil, err
	}

	var availableTaskIds []string

	result := make(map[string]string)
	for rows.Next(&result) {
		id := result[DbFieldTaskId]
		availableTaskIds = append(availableTaskIds, id)
	}

	err = rows.Close()
	if err != nil {
		return nil, errors.New("failed to close rows object: " + err.Error())
	}

	return availableTaskIds, nil
}

// Generates all required indices
func (controller *CouchbaseController) ensureIndices() error {

	/*
	The index we need to create will be specific to the query generated in `QueryNextTask`.
	 */

	/*
	We want to index only the fields we want to query.
	Therefore, we will just index the ExecuteAt field, because the WHERE clause of `QueryNextTask` contains
	only this field.
	  */
	fieldExecuteAt := fmt.Sprintf("`%s` ASC", DbFieldTaskExecuteAt)

	fieldsString := strings.Join([]string{
		fieldExecuteAt,
	}, ",")

	// We want the index to contain ONLY unlocked documents, so its WHERE clause should define it
	whereClause := fmt.Sprintf("META().`cas` <> -1")

	queryString := fmt.Sprintf("CREATE INDEX `%s` ON `%s` (%s) WHERE %s USING GSI",
		cbIndexNameTimedTask,
		controller.bucket.Name(),
		fieldsString,
		whereClause,
	)

	log.Printf("(ensureIndices) Executing query %s", queryString)

	indexQuery := gocb.NewN1qlQuery(queryString)

	/*
	This is a custom reimplementation of the createIndex method provided by Couchbase SDK, which lacks support
	for a custom WHERE clause in the index definition.
	Original one: https://github.com/couchbase/gocb/blob/master/bucketmgr.go#L262
	 */
	rows, err := controller.bucket.ExecuteN1qlQuery(indexQuery, nil)
	if err != nil {
		if strings.Contains(err.Error(), "already exist") {
			// Index already exists, so let's ignore this error
			// Trying to close the rows would throw the same error again, so return here
			return nil
		} else {
			return err
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}

	log.Printf("Index created successfully")

	return nil
}

// --- Operations on tasks ----

func (controller *CouchbaseController) InsertTask(task *Task) error {
	if task.Id == "" {
		return errors.New("task to insert needs to have an id")
	}

	_, err := controller.bucket.Insert(task.Id, task, 0)
	if err != nil {
		return err
	}

	return err
}

func (controller *CouchbaseController) GetAndLockTask(taskId string) (*Task, gocb.Cas, error) {
	if taskId == "" {
		return nil, 0, errors.New("task to lock needs to have an id")
	}

	/*
	Let's lock the task, and use the maximum available time to process it.
	By definition, using zero values for lock time will set the maximum available (currently 15 seconds).
	https://developer.couchbase.com/documentation/server/current/sdk/concurrent-mutations-cluster.html#story-h2-6
	 */
	task := new(Task)
	cas, err := controller.bucket.GetAndLock(taskId, 0, &task)
	if err != nil {
		return nil, 0, err
	}

	return task, cas, nil

	/*
	NOTE: it is possible to verify the unreliability of the system by NOT locking a document. There will be duplicates!
	Use the following code, instead of the locking one: cas, err := controller.bucket.Get(taskId, &task)
	 */
}

func (controller *CouchbaseController) RemoveTask(taskId string, cas gocb.Cas) error {
	if taskId == "" {
		return errors.New("task to remove needs to have an id")
	}

	_, err := controller.bucket.Remove(taskId, cas)
	if err != nil {
		return err
	}

	return nil
}
