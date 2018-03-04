package internal

import (
	"time"
	"github.com/satori/go.uuid"
	"errors"
)

const DbFieldTaskId = "Id"
const DbFieldTaskExecuteAt = "ExecuteAt"

type Task struct {
	Id string

	// The desired task execution time
	ExecuteAt int64

	// Task-specific content
	Content string
}

// Creates a new task to execute
func NewTask(executeAt time.Time, content string) (*Task, error) {
	if executeAt.IsZero() {
		return nil, errors.New("executeAt must not be a zero time")
	}

	taskUUID, err := uuid.NewV1()
	if err != nil {
		return nil, err
	}

	// Convert time.Time to int64 milliseconds
	executeAtMillis := executeAt.UnixNano() / int64(time.Millisecond)

	task := Task{
		Id:        taskUUID.String(),
		ExecuteAt: executeAtMillis,
		Content:   content,
	}

	return &task, nil
}
