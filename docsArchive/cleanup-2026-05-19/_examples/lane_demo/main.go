// Package main demonstrates the Lane scheduling system.
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/scheduler"
)

func main() {
	fmt.Println("=== Lane Scheduling Demo ===")
	fmt.Println()

	// Demo 1: Basic Lane Operations
	demoLaneOperations()

	// Demo 2: Priority Scheduling
	demoPriorityScheduling()

	// Demo 3: Interruptible Work
	demoInterruptibleWork()

	// Demo 4: Intent Lane Integration
	demoIntentLane()
}

func demoLaneOperations() {
	fmt.Println("--- Demo 1: Lane Operations ---")

	// Lane priorities
	fmt.Printf("SyncLane: %s (priority: %d)\n", scheduler.SyncLane, scheduler.SyncLane.Priority())
	fmt.Printf("InputLane: %s (priority: %d)\n", scheduler.InputLane, scheduler.InputLane.Priority())
	fmt.Printf("TransitionLane: %s (priority: %d)\n", scheduler.TransitionLane, scheduler.TransitionLane.Priority())
	fmt.Printf("IdleLane: %s (priority: %d)\n", scheduler.IdleLane, scheduler.IdleLane.Priority())

	// Merge lanes
	lanes := scheduler.MergeLanes(scheduler.InputLane, scheduler.TransitionLane, scheduler.IdleLane)
	fmt.Printf("\nMerged lanes: %s\n", lanes)

	// Pick highest/lowest priority
	highest := scheduler.PickHighestPriorityLane(lanes)
	lowest := scheduler.PickLowestPriorityLane(lanes)
	fmt.Printf("Highest priority: %s\n", highest)
	fmt.Printf("Lowest priority: %s\n", lowest)

	// Compare priorities
	if scheduler.InputLane.IsHigherPriorityThan(scheduler.TransitionLane) {
		fmt.Println("InputLane > TransitionLane ✓")
	}

	fmt.Println()
}

func demoPriorityScheduling() {
	fmt.Println("--- Demo 2: Priority Scheduling ---")

	// Create scheduler with callbacks
	s := scheduler.NewScheduler(
		scheduler.WithOnWorkStart(func(task *scheduler.ScheduledTask) {
			fmt.Printf("[START] Lane=%-12s ID=%d\n", task.Lane, task.ID)
		}),
		scheduler.WithOnWorkComplete(func(task *scheduler.ScheduledTask) {
			fmt.Printf("[DONE]  ID=%d\n", task.ID)
		}),
	)
	defer s.Shutdown()

	// Schedule tasks in different priorities (added in reverse order)
	fmt.Println("Scheduling tasks (Idle -> Transition -> Input):")

	s.ScheduleFunc(scheduler.IdleLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		fmt.Println("  -> Processing background task (IdleLane)")
		return true
	})

	s.ScheduleFunc(scheduler.TransitionLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		fmt.Println("  -> Processing transition task (TransitionLane)")
		return true
	})

	s.ScheduleFunc(scheduler.InputLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		fmt.Println("  -> Processing user input (InputLane)")
		return true
	})

	fmt.Println("\nExecuting (highest priority first):")
	s.Flush()

	fmt.Println()
}

func demoInterruptibleWork() {
	fmt.Println("--- Demo 3: Interruptible Work ---")

	s := scheduler.NewScheduler()
	defer s.Shutdown()

	processedItems := 0
	totalItems := 1000
	chunkSize := 200

	// Simulate large list processing with yielding
	task := s.ScheduleFunc(scheduler.TransitionLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		startTime := time.Now()
		itemsProcessed := 0

		for processedItems < totalItems && itemsProcessed < chunkSize {
			processedItems++
			itemsProcessed++

			// Simulate work
			time.Sleep(100 * time.Microsecond)
		}

		fmt.Printf("  Processed %d/%d items (chunk: %d)\n",
			processedItems, totalItems, itemsProcessed)

		// Check if we should yield (simulated: always yield after chunk)
		if processedItems < totalItems {
			// Simulate yielding due to time budget
			if time.Since(startTime) > 10*time.Millisecond {
				fmt.Println("  Yielding to allow other work...")
				return false // Not complete, will be resumed
			}
		}

		return processedItems >= totalItems
	})

	fmt.Printf("Processing %d items in chunks...\n", totalItems)

	// Execute until complete
	iterations := 0
	for s.HasPendingWork() && iterations < 10 {
		iterations++
		fmt.Printf("\nIteration %d:\n", iterations)
		s.PerformWork()

		// Simulate higher priority work arriving
		if iterations == 2 {
			fmt.Println("\nHigh-priority user input arrives!")
			s.ScheduleFunc(scheduler.InputLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
				fmt.Println("  -> [INTERRUPT] Handling user input NOW!")
				return true
			})
		}
	}

	_ = task // Task reference
	fmt.Printf("\nCompleted after %d iterations\n", iterations)
	fmt.Println()
}

func demoIntentLane() {
	fmt.Println("--- Demo 4: Intent Lane Integration ---")

	// Demonstrate intent wrapping
	fmt.Println("Lane wrappers:")

	// Simulated intent types
	inputIntent := mockIntent{typeName: "FieldChange", field: "username"}
	transitionIntent := mockIntent{typeName: "Navigate", field: "/settings"}
	idleIntent := mockIntent{typeName: "Analytics", field: "page_view"}

	// Wrap with lanes
	wrappedInput := scheduler.HighPriority(inputIntent)
	wrappedTransition := scheduler.LowPriority(transitionIntent)
	wrappedIdle := scheduler.BackgroundPriority(idleIntent)

	fmt.Printf("  Input intent:   Lane=%s\n", wrappedInput.Lane)
	fmt.Printf("  Navigate intent: Lane=%s\n", wrappedTransition.Lane)
	fmt.Printf("  Analytics intent: Lane=%s\n", wrappedIdle.Lane)

	// Auto-infer lane
	fmt.Println("\nAuto-inferred lanes:")
	fmt.Printf("  FieldChange -> %s\n", scheduler.InferLane(inputIntent))
	fmt.Printf("  Navigate    -> %s\n", scheduler.InferLane(transitionIntent))
	fmt.Printf("  Analytics   -> %s\n", scheduler.InferLane(idleIntent))

	// Batch intents
	fmt.Println("\nBatch scheduling:")
	batch := scheduler.NewIntentBatch(scheduler.TransitionLane, inputIntent, transitionIntent)
	batch.Add(idleIntent)
	fmt.Printf("  Batch size: %d intents\n", len(batch.Intents))
	fmt.Printf("  Batch lane: %s\n", batch.Lane)

	fmt.Println()
}

// mockIntent implements intent.Intent for demonstration
type mockIntent struct {
	typeName string
	field    string
}

func (m mockIntent) IntentType() string {
	return m.typeName
}

func (m mockIntent) String() string {
	return fmt.Sprintf("%s{field: %s}", m.typeName, m.field)
}
