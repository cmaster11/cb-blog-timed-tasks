package main

import (
	"fmt"
	"bufio"
	"os"
	"time"
	"cb-blog-timed-tasks/internal"
	"log"
	"cb-blog-timed-tasks/cmd/shared"
)

// Producer will continuously insert tasks to execute into Couchbase
func main() {
	couchbaseController := shared.SetupCouchbaseController()
	defer couchbaseController.Close()

	shouldExit := false
	exitChan := make(chan bool)

	go producerLoop(couchbaseController, &shouldExit, exitChan)

	go func() {
		fmt.Println("### Timed tasks - Producer ###")
		fmt.Println("Press 'Enter' to terminate the program...")
		fmt.Println("-----------------------------------------")

		bufio.NewReader(os.Stdin).ReadBytes('\n')
		shouldExit = true
	}()

	// Waits for exit
	<-exitChan
}

func producerLoop(couchbaseController *internal.CouchbaseController, shouldExit *bool, exitChan chan bool) {
	counter := 0

	for !*shouldExit {
		counter++

		// We want to execute the task few seconds from now
		executeAt := time.Now().Add(shared.ConfigProducerTaskExecuteAtDelay * time.Millisecond)
		content := fmt.Sprintf("Task n. %d", counter)

		task, err := internal.NewTask(executeAt, content)
		if err != nil {
			log.Printf("Failed to generate task n. %d %v", counter, err)
		}

		err = couchbaseController.InsertTask(task)
		if err != nil {
			log.Printf("Failed to insert task %v %v", task, err)
		}

		log.Printf("Inserted task n. %d", counter)

		// Pause
		time.Sleep(shared.ConfigProducerSleepDuration * time.Millisecond)

		// Automatically stop execution when max number of tasks has been reached
		if counter >= shared.ConfigProducerMaxTasksCount {
			break
		}
	}

	// Tells program to exit
	exitChan <- true

	fmt.Println("Exiting...")
}
