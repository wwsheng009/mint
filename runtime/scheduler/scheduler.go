// Package scheduler provides update batching and frame-split rendering.
//
// The scheduler manages component updates with the following features:
//   - Update merging: Multiple dirty updates are batched into a single render
//   - Time slicing: Large updates are split across multiple frames
//   - Priority processing: High-priority updates are processed first
//   - Dirty queue: Efficient tracking of components needing updates
package scheduler

import (
	"sync"
	"time"

	"github.com/wwsheng009/mint/runtime/priority"
)

// DirtyNode represents a node that needs rendering.
type DirtyNode struct {
	// Node is the component/node to render.
	Node interface{}

	// Priority is the dirty priority level.
	Priority priority.DirtyLevel

	// Timestamp when the node was marked dirty.
	Timestamp time.Time

	// LayoutDirty indicates if layout needs recalculation.
	LayoutDirty bool

	// PaintDirty indicates if painting is needed.
	PaintDirty bool
}

// UpdateBatch represents a batch of updates to be processed together.
type UpdateBatch struct {
	// Nodes contains the dirty nodes in this batch.
	Nodes []*DirtyNode

	// HighPriority count of high-priority nodes.
	HighPriority int

	// NormalPriority count of normal-priority nodes.
	NormalPriority int

	// LowPriority count of low-priority nodes.
	LowPriority int

	// Timestamp when this batch was created.
	Timestamp time.Time
}

// Scheduler manages component updates with batching and time slicing.
type Scheduler struct {
	mu sync.RWMutex

	// dirtyQueue contains nodes marked dirty but not yet processed.
	dirtyQueue map[string]*DirtyNode

	// processingSet tracks nodes currently being processed.
	processingSet map[string]struct{}

	// pendingBatch accumulates updates for the next frame.
	pendingBatch *UpdateBatch

	// isBatching indicates if batching mode is active.
	isBatching bool

	// defaultTimeBudget is the max time to spend per priority level.
	defaultTimeBudget time.Duration

	// maxBatchSize is the maximum number of updates to batch.
	maxBatchSize int

	// maxBatchDuration is how long to accumulate updates before forcing a flush.
	maxBatchDuration time.Duration

	// lastFlushTime tracks when the last batch was flushed.
	lastFlushTime time.Time
}

// New creates a new scheduler with default settings.
func New() *Scheduler {
	return &Scheduler{
		dirtyQueue:         make(map[string]*DirtyNode),
		processingSet:      make(map[string]struct{}),
		defaultTimeBudget:  2 * time.Millisecond,
		maxBatchSize:       1000,
		maxBatchDuration:   16 * time.Millisecond, // ~60fps
		lastFlushTime:      time.Now(),
	}
}

// NewWithBudget creates a scheduler with a custom time budget.
func NewWithBudget(budget time.Duration) *Scheduler {
	s := New()
	s.defaultTimeBudget = budget
	return s
}

// NewWithConfig creates a scheduler with custom configuration.
func NewWithConfig(timeBudget, maxBatchDuration time.Duration, maxBatchSize int) *Scheduler {
	s := New()
	s.defaultTimeBudget = timeBudget
	s.maxBatchDuration = maxBatchDuration
	s.maxBatchSize = maxBatchSize
	return s
}

// ==============================================================================
// Dirty Queue Management
// ==============================================================================

// MarkDirty marks a node as needing an update.
//
// If batching is enabled, the node is added to the pending batch.
// Otherwise, it's added to the dirty queue for immediate processing.
func (s *Scheduler) MarkDirty(nodeID string, node interface{}, level priority.DirtyLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.markDirtyUnsafe(nodeID, node, level, true, true)
}

// MarkLayoutDirty marks a node as needing layout recalculation.
func (s *Scheduler) MarkLayoutDirty(nodeID string, node interface{}, level priority.DirtyLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.markDirtyUnsafe(nodeID, node, level, true, false)
}

