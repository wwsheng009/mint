// Package standalone 提供独立运行的 DevTools 调试器
//
// 与嵌入式 DevTools 不同，standalone 调试器是一个独立的程序，
// 可以独立运行，不依赖于被调试程序的状态。
package standalone

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DebugEvent 调试事件
type DebugEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Frame     int                    `json:"frame"`
	Type      string                 `json:"type"`
	TargetID  string                 `json:"target_id"`
	Phase     string                 `json:"phase"`
	Data      map[string]interface{} `json:"data"`
}

// LogFileReader 日志文件读取器
type LogFileReader struct {
	mu      sync.RWMutex
	file    *os.File
	scanner *bufio.Scanner
	events  []*DebugEvent
	path    string
}

// NewLogFileReader 创建日志文件读取器
func NewLogFileReader(path string) (*LogFileReader, error) {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	r := &LogFileReader{
		events: make([]*DebugEvent, 0),
		path:   path,
	}

	// 如果文件存在，读取已有事件
	if _, err := os.Stat(path); err == nil {
		if err := r.loadExisting(); err != nil {
			return nil, fmt.Errorf("failed to load existing events: %w", err)
		}
	}

	return r, nil
}

// loadExisting 加载已有事件
func (r *LogFileReader) loadExisting() error {
	file, err := os.Open(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event DebugEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			r.mu.Lock()
			r.events = append(r.events, &event)
			r.mu.Unlock()
		}
	}

	return scanner.Err()
}

// Watch 监控日志文件变化
func (r *LogFileReader) Watch() error {
	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	r.file = file
	r.scanner = bufio.NewScanner(file)

	// 跳到文件末尾
	if _, err := file.Seek(0, 2); err != nil {
		return err
	}

	return nil
}

// Poll 轮询新事件
func (r *LogFileReader) Poll() []*DebugEvent {
	r.mu.Lock()
	defer r.mu.Unlock()

	newEvents := make([]*DebugEvent, 0)

	if r.scanner == nil {
		return newEvents
	}

	for r.scanner.Scan() {
		line := r.scanner.Text()
		if line == "" {
			continue
		}

		var event DebugEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			r.events = append(r.events, &event)
			newEvents = append(newEvents, &event)
		}
	}

	return newEvents
}

// GetAllEvents 获取所有事件
func (r *LogFileReader) GetAllEvents() []*DebugEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 返回副本
	events := make([]*DebugEvent, len(r.events))
	copy(events, r.events)
	return events
}

// GetEventsByType 按类型获取事件
func (r *LogFileReader) GetEventsByType(eventType string) []*DebugEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*DebugEvent, 0)
	for _, e := range r.events {
		if e.Type == eventType {
			events = append(events, e)
		}
	}
	return events
}

// GetEventsByTarget 按目标获取事件
func (r *LogFileReader) GetEventsByTarget(targetID string) []*DebugEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*DebugEvent, 0)
	for _, e := range r.events {
		if e.TargetID == targetID {
			events = append(events, e)
		}
	}
	return events
}

// GetEventsInTimeRange 获取时间范围内的事件
func (r *LogFileReader) GetEventsInTimeRange(start, end time.Time) []*DebugEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*DebugEvent, 0)
	for _, e := range r.events {
		if (e.Timestamp.Equal(start) || e.Timestamp.After(start)) &&
			(e.Timestamp.Equal(end) || e.Timestamp.Before(end)) {
			events = append(events, e)
		}
	}
	return events
}

