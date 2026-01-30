// Package devtools provides delta collectors for DevTools.
//
// This file contains all delta collectors that track changes incrementally.
package devtools

import (
	"sync"
	"sync/atomic"

	"github.com/wwsheng009/mint/runtime"
)

// LayoutCollector collects incremental layout changes.
type LayoutCollector struct {
	mu           sync.RWMutex
	enabled      uint32
	lastVersion  map[NodeID]uint32
	nodeRegistry map[NodeID]*runtime.LayoutNode
	deltaCh      chan *LayoutDelta
	currentFrame int
}

// NewLayoutCollector creates a new layout delta collector.
func NewLayoutCollector(deltaCh chan *LayoutDelta) *LayoutCollector {
	return &LayoutCollector{
		enabled:      0,
		lastVersion:  make(map[NodeID]uint32),
		nodeRegistry: make(map[NodeID]*runtime.LayoutNode),
		deltaCh:      deltaCh,
		currentFrame: 0,
	}
}

// Enable enables the collector.
func (lc *LayoutCollector) Enable() {
	atomic.StoreUint32(&lc.enabled, 1)
}

// Disable disables the collector.
func (lc *LayoutCollector) Disable() {
	atomic.StoreUint32(&lc.enabled, 0)
}

// IsEnabled returns true if the collector is enabled.
func (lc *LayoutCollector) IsEnabled() bool {
	return atomic.LoadUint32(&lc.enabled) != 0
}

// Collect collects incremental layout data.
func (lc *LayoutCollector) Collect(result *runtime.LayoutResult) {
	if !lc.IsEnabled() || result == nil {
		return
	}

	delta := &LayoutDelta{
		FrameID: FrameID(lc.currentFrame),
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	currentNodes := make(map[NodeID]bool)

	for _, box := range result.Boxes {
		if box.Node == nil {
			continue
		}

		nodeID := NodeID(box.NodeID)
		currentNodes[nodeID] = true

		lastVersion, exists := lc.lastVersion[nodeID]
		currentVersion := box.Node.GetLayoutVersion()

		if !exists {
			delta.Added = append(delta.Added, nodeID)
			lc.lastVersion[nodeID] = currentVersion
			lc.nodeRegistry[nodeID] = box.Node
			continue
		}

		if lastVersion != currentVersion {
			nodeDelta := lc.buildNodeDelta(box.Node, nodeID)
			if nodeDelta != nil {
				delta.Changed = append(delta.Changed, *nodeDelta)
			}
			lc.lastVersion[nodeID] = currentVersion
		}
	}

	for nodeID := range lc.lastVersion {
		if !currentNodes[nodeID] {
			delta.Removed = append(delta.Removed, nodeID)
			delete(lc.lastVersion, nodeID)
			delete(lc.nodeRegistry, nodeID)
		}
	}

	if len(delta.Added) > 0 || len(delta.Removed) > 0 || len(delta.Changed) > 0 {
		select {
		case lc.deltaCh <- delta:
		default:
		}
	}

	lc.currentFrame++
}

func (lc *LayoutCollector) buildNodeDelta(node *runtime.LayoutNode, nodeID NodeID) *NodeDelta {
	delta := &NodeDelta{
		ID:   nodeID,
		Mask: 0,
	}

	oldRect, hasOldRect := lc.getNodeRect(nodeID)
	newRect := Rect{
		X:      node.X,
		Y:      node.Y,
		Width:  node.MeasuredWidth,
		Height: node.MeasuredHeight,
	}

	if !hasOldRect || oldRect != newRect {
		delta.Rect = &newRect
		delta.Mask |= ChangeRect
	}

	oldZIndex := lc.getNodeZIndex(nodeID)
	if node.Style.ZIndex != oldZIndex {
		delta.ZIndex = &node.Style.ZIndex
		delta.Mask |= ChangeZ
	}

	if delta.Mask == 0 {
		return nil
	}

	return delta
}

func (lc *LayoutCollector) getNodeRect(id NodeID) (Rect, bool) {
	if node, ok := lc.nodeRegistry[id]; ok {
		return Rect{
			X:      node.X,
			Y:      node.Y,
			Width:  node.MeasuredWidth,
			Height: node.MeasuredHeight,
		}, true
	}
	return Rect{}, false
}

func (lc *LayoutCollector) getNodeZIndex(id NodeID) int {
	if node, ok := lc.nodeRegistry[id]; ok {
		return node.Style.ZIndex
	}
	return 0
}

// EventCollector collects events that occur during a frame.
type EventCollector struct {
	mu            sync.RWMutex
	enabled       uint32
	deltaCh       chan *EventDelta
	currentFrame  int
	currentEvents []EventEntry
}

// NewEventCollector creates a new event delta collector.
func NewEventCollector(deltaCh chan *EventDelta) *EventCollector {
	return &EventCollector{
		enabled:       0,
		deltaCh:       deltaCh,
		currentFrame:  0,
		currentEvents: make([]EventEntry, 0, 16),
	}
}

// Enable enables the collector.
func (ec *EventCollector) Enable() {
	atomic.StoreUint32(&ec.enabled, 1)
}

// Disable disables the collector.
func (ec *EventCollector) Disable() {
	atomic.StoreUint32(&ec.enabled, 0)
}

// IsEnabled returns true if the collector is enabled.
func (ec *EventCollector) IsEnabled() bool {
	return atomic.LoadUint32(&ec.enabled) != 0
}

// RecordEvent records a single event.
func (ec *EventCollector) RecordEvent(eventType, targetID, phase string, data map[string]interface{}) {
	if !ec.IsEnabled() {
		return
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.currentEvents = append(ec.currentEvents, EventEntry{
		Type:   eventType,
		Target: NodeID(targetID),
		Phase:  phase,
		Data:   data,
	})
}

// Flush flushes the current frame's events.
func (ec *EventCollector) Flush() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if len(ec.currentEvents) == 0 {
		ec.currentFrame++
		return
	}

	events := make([]EventEntry, len(ec.currentEvents))
	copy(events, ec.currentEvents)

	delta := &EventDelta{
		FrameID: FrameID(ec.currentFrame),
		Events:  events,
	}

	select {
	case ec.deltaCh <- delta:
	default:
	}

	ec.currentEvents = ec.currentEvents[:0]
	ec.currentFrame++
}
