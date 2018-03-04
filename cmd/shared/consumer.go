package shared

import (
	"cb-blog-timed-tasks/internal"
	"time"
	"log"
	"github.com/couchbase/gocb"
	"math/rand"
)

type ConsumerStats struct {
	totalFoundTaskIds            int
	totalFoundAlreadyLockedTasks int
	totalProcessedTasks          int
	couldNotLockAnyTaskTimes     int
	totalCyclesWithFoundTasks    int

	processedTaskIds []string
}

// Waits for a short amount of time, to not cause a useless querying overhead
func waitShort() {
	time.Sleep(100 * time.Millisecond)
}

// Actual task processing
func processTask(task *internal.Task) {
	// Simulate working time with random delay
	delay := time.Duration(ConfigConsumerProcessingTime + rand.Float32()*(ConfigConsumerProcessingTime*ConfigConsumerProcessingTimeRandomMultiplier))
	time.Sleep(delay * time.Millisecond)
	log.Printf("Processed task: %s", task.Content)
}

// Queries Couchbase to get available timed tasks, and processes them
func ConsumerLoop(couchbaseController *internal.CouchbaseController, consumersCount int, shouldExit *bool, exitChan chan bool, stats *ConsumerStats) {

	for !*shouldExit {
		/*
		The process is:
		1. Fetch a certain amount of next task ids equal to the number of currently running workers.
		2. Get the first unlocked task of the group and process it
		 */
		taskIds, err := couchbaseController.QueryNextTaskIds(consumersCount)
		if err != nil {
			log.Printf("Error fetching next task id %v", err)
			waitShort()
			continue
		}

		if len(taskIds) == 0 {
			log.Printf("No tasks found")

			if ConfigConsumerTerminateOnNoTasksAvailable {
				break
			}

			waitShort()
			continue
		}

		stats.totalCyclesWithFoundTasks++
		stats.totalFoundTaskIds += len(taskIds)

		var taskId string
		var task *internal.Task
		var lockedCAS gocb.Cas

		lockedCount := 0
		for _, taskId = range taskIds {
			// Lock and get the task, so that only this consumer will process it
			task, lockedCAS, err = couchbaseController.GetAndLockTask(taskId)
			if err != nil {
				lockedCount ++
				stats.totalFoundAlreadyLockedTasks++
				log.Printf("Error getting and locking task %s %v", taskId, err)
				continue
			}

			log.Printf("Successfully locked task %s (found %d locked tasks before)", taskId, lockedCount)
			break
		}

		if task == nil {
			log.Printf("Could not lock any task")
			stats.couldNotLockAnyTaskTimes++
			continue
		}

		processTask(task)

		stats.processedTaskIds = append(stats.processedTaskIds, taskId)
		stats.totalProcessedTasks++

		/*
		Remove the task from Couchbase.
		The task will be currently locked, which means we need to provide the current CAS value,
		so that the producer is authorized to remove it.
		  */
		err = couchbaseController.RemoveTask(taskId, lockedCAS)
		if err != nil {
			log.Printf("Error removing task %s %v", taskId, err)
			continue
		}

		log.Printf("Removed task %s", taskId)
	}

	// Tells program to exit
	exitChan <- true
}