// Close 关闭读取器
func (r *LogFileReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Clear 清空日志文件
func (r *LogFileReader) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = make([]*DebugEvent, 0)

	if err := os.Remove(r.path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// =============================================================================
// StandaloneDebugger 独立调试器
// =============================================================================

// StandaloneDebugger 独立调试器
type StandaloneDebugger struct {
	mu          sync.RWMutex
	logReader   *LogFileReader
	analysis    *AnalysisResult
	isWatching  bool
	stopWatch   chan struct{}
	logFilePath string
}

// AnalysisEvent 分析事件
type AnalysisEvent struct {
	Type    string `json:"type"`
	Count   int    `json:"count"`
	First   time.Time `json:"first"`
	Last    time.Time `json:"last"`
}

// TargetAnalysis 目标分析
type TargetAnalysis struct {
	TargetID   string                   `json:"target_id"`
	EventTypes map[string]int           `json:"event_types"`
	FirstEvent time.Time                `json:"first_event"`
	LastEvent  time.Time                `json:"last_event"`
	Data       map[string]interface{}   `json:"data"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	TotalEvents     int                       `json:"total_events"`
	EventTypes      map[string]int            `json:"event_types"`
	Targets         map[string]*TargetAnalysis `json:"targets"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         time.Time                 `json:"end_time"`
	Duration        time.Duration             `json:"duration"`
	EventsPerSecond float64                   `json:"events_per_second"`
}

// NewStandaloneDebugger 创建独立调试器
func NewStandaloneDebugger(logPath string) (*StandaloneDebugger, error) {
	logReader, err := NewLogFileReader(logPath)
	if err != nil {
		return nil, err
	}

	return &StandaloneDebugger{
		logReader:   logReader,
		stopWatch:   make(chan struct{}),
		logFilePath: logPath,
	}, nil
}

// Analyze 分析所有事件
func (d *StandaloneDebugger) Analyze() *AnalysisResult {
	events := d.logReader.GetAllEvents()

	if len(events) == 0 {
		return &AnalysisResult{
			EventTypes: make(map[string]int),
			Targets:    make(map[string]*TargetAnalysis),
		}
	}

	result := &AnalysisResult{
		TotalEvents: len(events),
		EventTypes:  make(map[string]int),
		Targets:     make(map[string]*TargetAnalysis),
		StartTime:   events[0].Timestamp,
		EndTime:     events[len(events)-1].Timestamp,
	}

	result.Duration = result.EndTime.Sub(result.StartTime)
	if result.Duration.Seconds() > 0 {
		result.EventsPerSecond = float64(result.TotalEvents) / result.Duration.Seconds()
	}

	for _, event := range events {
		// 统计事件类型
		result.EventTypes[event.Type]++

		// 统计目标
		if event.TargetID != "" {
			target, exists := result.Targets[event.TargetID]
			if !exists {
				target = &TargetAnalysis{
					TargetID:   event.TargetID,
					EventTypes: make(map[string]int),
					Data:       make(map[string]interface{}),
				}
				result.Targets[event.TargetID] = target
			}

			target.EventTypes[event.Type]++
			if target.FirstEvent.IsZero() || event.Timestamp.Before(target.FirstEvent) {
				target.FirstEvent = event.Timestamp
			}
			if target.LastEvent.IsZero() || event.Timestamp.After(target.LastEvent) {
				target.LastEvent = event.Timestamp
			}

			// 合并数据
			for k, v := range event.Data {
				target.Data[k] = v
			}
		}
	}

	d.mu.Lock()
	d.analysis = result
	d.mu.Unlock()

	return result
}

// GetAnalysis 获取分析结果
func (d *StandaloneDebugger) GetAnalysis() *AnalysisResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.analysis == nil {
		return d.Analyze()
	}
	return d.analysis
}

// WatchEvents 开始监控事件
func (d *StandaloneDebugger) WatchEvents(callback func([]*DebugEvent)) error {
	if err := d.logReader.Watch(); err != nil {
		return err
	}

	d.isWatching = true

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				events := d.logReader.Poll()
				if len(events) > 0 && callback != nil {
					callback(events)
				}
			case <-d.stopWatch:
				return
			}
		}
	}()

	return nil
}

// StopWatching 停止监控
func (d *StandaloneDebugger) StopWatching() {
	d.isWatching = false
	close(d.stopWatch)
}

// GetTimeline 获取时间线
func (d *StandaloneDebugger) GetTimeline() []*DebugEvent {
	return d.logReader.GetAllEvents()
}

// GetEventFlow 获取事件流（用于检测问题）
func (d *StandaloneDebugger) GetEventFlow() []string {
	events := d.logReader.GetAllEvents()

	flow := make([]string, 0, len(events))

	for i, e := range events {
		timestamp := e.Timestamp.Format("15:04:05.000")
		flow = append(flow, fmt.Sprintf("[%d] %s %-20s → %s", i, timestamp, e.Type, e.TargetID))
	}

	return flow
}

// DetectProblems 检测潜在问题
func (d *StandaloneDebugger) DetectProblems() []string {
	events := d.logReader.GetAllEvents()
	problems := make([]string, 0)

	if len(events) == 0 {
		return []string{"⚠️  没有检测到任何事件 - 程序可能未启动或 DevTools 未启用"}
	}

	// 检查事件类型平衡
	typeCounts := make(map[string]int)
	for _, e := range events {
		typeCounts[e.Type]++
	}

	// 检查焦点事件
	if focusEvents := typeCounts["focus"]; focusEvents == 0 {
		problems = append(problems, "⚠️  没有焦点事件 - 组件可能未实现 Focusable 接口")
	}

	// 检查鼠标事件
	if mouseEvents := typeCounts["mouse"]; mouseEvents == 0 {
		problems = append(problems, "⚠️  没有鼠标事件 - 命中测试可能失败")
	}

	// 检查组件添加
	if addEvents := typeCounts["component_add"]; addEvents == 0 {
		problems = append(problems, "⚠️  没有组件添加事件 - 组件可能未正确注册")
	}

	// 检查是否有鼠标点击但无回调
	mouseClicks := 0
	clickCallbacks := 0
	for _, e := range events {
		if e.Type == "mouse" && e.Data["type"] == "press" {
			mouseClicks++
		}
		if e.Type == "log" && strings.Contains(fmt.Sprint(e.Data["message"]), "CLICKED") {
			clickCallbacks++
		}
	}

	if mouseClicks > 0 && clickCallbacks == 0 {
		problems = append(problems, "⚠️  检测到鼠标点击但无回调 - onClick 可能未设置")
	}

	// 检查事件频率
	analysis := d.Analyze()
	if analysis.EventsPerSecond > 1000 {
		problems = append(problems, fmt.Sprintf("⚠️  事件频率过高 (%.1f 事件/秒) - 可能存在事件循环", analysis.EventsPerSecond))
	}

	// 检查是否有连续的相同事件
	for i := 1; i < len(events); i++ {
		if events[i].Type == events[i-1].Type &&
			events[i].TargetID == events[i-1].TargetID &&
			events[i].Timestamp.Sub(events[i-1].Timestamp) < 10*time.Millisecond {
			problems = append(problems, fmt.Sprintf("⚠️  检测到重复事件: %s → %s 在 %s 内重复",
				events[i].Type, events[i].TargetID, events[i].Timestamp.Sub(events[i-1].Timestamp)))
			break
		}
	}

	if len(problems) == 0 {
		problems = append(problems, "✅ 未检测到明显问题")
	}

	return problems
}

