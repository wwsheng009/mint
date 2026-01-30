package standalone

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger 日志记录器
//
// Logger 将事件写入日志文件，供独立调试器读取。
// 与嵌入式 DevTools 不同，Logger 只负责写入，不依赖任何运行时状态。
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	enabled  bool
	frame    int
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	LogDir    string // 日志目录
	SessionID string // 会话 ID（可选，默认自动生成）
}

// DefaultConfig 默认配置
func DefaultConfig() *LoggerConfig {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".mint", "devtools", "logs")

	return &LoggerConfig{
		LogDir: logDir,
	}
}

// NewLogger 创建新的日志记录器
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 确保目录存在
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 生成日志文件路径
	sessionID := config.SessionID
	if sessionID == "" {
		sessionID = time.Now().Format("20060102_150405")
	}
	logPath := filepath.Join(config.LogDir, fmt.Sprintf("session_%s.log", sessionID))

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := &Logger{
		file:    file,
		path:    logPath,
		enabled: true,
	}

	// 写入会话开始标记
	logger.Log("session", "system", "meta", map[string]interface{}{
		"event":  "session_start",
		"time":   time.Now().Format(time.RFC3339),
		"file":   logPath,
		"config": config,
	})

	return logger, nil
}

// Enable 启用日志
func (l *Logger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = true
}

// Disable 禁用日志
func (l *Logger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = false
}

// IsEnabled 检查是否启用
func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

// Log 记录事件
func (l *Logger) Log(eventType, targetID, phase string, data map[string]interface{}) error {
	if !l.IsEnabled() {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	event := DebugEvent{
		Timestamp: time.Now(),
		Frame:     l.frame,
		Type:      eventType,
		TargetID:  targetID,
		Phase:     phase,
		Data:      data,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := l.file.Write(append(bytes, '\n')); err != nil {
		return err
	}

	return l.file.Sync()
}

// BeginFrame 标记帧开始
func (l *Logger) BeginFrame() {
	l.mu.Lock()
	l.frame++
	l.mu.Unlock()
}

// EndFrame 标记帧结束
func (l *Logger) EndFrame() {
	// 可以在这里记录帧结束信息
}

// LogComponentAdd 记录组件添加
func (l *Logger) LogComponentAdd(componentID, componentType string, props map[string]interface{}) {
	data := map[string]interface{}{
		"component_type": componentType,
	}
	for k, v := range props {
		data[k] = v
	}
	l.Log("component_add", componentID, "target", data)
}

// LogMouseEvent 记录鼠标事件
func (l *Logger) LogMouseEvent(targetID string, x, y int, eventType, clickType string) {
	l.Log("mouse", targetID, "target", map[string]interface{}{
		"x":     x,
		"y":     y,
		"type":  eventType,
		"click": clickType,
	})
}

// LogFocusEvent 记录焦点事件
func (l *Logger) LogFocusEvent(targetID string, focused bool) {
	l.Log("focus", targetID, "target", map[string]interface{}{
		"focused": focused,
	})
}

// LogKeyEvent 记录键盘事件
func (l *Logger) LogKeyEvent(targetID string, key rune, mod string) {
	l.Log("key", targetID, "target", map[string]interface{}{
		"key":     string(key),
		"mod":     mod,
	})
}

// LogCustom 记录自定义事件
func (l *Logger) LogCustom(eventType, targetID string, data map[string]interface{}) {
	l.Log(eventType, targetID, "custom", data)
}

// LogError 记录错误
func (l *Logger) LogError(componentID, message string, err error) {
	data := map[string]interface{}{
		"message": message,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	l.Log("error", componentID, "error", data)
}

// LogMessage 记录普通消息
func (l *Logger) LogMessage(message string) {
	l.Log("log", "system", "info", map[string]interface{}{
		"message": message,
	})
}

// GetPath 获取日志文件路径
func (l *Logger) GetPath() string {
	return l.path
}

// Flush 刷新日志到磁盘
func (l *Logger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Sync()
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		// 写入会话结束标记
		event := DebugEvent{
			Timestamp: time.Now(),
			Type:      "session",
			TargetID:  "system",
			Phase:     "meta",
			Data: map[string]interface{}{
				"event": "session_end",
				"time":  time.Now().Format(time.RFC3339),
			},
		}
		bytes, _ := json.Marshal(event)
		l.file.Write(append(bytes, '\n'))
		l.file.Sync()

		err := l.file.Close()
		l.file = nil
		return err
	}

	return nil
}

// =============================================================================
// 便捷函数 - 全局单例
// =============================================================================

var (
	globalLogger *Logger
	globalOnce   sync.Once
)

// InitGlobalLogger 初始化全局日志记录器
func InitGlobalLogger(config *LoggerConfig) error {
	var err error
	globalOnce.Do(func() {
		globalLogger, err = NewLogger(config)
	})
	return err
}

// GetGlobalLogger 获取全局日志记录器
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger, _ = NewLogger(DefaultConfig())
	}
	return globalLogger
}

// Log 便捷函数 - 记录事件
func Log(eventType, targetID, phase string, data map[string]interface{}) error {
	logger := GetGlobalLogger()
	return logger.Log(eventType, targetID, phase, data)
}

// LogMessage 便捷函数 - 记录消息
func LogMessage(message string) {
	logger := GetGlobalLogger()
	logger.LogMessage(message)
}

// LogError 便捷函数 - 记录错误
func LogError(componentID, message string, err error) {
	logger := GetGlobalLogger()
	logger.LogError(componentID, message, err)
}
