// Package grid 提供调试辅助工具
package grid

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
)

// =============================================================================
// Debug Mode Control
// =============================================================================

// DebugMode 控制是否启用调试日志
var DebugMode = false

// InitDebugFromEnv 从环境变量初始化调试配置
func InitDebugFromEnv() {
	// 启用 Debug 模式
	if debugStr := os.Getenv("MINT_DEBUG_GRID"); debugStr == "true" || debugStr == "1" {
		DebugMode = true
		println("[Grid Debug] Debug mode enabled")
	}

	// 设置日志级别
	if levelStr := os.Getenv("MINT_LOG_LEVEL"); levelStr != "" {
		SetLogLevel(ParseLogLevel(levelStr))
	}
}

// =============================================================================
// Debug Printf Helpers
// =============================================================================

// DebugPrintf 条件打印（仅在 DebugMode=true 时输出）
func DebugPrintf(format string, args ...interface{}) {
	if DebugMode {
		log.Printf("[Grid-Debug] "+format, args...)
	}
}

// InfoPrintf 打印 INFO 级别日志
func InfoPrintf(format string, args ...interface{}) {
	if currentLogLevel <= LogLevelInfo {
		log.Printf("[Grid-INFO] "+format, args...)
	}
}

// WarnPrintf 打印 WARN 级别日志
func WarnPrintf(format string, args ...interface{}) {
	if currentLogLevel <= LogLevelWarn {
		log.Printf("[Grid-WARN] "+format, args...)
	}
}

// ErrorPrintf 打印 ERROR 级别日志
func ErrorPrintf(format string, args ...interface{}) {
	if currentLogLevel <= LogLevelError {
		log.Printf("[Grid-ERROR] "+format, args...)
	}
}

// =============================================================================
// Formatting Helpers
// =============================================================================