// GetStatistics 获取统计信息
func (d *StandaloneDebugger) GetStatistics() map[string]interface{} {
	analysis := d.Analyze()

	stats := map[string]interface{}{
		"total_events":      analysis.TotalEvents,
		"duration":          analysis.Duration.String(),
		"events_per_second": analysis.EventsPerSecond,
		"event_types":       analysis.EventTypes,
		"target_count":      len(analysis.Targets),
		"start_time":        analysis.StartTime.Format("2006-01-02 15:04:05"),
		"end_time":          analysis.EndTime.Format("2006-01-02 15:04:05"),
	}

	return stats
}

// ExportReport 导出报告
func (d *StandaloneDebugger) ExportReport() string {
	analysis := d.Analyze()
	problems := d.DetectProblems()

	var sb strings.Builder

	sb.WriteString("╔═══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                        DevTools 调试报告                            ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════════════╝\n\n")

	// 基本统计
	sb.WriteString("📊 基本统计\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  总事件数:     %d\n", analysis.TotalEvents))
	sb.WriteString(fmt.Sprintf("  运行时长:     %s\n", analysis.Duration))
	sb.WriteString(fmt.Sprintf("  事件频率:     %.2f 事件/秒\n", analysis.EventsPerSecond))
	sb.WriteString(fmt.Sprintf("  目标组件数:   %d\n", len(analysis.Targets)))
	sb.WriteString(fmt.Sprintf("  开始时间:     %s\n", analysis.StartTime.Format("15:04:05")))
	sb.WriteString(fmt.Sprintf("  结束时间:     %s\n\n", analysis.EndTime.Format("15:04:05")))

	// 事件类型统计
	sb.WriteString("📋 事件类型分布\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")

	// 按数量排序
	typePairs := make([][2]interface{}, 0, len(analysis.EventTypes))
	for k, v := range analysis.EventTypes {
		typePairs = append(typePairs, [2]interface{}{k, v})
	}
	sort.Slice(typePairs, func(i, j int) bool {
		return typePairs[i][1].(int) > typePairs[j][1].(int)
	})

	for _, pair := range typePairs {
		eventType := pair[0].(string)
		count := pair[1].(int)
		percentage := float64(count) / float64(analysis.TotalEvents) * 100
		sb.WriteString(fmt.Sprintf("  %-20s %6d  (%5.1f%%)\n", eventType, count, percentage))
	}
	sb.WriteString("\n")

	// 目标组件
	sb.WriteString("🎯 目标组件\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")

	targetList := make([]*TargetAnalysis, 0, len(analysis.Targets))
	for _, t := range analysis.Targets {
		targetList = append(targetList, t)
	}
	sort.Slice(targetList, func(i, j int) bool {
		return len(targetList[i].EventTypes) > len(targetList[j].EventTypes)
	})

	for _, target := range targetList {
		sb.WriteString(fmt.Sprintf("  %s\n", target.TargetID))
		eventTypeList := make([]string, 0, len(target.EventTypes))
		for et := range target.EventTypes {
			eventTypeList = append(eventTypeList, et)
		}
		sort.Strings(eventTypeList)
		for _, et := range eventTypeList {
			sb.WriteString(fmt.Sprintf("    - %s: %d\n", et, target.EventTypes[et]))
		}
	}
	sb.WriteString("\n")

	// 问题检测
	sb.WriteString("🔍 问题检测\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")
	for _, p := range problems {
		sb.WriteString(fmt.Sprintf("  %s\n", p))
	}
	sb.WriteString("\n")

	// 事件流
	sb.WriteString("📈 事件流（最近 20 条）\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")

	flow := d.GetEventFlow()
	start := len(flow) - 20
	if start < 0 {
		start = 0
	}
	for i := start; i < len(flow); i++ {
		sb.WriteString(fmt.Sprintf("  %s\n", flow[i]))
	}

	return sb.String()
}

// Clear 清空所有数据
func (d *StandaloneDebugger) Clear() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.analysis = nil
	return d.logReader.Clear()
}

// Close 关闭调试器
func (d *StandaloneDebugger) Close() error {
	d.StopWatching()
	return d.logReader.Close()
}
