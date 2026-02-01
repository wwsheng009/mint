// replay/sandbox.go - 回放沙箱实现
package replay

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

// ReplaySandbox 回放沙箱
type ReplaySandbox struct {
	mu     sync.RWMutex

	player *Player
	buffer *paint.Buffer
	config *sandbox.Config
}

// New 创建回放沙箱
func New(events []platform.RawInput, width, height int) *ReplaySandbox {
	config := sandbox.ReplayConfig()
	config.Width = width
	config.Height = height

	return &ReplaySandbox{
		player: NewPlayer(events),
		buffer: paint.NewBuffer(width, height),
		config: config,
	}
}

// Initialize 初始化沙箱
func (rs *ReplaySandbox) Initialize(config *sandbox.Config) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if config != nil {
		rs.config = config
		rs.buffer = paint.NewBuffer(config.Width, config.Height)
	}

	return nil
}

// Run 运行回放
func (rs *ReplaySandbox) Run() error {
	return rs.player.Play()
}

// Pause 暂停回放
func (rs *ReplaySandbox) Pause() error {
	return rs.player.Pause()
}

// Resume 恢复回放
func (rs *ReplaySandbox) Resume() error {
	return rs.player.Play()
}

// Close 关闭沙箱
func (rs *ReplaySandbox) Close() error {
	rs.player.Stop()
	return nil
}

// State 获取当前状态
func (rs *ReplaySandbox) State() sandbox.State {
	if rs.player.IsPlaying() {
		return sandbox.StateRunning
	}
	return sandbox.StatePaused
}

// Type 获取沙箱类型
func (rs *ReplaySandbox) Type() sandbox.SandboxType {
	return sandbox.TypeReplay
}

// Config 获取配置
func (rs *ReplaySandbox) Config() *sandbox.Config {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.config
}

// Buffer 获取渲染缓冲区
func (rs *ReplaySandbox) Buffer() *paint.Buffer {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.buffer
}

// SetBuffer 设置渲染缓冲区
func (rs *ReplaySandbox) SetBuffer(buf *paint.Buffer) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.buffer = buf
}

// Resize 调整缓冲区大小
func (rs *ReplaySandbox) Resize(width, height int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.buffer = paint.NewBuffer(width, height)
	rs.config.Width = width
	rs.config.Height = height
}

// Size 获取当前尺寸
func (rs *ReplaySandbox) Size() (int, int) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.config.Width, rs.config.Height
}

// SetSpeed 设置回放速度
func (rs *ReplaySandbox) SetSpeed(speed float64) {
	rs.player.SetSpeed(speed)
}

// GetSpeed 获取回放速度
func (rs *ReplaySandbox) GetSpeed() float64 {
	return rs.player.Speed()
}

// Step 前进一步
func (rs *ReplaySandbox) Step() (platform.RawInput, error) {
	return rs.player.Next()
}

// StepBack 后退一步
func (rs *ReplaySandbox) StepBack() (platform.RawInput, error) {
	return rs.player.Previous()
}
