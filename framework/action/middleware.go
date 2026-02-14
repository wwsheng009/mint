package action

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/internal/log"
)

// ============================================================================
// Logging Middleware
// ============================================================================

// LoggingMiddleware 日志中间件，记录所有 Action 的分发和处理情况
type LoggingMiddleware struct {
	enabled bool
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		enabled: os.Getenv("ACTION_DEBUG") == "true",
	}
}

// SetEnabled 设置是否启用日志
func (m *LoggingMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Name 返回中间件名称
func (m *LoggingMiddleware) Name() string {
	return "logging"
}

// Before 记录 Action 开始分发
func (m *LoggingMiddleware) Before(action *Action) *Action {
	if !m.enabled {
		return action
	}

	log.UILogger.Debug("[Action] ↓ Dispatch: Type=%s, TargetID=%d, Source=%s",
		action.Type, action.TargetID, action.Source)

	// 记录开始时间到 Meta
	if action.Meta == nil {
		action.Meta = make(map[string]interface{})
	}
	action.Meta["_start"] = time.Now()

	return action
}

// After 记录 Action 处理结果
func (m *LoggingMiddleware) After(action *Action, result *RouterResult) {
	if !m.enabled {
		return
	}

	// 计算处理时长
	var duration time.Duration
	if start, ok := action.Meta["_start"].(time.Time); ok {
		duration = time.Since(start)
	}

	handled := "✓"
	if !result.Handled {
		handled = "✗"
	}

	phase := "None"
	if result.Phase == ActionPhaseCapture {
		phase = "Capture"
	} else if result.Phase == ActionPhaseTarget {
		phase = "Target"
	} else if result.Phase == ActionPhaseBubble {
		phase = "Bubble"
	}

	log.UILogger.Debug("[Action] ↑ Complete: Type=%s %s, Phase=%s, Duration=%v",
		action.Type, handled, phase, duration)
}

// ============================================================================
// Throttle Middleware
// ============================================================================

// ThrottleMiddleware 节流中间件，防止 Action 触发过于频繁
type ThrottleMiddleware struct {
	interval   time.Duration
	lastAction map[ActionType]time.Time
	mu         sync.Mutex
}

// NewThrottleMiddleware 创建节流中间件
func NewThrottleMiddleware(interval time.Duration) *ThrottleMiddleware {
	return &ThrottleMiddleware{
		interval:   interval,
		lastAction: make(map[ActionType]time.Time),
	}
}

// Name 返回中间件名称
func (m *ThrottleMiddleware) Name() string {
	return "throttle"
}

// Before 检查是否需要拦截（触发过于频繁）
func (m *ThrottleMiddleware) Before(action *Action) *Action {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	last, exists := m.lastAction[action.Type]

	// 某些 Action 类型不进行节流（如取消、退出等）
	switch action.Type {
	case ActionCancel, ActionQuit, ActionNavigateEnd:
		// 这些操作不节流
		m.lastAction[action.Type] = now
		return action
	}

	// 检查是否触发太频繁
	if exists && now.Sub(last) < m.interval {
		// 触发太频繁，拦截
		if os.Getenv("ACTION_DEBUG") == "true" {
			log.UILogger.Debug("[Action] ⚠ Throttled: Type=%s (last: %v ago)",
				action.Type, now.Sub(last))
		}
		return nil
	}

	// 更新最后触发时间
	m.lastAction[action.Type] = now
	return action
}

// After 空实现
func (m *ThrottleMiddleware) After(action *Action, result *RouterResult) {
	// 节流中间件不需要 After 操作
}

// ============================================================================
// Validation Middleware
// ============================================================================

// ActionValidator Action 验证函数类型
type ActionValidator func(action *Action) error

// ValidationMiddleware 验证中间件，在 Action 分发前进行验证
type ValidationMiddleware struct {
	validators map[ActionType]ActionValidator
	mu         sync.RWMutex
}

// NewValidationMiddleware 创建验证中间件
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validators: make(map[ActionType]ActionValidator),
	}
}

// Name 返回中间件名称
func (m *ValidationMiddleware) Name() string {
	return "validation"
}

// RegisterValidator 注册验证器
func (m *ValidationMiddleware) RegisterValidator(actionType ActionType, validator ActionValidator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validators[actionType] = validator
}

// UnregisterValidator 取消注册验证器
func (m *ValidationMiddleware) UnregisterValidator(actionType ActionType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.validators, actionType)
}

// Before 验证 Action
func (m *ValidationMiddleware) Before(action *Action) *Action {
	m.mu.RLock()
	validator, exists := m.validators[action.Type]
	m.mu.RUnlock()

	if !exists {
		return action
	}

	if err := validator(action); err != nil {
		log.UILogger.Debug("[Action] ✗ Validation failed: Type=%s, Error=%v",
			action.Type, err)
		return nil // 验证失败，拦截
	}

	return action
}

// After 空实现
func (m *ValidationMiddleware) After(action *Action, result *RouterResult) {
	// 验证中间件不需要 After 操作
}

// ============================================================================
// Metrics Middleware
// ============================================================================

