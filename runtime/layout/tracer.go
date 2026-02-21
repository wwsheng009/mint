package layout

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Tracer - Constraint Propagation Tracer
// ============================================================================

// Tracer 追踪约束的传播过程
type Tracer struct {
	mu       sync.Mutex
	enabled  bool
	entries  []TraceEntry
	compact  bool
	showPath bool
}

// TraceEntry 记录一次约束传播
type TraceEntry struct {
	Seq        int           // 序列号
	Timestamp  time.Time     // 时间戳
	From       string        // 来源组件 ID
	To         string        // 目标组件 ID
	Path       string        // 完整路径
	Input      Constraints   // 输入约束
	Output     Constraints   // 输出约束
	Dimension  Size          // 测量结果
	Reason     string        // 约束修改原因
}

// NewTracer 创建新的追踪器
func NewTracer() *Tracer {
	return &Tracer{
		enabled: true,
		entries: make([]TraceEntry, 0),
		compact: false,
		showPath: true,
	}
}

// ============================================================================
// 全局追踪器
// ============================================================================

var (
	globalTracer     *Tracer
	globalTracerOnce sync.Once
)

// GetGlobalTracer 获取全局追踪器实例
func GetGlobalTracer() *Tracer {
	globalTracerOnce.Do(func() {
		globalTracer = NewTracer()
		globalTracer.enabled = false // 默认禁用
	})
	return globalTracer
}

// EnableTracer 启用全局追踪器
func EnableTracer() {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = true
	t.entries = nil
}

// DisableTracer 禁用全局追踪器
func DisableTracer() {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = false
}

// IsTracerEnabled 检查追踪器是否启用
func IsTracerEnabled() bool {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enabled
}

// TraceMeasuring 追踪测量过程的约束传递
func TraceMeasuring(from, to, path string, input, output Constraints, resultSize Size, reason string) {
	if !IsTracerEnabled() {
		return
	}

	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()

	seq := len(t.entries)
	
	entry := TraceEntry{
		Seq:       seq,
		Timestamp: time.Now(),
		From:      from,
		To:        to,
		Path:      path,
		Input:     input,
		Output:    output,
		Dimension: resultSize,
		Reason:    reason,
	}

	t.entries = append(t.entries, entry)
}

// ============================================================================
// 输出方法
// ============================================================================

// Dump 返回追踪日志
func DumpTrace() string {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dump()
}

func (t *Tracer) dump() string {
	if !t.enabled || len(t.entries) == 0 {
		return "No trace data available (tracing is disabled or no entries recorded)"
	}

	var buf strings.Builder

	// 标题
	buf.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    Constraint Propagation Trace               ║\n")
	buf.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

	// 输出每个条目
	for _, entry := range t.entries {
		buf.WriteString(t.formatEntry(entry))
	}

	return buf.String()
}

func (t *Tracer) formatEntry(entry TraceEntry) string {
	var buf strings.Builder

	// 序号
	buf.WriteString(fmt.Sprintf("Step %d\n", entry.Seq))

	// 路径
	if t.showPath && entry.Path != "" {
		buf.WriteString(fmt.Sprintf("  Path: %s\n", entry.Path))
	}

	// 组件 ID
	buf.WriteString(fmt.Sprintf("  %s → %s\n", entry.From, entry.To))

	// 约束
	buf.WriteString(fmt.Sprintf("  Input:    %s\n", formatConstraints(entry.Input)))
	buf.WriteString(fmt.Sprintf("  Output:   %s\n", formatConstraints(entry.Output)))

	// 尺寸
	if entry.Dimension.Width > 0 || entry.Dimension.Height > 0 {
		buf.WriteString(fmt.Sprintf("  Dimension: %s\n", formatSize(entry.Dimension)))
	}

	// 原因
	if entry.Reason != "" {
		buf.WriteString(fmt.Sprintf("  Reason:   %s\n", entry.Reason))
	}

	// 检测问题
	if entry.Dimension.Height > entry.Input.MaxHeight && entry.Input.MaxHeight > 0 && entry.Input.MaxHeight < MaxInt {
		buf.WriteString(fmt.Sprintf("  ⚠️  Height %d exceeds MaxHeight %d\n",
			entry.Dimension.Height, entry.Input.MaxHeight))
	}

	buf.WriteString("\n")
	return buf.String()
}

// ============================================================================
// 配置方法
// ============================================================================

// SetCompactMode 设置紧凑模式
func SetCompactMode(compact bool) {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.compact = compact
}

// SetShowPath 设置是否显示路径
func SetShowPath(show bool) {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.showPath = show
}

// =============================================================================
// 辅助函数
// =============================================================================

func formatConstraints(c Constraints) string {
	return fmt.Sprintf("{%d..%d} × {%d..%d}",
		c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

func formatSize(s Size) string {
	return fmt.Sprintf("%dw × %dh", s.Width, s.Height)
}

// =============================================================================
// 清除方法
// =============================================================================

// ClearTrace 清除追踪数据
func ClearTrace() {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = nil
}

// GetTraceEntries 获取追踪条目（用于测试）
func GetTraceEntries() []TraceEntry {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]TraceEntry{}, t.entries...)
}

// GetEntryCount 获取追踪条目数
func GetEntryCount() int {
	t := GetGlobalTracer()
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}
