// Package main demonstrates interruptible rendering with Lane scheduling.
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/scheduler"
)

func main() {
	fmt.Println("=== Interruptible Rendering Demo ===")
	fmt.Println()
	fmt.Println("This demo shows how low-priority work can be interrupted")
	fmt.Println("by high-priority user input using the Lane scheduling system.")
	fmt.Println()

	// Demo 1: Basic interruption
	demoBasicInterruption()

	// Demo 2: Large list processing
	demoLargeListProcessing()

	// Demo 3: Priority ordering
	demoPriorityOrdering()
}

func demoBasicInterruption() {
	fmt.Println("--- Demo 1: Basic Interruption ---")

	s := scheduler.NewScheduler(
		scheduler.WithOnWorkStart(func(task *scheduler.ScheduledTask) {
			fmt.Printf("[START] Lane=%s ID=%d\n", task.Lane, task.ID)
		}),
		scheduler.WithOnWorkYield(func(task *scheduler.ScheduledTask) {
			fmt.Printf("[YIELD] Task %d yielded control\n", task.ID)
		}),
		scheduler.WithOnWorkComplete(func(task *scheduler.ScheduledTask) {
			fmt.Printf("[DONE] Task %d completed\n", task.ID)
		}),
	)
	defer s.Shutdown()

	// Schedule a low-priority background task
	backgroundTask := s.ScheduleFunc(scheduler.IdleLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		fmt.Println("  Processing background work...")
		time.Sleep(50 * time.Millisecond)

		// Simulate checking for interruption
		if shouldYield() {
			fmt.Println("  Background work yielded to higher priority!")
			return false // Not complete, will resume later
		}

		return true
	})

	// Schedule a high-priority user input (after background starts)
	go func() {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("\n  [User Input Arrives!]")
		s.ScheduleFunc(scheduler.InputLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
			fmt.Println("  Processing user input immediately!")
			return true
		})
	}()

	// Run scheduler
	time.Sleep(20 * time.Millisecond)
	fmt.Println("\nScheduler state:")
	fmt.Printf("  Pending lanes: %s\n", s.GetPendingLanes())
	fmt.Printf("  Input queue: %d\n", s.GetQueueLength(scheduler.InputLane))
	fmt.Printf("  Idle queue: %d\n", s.GetQueueLength(scheduler.IdleLane))

	_ = backgroundTask
	s.Flush()
	fmt.Println()
}

func demoLargeListProcessing() {
	fmt.Println("--- Demo 2: Large List Processing ---")

	s := scheduler.NewScheduler()
	defer s.Shutdown()

	const totalItems = 10000
	processedItems := 0
	chunkSize := 1000

	fmt.Printf("Processing %d items in chunks of %d...\n", totalItems, chunkSize)

	// Process large list in chunks with yielding
	startTime := time.Now()

	for processedItems < totalItems {
		task := s.ScheduleFunc(scheduler.TransitionLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
			itemsThisChunk := 0

			for processedItems < totalItems && itemsThisChunk < chunkSize {
				processedItems++
				itemsThisChunk++
			}

			fmt.Printf("  Processed %d/%d items\n", processedItems, totalItems)

			// Check if we should yield
			if processedItems < totalItems && shouldYield() {
				return false // Yield and resume later
			}

			return processedItems >= totalItems
		})

		s.PerformWork()

		if processedItems >= totalItems {
			task.Cancel() // Clean up
			break
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Completed in %v\n", elapsed)
	fmt.Println()
}

func demoPriorityOrdering() {
	fmt.Println("--- Demo 3: Priority Ordering ---")

	s := scheduler.NewScheduler()
	defer s.Shutdown()

	executionOrder := make([]string, 0)

	// Add tasks in reverse priority order
	s.ScheduleFunc(scheduler.IdleLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		executionOrder = append(executionOrder, "IdleLane")
		return true
	})

	s.ScheduleFunc(scheduler.TransitionLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		executionOrder = append(executionOrder, "TransitionLane")
		return true
	})

	s.ScheduleFunc(scheduler.DefaultLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		executionOrder = append(executionOrder, "DefaultLane")
		return true
	})

	s.ScheduleFunc(scheduler.InputLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		executionOrder = append(executionOrder, "InputLane")
		return true
	})

	s.ScheduleFunc(scheduler.SyncLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		executionOrder = append(executionOrder, "SyncLane")
		return true
	})

	fmt.Println("Tasks scheduled: IdleLane -> TransitionLane -> DefaultLane -> InputLane -> SyncLane")
	fmt.Println()

	s.Flush()

	fmt.Println("Execution order (highest priority first):")
	for i, lane := range executionOrder {
		fmt.Printf("  %d. %s\n", i+1, lane)
	}

	fmt.Println()
	fmt.Println("Key insight: Higher priority lanes execute first,")
	fmt.Println("ensuring user input is never blocked by background work.")
	fmt.Println()
}