// FormatConstraints 格式化约束
func FormatConstraints(c layout.Constraints) string {
	return fmt.Sprintf("W:[%d..%d] H:[%d..%d]",
		c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

// FormatSize 格式化尺寸
func FormatSize(s layout.Size) string {
	return fmt.Sprintf("%dx%d", s.Width, s.Height)
}

// FormatDimensions 格式化 Grid 的列/行尺寸
func FormatDimensions(colWidths, rowHeights []int, colGap, rowGap int) string {
	cols := fmt.Sprintf("Cols(%d): %v (gap=%d)",
		len(colWidths), colWidths, colGap)
	rows := fmt.Sprintf("Rows(%d): %v (gap=%d)",
		len(rowHeights), rowHeights, rowGap)
	return cols + "\n" + rows
}

// FormatPoint 格式化坐标点
func FormatPoint(x, y int) string {
	return fmt.Sprintf("(%d,%d)", x, y)
}

// FormatRect 格式化矩形区域
func FormatRect(x, y, w, h int) string {
	return fmt.Sprintf("[%d,%d] %dx%d", x, y, w, h)
}

// =============================================================================
// Layout Information Printing
// =============================================================================

// PrintMeasurementInfo 打印测量信息
func PrintMeasurementInfo(gridId string, input, output layout.Constraints, size layout.Size) {
	DebugPrintf("Grid[%s] Measurement:")
	DebugPrintf("  Input:  %s", FormatConstraints(input))
	DebugPrintf("  Output: %s", FormatConstraints(output))
	DebugPrintf("  Size:   %s", FormatSize(size))
}

// PrintLayoutASCII 打印 ASCII 版本的布局视图
func PrintLayoutASCII(gridId string, colWidths, rowHeights []int, colGap, rowGap int) {
	DebugPrintf("=== ASCII Layout View (%s) ===", gridId)

	if len(colWidths) == 0 || len(rowHeights) == 0 {
		DebugPrintf("(empty grid)")
		return
	}

	// 打印列宽度
	colsLine := "Col widths: "
	for i, w := range colWidths {
		if i > 0 {
			colsLine += " "
		}
		colsLine += fmt.Sprintf("[%d:%d]", i, w)
	}
	DebugPrintf(colsLine)

	// 打印行高度
	rowsLine := "Row heights: "
	for i, h := range rowHeights {
		if i > 0 {
			rowsLine += " "
		}
		rowsLine += fmt.Sprintf("[%d:%d]", i, h)
	}
	DebugPrintf(rowsLine)

	// 打印网格 ASCII 表示
	PrintGridASCII(colWidths, rowHeights, colGap, rowGap)
}

// PrintGridASCII 打印简单的网格 ASCII 表示
func PrintGridASCII(colWidths, rowHeights []int, colGap, rowGap int) {
	numCols := len(colWidths)
	numRows := len(rowHeights)

	if numCols == 0 || numRows == 0 {
		return
	}

	// 打印顶边框
	topBorder := "┌"
	for i := 0; i < numCols; i++ {
		for j := 0; j < colWidths[i]; j++ {
			topBorder += "─"
		}
		if i < numCols-1 {
			for g := 0; g < colGap; g++ {
				topBorder += " "
			}
			topBorder += "┬"
		}
	}
	topBorder += "┐"
	DebugPrintf(topBorder)

	// 打印每个格子
	for row := 0; row < numRows; row++ {
		contentLine := "│"
		for col := 0; col < numCols; col++ {
			cellWidth := colWidths[col]
			cellHeight := rowHeights[row]
			cellLabel := fmt.Sprintf("%dx%d", cellWidth, cellHeight)

			// 添加格子标签
			contentLine += cellLabel

			// 填充剩余空格
			for w := len(cellLabel); w < cellWidth; w++ {
				contentLine += " "
			}
			if col < numCols-1 {
				for g := 0; g < colGap; g++ {
					contentLine += " "
				}
				contentLine += "│"
			}
		}
		contentLine += "│"
		DebugPrintf(contentLine)

		// 打印分隔线（如果不是最后一行）
		if row < numRows-1 {
			sepLine := "├"
			for i := 0; i < numCols; i++ {
				for j := 0; j < colWidths[i]; j++ {
					sepLine += "─"
				}
				if i < numCols-1 {
					for g := 0; g < colGap; g++ {
						sepLine += " "
					}
					sepLine += "┼"
				}
			}
			sepLine += "┤"
			DebugPrintf(sepLine)
		}
	}

	// 打印底边框
	bottomBorder := "└"
	for i := 0; i < numCols; i++ {
		for j := 0; j < colWidths[i]; j++ {
			bottomBorder += "─"
		}
		if i < numCols-1 {
			for g := 0; g < colGap; g++ {
				bottomBorder += " "
			}
			bottomBorder += "┴"
		}
	}
	bottomBorder += "┘"
	DebugPrintf(bottomBorder)
}

// PrintCellBordersASCII 打印带边框的网格 ASCII 表示
func PrintCellBordersASCII(gridId string, inst *Instance) {
	if inst == nil || !inst.showCellBorders {
		PrintLayoutASCII(gridId, inst.colWidths, inst.rowHeights, inst.columnGap, inst.rowGap)
		return
	}

	DebugPrintf("=== Cell Borders View (%s) ===", gridId)
	DebugPrintf("Border style: %s", inst.cellBorderStyle)
	DebugPrintf("Rounded: %v", inst.cellBorderRounded)

	// 使用更详细的边框字符样式
	PrintGridBorderASCII(inst.colWidths, inst.rowHeights, inst.columnGap, inst.rowGap, inst.cellBorderStyle)
}

// PrintGridBorderASCII 打印带样式边框的网格
func PrintGridBorderASCII(colWidths, rowHeights []int, colGap, rowGap int, borderStyle string) {
	numCols := len(colWidths)
	numRows := len(rowHeights)

	if numCols == 0 || numRows == 0 {
		return
	}

	// 获取边框字符
	chars := cellBorderChars[borderStyle]
	if chars.horizontal == "" {
		chars = cellBorderChars["single"]
	}

	// 打印顶边框
	topBorder := " " + chars.topLeft
	for i := 0; i < numCols; i++ {
		for j := 0; j < colWidths[i]; j++ {
			topBorder += chars.horizontal
		}
		if i < numCols-1 {
			for g := 0; g < colGap; g++ {
				topBorder += " "
			}
			topBorder += " "
			topBorder += chars.topCross
			topBorder += " "
		}
	}
	topBorder += chars.topRight
	DebugPrintf(topBorder)

	// 打印格子分隔
	for row := 0; row < numRows; row++ {
		// 打印格子内容行
		contentLine := " " + chars.vertical + " "
		for col := 0; col < numCols; col++ {
			cellWidth := colWidths[col]
			label := fmt.Sprintf("%dx%d", cellWidth, rowHeights[row])
			contentLine += label
			for w := len(label); w < cellWidth; w++ {
				contentLine += " "
			}
			if col < numCols-1 {
				for g := 0; g < colGap; g++ {
					contentLine += " "
				}
				contentLine += " "
				contentLine += chars.vertical
				contentLine += " "
			}
		}
		contentLine += " "
		contentLine += chars.vertical
		DebugPrintf(contentLine)

		// 打印行分隔线
		if row < numRows-1 {
			sepLine := " " + chars.leftCross
			for i := 0; i < numCols; i++ {
				for j := 0; j < colWidths[i]; j++ {
					sepLine += chars.horizontal
				}
				if i < numCols-1 {
					for g := 0; g < colGap; g++ {
						sepLine += " "
					}
					sepLine += " "
					sepLine += chars.cross
					sepLine += " "
				}
			}
			sepLine += chars.rightCross
			DebugPrintf(sepLine)
		}
	}

	// 打印底边框
	bottomBorder := " " + chars.bottomLeft
	for i := 0; i < numCols; i++ {
		for j := 0; j < colWidths[i]; j++ {
			bottomBorder += chars.horizontal
		}
		if i < numCols-1 {
			for g := 0; g < colGap; g++ {
				bottomBorder += " "
			}
			bottomBorder += " "
			bottomBorder += chars.bottomCross
			bottomBorder += " "
		}
	}
	bottomBorder += chars.bottomRight
	DebugPrintf(bottomBorder)
}

// =============================================================================
// Constraint Validation
// =============================================================================

// ValidateConstraints 验证约束的有效性
func ValidateConstraints(c layout.Constraints) error {
	if c.MinWidth < 0 || c.MinHeight < 0 {
		return fmt.Errorf("negative minimum size: %d x %d", c.MinWidth, c.MinHeight)
	}
	if c.MaxWidth < c.MinWidth {
		return fmt.Errorf("MaxWidth(%d) < MinWidth(%d)", c.MaxWidth, c.MinWidth)
	}
	if c.MaxHeight < c.MinHeight {
		return fmt.Errorf("MaxHeight(%d) < MinHeight(%d)", c.MaxHeight, c.MinHeight)
	}
	return nil
}

// ValidateSize 验证尺寸是否有效
func ValidateSize(s layout.Size) error {
	if s.Width < 0 || s.Height < 0 {
		return fmt.Errorf("negative size: %d x %d", s.Width, s.Height)
	}
	return nil
}

// ValidateSizeAgainstConstraints 验证尺寸是否符合约束
func ValidateSizeAgainstConstraints(s layout.Size, c layout.Constraints) error {
	if s.Width < c.MinWidth {
		return fmt.Errorf("Width(%d) < MinWidth(%d)", s.Width, c.MinWidth)
	}
	if s.Height < c.MinHeight {
		return fmt.Errorf("Height(%d) < MinHeight(%d)", s.Height, c.MinHeight)
	}
	if s.Width > c.MaxWidth && c.MaxWidth < layout.MaxInt {
		return fmt.Errorf("Width(%d) > MaxWidth(%d)", s.Width, c.MaxWidth)
	}
	if s.Height > c.MaxHeight && c.MaxHeight < layout.MaxInt {
		return fmt.Errorf("Height(%d) > MaxHeight(%d)", s.Height, c.MaxHeight)
	}
	return nil
}

// =============================================================================
// Trace Helpers
// =============================================================================

// TraceMeasurement 追踪测量过程（集成到 layout.Tracer）
func (inst *Instance) TraceMeasurement(
	step string,
	input, output layout.Constraints,
	size layout.Size,
	reason string,
) {
	if !layout.IsTracerEnabled() {
		return
	}

	gridId := fmt.Sprintf("grid-%s", inst.key)
	path := fmt.Sprintf("root/grids/%s/%s", inst.key, step)

	layout.TraceMeasuring(
		"parent", gridId, path,
		input, output, size, reason,
	)
}

// =============================================================================
// Log Level Control
// =============================================================================

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

var currentLogLevel = LogLevelInfo

// SetLogLevel 设置日志级别
func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}

// GetLogLevel 获取当前日志级别
func GetLogLevel() LogLevel {
	return currentLogLevel
}

// ParseLogLevel 解析日志级别字符串
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "ERROR":
		return LogLevelError
	case "WARN", "WARNING":
		return LogLevelWarn
	case "INFO":
		return LogLevelInfo
	case "DEBUG":
		return LogLevelDebug
	case "TRACE":
		return LogLevelTrace
	default:
		return LogLevelInfo
	}
}

