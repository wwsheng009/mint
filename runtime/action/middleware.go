package action

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/internal/log"
)

// ============================================================================
// ActionMiddleware Interface
// (Already defined in router.go, repeated here for reference only)
// ============================================================================

// MiddlewareChain is defined in router.go
// type MiddlewareChain struct { middlewares []ActionMiddleware }

// ============================================================================
// 1. Logging Middleware
// ============================================================================

// LoggingMiddleware records action dispatch and handling
type LoggingMiddleware struct {
	enabled bool
}

// NewLoggingMiddleware creates a logging middleware
// Use ACTION_DEBUG=true environment variable to enable by default
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		enabled: os.Getenv("ACTION_DEBUG") == "true",
	}
}

// SetEnabled sets logging enabled state
func (m *LoggingMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Name returns middleware name
func (m *LoggingMiddleware) Name() string {
	return "logging"
}

// Before logs action dispatch start
func (m *LoggingMiddleware) Before(action *Action) *Action {
	if !m.enabled {
		return action
	}

	log.PaintLogger.Debug("[Action] ↓ Dispatch: Type=%s, Target=%s/%d, Source=%s\n",
		action.Type, action.Target, action.TargetID, action.Source)

	// Record start time in Meta
	if action.Meta == nil {
		action.Meta = make(map[string]interface{})
	}
	action.Meta["_start"] = time.Now()

	return action
}

// After logs action handling result
func (m *LoggingMiddleware) After(action *Action, result *RouterResult) {
	if !m.enabled {
		return
	}

	// Calculate duration
	var duration time.Duration
	if start, ok := action.Meta["_start"].(time.Time); ok {
		duration = time.Since(start)
	}

	handled := "✓"
	if !result.Handled {
		handled = "✗"
	}

	phaseStr := result.Phase.String()
	log.PaintLogger.Debug("[Action] ↑ Complete: Type=%s %s, Phase=%s, Duration=%v\n",
		action.Type, handled, phaseStr, duration)
}

// ============================================================================
// 2. Throttle Middleware
// ============================================================================

// ThrottleMiddleware prevents action triggers from being too frequent
type ThrottleMiddleware struct {
	interval   time.Duration
	lastAction map[ActionType]time.Time
	mu         sync.Mutex
}

// NewThrottleMiddleware creates a throttle middleware
// interval is the minimum time between identical actions
func NewThrottleMiddleware(interval time.Duration) *ThrottleMiddleware {
	return &ThrottleMiddleware{
		interval:   interval,
		lastAction: make(map[ActionType]time.Time),
	}
}

// Name returns middleware name
func (m *ThrottleMiddleware) Name() string {
	return "throttle"
}

// Before checks if action should be throttled
func (m *ThrottleMiddleware) Before(action *Action) *Action {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	last, exists := m.lastAction[action.Type]

	// Certain actions should not be throttled (e.g., cancel, quit)
	switch action.Type {
	case ActionCancel, ActionQuit, ActionNavigateEnd:
		m.lastAction[action.Type] = now
		return action
	}

	if action.IsNavigation() {
		m.lastAction[action.Type] = now
		return action
	}

	// Text editing actions must preserve every event. Terminal paste on some
	// platforms arrives as a rapid stream of key presses, and throttling these
	// actions drops characters.
	if isUnthrottledEditingAction(action.Type) {
		m.lastAction[action.Type] = now
		return action
	}

	// Check if triggered too frequently
	if exists && now.Sub(last) < m.interval {
		log.ActionLogger.Debug("[Action] ⚠ Throttled: Type=%s (last: %v ago)\n",
			action.Type, now.Sub(last))

		return nil
	}

	// Update last trigger time
	m.lastAction[action.Type] = now
	return action
}

func isUnthrottledEditingAction(actionType ActionType) bool {
	switch actionType {
	case ActionInputChar,
		ActionInputText,
		ActionPaste,
		ActionBackspace,
		ActionDeleteChar,
		ActionDeleteWord,
		ActionDeleteLine,
		ActionCursorHome,
		ActionCursorEnd,
		ActionCursorUp,
		ActionCursorDown,
		ActionCursorLeft,
		ActionCursorRight,
		ActionCursorWordLeft,
		ActionCursorWordRight,
		ActionSelectAll,
		ActionSelectWord,
		ActionSelectLine:
		return true
	default:
		return false
	}
}

// After is no-op for throttle middleware
func (m *ThrottleMiddleware) After(action *Action, result *RouterResult) {
	// No-op
}

// ============================================================================
// 3. Validation Middleware
// ============================================================================

// ActionValidator is a function type for action validation
type ActionValidator func(action *Action) error

// ValidationMiddleware validates actions before dispatch
type ValidationMiddleware struct {
	validators map[ActionType]ActionValidator
	mu         sync.RWMutex
}

// NewValidationMiddleware creates a validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validators: make(map[ActionType]ActionValidator),
	}
}

