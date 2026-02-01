// real/sandbox.go - 真实终端沙箱实现
package real

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/adapter"
)

// RealSandbox 真实终端沙箱
type RealSandbox struct {
	mu sync.RWMutex

	lifecycle *sandbox.Lifecycle
	config    *sandbox.Config
	buffer    *paint.Buffer

	// 输入适配器
	input *adapter.InputAdapter

	// 事件系统
	injector *sandbox.EventInjector
	recorder *sandbox.EventRecorder

	// 快照
	snapMgr *sandbox.SnapshotManager

	// 停止信号
	stopCh chan struct{}
}

// New 创建真实沙箱
func New(width, height int) (*RealSandbox, error) {
	config := sandbox.RealConfig()
	config.Width = width
	config.Height = height

	input, err := adapter.NewInputAdapter()
	if err != nil {
		return nil, err
	}

	rs := &RealSandbox{
		lifecycle: sandbox.NewLifecycle(),
		config:    config,
		buffer:    paint.NewBuffer(width, height),
		input:     input,
		injector:  sandbox.NewEventInjector(sandbox.InjectProhibited),
		recorder:  sandbox.NewEventRecorder(config.Event.RecordMaxLen),
		snapMgr:   sandbox.NewSnapshotManager(config.Snapshot.MaxCount),
		stopCh:    make(chan struct{}),
	}

	rs.injector.SetRecorder(rs.recorder)

	return rs, nil
}

// ==============================================================================
// Sandbox Interface
// ==============================================================================

// Initialize 初始化沙箱
func (rs *RealSandbox) Initialize(config *sandbox.Config) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if config != nil {
		rs.config = config
		rs.buffer = paint.NewBuffer(config.Width, config.Height)
	}

	return rs.lifecycle.Transition(sandbox.StateInitialized)
}

// Run 运行沙箱主循环
func (rs *RealSandbox) Run() error {
	if err := rs.lifecycle.Transition(sandbox.StateRunning); err != nil {
		return err
	}

	// 启动输入读取
	if err := rs.input.Start(); err != nil {
		return err
	}

	// 事件循环
	go rs.eventLoop()

	return nil
}

// eventLoop 事件循环
func (rs *RealSandbox) eventLoop() {
	for {
		select {
		case <-rs.stopCh:
			return
		case event := <-rs.input.Events():
			rs.handleEvent(event)
		}
	}
}

// handleEvent 处理事件
func (rs *RealSandbox) handleEvent(event platform.RawInput) {
	// 录制事件
	rs.recorder.Record(event)

	// 处理窗口调整
	if event.Type == platform.InputResize {
		rs.Resize(event.Width, event.Height)
	}
}

// Pause 暂停沙箱
func (rs *RealSandbox) Pause() error {
	return rs.lifecycle.Transition(sandbox.StatePaused)
}

// Resume 恢复沙箱
func (rs *RealSandbox) Resume() error {
	return rs.lifecycle.Transition(sandbox.StateRunning)
}

// Close 关闭沙箱
func (rs *RealSandbox) Close() error {
	close(rs.stopCh)
	rs.input.Stop()
	platform.RestoreTerminal()
	return rs.lifecycle.Transition(sandbox.StateStopped)
}

// State 获取当前状态
func (rs *RealSandbox) State() sandbox.State {
	return rs.lifecycle.State()
}

// Type 获取沙箱类型
func (rs *RealSandbox) Type() sandbox.SandboxType {
	return sandbox.TypeReal
}

// Config 获取配置
func (rs *RealSandbox) Config() *sandbox.Config {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.config
}

// Buffer 获取渲染缓冲区
func (rs *RealSandbox) Buffer() *paint.Buffer {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.buffer
}

// SetBuffer 设置渲染缓冲区
func (rs *RealSandbox) SetBuffer(buf *paint.Buffer) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.buffer = buf
}

// Resize 调整缓冲区大小
func (rs *RealSandbox) Resize(width, height int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.buffer = paint.NewBuffer(width, height)
	rs.config.Width = width
	rs.config.Height = height
}

// Size 获取当前尺寸
func (rs *RealSandbox) Size() (int, int) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.config.Width, rs.config.Height
}

// ==============================================================================
// EventSource Interface
// ==============================================================================

// Events 返回事件通道
func (rs *RealSandbox) Events() <-chan platform.RawInput {
	return rs.input.Events()
}

// ==============================================================================
// Snapshotter Interface
// ==============================================================================

// Snapshot 创建快照
func (rs *RealSandbox) Snapshot(level sandbox.SnapshotLevel, tags ...string) (*sandbox.Snapshot, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	return rs.snapMgr.Create(level, rs.buffer, rs.recorder.Events(), nil, tags...)
}

// Restore 恢复快照
func (rs *RealSandbox) Restore(snap *sandbox.Snapshot) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// 真实环境软恢复：重新渲染
	if snap.Buffer != nil {
		rs.buffer = paint.NewBuffer(snap.Buffer.Width, snap.Buffer.Height)
		for y := 0; y < snap.Buffer.Height; y++ {
			copy(rs.buffer.Cells[y], snap.Buffer.Cells[y])
		}
	}

	return nil
}

// ListSnapshots 列出所有快照
func (rs *RealSandbox) ListSnapshots() []*sandbox.SnapshotMetadata {
	return rs.snapMgr.List()
}

// RecordedEvents 获取录制的事件
func (rs *RealSandbox) RecordedEvents() []platform.RawInput {
	return rs.recorder.Events()
}
