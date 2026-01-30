// standalone-debugger 是一个独立的 DevTools 调试器
//
// 使用方法:
//
//	# 启动调试器，监控日志文件
//	go run github.com/wwsheng009/mint/devtools/standalone@latest
//
//	# 指定日志文件路径
//	go run github.com/wwsheng009/mint/devtools/standalone@latest -log <path>
//
//	# 只分析不监控
//	go run github.com/wwsheng009/mint/devtools/standalone@latest -analyze-only
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	logPath      string
	analyzeOnly  bool
	watchMode    bool
	exportReport bool
	showHelp     bool
)

func init() {
	// 设置默认日志路径
	homeDir, _ := os.UserHomeDir()
	defaultLogDir := filepath.Join(homeDir, ".mint", "devtools", "logs")

	// 查找最新的日志文件
	defaultLogPath := findLatestLog(defaultLogDir)

	flag.StringVar(&logPath, "log", defaultLogPath, "日志文件路径")
	flag.BoolVar(&analyzeOnly, "analyze-only", false, "只分析不监控")
	flag.BoolVar(&watchMode, "watch", true, "监控模式（自动）")
	flag.BoolVar(&exportReport, "report", false, "导出完整报告")
	flag.BoolVar(&showHelp, "help", false, "显示帮助")
}

func main() {
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	// 导入独立调试器
	d := &Debugger{
		logPath: logPath,
	}

	if exportReport {
		d.ExportFullReport()
		return
	}

	// 启动调试器
	d.Start()
}

// Debugger 调试器
type Debugger struct {
	logPath      string
	lastEventCount int
}

// Start 启动调试器
func (d *Debugger) Start() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Mint DevTools 独立调试器                        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 检查日志文件
	if d.logPath == "" {
		fmt.Println("❌ 错误: 未找到日志文件")
		fmt.Println()
		fmt.Println("请使用 -log 参数指定日志文件路径，或确保被调试程序已启动并写入日志。")
		fmt.Println()
		fmt.Println("日志文件位置示例:")
		homeDir, _ := os.UserHomeDir()
		fmt.Printf("  %s/.mint/devtools/logs/session_20240130_123456.log\n", homeDir)
		return
	}

	fmt.Printf("📂 日志文件: %s\n", d.logPath)
	fmt.Println()

	// 分析已有事件
	d.Analyze()

	if analyzeOnly {
		return
	}

	// 监控模式
	fmt.Println("👀 监控模式: 已启动")
	fmt.Println("   等待新事件... (按 Ctrl+C 退出)")
	fmt.Println()

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 监控循环
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.CheckNewEvents()
		case <-sigCh:
			fmt.Println("\n\n👋 收到退出信号，正在关闭...")
			d.PrintSummary()
			return
		}
	}
}

// Analyze 分析日志
func (d *Debugger) Analyze() {
	events, err := d.readEvents()
	if err != nil {
		fmt.Printf("❌ 读取日志失败: %v\n", err)
		return
	}

	d.lastEventCount = len(events)

	if len(events) == 0 {
		fmt.Println("⚠️  日志文件为空，等待事件写入...")
		fmt.Println()
		return
	}

	fmt.Println("📊 事件统计")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	// 统计事件
	typeCounts := make(map[string]int)
	targetCounts := make(map[string]int)
	firstEvent := events[0].Timestamp
	lastEvent := events[len(events)-1].Timestamp

	for _, e := range events {
		typeCounts[e.Type]++
		if e.TargetID != "" {
			targetCounts[e.TargetID]++
		}
	}

	duration := lastEvent.Sub(firstEvent)
	eventsPerSec := float64(len(events)) / duration.Seconds()

	fmt.Printf("  总事件数:     %d\n", len(events))
	fmt.Printf("  事件类型:     %d\n", len(typeCounts))
	fmt.Printf("  目标组件数:   %d\n", len(targetCounts))
	fmt.Printf("  运行时长:     %s\n", duration.Round(time.Millisecond))
	fmt.Printf("  事件频率:     %.2f 事件/秒\n", eventsPerSec)
	fmt.Println()

	fmt.Println("📋 事件类型分布")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	for t, count := range typeCounts {
		pct := float64(count) / float64(len(events)) * 100
		fmt.Printf("  %-20s %6d  (%5.1f%%)\n", t, count, pct)
	}
	fmt.Println()

	if len(targetCounts) > 0 {
		fmt.Println("🎯 目标组件")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for target, count := range targetCounts {
			fmt.Printf("  %-30s %6d 事件\n", target, count)
		}
		fmt.Println()
	}

	// 问题检测
	d.DetectProblems(events)

	// 显示最近事件
	d.ShowRecentEvents(events, 10)
}