// Name returns middleware name
func (m *ValidationMiddleware) Name() string {
	return "validation"
}

// RegisterValidator registers a validator for an action type
func (m *ValidationMiddleware) RegisterValidator(actionType ActionType, validator ActionValidator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validators[actionType] = validator
}

// UnregisterValidator unregisters a validator
func (m *ValidationMiddleware) UnregisterValidator(actionType ActionType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.validators, actionType)
}

// Before validates the action
func (m *ValidationMiddleware) Before(action *Action) *Action {
	m.mu.RLock()
	validator, exists := m.validators[action.Type]
	m.mu.RUnlock()

	if !exists {
		return action
	}

	if err := validator(action); err != nil {
		log.PaintLogger.Debug("[Action] ✗ Validation failed: Type=%s, Error=%v\n",
			action.Type, err)
		return nil // Validation failed, intercept
	}

	return action
}

// After is no-op for validation middleware
func (m *ValidationMiddleware) After(action *Action, result *RouterResult) {
	// No-op
}

// ============================================================================
// 4. Metrics Middleware
// ============================================================================

// MetricsMiddleware collects action processing statistics
type MetricsMiddleware struct {
	// Action counts
	actionCounts map[ActionType]int64
	// Duration statistics
	durations map[ActionType][]time.Duration
	// Error counts
	errorCounts map[ActionType]int64

	mu sync.RWMutex
}

// NewMetricsMiddleware creates a metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		actionCounts: make(map[ActionType]int64),
		durations:    make(map[ActionType][]time.Duration),
		errorCounts:  make(map[ActionType]int64),
	}
}

// Name returns middleware name
func (m *MetricsMiddleware) Name() string {
	return "metrics"
}

// Before records start time and increments count
func (m *MetricsMiddleware) Before(action *Action) *Action {
	m.mu.Lock()
	m.actionCounts[action.Type]++
	m.mu.Unlock()

	// Record start time
	if action.Meta == nil {
		action.Meta = make(map[string]interface{})
	}
	action.Meta["_metrics_start"] = time.Now()

	return action
}

// After records duration
func (m *MetricsMiddleware) After(action *Action, result *RouterResult) {
	start, ok := action.Meta["_metrics_start"].(time.Time)
	if !ok {
		return
	}
	duration := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Record duration (keep last 100 samples)
	durations := m.durations[action.Type]
	durations = append(durations, duration)
	if len(durations) > 100 {
		durations = durations[1:]
	}
	m.durations[action.Type] = durations

	// Record unhandled actions as errors
	if !result.Handled {
		m.errorCounts[action.Type]++
	}
}

// GetActionCount returns action count for a type
func (m *MetricsMiddleware) GetActionCount(actionType ActionType) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionCounts[actionType]
}

// GetAllActionCounts returns all action counts
func (m *MetricsMiddleware) GetAllActionCounts() map[ActionType]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[ActionType]int64, len(m.actionCounts))
	for k, v := range m.actionCounts {
		result[k] = v
	}
	return result
}

