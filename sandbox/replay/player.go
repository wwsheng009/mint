// replay/player.go - 事件回放器
package replay

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

// Player 事件回放器
type Player struct {
	mu      sync.RWMutex
	events  []platform.RawInput
	index   int
	speed   float64
	playing bool
}

// NewPlayer 创建回放器
func NewPlayer(events []platform.RawInput) *Player {
	if events == nil {
		events = make([]platform.RawInput, 0)
	}
	return &Player{
		events:  events,
		index:   0,
		speed:   1.0,
		playing: false,
	}
}

// Play 开始回放
func (p *Player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playing = true
	return nil
}

// Pause 暂停回放
func (p *Player) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playing = false
	return nil
}

// Stop 停止回放
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playing = false
	p.index = 0
	return nil
}

// Seek 跳转到指定索引
func (p *Player) Seek(index int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index >= len(p.events) {
		return sandbox.ErrSnapshotNotFound
	}

	p.index = index
	return nil
}

// Next 下一个事件
func (p *Player) Next() (platform.RawInput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.index >= len(p.events) {
		return platform.RawInput{}, sandbox.ErrQueueEmpty
	}

	event := p.events[p.index]
	p.index++
	return event, nil
}

// Previous 上一个事件
func (p *Player) Previous() (platform.RawInput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.index <= 0 {
		return platform.RawInput{}, sandbox.ErrQueueEmpty
	}

	p.index--
	return p.events[p.index], nil
}

// Current 当前事件
func (p *Player) Current() (platform.RawInput, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.index < 0 || p.index >= len(p.events) {
		return platform.RawInput{}, sandbox.ErrQueueEmpty
	}

	return p.events[p.index], nil
}

// SetSpeed 设置回放速度
func (p *Player) SetSpeed(speed float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.speed = speed
}

// Speed 获取回放速度
func (p *Player) Speed() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.speed
}

// IsPlaying 是否正在播放
func (p *Player) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playing
}

// Index 当前索引
func (p *Player) Index() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.index
}

// Length 事件总数
func (p *Player) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.events)
}

// HasNext 是否有下一个事件
func (p *Player) HasNext() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.index < len(p.events)
}

// HasPrevious 是否有上一个事件
func (p *Player) HasPrevious() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.index > 0
}

// Reset 重置到开始
func (p *Player) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.index = 0
	p.playing = false
	p.speed = 1.0
}