// MarkPaintDirty marks a node as needing repainting.
func (s *Scheduler) MarkPaintDirty(nodeID string, node interface{}, level priority.DirtyLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.markDirtyUnsafe(nodeID, node, level, false, true)
}

// markDirtyUnsafe is the internal implementation without locking.
func (s *Scheduler) markDirtyUnsafe(nodeID string, node interface{}, level priority.DirtyLevel, layout, paint bool) {
	// Check if already being processed
	if _, processing := s.processingSet[nodeID]; processing {
		return
	}

	// Create or update the dirty node
	dirtyNode := &DirtyNode{
		Node:        node,
		Priority:    level,
		Timestamp:   time.Now(),
		LayoutDirty: layout,
		PaintDirty:  paint,
	}

	// When batching, only add to batch (not dirty queue)
	if s.isBatching {
		// Check if already in pending batch
		if s.pendingBatch != nil {
			for _, existing := range s.pendingBatch.Nodes {
				if getNodeID(existing.Node) == nodeID {
					existing.LayoutDirty = existing.LayoutDirty || layout
					existing.PaintDirty = existing.PaintDirty || paint
					if level < existing.Priority {
						existing.Priority = level
					}
					existing.Timestamp = time.Now()
					return
				}
			}
		}
		s.addToBatch(nodeID, dirtyNode)
		return
	}

	// Not batching: add to dirty queue
	if existing, ok := s.dirtyQueue[nodeID]; ok {
		existing.LayoutDirty = existing.LayoutDirty || layout
		existing.PaintDirty = existing.PaintDirty || paint
		// Upgrade priority if new level is higher
		if level < existing.Priority {
			existing.Priority = level
		}
		existing.Timestamp = time.Now()
	} else {
		s.dirtyQueue[nodeID] = dirtyNode
	}
}

// IsDirty checks if a node is currently marked dirty.
func (s *Scheduler) IsDirty(nodeID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.dirtyQueue[nodeID]
	return ok
}

// ClearDirty removes a node from the dirty queue.
func (s *Scheduler) ClearDirty(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.dirtyQueue, nodeID)
	delete(s.processingSet, nodeID)
}

// ==============================================================================
// Batch Management
// ==============================================================================

// BeginBatch starts accumulating updates for batched processing.
//
// While batching is active, MarkDirty calls add nodes to a pending batch
// instead of immediately queuing them. Call FlushBatch to process the batch.
func (s *Scheduler) BeginBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isBatching = true
	if s.pendingBatch == nil {
		s.pendingBatch = &UpdateBatch{
			Nodes:     make([]*DirtyNode, 0, s.maxBatchSize),
			Timestamp: time.Now(),
		}
	}
}

// EndBatch stops batching and optionally flushes the accumulated updates.
func (s *Scheduler) EndBatch(flush bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isBatching = false

	if flush && s.pendingBatch != nil && len(s.pendingBatch.Nodes) > 0 {
		// Move batched nodes to dirty queue
		for _, node := range s.pendingBatch.Nodes {
			nodeID := getNodeID(node.Node)
			s.dirtyQueue[nodeID] = node
		}
	}

	s.pendingBatch = nil
	s.lastFlushTime = time.Now()
}

// FlushBatch processes the accumulated batch of updates.
//
// Returns true if any updates were flushed.
func (s *Scheduler) FlushBatch() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pendingBatch == nil || len(s.pendingBatch.Nodes) == 0 {
		return false
	}

	// Move batched nodes to dirty queue
	for _, node := range s.pendingBatch.Nodes {
		nodeID := getNodeID(node.Node)
		s.dirtyQueue[nodeID] = node
	}

	s.pendingBatch = &UpdateBatch{
		Nodes:     make([]*DirtyNode, 0, s.maxBatchSize),
		Timestamp: time.Now(),
	}
	s.lastFlushTime = time.Now()

	return true
}

// ShouldFlush returns true if the batch should be flushed based on size or time.
func (s *Scheduler) ShouldFlush() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.pendingBatch == nil {
		return false
	}

	// Flush if batch is full
	if len(s.pendingBatch.Nodes) >= s.maxBatchSize {
		return true
	}

	// Flush if max duration elapsed
	if time.Since(s.pendingBatch.Timestamp) >= s.maxBatchDuration {
		return true
	}

	return false
}

