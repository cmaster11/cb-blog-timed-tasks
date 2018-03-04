package main

import (
	"fmt"
	"bufio"
	"os"
	"cb-blog-timed-tasks/cmd/shared"
)

func main() {
	couchbaseController := shared.SetupCouchbaseController()
	defer couchbaseController.Close()

	shouldExit := false
	exitChan := make(chan bool)

	stats := new(shared.ConsumerStats)
	/*
	We're still using shared.ConfigConsumersCount here, so that this program can be externally parallelized
	with tools other than Goroutines.
	  */
	go shared.ConsumerLoop(couchbaseController, shared.ConfigConsumersCount, &shouldExit, exitChan, stats)

	go func() {
		fmt.Println("### Timed tasks - Consumer ###")
		fmt.Println("Press 'Enter' to terminate the program...")
		fmt.Println("-----------------------------------------")

		bufio.NewReader(os.Stdin).ReadBytes('\n')
		shouldExit = true
	}()

	// Waits for exit
	<-exitChan

	shared.PrintStats([]*shared.ConsumerStats{stats})

	fmt.Println("Exiting...")
}