// MetricsMiddleware 指标中间件，收集 Action 处理的统计数据
type MetricsMiddleware struct {
	// Action 计数
	actionCounts map[ActionType]int64
	// 处理时长统计
	durations map[ActionType][]time.Duration
	// 错误计数
	errorCounts map[ActionType]int64

	mu sync.RWMutex
}

// NewMetricsMiddleware 创建指标中间件
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		actionCounts: make(map[ActionType]int64),
		durations:    make(map[ActionType][]time.Duration),
		errorCounts:  make(map[ActionType]int64),
	}
}

// Name 返回中间件名称
func (m *MetricsMiddleware) Name() string {
	return "metrics"
}

// Before 记录开始时间
func (m *MetricsMiddleware) Before(action *Action) *Action {
	m.mu.Lock()
	m.actionCounts[action.Type]++
	m.mu.Unlock()

	// 记录开始时间
	if action.Meta == nil {
		action.Meta = make(map[string]interface{})
	}
	action.Meta["_metrics_start"] = time.Now()

	return action
}

// After 记录处理时长
func (m *MetricsMiddleware) After(action *Action, result *RouterResult) {
	start, ok := action.Meta["_metrics_start"].(time.Time)
	if !ok {
		return
	}
	duration := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	// 记录时长（保留最近 100 个样本）
	durations := m.durations[action.Type]
	durations = append(durations, duration)
	if len(durations) > 100 {
		durations = durations[1:]
	}
	m.durations[action.Type] = durations

	// 记录未处理的 Action
	if !result.Handled {
		m.errorCounts[action.Type]++
	}
}

// GetActionCount 获取指定 Action 类型的计数
func (m *MetricsMiddleware) GetActionCount(actionType ActionType) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.actionCounts[actionType]
}

// GetAllActionCounts 获取所有 Action 计数
func (m *MetricsMiddleware) GetAllActionCounts() map[ActionType]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[ActionType]int64, len(m.actionCounts))
	for k, v := range m.actionCounts {
		result[k] = v
	}
	return result
}

// GetAverageDuration 获取指定 Action 类型的平均处理时长
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

// GetErrorCount 获取指定 Action 类型的错误计数
func (m *MetricsMiddleware) GetErrorCount(actionType ActionType) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorCounts[actionType]
}

// Reset 重置所有统计数据
func (m *MetricsMiddleware) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.actionCounts = make(map[ActionType]int64)
	m.durations = make(map[ActionType][]time.Duration)
	m.errorCounts = make(map[ActionType]int64)
}

// FormatStats 格式化统计信息
func (m *MetricsMiddleware) FormatStats() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("Action Metrics:\n")

	// Action 计数
	sb.WriteString("\nAction Counts:\n")
	for actionType, count := range m.actionCounts {
		if count > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", actionType, count))
		}
	}

	// 平均时长
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

	// 错误计数
	sb.WriteString("\nError Counts:\n")
	for actionType, count := range m.errorCounts {
		if count > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", actionType, count))
		}
	}

	return sb.String()
}

// ============================================================================
// Recovery Middleware
// ============================================================================

// RecoveryMiddleware 恢复中间件，捕获 Action 处理过程中的 panic
type RecoveryMiddleware struct {
	onPanic func(action *Action, recovered interface{})
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{
		onPanic: func(action *Action, recovered interface{}) {
			log.UILogger.Debug("[Action] ⚠ Panic recovered: Type=%s, Error=%v",
				action.Type, recovered)
		},
	}
}

// Name 返回中间件名称
func (m *RecoveryMiddleware) Name() string {
	return "recovery"
}

// SetOnPanic 设置 panic 回调
func (m *RecoveryMiddleware) SetOnPanic(fn func(action *Action, recovered interface{})) {
	m.onPanic = fn
}

// Before 空实现（恢复在 After 中处理）
func (m *RecoveryMiddleware) Before(action *Action) *Action {
	return action
}

// After 检查并恢复 panic
func (m *RecoveryMiddleware) After(action *Action, result *RouterResult) {
	// 这个中间件需要在 Router 层面实现 panic 捕获
	// 这里只是占位符，实际实现需要修改 Router.Dispatch()
}

// ============================================================================
// Built-in Middleware Chains
// ============================================================================

// DefaultMiddlewareChain 返回默认的中间件链
// 包含：日志、节流、验证、恢复
func DefaultMiddlewareChain() *MiddlewareChain {
	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		NewThrottleMiddleware(16*time.Millisecond), // ~60fps 限制
		NewValidationMiddleware(),
		NewMetricsMiddleware(),
		NewLoggingMiddleware(),
	)
}

// DebugMiddlewareChain 返回调试模式的中间件链
// 包含：日志（启用）、指标
func DebugMiddlewareChain() *MiddlewareChain {
	logging := NewLoggingMiddleware()
	logging.enabled = true

	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		logging,
		NewMetricsMiddleware(),
	)
}

// ProductionMiddlewareChain 返回生产环境的中间件链
// 包含：节流、恢复
func ProductionMiddlewareChain() *MiddlewareChain {
	return NewMiddlewareChain(
		NewRecoveryMiddleware(),
		NewThrottleMiddleware(16*time.Millisecond),
	)
}