// GetBatchSize returns the current batch size.
func (s *Scheduler) GetBatchSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.pendingBatch == nil {
		return 0
	}
	return len(s.pendingBatch.Nodes)
}

// addToBatch adds a dirty node to the pending batch.
func (s *Scheduler) addToBatch(nodeID string, node *DirtyNode) {
	if s.pendingBatch == nil {
		s.pendingBatch = &UpdateBatch{
			Nodes:     make([]*DirtyNode, 0, s.maxBatchSize),
			Timestamp: time.Now(),
		}
	}

	// Check if already in batch
	for _, existing := range s.pendingBatch.Nodes {
		if getNodeID(existing.Node) == nodeID {
			existing.LayoutDirty = existing.LayoutDirty || node.LayoutDirty
			existing.PaintDirty = existing.PaintDirty || node.PaintDirty
			return
		}
	}

	s.pendingBatch.Nodes = append(s.pendingBatch.Nodes, node)

	switch node.Priority {
	case priority.DirtyHigh:
		s.pendingBatch.HighPriority++
	case priority.DirtyNormal:
		s.pendingBatch.NormalPriority++
	case priority.DirtyLow:
		s.pendingBatch.LowPriority++
	}
}

// ==============================================================================
// Update Processing
// ==============================================================================

// Renderer defines how nodes are rendered.
type Renderer interface {
	Layout(node interface{})
	Paint(node interface{})
}

// ProcessOptions controls how updates are processed.
type ProcessOptions struct {
	// TimeBudget limits how long to spend processing.
	TimeBudget time.Duration

	// MaxNodes limits the number of nodes to process.
	MaxNodes int

	// PriorityLevels specifies which priorities to process.
	// If empty, all levels are processed.
	PriorityLevels []priority.DirtyLevel
}

// ProcessResult contains statistics about processed updates.
type ProcessResult struct {
	// Processed is the total number of nodes processed.
	Processed int

	// OutOfTime indicates if processing stopped due to time budget.
	OutOfTime bool

	// ByPriority counts processed nodes by priority level.
	ByPriority map[priority.DirtyLevel]int

	// Remaining is the count of dirty nodes left to process.
	Remaining int
}

// DefaultProcessOptions returns default processing options.
func DefaultProcessOptions() ProcessOptions {
	return ProcessOptions{
		TimeBudget:     0, // No limit by default
		MaxNodes:       0,
		PriorityLevels: nil,
	}
}

// ProcessNext processes the next batch of dirty updates.
//
// Nodes are processed in priority order: High → Normal → Low.
// Time budget is respected per priority level.
func (s *Scheduler) ProcessNext(renderer Renderer, opts ProcessOptions) ProcessResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Auto-flush batch if needed
	if s.isBatching && s.ShouldFlush() {
		s.flushBatchUnsafe()
	}

	result := ProcessResult{
		ByPriority: map[priority.DirtyLevel]int{
			priority.DirtyHigh:   0,
			priority.DirtyNormal: 0,
			priority.DirtyLow:    0,
		},
	}

	// Determine which levels to process
	levels := s.priorityLevels(opts)

	// Process each priority level
	for _, level := range levels {
		levelResult := s.processLevel(renderer, level, opts)
		result.Processed += levelResult.Processed
		result.ByPriority[level] = levelResult.Processed

		if levelResult.OutOfTime {
			result.OutOfTime = true
			break
		}
	}

	result.Remaining = len(s.dirtyQueue)
	return result
}

// priorityLevels returns the priority levels to process.
func (s *Scheduler) priorityLevels(opts ProcessOptions) []priority.DirtyLevel {
	if len(opts.PriorityLevels) > 0 {
		return opts.PriorityLevels
	}
	return []priority.DirtyLevel{
		priority.DirtyHigh,
		priority.DirtyNormal,
		priority.DirtyLow,
	}
}

