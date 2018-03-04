package main

import (
	"fmt"
	"bufio"
	"os"
	"cb-blog-timed-tasks/cmd/shared"
)

// Runs multiple consumers in parallel
func main() {
	couchbaseController := shared.SetupCouchbaseController()
	defer couchbaseController.Close()

	shouldExit := false

	consumersCount := shared.ConfigConsumersCount
	exitChan := make(chan bool)

	// Gather all consumers statistics
	consumersStats := make([]*shared.ConsumerStats, consumersCount)
	for i := 0; i < consumersCount; i++ {
		consumersStats[i] = new(shared.ConsumerStats)
		go shared.ConsumerLoop(couchbaseController, consumersCount, &shouldExit, exitChan, consumersStats[i])
	}

	go func() {
		fmt.Println("### Timed tasks - Consumer cluster ###")
		fmt.Println("Press 'Enter' to terminate the program...")
		fmt.Println("-----------------------------------------")

		bufio.NewReader(os.Stdin).ReadBytes('\n')
		shouldExit = true
	}()

	// Waits for exit from all channels
	for i := 0; i < consumersCount; i++ {
		<-exitChan
	}

	shared.PrintStats(consumersStats)

	fmt.Println("Exiting...")
}