// GetAverageDuration returns average processing duration for an action type
func (m *MetricsMiddleware) GetAverageDuration(actionType ActionType) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	durations := m.durations[actionType]
	if len(durations) == 0 {
		return 0
	}

	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

// GetErrorCount returns error count for an action type
func (m *MetricsMiddleware) GetErrorCount(actionType ActionType) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorCounts[actionType]
}

// Reset resets all statistics
func (m *MetricsMiddleware) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.actionCounts = make(map[ActionType]int64)
	m.durations = make(map[ActionType][]time.Duration)
	m.errorCounts = make(map[ActionType]int64)
}

// FormatStats formats statistics as string
func (m *MetricsMiddleware) FormatStats() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("Action Metrics:\n")

	// Action counts
	sb.WriteString("\nAction Counts:\n")
	for actionType, count := range m.actionCounts {
		if count > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", actionType, count))
		}
	}

	// Average durations
	sb.WriteString("\nAverage Durations:\n")
	for actionType, durations := range m.durations {
		if len(durations) > 0 {
			var sum time.Duration
			for _, d := range durations {
				sum += d
			}
			avg := sum / time.Duration(len(durations))
			sb.WriteString(fmt.Sprintf("  %s: %v\n", actionType, avg))
		}
	}

	// Error counts
	sb.WriteString("\nError Counts:\n")
	for actionType, count := range m.errorCounts {
		if count > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", actionType, count))
		}
	}

	return sb.String()
}

// ============================================================================
// 5. Recovery Middleware
// ============================================================================

// RecoveryMiddleware catches panics during action processing
type RecoveryMiddleware struct {
	onPanic func(action *Action, recovered interface{})
}

// NewRecoveryMiddleware creates a recovery middleware
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{
		onPanic: func(action *Action, recovered interface{}) {
			log.PaintLogger.Debug("[Action] ⚠ Panic recovered: Type=%s, Error=%v\n",
				action.Type, recovered)
		},
	}
}

// Name returns middleware name
func (m *RecoveryMiddleware) Name() string {
	return "recovery"
}

// SetOnPanic sets panic callback
func (m *RecoveryMiddleware) SetOnPanic(fn func(action *Action, recovered interface{})) {
	m.onPanic = fn
}

// Before is no-op (recovery handled elsewhere)
func (m *RecoveryMiddleware) Before(action *Action) *Action {
	return action
}

// After checks for and recovers from panics
// Note: Actual panic recovery needs to be implemented in Router.Dispatch()
func (m *RecoveryMiddleware) After(action *Action, result *RouterResult) {
	// Placeholder: actual implementation needs Router-level panic catching
}

// ============================================================================
// 6. Profiling Middleware (NEW)
// ============================================================================

// ProfilingMiddleware measures action processing performance with detailed metrics
type ProfilingMiddleware struct {
	enabled     bool
	actionCall  stacks // Stack traces for each action type
	cpuProfiled bool
	memProfiled bool

	mu sync.RWMutex
}

type stacks map[ActionType][]SampledStack

// SampledStack represents a sampled stack trace
type SampledStack struct {
	Stack    []uintptr
	Count    int
	Duration time.Duration
}

// NewProfilingMiddleware creates a profiling middleware
func NewProfilingMiddleware() *ProfilingMiddleware {
	return &ProfilingMiddleware{
		enabled:    true,
		actionCall: make(stacks),
	}
}

// Name returns middleware name
func (m *ProfilingMiddleware) Name() string {
	return "profiling"
}