// processLevel processes all dirty nodes at a given priority level.
func (s *Scheduler) processLevel(renderer Renderer, level priority.DirtyLevel, opts ProcessOptions) ProcessResult {
	result := ProcessResult{OutOfTime: false}

	// Set time budget
	budget := s.defaultTimeBudget
	if opts.TimeBudget > 0 {
		budget = opts.TimeBudget
	}
	start := time.Now()

	// Collect nodes at this level
	nodes := s.collectByLevel(level)

	// Process nodes
	for _, node := range nodes {
		// Check max nodes limit
		if opts.MaxNodes > 0 && result.Processed >= opts.MaxNodes {
			break
		}

		// Check time budget
		if budget > 0 && time.Since(start) > budget {
			result.OutOfTime = true
			break
		}

		// Mark as processing
		nodeID := getNodeID(node.Node)
		s.processingSet[nodeID] = struct{}{}

		// Process the node
		if node.LayoutDirty {
			renderer.Layout(node.Node)
		}
		if node.PaintDirty {
			renderer.Paint(node.Node)
		}

		// Remove from dirty queue
		delete(s.dirtyQueue, nodeID)
		delete(s.processingSet, nodeID)

		result.Processed++
	}

	return result
}

// collectByLevel collects all dirty nodes at a given priority level.
func (s *Scheduler) collectByLevel(level priority.DirtyLevel) []*DirtyNode {
	result := make([]*DirtyNode, 0, len(s.dirtyQueue))

	for _, node := range s.dirtyQueue {
		if node.Priority == level {
			result = append(result, node)
		}
	}

	return result
}

// flushBatchUnsafe flushes the pending batch without locking.
func (s *Scheduler) flushBatchUnsafe() {
	if s.pendingBatch == nil || len(s.pendingBatch.Nodes) == 0 {
		return
	}

	for _, node := range s.pendingBatch.Nodes {
		nodeID := getNodeID(node.Node)
		s.dirtyQueue[nodeID] = node
	}

	s.pendingBatch = &UpdateBatch{
		Nodes:     make([]*DirtyNode, 0, s.maxBatchSize),
		Timestamp: time.Now(),
	}
	s.lastFlushTime = time.Now()
}

// ==============================================================================
// Statistics and Configuration
// ==============================================================================

// DirtyCount returns the number of dirty nodes by priority level.
func (s *Scheduler) DirtyCount() map[priority.DirtyLevel]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := map[priority.DirtyLevel]int{
		priority.DirtyHigh:   0,
		priority.DirtyNormal: 0,
		priority.DirtyLow:    0,
	}

	for _, node := range s.dirtyQueue {
		counts[node.Priority]++
	}

	return counts
}

// TotalDirtyCount returns the total number of dirty nodes.
func (s *Scheduler) TotalDirtyCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.dirtyQueue)
}

// IsBatching returns true if batching is currently active.
func (s *Scheduler) IsBatching() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isBatching
}

// SetTimeBudget sets the default time budget per priority level.
func (s *Scheduler) SetTimeBudget(budget time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultTimeBudget = budget
}

// SetMaxBatchSize sets the maximum batch size.
func (s *Scheduler) SetMaxBatchSize(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxBatchSize = size
}

// SetMaxBatchDuration sets the maximum batch duration.
func (s *Scheduler) SetMaxBatchDuration(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxBatchDuration = duration
}

// Clear removes all dirty nodes and resets batching state.
func (s *Scheduler) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dirtyQueue = make(map[string]*DirtyNode)
	s.processingSet = make(map[string]struct{})
	s.pendingBatch = nil
	s.isBatching = false
	s.lastFlushTime = time.Now()
}

// ==============================================================================
// Helper Functions
// ==============================================================================

// getNodeID extracts a node ID from various node types.
func getNodeID(node interface{}) string {
	// Try common interface methods
	if ider, ok := node.(interface{ ID() string }); ok {
		return ider.ID()
	}

	// Fallback to string representation
	return ""
}
