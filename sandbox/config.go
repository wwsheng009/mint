// sandbox/config.go - 配置系统
package sandbox

import "time"

// Config 沙箱配置
type Config struct {
	// 基础配置
	Width  int
	Height int
	Title  string
	FPS    int

	// 事件配置
	Event EventConfig

	// 快照配置
	Snapshot SnapshotConfig

	// 性能配置
	Performance PerformanceConfig
}

// EventConfig 事件配置
type EventConfig struct {
	QueueMaxSize   int              // 最大队列长度 (默认 10000)
	QueueMaxMemory int64            // 最大内存占用 (默认 100MB)
	EvictPolicy    EvictPolicy      // 淘汰策略
	Strategy       InjectionStrategy // 注入策略
	RecordEnabled  bool             // 是否启用录制
	RecordMaxLen   int              // 录制最大长度
}

// SnapshotConfig 快照配置
type SnapshotConfig struct {
	AutoSnapshot bool           // 自动快照
	Interval     time.Duration  // 快照间隔
	MaxCount     int            // 最大快照数
	Level        SnapshotLevel  // 默认快照级别
	PersistPath  string         // 持久化路径
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	Throttle      bool          // 节流
	MaxFPS        int           // 最大帧率
	RenderTimeout time.Duration // 渲染超时
	Profile       bool          // 性能分析
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Width:  80,
		Height: 24,
		FPS:    60,

		Event: EventConfig{
			QueueMaxSize:   10000,
			QueueMaxMemory: 100 * 1024 * 1024, // 100MB
			EvictPolicy:    EvictOldest,
			Strategy:       InjectAllowed,
			RecordEnabled:  false,
			RecordMaxLen:   10000,
		},

		Snapshot: SnapshotConfig{
			AutoSnapshot: false,
			MaxCount:     100,
			Level:        SnapshotStandard,
		},

		Performance: PerformanceConfig{
			Throttle:      true,
			MaxFPS:        60,
			RenderTimeout: 100 * time.Millisecond,
			Profile:       false,
		},
	}
}

// RealConfig 真实环境配置
func RealConfig() *Config {
	cfg := DefaultConfig()
	cfg.Event.Strategy = InjectProhibited
	cfg.Event.RecordEnabled = true
	return cfg
}

// MockConfig 模拟环境配置
func MockConfig() *Config {
	cfg := DefaultConfig()
	cfg.Event.Strategy = InjectAllowed
	cfg.Performance.Throttle = false // 测试时不节流
	return cfg
}

// ReplayConfig 回放环境配置
func ReplayConfig() *Config {
	cfg := DefaultConfig()
	cfg.Event.Strategy = InjectRecorded
	return cfg
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Width <= 0 {
		c.Width = 80
	}
	if c.Height <= 0 {
		c.Height = 24
	}
	if c.FPS <= 0 {
		c.FPS = 60
	}
	if c.Event.QueueMaxSize <= 0 {
		c.Event.QueueMaxSize = 10000
	}
	return nil
}

// Clone 克隆配置
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}