// CheckNewEvents 检查新事件
func (d *Debugger) CheckNewEvents() {
	events, err := d.readEvents()
	if err != nil {
		return
	}

	if len(events) > d.lastEventCount {
		newEvents := events[d.lastEventCount:]
		d.lastEventCount = len(events)

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("\n[%s] 📨 收到 %d 个新事件\n", timestamp, len(newEvents))

		// 显示新事件
		for _, e := range newEvents {
			ts := e.Timestamp.Format("15:04:05.000")
			fmt.Printf("  %s %-20s → %s", ts, e.Type, e.TargetID)

			// 显示关键数据
			if msg, ok := e.Data["message"].(string); ok && len(msg) < 50 {
				fmt.Printf(" [%s]", msg)
			} else if e.Type == "mouse" {
				if x, ok := e.Data["x"].(float64); ok {
					if y, ok := e.Data["y"].(float64); ok {
						fmt.Printf(" [位置: (%.0f, %.0f)]", x, y)
					}
				}
			} else if e.Type == "focus" {
				if focused, ok := e.Data["focused"].(bool); ok {
					if focused {
						fmt.Print(" [获得焦点]")
					} else {
						fmt.Print(" [失去焦点]")
					}
				}
			}

			fmt.Println()
		}

		// 检测问题
		fmt.Println()
		d.DetectProblems(events)
	}
}

// DetectProblems 检测问题
func (d *Debugger) DetectProblems(events []DebugEvent) {
	problems := make([]string, 0)

	typeCounts := make(map[string]int)
	for _, e := range events {
		typeCounts[e.Type]++
	}

	// 检查各种问题
	if typeCounts["component_add"] == 0 {
		problems = append(problems, "⚠️  没有组件添加事件 - 组件可能未正确注册")
	}

	if typeCounts["focus"] == 0 && len(events) > 10 {
		problems = append(problems, "⚠️  没有焦点事件 - 组件可能未实现 Focusable 接口")
	}

	if typeCounts["mouse"] == 0 && len(events) > 10 {
		problems = append(problems, "⚠️  没有鼠标事件 - 命中测试可能失败")
	}

	// 检查鼠标点击但无回调
	clicks := 0
	callbacks := 0
	for _, e := range events {
		if e.Type == "mouse" && e.Data["type"] == "press" {
			clicks++
		}
		if e.Type == "log" {
			if msg, ok := e.Data["message"].(string); ok {
				if strings.Contains(msg, "CLICKED") {
					callbacks++
				}
			}
		}
	}
	if clicks > 0 && callbacks == 0 {
		problems = append(problems, "⚠️  检测到鼠标点击但无回调 - onClick 可能未设置")
	}

	if len(problems) == 0 {
		problems = append(problems, "✅ 未检测到明显问题")
	}

	fmt.Println("🔍 问题检测")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	for _, p := range problems {
		fmt.Printf("  %s\n", p)
	}
	fmt.Println()
}

// ShowRecentEvents 显示最近事件
func (d *Debugger) ShowRecentEvents(events []DebugEvent, count int) {
	fmt.Printf("📈 最近事件 (显示 %d 条)\n", count)
	fmt.Println("─────────────────────────────────────────────────────────────────")

	start := len(events) - count
	if start < 0 {
		start = 0
	}

	for i := start; i < len(events); i++ {
		e := events[i]
		ts := e.Timestamp.Format("15:04:05.000")
		fmt.Printf("  %s %-20s → %s\n", ts, e.Type, e.TargetID)
	}
	fmt.Println()
}

// PrintSummary 打印总结
func (d *Debugger) PrintSummary() {
	events, _ := d.readEvents()

	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("                              总结                                  ")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Printf("总事件数: %d\n", len(events))
	fmt.Println("感谢使用 Mint DevTools!")
}

// ExportFullReport 导出完整报告
func (d *Debugger) ExportFullReport() {
	events, err := d.readEvents()
	if err != nil {
		fmt.Printf("❌ 读取日志失败: %v\n", err)
		return
	}

	// 生成报告
	fmt.Println(d.GenerateReport(events))
}