// SetLogLevelFromEnv 从环境变量设置日志级别
func SetLogLevelFromEnv() {
	if levelStr := os.Getenv("MINT_LOG_LEVEL"); levelStr != "" {
		SetLogLevel(ParseLogLevel(levelStr))
	}
}

// =============================================================================
// Debug Statistics
// =============================================================================

// DebugStats 调试统计信息
type DebugStats struct {
	MeasureCallCount  int    // Measure 调用次数
	PaintCallCount    int    // Paint 调用次数
	TotalMeasureTime  int64  // 总测量时间 (纳秒)
	MinMeasureTime    int64  // 最小测量时间 (纳秒)
	MaxMeasureTime    int64  // 最大测量时间 (纳秒)
	AvgMeasureTime    int64  // 平均测量时间 (纳秒)
	LastMeasureSize   string // 最后一次测量尺寸
	LastConstraint    string // 最后一次约束
}

var (
	globalStats DebugStats
)

// GetDebugStats 获取调试统计信息
func GetDebugStats() DebugStats {
	stats := globalStats
	// 计算平均时间
	if stats.MeasureCallCount > 0 {
		stats.AvgMeasureTime = stats.TotalMeasureTime / int64(stats.MeasureCallCount)
	}
	return stats
}

// ResetDebugStats 重置调试统计
func ResetDebugStats() {
	globalStats = DebugStats{
		MinMeasureTime: 1<<63 - 1, // 最大 int64
	}
}

