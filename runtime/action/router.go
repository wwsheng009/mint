package action

import (
	"fmt"
	"sort"

	"github.com/wwsheng009/mint/runtime"
)

// ============================================================================
// Middleware and Global Handler Interfaces
// ============================================================================

// ActionMiddleware defines middleware for action processing
type ActionMiddleware interface {
	// Name returns the middleware name
	Name() string

	// Before is called before action dispatch
	// Returns nil to intercept (stop) the action
	// Returns modified action to continue dispatch
	Before(action *Action) *Action

	// After is called after action dispatch
	After(action *Action, result *RouterResult)
}

// GlobalActionHandler handles actions without specific targets
type GlobalActionHandler interface {
	// HandleGlobalAction processes an action without target
	HandleGlobalAction(action *Action) bool

	// Priority returns handler priority (higher = earlier)
	Priority() int
}

// MiddlewareChain manages middleware execution order
type MiddlewareChain struct {
	middlewares []ActionMiddleware
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(middlewares ...ActionMiddleware) *MiddlewareChain {
	return &MiddlewareChain{middlewares: middlewares}
}

// Before calls all middleware Before methods in order
func (c *MiddlewareChain) Before(action *Action) *Action {
	for _, mw := range c.middlewares {
		if action == nil {
			return nil
		}
		action = mw.Before(action)
	}
	return action
}

// After calls all middleware After methods in reverse order
func (c *MiddlewareChain) After(action *Action, result *RouterResult) {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		c.middlewares[i].After(action, result)
	}
}

// Middlewares returns the list of middlewares
func (c *MiddlewareChain) Middlewares() []ActionMiddleware {
	return c.middlewares
}

// Add appends a middleware to the chain
func (c *MiddlewareChain) Add(middleware ActionMiddleware) {
	c.middlewares = append(c.middlewares, middleware)
}

// ============================================================================
// Router - Three-Phase Action Dispatch
// ============================================================================

// Router implements three-phase action dispatching:
//   1. Capture Phase: From root to target, handlers by priority
//   2. Target Phase: At the target component
//   3. Bubble Phase: From target to root
//
// Usage:
//   router := NewRouter(rootNode)
//   router.AddCaptureHandler(inspector, PriorityHigh)
//   router.SetMiddleware(NewMiddlewareChain(loggingMW))
//   result := router.Dispatch(action)
type Router struct {
	// Root is the component tree root node
	Root *runtime.LayoutNode

	// CaptureHandlers executed before reaching target (sorted by priority)
	CaptureHandlers []*CaptureHandlerEntry

	// BubbleHandlers executed after target phase
	BubbleHandlers []*BubbleHandlerEntry

	// TargetHandlers maps target IDs to handlers (supports both string and uint64)
	TargetHandlers map[string]*TargetHandlerEntry

	// GlobalHandlers handle actions without targets
	GlobalHandlers []GlobalActionHandler

	// Middleware chain for cross-cutting concerns
	Middleware *MiddlewareChain
}

// CaptureHandlerEntry represents a capture phase handler with priority
type CaptureHandlerEntry struct {
	Handler  CaptureActionHandler
	Priority int
	ID       string
}

// BubbleHandlerEntry represents a bubble phase handler
type BubbleHandlerEntry struct {
	Handler BubbleActionHandler
	ID      string
}

// TargetHandlerEntry represents a target handler
type TargetHandlerEntry struct {
	Handler Target
	Target  string // Semantic target ID
}

// CaptureActionHandler handles actions during capture phase
type CaptureActionHandler interface {
	// HandleCapture processes action at capture phase
	// Returns true to stop propagation
	HandleCapture(act *Action, target *runtime.LayoutNode) bool

	// Priority returns handler priority (higher = earlier)
	Priority() int
}

// BubbleActionHandler handles actions during bubble phase
type BubbleActionHandler interface {
	// HandleBubble processes action at bubble phase
	// Returns true to stop propagation
	HandleBubble(act *Action, target *runtime.LayoutNode) bool
}

// RouterResult represents action dispatch result
type RouterResult struct {
	// Handled indicates if any handler processed the action
	Handled bool

	// Stopped indicates if action propagation was stopped
	Stopped bool

	// Phase indicates which phase handled the action
	Phase ActionPhase
}

// ActionPhase represents dispatch phase
type ActionPhase int

const (
	// ActionPhaseNone indicates not started
	ActionPhaseNone ActionPhase = iota

	// ActionPhaseCapture is capture phase (root → target)
	ActionPhaseCapture

	// ActionPhaseTarget is target phase (at target)
	ActionPhaseTarget

	// ActionPhaseBubble is bubble phase (target → root)
	ActionPhaseBubble
)

// String returns phase name
func (p ActionPhase) String() string {
	switch p {
	case ActionPhaseNone:
		return "None"
	case ActionPhaseCapture:
		return "Capture"
	case ActionPhaseTarget:
		return "Target"
	case ActionPhaseBubble:
		return "Bubble"
	default:
		return "Unknown"
	}
}

