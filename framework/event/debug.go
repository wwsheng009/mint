package event

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/internal/log"
)

// EventLogger 记录和可视化事件流
type EventLogger struct {
	mu       sync.Mutex
	entries  []*EventLogEntry
	enabled  bool
	maxSize  int
}

// EventLogEntry 表示一个事件日志条目
type EventLogEntry struct {
	Timestamp time.Time
	Phase     string
	EventType string
	TargetID  string
	Details   string
	Duration  time.Duration
}

// NewEventLogger 创建新的事件日志记录器
func NewEventLogger() *EventLogger {
	return &EventLogger{
		entries: make([]*EventLogEntry, 0),
		enabled: log.EventLogger.Enabled(),
		maxSize: 1000, // 最多保存 1000 条日志
	}
}

// Log 记录一个事件
func (l *EventLogger) Log(phase, eventType, targetID, details string) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := &EventLogEntry{
		Timestamp: time.Now(),
		Phase:     phase,
		EventType: eventType,
		TargetID:  targetID,
		Details:   details,
	}

	l.entries = append(l.entries, entry)

	// 限制日志大小
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[1:]
	}
}

// LogWithDuration 记录一个事件及其持续时间
func (l *EventLogger) LogWithDuration(phase, eventType, targetID, details string, duration time.Duration) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := &EventLogEntry{
		Timestamp: time.Now(),
		Phase:     phase,
		EventType: eventType,
		TargetID:  targetID,
		Details:   details,
		Duration:  duration,
	}

	l.entries = append(l.entries, entry)

	if len(l.entries) > l.maxSize {
		l.entries = l.entries[1:]
	}
}

// GetEntries 获取所有日志条目
func (l *EventLogger) GetEntries() []*EventLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 返回副本
	entries := make([]*EventLogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// Clear 清空日志
func (l *EventLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = make([]*EventLogEntry, 0)
}

// Enable 启用日志记录
func (l *EventLogger) Enable() {
	l.enabled = true
}

// Disable 禁用日志记录
func (l *EventLogger) Disable() {
	l.enabled = false
}

// IsEnabled 检查是否启用
func (l *EventLogger) IsEnabled() bool {
	return l.enabled
}

// Dump 转储日志为字符串
func (l *EventLogger) Dump() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 {
		return "Event Log: <empty>"
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Event Log (%d entries):\n", len(l.entries)))
	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n")

	for i, entry := range l.entries {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i, entry.Timestamp.Format("15:04:05.000")))
		builder.WriteString(fmt.Sprintf("    Phase: %s\n", entry.Phase))
		builder.WriteString(fmt.Sprintf("    Type: %s\n", entry.EventType))
		if entry.TargetID != "" {
			builder.WriteString(fmt.Sprintf("    Target: %s\n", entry.TargetID))
		}
		if entry.Details != "" {
			builder.WriteString(fmt.Sprintf("    Details: %s\n", entry.Details))
		}
		if entry.Duration > 0 {
			builder.WriteString(fmt.Sprintf("    Duration: %v\n", entry.Duration))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// Visualize 创建事件流的可视化
func (l *EventLogger) Visualize() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 {
		return "Event Flow: <empty>"
	}

	var builder strings.Builder

	builder.WriteString("Event Flow Visualization:\n")
	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n")

	for _, entry := range l.entries {
		// 格式: [TIME] PHASE -> TYPE (TARGET)
		builder.WriteString(fmt.Sprintf("[%s] %s -> %s",
			entry.Timestamp.Format("15:04:05.000"),
			entry.Phase,
			entry.EventType))

		if entry.TargetID != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", entry.TargetID))
		}

		if entry.Duration > 0 {
			builder.WriteString(fmt.Sprintf(" [%v]", entry.Duration))
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// GetStats 获取事件统计信息
func (l *EventLogger) GetStats() map[string]interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats := map[string]interface{}{
		"total": len(l.entries),
	}

	// 按阶段统计
	phaseCounts := make(map[string]int)
	for _, entry := range l.entries {
		phaseCounts[entry.Phase]++
	}
	stats["by_phase"] = phaseCounts

	// 按类型统计
	typeCounts := make(map[string]int)
	for _, entry := range l.entries {
		typeCounts[entry.EventType]++
	}
	stats["by_type"] = typeCounts

	// 按目标统计
	targetCounts := make(map[string]int)
	for _, entry := range l.entries {
		if entry.TargetID != "" {
			targetCounts[entry.TargetID]++
		}
	}
	stats["by_target"] = targetCounts

	// 平均持续时间
	var totalDuration time.Duration
	countWithDuration := 0
	for _, entry := range l.entries {
		if entry.Duration > 0 {
			totalDuration += entry.Duration
			countWithDuration++
		}
	}
	if countWithDuration > 0 {
		stats["avg_duration"] = totalDuration / time.Duration(countWithDuration)
	}

	return stats
}

// SaveToFile 将日志保存到文件
func (l *EventLogger) SaveToFile(filename string) error {
	return os.WriteFile(filename, []byte(l.Dump()), 0644)
}

// Filter 按条件过滤日志
func (l *EventLogger) Filter(filterFunc func(*EventLogEntry) bool) []*EventLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	filtered := make([]*EventLogEntry, 0)
	for _, entry := range l.entries {
		if filterFunc(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetLastN 获取最后 N 条日志
func (l *EventLogger) GetLastN(n int) []*EventLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) <= n {
		// 返回副本
		entries := make([]*EventLogEntry, len(l.entries))
		copy(entries, l.entries)
		return entries
	}

	start := len(l.entries) - n
	entries := make([]*EventLogEntry, n)
	copy(entries, l.entries[start:])
	return entries
}

// GetByPhase 获取特定阶段的日志
func (l *EventLogger) GetByPhase(phase string) []*EventLogEntry {
	return l.Filter(func(entry *EventLogEntry) bool {
		return entry.Phase == phase
	})
}

// GetByTarget 获取特定目标的日志
func (l *EventLogger) GetByTarget(targetID string) []*EventLogEntry {
	return l.Filter(func(entry *EventLogEntry) bool {
		return entry.TargetID == targetID
	})
}

// GetByType 获取特定类型的日志
func (l *EventLogger) GetByType(eventType string) []*EventLogEntry {
	return l.Filter(func(entry *EventLogEntry) bool {
		return entry.EventType == eventType
	})
}

// PrintStats 打印统计信息到 stderr
func (l *EventLogger) PrintStats() {
	stats := l.GetStats()
	log.UILogger.Debug("Event Logger Stats:\n")
	log.UILogger.Debug("  Total Events: %v\n", stats["total"])
	log.UILogger.Debug("  By Phase: %v\n", stats["by_phase"])
	log.UILogger.Debug("  By Type: %v\n", stats["by_type"])
	log.UILogger.Debug("  By Target: %v\n", stats["by_target"])
	if avgDuration, ok := stats["avg_duration"]; ok {
		log.UILogger.Debug("  Avg Duration: %v\n", avgDuration)
	}
}

// 全局事件日志记录器
var globalEventLogger = NewEventLogger()

// GetGlobalLogger 获取全局事件日志记录器
func GetGlobalLogger() *EventLogger {
	return globalEventLogger
}

// LogEvent 是全局日志记录辅助函数
func LogEvent(phase, eventType, targetID, details string) {
	globalEventLogger.Log(phase, eventType, targetID, details)
}

// LogEventWithDuration 是全局日志记录辅助函数（带持续时间）
func LogEventWithDuration(phase, eventType, targetID, details string, duration time.Duration) {
	globalEventLogger.LogWithDuration(phase, eventType, targetID, details, duration)
}