// RecordMeasureCall 记录 Measure 调用（由 Instance 内部调用）
func RecordMeasureCall(constraints layout.Constraints, size layout.Size, duration int64) {
	globalStats.MeasureCallCount++
	globalStats.TotalMeasureTime += duration
	globalStats.LastConstraint = FormatConstraints(constraints)
	globalStats.LastMeasureSize = FormatSize(size)

	if duration < globalStats.MinMeasureTime {
		globalStats.MinMeasureTime = duration
	}
	if duration > globalStats.MaxMeasureTime {
		globalStats.MaxMeasureTime = duration
	}
}

// PrintDebugStats 打印调试统计信息
func PrintDebugStats() {
	stats := GetDebugStats()
	if stats.MeasureCallCount == 0 {
		DebugPrintf("DebugStats: No measurements recorded")
		return
	}

	DebugPrintf("=== Grid Debug Statistics ===")
	DebugPrintf("Measure calls: %d", stats.MeasureCallCount)
	DebugPrintf("Total time: %.2f ms", float64(stats.TotalMeasureTime)/1e6)
	DebugPrintf("Avg time: %.2f ms", float64(stats.AvgMeasureTime)/1e6)
	DebugPrintf("Min time: %.2f ms", float64(stats.MinMeasureTime)/1e6)
	DebugPrintf("Max time: %.2f ms", float64(stats.MaxMeasureTime)/1e6)
	DebugPrintf("Last constraint: %s", stats.LastConstraint)
	DebugPrintf("Last size: %s", stats.LastMeasureSize)
}

// =============================================================================
// Utility Functions
// =============================================================================

// PrintTree 打印组件树（递归）
func PrintTree(node *VNode, indent int) {
	indentStr := strings.Repeat("  ", indent)
	fmt.Printf("%sGrid[key=%s, cols=%d, rows=%d]\n",
		indentStr, node.key, len(node.columns), len(node.rows))

	for i := range node.cells {
		if node.cells[i].Child != nil {
			fmt.Printf("%s  Cell[%d,%d]: ", indentStr, node.cells[i].Row, node.cells[i].Col)
			// 这里可以递归打印子组件
			fmt.Printf("%v\n", node.cells[i].Child)
		}
	}
}

// GetEnvVar 获取环境变量整数值
func GetEnvVar(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// IsEnvVarTrue 检查环境变量是否为 true
func IsEnvVarTrue(key string) bool {
	return os.Getenv(key) == "true" || os.Getenv(key) == "1"
}