// NewRouter creates a new router
func NewRouter(root *runtime.LayoutNode) *Router {
	return &Router{
		Root:            root,
		CaptureHandlers: make([]*CaptureHandlerEntry, 0),
		BubbleHandlers:  make([]*BubbleHandlerEntry, 0),
		TargetHandlers:  make(map[string]*TargetHandlerEntry),
		GlobalHandlers:  make([]GlobalActionHandler, 0),
		Middleware:      NewMiddlewareChain(),
	}
}

// SetMiddleware sets the middleware chain
func (r *Router) SetMiddleware(chain *MiddlewareChain) {
	r.Middleware = chain
}

// AddMiddleware appends a middleware
func (r *Router) AddMiddleware(middleware ActionMiddleware) {
	if r.Middleware == nil {
		r.Middleware = NewMiddlewareChain()
	}
	r.Middleware.Add(middleware)
}

// AddGlobalHandler adds a global handler
func (r *Router) AddGlobalHandler(handler GlobalActionHandler) {
	r.GlobalHandlers = append(r.GlobalHandlers, handler)
	sort.Slice(r.GlobalHandlers, func(i, j int) bool {
		return r.GlobalHandlers[i].Priority() > r.GlobalHandlers[j].Priority()
	})
}

// AddCaptureHandler adds a capture phase handler
func (r *Router) AddCaptureHandler(handler CaptureActionHandler, id string) {
	entry := &CaptureHandlerEntry{
		Handler:  handler,
		Priority: handler.Priority(),
		ID:       id,
	}
	r.CaptureHandlers = append(r.CaptureHandlers, entry)
	r.sortCaptureHandlers()
}

// AddBubbleHandler adds a bubble phase handler
func (r *Router) AddBubbleHandler(handler BubbleActionHandler, id string) {
	entry := &BubbleHandlerEntry{
		Handler: handler,
		ID:      id,
	}
	r.BubbleHandlers = append(r.BubbleHandlers, entry)
}

// RegisterTarget registers a handler for a target ID
func (r *Router) RegisterTarget(targetID string, handler Target) {
	r.TargetHandlers[targetID] = &TargetHandlerEntry{
		Handler: handler,
		Target:  targetID,
	}
}

// UnregisterTarget removes a handler
func (r *Router) UnregisterTarget(targetID string) {
	delete(r.TargetHandlers, targetID)
}

// RemoveCaptureHandler removes a capture handler by ID
func (r *Router) RemoveCaptureHandler(id string) {
	for i, entry := range r.CaptureHandlers {
		if entry.ID == id {
			r.CaptureHandlers = append(r.CaptureHandlers[:i], r.CaptureHandlers[i+1:]...)
			return
		}
	}
}

// RemoveBubbleHandler removes a bubble handler by ID
func (r *Router) RemoveBubbleHandler(id string) {
	for i, entry := range r.BubbleHandlers {
		if entry.ID == id {
			r.BubbleHandlers = append(r.BubbleHandlers[:i], r.BubbleHandlers[i+1:]...)
			return
		}
	}
}

// sortCaptureHandlers sorts capture handlers by priority (descending)
func (r *Router) sortCaptureHandlers() {
	sort.Slice(r.CaptureHandlers, func(i, j int) bool {
		return r.CaptureHandlers[i].Priority > r.CaptureHandlers[j].Priority
	})
}

// GetRoot returns the root node
func (r *Router) GetRoot() *runtime.LayoutNode {
	return r.Root
}

// SetRoot sets the root node
func (r *Router) SetRoot(root *runtime.LayoutNode) {
	r.Root = root
}

// GetCaptureHandlers returns all capture handlers
func (r *Router) GetCaptureHandlers() []*CaptureHandlerEntry {
	return r.CaptureHandlers
}

// GetBubbleHandlers returns all bubble handlers
func (r *Router) GetBubbleHandlers() []*BubbleHandlerEntry {
	return r.BubbleHandlers
}

// GetTargetHandlers returns all target handlers
func (r *Router) GetTargetHandlers() map[string]*TargetHandlerEntry {
	return r.TargetHandlers
}

// Dispatch dispatches action through three phases
func (r *Router) Dispatch(act *Action) *RouterResult {
	result := &RouterResult{
		Handled: false,
		Stopped: false,
		Phase:   ActionPhaseNone,
	}

	if act == nil {
		return result
	}

	// Apply middleware (Before)
	if r.Middleware != nil {
		act = r.Middleware.Before(act)
		if act == nil {
			result.Handled = true
			result.Stopped = true
			return result
		}
	}

	// Check for stopped flag
	if act.IsStopped() {
		result.Handled = true
		result.Stopped = true
		r.callMiddlewareAfter(act, result)
		return result
	}

	// Global handlers (no target)
	if act.Target == "" && act.TargetID == 0 && len(r.GlobalHandlers) > 0 {
		for _, handler := range r.GlobalHandlers {
			if handler.HandleGlobalAction(act) {
				result.Handled = true
				r.callMiddlewareAfter(act, result)
				return result
			}
		}
	}

	// Find target node
	var targetNode *runtime.LayoutNode
	if act.Target != "" {
		targetNode = r.findNodeByStringID(act.Target)
	} else if act.TargetID != 0 {
		targetNode = r.findNodeByUint64ID(act.TargetID)
	}

	if targetNode == nil {
		targetNode = r.Root
	}

	// Phase 1: Capture
	if r.capturePhase(act, targetNode, result) {
		r.callMiddlewareAfter(act, result)
		return result
	}

	// Phase 2: Target
	if r.targetPhase(act, targetNode, result) {
		r.callMiddlewareAfter(act, result)
		return result
	}

	// Phase 3: Bubble
	if r.bubblePhase(act, targetNode, result) {
		r.callMiddlewareAfter(act, result)
		return result
	}

	r.callMiddlewareAfter(act, result)
	return result
}

