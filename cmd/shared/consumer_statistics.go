package shared

import (
	"fmt"
	"text/tabwriter"
	"os"
)

// Statistics
func GetStatsRow(stats *ConsumerStats) string {
	/*
	How the system works:

	1. Get list of available tasks
	2. Get the first lockable task of the list, discard the others

	Efficiency can be:
	- How many loops were effective (did worker manage to lock a task on every loop?)
	- How many lock tries have been made before a task could be locked
	 */

	totalProcessedTasks := float32(stats.totalProcessedTasks)
	totalFoundAlreadyLockedTasks := float32(stats.totalFoundAlreadyLockedTasks)
	totalCyclesWithFoundTasks := float32(stats.totalCyclesWithFoundTasks)

	// Calculate efficiencies
	loopEfficiency := (totalProcessedTasks / totalCyclesWithFoundTasks) * 100
	lockEfficiency := totalProcessedTasks / (totalProcessedTasks + totalFoundAlreadyLockedTasks) * 100

	return fmt.Sprintf("%d\t%d\t%d\t%d\t%f\t%f\t",
		stats.totalFoundTaskIds,
		stats.totalFoundAlreadyLockedTasks,
		stats.totalProcessedTasks,
		stats.couldNotLockAnyTaskTimes,
		loopEfficiency,
		lockEfficiency,
	)
}

func GetStatsHeader() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t",
		"totFound",
		"totLockedFound",
		"totProcessed",
		"couldNotLock",
		"loopEfficiency (%)",
		"lockEfficiency (%)",
	)
}
func PrintStats(statsArray []*ConsumerStats) {
	// Calculate duplicate task ids
	processedTasksIds := make(map[string]int)
	for _, stats := range statsArray {
		for _, taskId := range stats.processedTaskIds {
			processedTasksIds[taskId] ++
		}
	}

	totalDuplicateTasksExecutions := 0
	for _, count := range processedTasksIds {
		if count > 1 {
			totalDuplicateTasksExecutions += count
		}
	}

	// Statistics description
	fmt.Print("\n\n### Stats ###\n\n")
	fmt.Println("- totFound: total task ids found")
	fmt.Println("- totLockedFound: total already locked tasks found")
	fmt.Println("- totProcessed: total processed tasks")
	fmt.Println("- couldNotLock: how many loops a consumer has not been able to lock any task available in 'found task ids' list")
	fmt.Println("- loopEfficiency: defines how many times a loop processed one task (of loops with found tasks)")
	fmt.Println("- lockEfficiency: defines how many lock tries have been made before a task got processed (the lower, the worse: more tries have been made)")
	fmt.Println()

	fmt.Println("NOTE: `couldNotLock` and `lockEfficiency` depend on the concurrency factor between consumers. " +
		"When concurrency grows, a consumer will find more locked tasks:")
	fmt.Println("- As task processing time becomes shorter, concurrency level grows.")
	fmt.Println("- As number of concurrent workers grow, concurrency level grows.")
	fmt.Println()

	fmt.Printf("Total duplicate tasks executions: %d <-- Really important! This value MUST be 0 for the system to be reliable.\n", totalDuplicateTasksExecutions)
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug|tabwriter.AlignRight)

	fmt.Fprintln(w, GetStatsHeader())

	for _, stats := range statsArray {
		fmt.Fprintln(w, GetStatsRow(stats))
	}

	w.Flush()
}
