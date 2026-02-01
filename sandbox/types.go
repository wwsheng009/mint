// sandbox/types.go - 核心类型定义
package sandbox

import (
	"time"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
)

// SandboxType 沙箱类型
type SandboxType int

const (
	TypeReal   SandboxType = iota // 真实终端环境
	TypeMock                      // 模拟测试环境
	TypeReplay                    // 回放环境
)

// String 返回沙箱类型的字符串表示
func (t SandboxType) String() string {
	switch t {
	case TypeReal:
		return "real"
	case TypeMock:
		return "mock"
	case TypeReplay:
		return "replay"
	default:
		return "unknown"
	}
}

// State 沙箱状态
type State int

const (
	StateStopped State = iota
	StateInitialized
	StateRunning
	StatePaused
	StateError
)

// String 返回状态的字符串表示
func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateInitialized:
		return "initialized"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// Phase 生命周期阶段
type Phase int

const (
	PhaseBefore Phase = iota // 状态转换前
	PhaseAfter               // 状态转换后
)

// HookKey 钩子键
type HookKey struct {
	State State
	Phase Phase
}

// InjectionStrategy 事件注入策略
type InjectionStrategy int

const (
	InjectProhibited InjectionStrategy = iota // 禁止注入 (真实环境)
	InjectAllowed                              // 允许注入 (测试环境)
	InjectRecorded                             // 仅录制 (录制模式)
)

// String 返回注入策略的字符串表示
func (s InjectionStrategy) String() string {
	switch s {
	case InjectProhibited:
		return "prohibited"
	case InjectAllowed:
		return "allowed"
	case InjectRecorded:
		return "recorded"
	default:
		return "unknown"
	}
}

// EvictPolicy 事件淘汰策略
type EvictPolicy int

const (
	EvictOldest     EvictPolicy = iota // 淘汰最旧的
	EvictByPriority                    // 按优先级淘汰
	EvictPersist                       // 持久化到磁盘
)

// String 返回淘汰策略的字符串表示
func (p EvictPolicy) String() string {
	switch p {
	case EvictOldest:
		return "oldest"
	case EvictByPriority:
		return "priority"
	case EvictPersist:
		return "persist"
	default:
		return "unknown"
	}
}

// SnapshotLevel 快照级别
type SnapshotLevel int

const (
	SnapshotMinimal  SnapshotLevel = iota // 仅渲染缓冲区
	SnapshotStandard                      // 缓冲区+事件历史
	SnapshotFull                          // 包括应用状态
)

// String 返回快照级别的字符串表示
func (l SnapshotLevel) String() string {
	switch l {
	case SnapshotMinimal:
		return "minimal"
	case SnapshotStandard:
		return "standard"
	case SnapshotFull:
		return "full"
	default:
		return "unknown"
	}
}

// InputEvent 统一输入事件 (包装 platform.RawInput)
type InputEvent struct {
	Raw       platform.RawInput
	Injected  bool      // 是否为注入事件
	Timestamp time.Time
}

// BufferWrapper 缓冲区包装器 (复用 paint.Buffer)
type BufferWrapper struct {
	*paint.Buffer
	history   []*paint.Buffer // 历史快照
	maxHistory int
}

// NewBufferWrapper 创建缓冲区包装器
func NewBufferWrapper(buf *paint.Buffer, maxHistory int) *BufferWrapper {
	return &BufferWrapper{
		Buffer:     buf,
		history:    make([]*paint.Buffer, 0, maxHistory),
		maxHistory: maxHistory,
	}
}

// SaveSnapshot 保存当前缓冲区到历史
func (bw *BufferWrapper) SaveSnapshot() {
	if bw.Buffer == nil || bw.maxHistory <= 0 {
		return
	}

	snap := paint.NewBuffer(bw.Buffer.Width, bw.Buffer.Height)
	for y := 0; y < bw.Buffer.Height; y++ {
		copy(snap.Cells[y], bw.Buffer.Cells[y])
	}

	bw.history = append(bw.history, snap)

	// 淘汰旧快照
	if len(bw.history) > bw.maxHistory {
		bw.history = bw.history[1:]
	}
}

// History 返回历史快照
func (bw *BufferWrapper) History() []*paint.Buffer {
	return bw.history
}

// ClearHistory 清空历史
func (bw *BufferWrapper) ClearHistory() {
	bw.history = bw.history[:0]
}