// callMiddlewareAfter calls middleware After method
func (r *Router) callMiddlewareAfter(act *Action, result *RouterResult) {
	if r.Middleware != nil {
		r.Middleware.After(act, result)
	}
}

// capturePhase executes capture phase
func (r *Router) capturePhase(act *Action, target *runtime.LayoutNode, result *RouterResult) bool {
	result.Phase = ActionPhaseCapture

	for _, entry := range r.CaptureHandlers {
		if entry.Handler == nil {
			continue
		}

		stopped := entry.Handler.HandleCapture(act, target)
		if stopped {
			result.Handled = true
			result.Stopped = true
			return true
		}
	}
	return false
}

// targetPhase executes target phase
func (r *Router) targetPhase(act *Action, target *runtime.LayoutNode, result *RouterResult) bool {
	result.Phase = ActionPhaseTarget

	// No target ID, skip
	if act.Target == "" && act.TargetID == 0 {
		return false
	}

	// Find handler by target ID
	var entry *TargetHandlerEntry
	var ok bool

	if act.Target != "" {
		entry, ok = r.TargetHandlers[act.Target]
	} else if act.TargetID != 0 {
		// Try string representation of uint64
		entry, ok = r.TargetHandlers[fmt.Sprintf("%d", act.TargetID)]
	}

	if !ok || entry == nil || entry.Handler == nil {
		return false
	}

	// Call HandleAction
	handled := entry.Handler.HandleAction(act)
	if handled {
		result.Handled = true
		return true
	}

	return false
}

// bubblePhase executes bubble phase
func (r *Router) bubblePhase(act *Action, target *runtime.LayoutNode, result *RouterResult) bool {
	result.Phase = ActionPhaseBubble

	// Global bubble handlers
	for _, entry := range r.BubbleHandlers {
		if entry.Handler == nil {
			continue
		}

		stopped := entry.Handler.HandleBubble(act, target)
		if stopped {
			result.Handled = true
			result.Stopped = true
			return true
		}
	}

	// Bubble up the parent chain
	current := target
	for current != nil {
		// Check if node component implements Target
		if current.Component != nil && current.Component.Instance != nil {
			if targetHandler, ok := current.Component.Instance.(Target); ok {
				handled := targetHandler.HandleAction(act)
				if handled {
					result.Handled = true
					return true
				}
			}
		}
		current = current.Parent
	}

	return false
}

// findNodeByStringID finds node by string ID
func (r *Router) findNodeByStringID(id string) *runtime.LayoutNode {
	if r.Root == nil {
		return nil
	}
	return r.findNodeRecursive(r.Root, id)
}

// findNodeByUint64ID finds node by uint64 ID
func (r *Router) findNodeByUint64ID(id uint64) *runtime.LayoutNode {
	if r.Root == nil {
		return nil
	}
	return r.findNodeRecursiveByUint64(r.Root, id)
}

// findNodeRecursive recursively finds node by string ID
func (r *Router) findNodeRecursive(node *runtime.LayoutNode, id string) *runtime.LayoutNode {
	if node == nil {
		return nil
	}

	if node.ID == id {
		return node
	}

	for _, child := range node.Children {
		if found := r.findNodeRecursive(child, id); found != nil {
			return found
		}
	}

	return nil
}

// findNodeRecursiveByUint64 recursively finds node by uint64 ID
func (r *Router) findNodeRecursiveByUint64(node *runtime.LayoutNode, id uint64) *runtime.LayoutNode {
	if node == nil {
		return nil
	}

	nodeID := StringToTargetID(node.ID)
	if nodeID == id {
		return node
	}

	for _, child := range node.Children {
		if found := r.findNodeRecursiveByUint64(child, id); found != nil {
			return found
		}
	}

	return nil
}

// BuildTargetRegistry traverses component tree and registers all Targets
func (r *Router) BuildTargetRegistry() {
	if r.Root == nil {
		return
	}

	r.TargetHandlers = make(map[string]*TargetHandlerEntry)
	r.registerNodeRecursive(r.Root)
}

// registerNodeRecursive recursively registers nodes
func (r *Router) registerNodeRecursive(node *runtime.LayoutNode) {
	if node == nil {
		return
	}

	if node.Component != nil && node.Component.Instance != nil {
		if target, ok := node.Component.Instance.(Target); ok && node.ID != "" {
			r.RegisterTarget(node.ID, target)
		}
	}

	for _, child := range node.Children {
		r.registerNodeRecursive(child)
	}
}