// SetEnabled enables/disables profiling
func (m *ProfilingMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Before captures call stack
func (m *ProfilingMiddleware) Before(action *Action) *Action {
	if !m.enabled {
		return action
	}

	// Capture call stack (depth limited to avoid overhead)
	pcs := make([]uintptr, 8)
	n := runtime.Callers(2, pcs)

	sampled := SampledStack{
		Stack: pcs[:n],
		Count: 1,
	}

	m.mu.Lock()
	m.actionCall[action.Type] = append(m.actionCall[action.Type], sampled)
	m.mu.Unlock()

	// Record start time
	if action.Meta == nil {
		action.Meta = make(map[string]interface{})
	}
	action.Meta["_profile_start"] = time.Now()

	return action
}

// After records duration
func (m *ProfilingMiddleware) After(action *Action, result *RouterResult) {
	if !m.enabled {
		return
	}

	start, ok := action.Meta["_profile_start"].(time.Time)
	if !ok {
		return
	}
	duration := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update the last sample's duration
	stacks := m.actionCall[action.Type]
	if len(stacks) > 0 {
		stacks[len(stacks)-1].Duration = duration
	}
}

// GetHotSpots returns action types with highest average duration
func (m *ProfilingMiddleware) GetHotSpots(limit int) []ActionType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type avgDuration struct {
		actionType ActionType
		duration   time.Duration
		count      int
	}

	avgs := make([]avgDuration, 0, len(m.actionCall))
	for actionType, samples := range m.actionCall {
		var total time.Duration
		count := 0
		for _, sample := range samples {
			total += sample.Duration
			count += sample.Count
		}
		if count > 0 {
			avgs = append(avgs, avgDuration{
				actionType: actionType,
				duration:   total / time.Duration(count),
				count:      count,
			})
		}
	}

	// Sort by duration descending
	for i := 0; i < len(avgs)-1; i++ {
		for j := i + 1; j < len(avgs); j++ {
			if avgs[i].duration < avgs[j].duration {
				avgs[i], avgs[j] = avgs[j], avgs[i]
			}
		}
	}

	result := make([]ActionType, 0, limit)
	for i, avg := range avgs {
		if i >= limit {
			break
		}
		result = append(result, avg.actionType)
	}

	return result
}

// ============================================================================
// 7. Caching Middleware (NEW)
// ============================================================================

// CachingMiddleware caches action handler results for performance
type CachingMiddleware struct {
	enabled bool
	cache   map[cacheKey]interface{}
	ttl     time.Duration
	maxSize int
	mu      sync.RWMutex
}

type cacheKey struct {
	actionType ActionType
	targetID   string
	targetU64  uint64
}

// cachedResult represents a cached result with expiration
type cachedResult struct {
	value     interface{}
	expiresAt time.Time
}

// NewCachingMiddleware creates a caching middleware
func NewCachingMiddleware(ttl time.Duration, maxSize int) *CachingMiddleware {
	return &CachingMiddleware{
		enabled: true,
		ttl:     ttl,
		maxSize: maxSize,
		cache:   make(map[cacheKey]interface{}),
	}
}

// Name returns middleware name
func (m *CachingMiddleware) Name() string {
	return "caching"
}

// SetEnabled enables/disables caching
func (m *CachingMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Before checks cache for existing result
func (m *CachingMiddleware) Before(action *Action) *Action {
	if !m.enabled {
		return action
	}

	key := cacheKey{
		actionType: action.Type,
		targetID:   action.Target,
		targetU64:  action.TargetID,
	}

	m.mu.RLock()
	result, exists := m.cache[key]
	m.mu.RUnlock()

	if exists {
		// If cache hit, we'd typically return a modified action with the cached result
		// For simplicity, we'll just mark it in the action
		if action.Meta == nil {
			action.Meta = make(map[string]interface{})
		}
		action.Meta["_cached_result"] = result
	}

	return action
}

// After caches the result
func (m *CachingMiddleware) After(action *Action, result *RouterResult) {
	if !m.enabled {
		return
	}

	// Only cache successful results
	if !result.Handled {
		return
	}

	key := cacheKey{
		actionType: action.Type,
		targetID:   action.Target,
		targetU64:  action.TargetID,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Evict if cache is full
	if len(m.cache) >= m.maxSize {
		// Simple eviction: delete first key
		for k := range m.cache {
			delete(m.cache, k)
			break
		}
	}

	m.cache[key] = cachedResult{
		value:     result,
		expiresAt: time.Now().Add(m.ttl),
	}
}

// Clear clears the cache
func (m *CachingMiddleware) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[cacheKey]interface{})
}