// GenerateReport 生成报告
func (d *Debugger) GenerateReport(events []DebugEvent) string {
	var sb strings.Builder

	sb.WriteString("╔═══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    Mint DevTools 调试报告                          ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════════════╝\n\n")

	if len(events) == 0 {
		sb.WriteString("⚠️  没有检测到任何事件\n")
		sb.WriteString("\n可能的原因:\n")
		sb.WriteString("  1. 日志文件为空\n")
		sb.WriteString("  2. 被调试程序未启动\n")
		sb.WriteString("  3. DevTools Logger 未正确初始化\n")
		return sb.String()
	}

	// 统计
	typeCounts := make(map[string]int)
	targetCounts := make(map[string]int)
	targetEvents := make(map[string][]DebugEvent)
	firstEvent := events[0].Timestamp
	lastEvent := events[len(events)-1].Timestamp

	for _, e := range events {
		typeCounts[e.Type]++
		if e.TargetID != "" {
			targetCounts[e.TargetID]++
			targetEvents[e.TargetID] = append(targetEvents[e.TargetID], e)
		}
	}

	duration := lastEvent.Sub(firstEvent)
	eventsPerSec := float64(len(events)) / duration.Seconds()

	// 基本统计
	sb.WriteString("📊 基本统计\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  总事件数:     %d\n", len(events)))
	sb.WriteString(fmt.Sprintf("  运行时长:     %s\n", duration))
	sb.WriteString(fmt.Sprintf("  事件频率:     %.2f 事件/秒\n", eventsPerSec))
	sb.WriteString(fmt.Sprintf("  开始时间:     %s\n", firstEvent.Format("15:04:05")))
	sb.WriteString(fmt.Sprintf("  结束时间:     %s\n\n", lastEvent.Format("15:04:05")))

	// 事件类型
	sb.WriteString("📋 事件类型分布\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")
	for t, count := range typeCounts {
		pct := float64(count) / float64(len(events)) * 100
		sb.WriteString(fmt.Sprintf("  %-20s %6d  (%5.1f%%)\n", t, count, pct))
	}
	sb.WriteString("\n")

	// 目标组件详情
	sb.WriteString("🎯 目标组件详情\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")
	for target, count := range targetCounts {
		sb.WriteString(fmt.Sprintf("\n  %s (%d 事件)\n", target, count))
		for _, e := range targetEvents[target] {
			ts := e.Timestamp.Format("15:04:05.000")
			sb.WriteString(fmt.Sprintf("    %s %-20s", ts, e.Type))
			if msg, ok := e.Data["message"].(string); ok {
				sb.WriteString(fmt.Sprintf(" [%s]", truncate(msg, 40)))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// 完整事件流
	sb.WriteString("📈 完整事件流\n")
	sb.WriteString("─────────────────────────────────────────────────────────────────\n")
	for i, e := range events {
		ts := e.Timestamp.Format("15:04:05.000")
		sb.WriteString(fmt.Sprintf("[%4d] %s %-20s → %s\n", i, ts, e.Type, e.TargetID))
	}

	return sb.String()
}

// readEvents 读取事件
func (d *Debugger) readEvents() ([]DebugEvent, error) {
	if d.logPath == "" {
		return []DebugEvent{}, nil
	}

	file, err := os.Open(d.logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	events := make([]DebugEvent, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event DebugEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			events = append(events, event)
		}
	}

	return events, scanner.Err()
}

// DebugEvent 调试事件（本地版本，用于 standalone）
type DebugEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Frame     int                    `json:"frame"`
	Type      string                 `json:"type"`
	TargetID  string                 `json:"target_id"`
	Phase     string                 `json:"phase"`
	Data      map[string]interface{} `json:"data"`
}

// findLatestLog 查找最新日志文件
func findLatestLog(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var latest os.FileInfo
	var latestName string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasPrefix(e.Name(), "session_") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if latest == nil || info.ModTime().After(latest.ModTime()) {
			latest = info
			latestName = e.Name()
		}
	}

	if latestName != "" {
		return filepath.Join(dir, latestName)
	}
	return ""
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// printHelp 打印帮助
func printHelp() {
	fmt.Println("Mint DevTools 独立调试器")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  standalone-debugger [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -log <path>      日志文件路径 (默认: 自动查找最新)")
	fmt.Println("  -analyze-only    只分析不监控")
	fmt.Println("  -report          导出完整报告")
	fmt.Println("  -watch           监控模式 (默认启用)")
	fmt.Println("  -help            显示此帮助")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 分析最新日志")
	fmt.Println("  standalone-debugger")
	fmt.Println()
	fmt.Println("  # 指定日志文件")
	fmt.Println("  standalone-debugger -log /path/to/session.log")
	fmt.Println()
	fmt.Println("  # 导出报告")
	fmt.Println("  standalone-debugger -log /path/to/session.log -report")
	fmt.Println()
	fmt.Println("在被调试程序中使用 Logger:")
	fmt.Println(`  import "github.com/wwsheng009/mint/devtools/standalone"`)
	fmt.Println(`  logger := standalone.NewLogger(nil)`)
	fmt.Println(`  logger.LogComponentAdd("btn1", "Button", map[string]interface{}{"text": "Click"})`)
	fmt.Println(`  logger.LogMouseEvent("btn1", 10, 5, "press", "left")`)
}
