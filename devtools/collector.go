// Package devtools provides delta collectors for DevTools.
//
// This file contains all delta collectors that track changes incrementally.
package devtools

import (
	"sync"
	"sync/atomic"
	"time"
)

// LayoutCollector collects incremental layout changes.
type LayoutCollector struct {
	mu           sync.RWMutex
	enabled      uint32
	lastVersion  map[NodeID]uint32
	nodeRegistry map[NodeID]*LayoutNodeAdapter
	deltaCh      chan *LayoutDelta
	currentFrame int

	// P0-3: 新增清理机制
	lastCleanupTime time.Time
	cleanupInterval time.Duration
	nodeLastSeen    map[NodeID]time.Time
}

// NewLayoutCollector creates a new layout delta collector.
func NewLayoutCollector(deltaCh chan *LayoutDelta) *LayoutCollector {
	return &LayoutCollector{
		enabled:         0,
		lastVersion:     make(map[NodeID]uint32),
		nodeRegistry:    make(map[NodeID]*LayoutNodeAdapter),
		nodeLastSeen:    make(map[NodeID]time.Time),
		deltaCh:         deltaCh,
		currentFrame:    0,
		cleanupInterval: 30 * time.Second, // 每30秒清理一次
		lastCleanupTime: time.Now(),
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
func (lc *LayoutCollector) Collect(result interface{}) {
	if !lc.IsEnabled() || result == nil {
		return
	}

	// P0-3: 定期清理过期节点
	lc.cleanup()

	// 使用适配器处理结果
	adapter := AdaptLayoutResult(result)

	delta := &LayoutDelta{
		FrameID: FrameID(lc.currentFrame),
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	now := time.Now()
	currentNodes := make(map[NodeID]bool)

	for _, box := range adapter.Boxes() {
		if box.Node == nil {
			continue
		}

		nodeID := NodeID(box.NodeID)
		currentNodes[nodeID] = true

		// P0-3: 更新访问时间
		lc.nodeLastSeen[nodeID] = now

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

	// 检测删除的节点，同时清理
	for nodeID := range lc.lastVersion {
		if !currentNodes[nodeID] {
			delta.Removed = append(delta.Removed, nodeID)
			delete(lc.lastVersion, nodeID)
			delete(lc.nodeRegistry, nodeID)
			delete(lc.nodeLastSeen, nodeID)
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

// cleanup 清理过期的节点记录 (P0-3)
func (lc *LayoutCollector) cleanup() {
	now := time.Now()
	if now.Sub(lc.lastCleanupTime) < lc.cleanupInterval {
		return
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// 清理超过 5 分钟未访问的节点
	staleTime := now.Add(-5 * time.Minute)

	for nodeID, lastSeen := range lc.nodeLastSeen {
		if lastSeen.Before(staleTime) {
			delete(lc.lastVersion, nodeID)
			delete(lc.nodeRegistry, nodeID)
			delete(lc.nodeLastSeen, nodeID)
		}
	}

	lc.lastCleanupTime = now
}

func (lc *LayoutCollector) buildNodeDelta(node *LayoutNodeAdapter, nodeID NodeID) *NodeDelta {
	delta := &NodeDelta{
		ID:   nodeID,
		Mask: 0,
	}

	oldRect, hasOldRect := lc.getNodeRect(nodeID)
	newRect := Rect{
		X:      node.GetX(),
		Y:      node.GetY(),
		Width:  node.GetMeasuredWidth(),
		Height: node.GetMeasuredHeight(),
	}

	if !hasOldRect || oldRect != newRect {
		delta.Rect = &newRect
		delta.Mask |= ChangeRect
	}

	oldZIndex := lc.getNodeZIndex(nodeID)
	newZIndex := node.GetStyle().GetZIndex()
	if newZIndex != oldZIndex {
		delta.ZIndex = &newZIndex
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
			X:      node.GetX(),
			Y:      node.GetY(),
			Width:  node.GetMeasuredWidth(),
			Height: node.GetMeasuredHeight(),
		}, true
	}
	return Rect{}, false
}

func (lc *LayoutCollector) getNodeZIndex(id NodeID) int {
	if node, ok := lc.nodeRegistry[id]; ok {
		return node.GetStyle().GetZIndex()
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