// Cleanup removes expired entries
func (m *CachingMiddleware) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, value := range m.cache {
		if cached, ok := value.(cachedResult); ok && now.After(cached.expiresAt) {
			delete(m.cache, key)
		}
	}
}

// ============================================================================
// 8. Audit Middleware (NEW)
// ============================================================================

// AuditMiddleware logs all actions for audit/compliance purposes
type AuditMiddleware struct {
	enabled bool
	entries []AuditEntry
	maxSize int
	mu      sync.RWMutex
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Action    *Action
	Result    *RouterResult
	Timestamp time.Time
	User      string // User identifier (if available)
}

// NewAuditMiddleware creates an audit middleware
func NewAuditMiddleware(maxSize int) *AuditMiddleware {
	return &AuditMiddleware{
		enabled: true,
		maxSize: maxSize,
	}
}

// Name returns middleware name
func (m *AuditMiddleware) Name() string {
	return "audit"
}

// SetEnabled enables/disables auditing
func (m *AuditMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Before is no-op for audit
func (m *AuditMiddleware) Before(action *Action) *Action {
	return action
}

// After logs action to audit log
func (m *AuditMiddleware) After(action *Action, result *RouterResult) {
	if !m.enabled {
		return
	}

	entry := AuditEntry{
		Action:    action.Clone(),
		Result:    result,
		Timestamp: time.Now(),
		User:      getMetaString(action, "user_id"),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = append(m.entries, entry)

	// Limit size
	if len(m.entries) > m.maxSize {
		m.entries = m.entries[len(m.entries)-m.maxSize:]
	}
}

// GetAuditLog returns audit log entries
func (m *AuditMiddleware) GetAuditLog() []AuditEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]AuditEntry, len(m.entries))
	copy(entries, m.entries)
	return entries
}

// GetAuditLogForActionType returns entries for specific action type
func (m *AuditMiddleware) GetAuditLogForActionType(actionType ActionType) []AuditEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []AuditEntry
	for _, entry := range m.entries {
		if entry.Action.Type == actionType {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Clear clears audit log
func (m *AuditMiddleware) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = m.entries[:0]
}

// ============================================================================
// Built-in Middleware Chains
// ============================================================================

// DefaultMiddlewareChain returns default middleware chain
// Contains: Recovery, Throttle, Validation, Metrics, Logging
func DefaultMiddlewareChain() *MiddlewareChain {
	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		NewThrottleMiddleware(16*time.Millisecond), // ~60fps limit
		NewValidationMiddleware(),
		NewMetricsMiddleware(),
		NewLoggingMiddleware(),
	)
}

// FullMiddlewareChain returns full middleware chain with all 8 middlewares
func FullMiddlewareChain() *MiddlewareChain {
	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		NewThrottleMiddleware(16*time.Millisecond),
		NewValidationMiddleware(),
		NewCachingMiddleware(5*time.Minute, 1000),
		NewProfilingMiddleware(),
		NewMetricsMiddleware(),
		NewAuditMiddleware(5000),
		NewLoggingMiddleware(),
	)
}

// DebugMiddlewareChain returns debug mode middleware chain
func DebugMiddlewareChain() *MiddlewareChain {
	logging := NewLoggingMiddleware()
	logging.enabled = true

	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		logging,
		NewProfilingMiddleware(),
		NewMetricsMiddleware(),
	)
}

// ProductionMiddlewareChain returns production environment middleware chain
func ProductionMiddlewareChain() *MiddlewareChain {
	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		NewThrottleMiddleware(16*time.Millisecond),
		NewMetricsMiddleware(),
		NewAuditMiddleware(10000),
	)
}

// MinimalMiddlewareChain returns minimal middleware chain
func MinimalMiddlewareChain() *MiddlewareChain {
	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		NewThrottleMiddleware(16*time.Millisecond),
	)
}

// getMetaString is a helper to get meta value as string
func getMetaString(action *Action, key string) string {
	if action.Meta == nil {
		return ""
	}
	if val, ok := action.Meta[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
