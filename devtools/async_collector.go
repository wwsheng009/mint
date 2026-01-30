// Package devtools provides the asynchronous collector for DevTools.
//
// The AsyncCollector coordinates all delta collectors and processes
// events from the event bus in background goroutines.
package devtools

import (
	"sync"
	"time"
)

// AsyncCollector coordinates all delta collectors.
// It runs background goroutines to process debug events asynchronously.
type AsyncCollector struct {
	// Event bus
	eventBus *EventBus

	// Collectors
	layoutCollector  *LayoutCollector
	eventCollector   *EventCollector

	// Output channels
	layoutDeltaCh  chan *LayoutDelta
	eventDeltaCh   chan *EventDelta

	// Output sink (e.g., WebSocket, TUI panel)
	outputCh       chan<- *DebugMessage

	// Frame tracking
	currentFrame   FrameID
	frameStartTime time.Time

	// Control
	mu     sync.RWMutex
	running bool
	done   chan struct{}
	wg     sync.WaitGroup
}

// NewAsyncCollector creates a new async collector.
func NewAsyncCollector(outputCh chan<- *DebugMessage) *AsyncCollector {
	layoutDeltaCh := make(chan *LayoutDelta, 32)
	eventDeltaCh := make(chan *EventDelta, 32)

	return &AsyncCollector{
		eventBus:        NewEventBus(4096),
		layoutCollector:  NewLayoutCollector(layoutDeltaCh),
		eventCollector:   NewEventCollector(eventDeltaCh),
		layoutDeltaCh:   layoutDeltaCh,
		eventDeltaCh:    eventDeltaCh,
		outputCh:        outputCh,
		currentFrame:    0,
		frameStartTime:  time.Now(),
		done:            make(chan struct{}),
		running:         false,
	}
}

// Start starts the async collector and all background goroutines.
func (ac *AsyncCollector) Start() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.running {
		return
	}

	ac.running = true
	ac.eventBus.Enable()
	ac.layoutCollector.Enable()
	ac.eventCollector.Enable()

	ac.wg.Add(3)
	go ac.processLayoutDeltas()
	go ac.processEventDeltas()
	go ac.processFrameTimeline()
}

// Stop stops the async collector and all background goroutines.
func (ac *AsyncCollector) Stop() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if !ac.running {
		return
	}

	ac.running = false
	ac.layoutCollector.Disable()
	ac.eventCollector.Disable()
	ac.eventBus.Disable()

	// Signal goroutines to stop
	close(ac.done)

	// Close delta channels to unblock consumers
	close(ac.layoutDeltaCh)
	close(ac.eventDeltaCh)

	// Wait for all goroutines to exit
	ac.wg.Wait()

	// Close event bus (stops dispatch loops)
	ac.eventBus.Close()
}

// GetEventBus returns the event bus.
func (ac *AsyncCollector) GetEventBus() *EventBus {
	return ac.eventBus
}

// GetLayoutCollector returns the layout collector.
func (ac *AsyncCollector) GetLayoutCollector() *LayoutCollector {
	return ac.layoutCollector
}

// GetEventCollector returns the event collector.
func (ac *AsyncCollector) GetEventCollector() *EventCollector {
	return ac.eventCollector
}

// processLayoutDeltas processes layout deltas.
func (ac *AsyncCollector) processLayoutDeltas() {
	defer ac.wg.Done()

	for {
		select {
		case <-ac.done:
			return
		case delta, ok := <-ac.layoutDeltaCh:
			if !ok {
				// Channel closed, exit
				return
			}
			if delta != nil && ac.outputCh != nil {
				select {
				case ac.outputCh <- &DebugMessage{
					Type:    MsgLayoutDelta,
					Payload: delta,
				}:
				case <-ac.done:
					return
				}
			}
		}
	}
}

// processEventDeltas processes event deltas.
func (ac *AsyncCollector) processEventDeltas() {
	defer ac.wg.Done()

	for {
		select {
		case <-ac.done:
			return
		case delta, ok := <-ac.eventDeltaCh:
			if !ok {
				// Channel closed, exit
				return
			}
			if delta != nil && ac.outputCh != nil {
				select {
				case ac.outputCh <- &DebugMessage{
					Type:    MsgEventDelta,
					Payload: delta,
				}:
				case <-ac.done:
					return
				}
			}
		}
	}
}

// processFrameTimeline generates frame timeline messages.
func (ac *AsyncCollector) processFrameTimeline() {
	defer ac.wg.Done()

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ac.done:
			return
		case <-ticker.C:
			// Frame timeline generated at EndFrame
		}
	}
}

// BeginFrame marks the beginning of a new frame.
func (ac *AsyncCollector) BeginFrame() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.currentFrame++
	ac.frameStartTime = time.Now()
}

// EndFrame marks the end of the current frame.
func (ac *AsyncCollector) EndFrame() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.eventCollector.Flush()

	if ac.outputCh != nil {
		ac.outputCh <- &DebugMessage{
			Type: MsgFrameTimeline,
			Payload: map[string]interface{}{
				"frameID":   ac.currentFrame,
				"startTime": ac.frameStartTime,
				"endTime":   time.Now(),
				"duration":  time.Since(ac.frameStartTime).Nanoseconds(),
			},
		}
	}
}

// IsRunning returns true if the async collector is running.
func (ac *AsyncCollector) IsRunning() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.running
}

// GetCurrentFrame returns the current frame ID.
func (ac *AsyncCollector) GetCurrentFrame() FrameID {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.currentFrame
}
