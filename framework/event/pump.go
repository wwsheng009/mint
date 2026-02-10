package event

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/platform"
)

// EventSource 事件源接口
// Pump 从 EventSource 读取原始输入，转换为框架事件
//
// 实现者：
//   - PlatformEventSource: 包装 platform.InputReader (真实终端)
//   - SandboxEventSource: 包装 sandbox.MockSandbox (测试环境)
//   - ChannelEventSource: 直接从通道读取 (最简单)
type EventSource interface {
	// Start 启动事件源，返回事件通道
	Start() (<-chan platform.RawInput, error)

	// Stop 停止事件源
	Stop() error
}

// Pump reads raw input from an EventSource and converts to Msg.
type Pump struct {
	source EventSource
	messages chan runtimemsg.Msg // Changed from events chan Event
	quit   chan struct{}
	running int32 // Use atomic for cross-goroutine visibility (0=stopped, 1=running)
	mu     sync.RWMutex // Protects messages channel from close while sending
	wg     sync.WaitGroup // Waits for convertLoop to exit

	// HitMap for mouse event hit testing (set by App after each render)
	hitMap *event.HitMap
	hitMapMu sync.RWMutex // Protects hitMap from concurrent access
}

// NewPump creates a new event pump with a platform input reader.
func NewPump(reader platform.InputReader) *Pump {
	return &Pump{
		source: &PlatformEventSource{reader: reader},
		messages: make(chan runtimemsg.Msg, 100), // Changed from events
		quit:   make(chan struct{}),
		running: 0,
	}
}

// NewPumpWithSource creates a new event pump with a custom EventSource.
// This allows using MockSandbox or other test event sources.
func NewPumpWithSource(source EventSource) *Pump {
	return &Pump{
		source: source,
		messages: make(chan runtimemsg.Msg, 100), // Changed from events
		quit:   make(chan struct{}),
		running: 0,
	}
}

// Start starts the event pump.
func (p *Pump) Start() error {
	if atomic.LoadInt32(&p.running) != 0 {
		return nil
	}

	// Start event source and get input channel
	rawInputs, err := p.source.Start()
	if err != nil {
		return err
	}

	atomic.StoreInt32(&p.running, 1)

	// DEBUG: 打印启动信息
	if os.Getenv("TUI_DEBUG_PUMP") == "true" {
		fmt.Fprintf(os.Stderr, "[PUMP] Started, convertLoop running...\n")
	}

	// Start conversion loop
	p.wg.Add(1)
	go p.convertLoop(rawInputs)

	return nil
}

// convertLoop converts raw inputs to messages.
func (p *Pump) convertLoop(rawInputs <-chan platform.RawInput) {
	defer p.wg.Done()
	for {
		select {
		case <-p.quit:
			return

		case raw, ok := <-rawInputs:
			if !ok {
				return
			}
			msg := p.convertToMsg(raw)
			if msg != nil {
				// Safe to send because Stop() waits for this goroutine to exit
				select {
				case p.messages <- msg:
				case <-p.quit:
					return
				}
			}
		}
	}
}

// convertToMsg converts raw input to Msg.
func (p *Pump) convertToMsg(raw platform.RawInput) runtimemsg.Msg {
	switch raw.Type {
	case platform.InputKeyPress:
		return p.convertToKeyMsg(raw)

	case platform.InputResize:
		return p.convertToResizeMsg(raw)

	case platform.InputMouse:
		return p.convertToMouseMsg(raw)

	default:
		return nil
	}
}

// convertToKeyMsg converts keyboard raw input to KeyMsg.
func (p *Pump) convertToKeyMsg(raw platform.RawInput) runtimemsg.Msg {
	// Convert Modifiers
	var modifiers runtimemsg.Modifiers
	if raw.Modifiers&platform.ModAlt != 0 {
		modifiers.Alt = true
	}
	if raw.Modifiers&platform.ModCtrl != 0 {
		modifiers.Ctrl = true
	}
	if raw.Modifiers&platform.ModShift != 0 {
		modifiers.Shift = true
	}

	return runtimemsg.NewKeyMsg(
		raw.Key,
		raw.Special,
		modifiers,
	)
}

// convertToResizeMsg converts resize raw input to ResizeMsg.
func (p *Pump) convertToResizeMsg(raw platform.RawInput) runtimemsg.Msg {
	// Create a simple resize message
	return &runtimemsg.BaseMsg{
		TypeValue:      runtimemsg.MsgTypeResize,
		TimestampValue: time.Now(),
	}
	// Note: Width/Height info would need ResizeMsg struct, but for now just type is enough
}

// convertToMouseMsg converts mouse raw input to MouseMsg.
func (p *Pump) convertToMouseMsg(raw platform.RawInput) runtimemsg.Msg {
	// Convert mouse button
	var button runtimemsg.MouseButton
	switch raw.MouseButton {
	case platform.MouseLeft:
		button = runtimemsg.MouseLeft
	case platform.MouseMiddle:
		button = runtimemsg.MouseMiddle
	case platform.MouseRight:
		button = runtimemsg.MouseRight
	default:
		button = runtimemsg.MouseButtonUnknown
	}

	// Convert mouse action
	var action runtimemsg.MouseAction
	switch raw.MouseAction {
	case platform.MousePress:
		action = runtimemsg.MouseActionPress
	case platform.MouseRelease:
		action = runtimemsg.MouseActionRelease
	case platform.MouseMotion:
		action = runtimemsg.MouseActionMove
	case platform.MouseWheelUp:
		action = runtimemsg.MouseActionWheel
	case platform.MouseWheelDown:
		action = runtimemsg.MouseActionWheel
	default:
		action = runtimemsg.MouseActionUnknown
	}

	// Create MouseMsg with basic fields
	mouseMsg := &runtimemsg.MouseMsg{
		X:      raw.MouseX,
		Y:      raw.MouseY,
		Button: button,
		Action: action,
	}

	// Phase 1-6: Fill in hit testing information from HitMap
	p.hitMapMu.RLock()
	hitMap := p.hitMap
	p.hitMapMu.RUnlock()

	if hitMap != nil {
		// Perform hit testing
		entry := hitMap.HitTest(raw.MouseX, raw.MouseY)
		if entry != nil {
			mouseMsg.TargetID = entry.NodeID
			// Calculate local coordinates using the entry's LocalXY function
			localX, localY := entry.LocalXY(raw.MouseX, raw.MouseY)
			mouseMsg.LocalX = localX
			mouseMsg.LocalY = localY
		}
	}

	// Calculate Delta for wheel events
	if raw.MouseAction == platform.MouseWheelUp {
		mouseMsg.Delta = 1
	} else if raw.MouseAction == platform.MouseWheelDown {
		mouseMsg.Delta = -1
	}

	return mouseMsg
}

// Stop stops the event pump.
func (p *Pump) Stop() {
	if atomic.LoadInt32(&p.running) == 0 {
		return
	}

	atomic.StoreInt32(&p.running, 0)

	// Send quit signal first (signals convertLoop to stop)
	close(p.quit)

	// Stop event source
	if p.source != nil {
		p.source.Stop()
	}

	// Wait for convertLoop to exit (prevents send on closed channel)
	p.wg.Wait()

	// Now safe to close messages channel
	p.mu.Lock()
	close(p.messages)
	p.mu.Unlock()
}

// Events returns the messages channel.
func (p *Pump) Events() <-chan runtimemsg.Msg {
	return p.messages
}

// IsRunning checks if the pump is running.
func (p *Pump) IsRunning() bool {
	return atomic.LoadInt32(&p.running) != 0
}

// PumpWithTimeout gets a message with timeout.
func (p *Pump) PumpWithTimeout(timeout time.Duration) (runtimemsg.Msg, bool) {
	select {
	case message := <-p.messages:
		return message, true
	case <-time.After(timeout):
		return nil, false
	}
}

// ============================================================================
// EventSource 实现
// ============================================================================

// PlatformEventSource 包装 platform.InputReader 为 EventSource
type PlatformEventSource struct {
	reader     platform.InputReader
	rawInputs  chan platform.RawInput
}

// Start 启动平台输入源
func (s *PlatformEventSource) Start() (<-chan platform.RawInput, error) {
	s.rawInputs = make(chan platform.RawInput, 50)
	if err := s.reader.Start(s.rawInputs); err != nil {
		return nil, err
	}
	return s.rawInputs, nil
}

// Stop 停止平台输入源
func (s *PlatformEventSource) Stop() error {
	if s.reader != nil {
		return s.reader.Stop()
	}
	return nil
}

// ChannelEventSource 直接从通道读取事件 (最简单的 EventSource)
type ChannelEventSource struct {
	ch chan platform.RawInput
}

// NewChannelEventSource 创建通道事件源
func NewChannelEventSource(ch chan platform.RawInput) *ChannelEventSource {
	return &ChannelEventSource{ch: ch}
}

// Start 返回事件通道
func (s *ChannelEventSource) Start() (<-chan platform.RawInput, error) {
	return s.ch, nil
}

// Stop 停止 (无操作)
func (s *ChannelEventSource) Stop() error {
	return nil
}

// SetHitMap sets the HitMap for mouse event hit testing.
// This should be called by App after each render to ensure hit testing
// uses the latest layout information.
func (p *Pump) SetHitMap(hitMap *event.HitMap) {
	p.hitMapMu.Lock()
	defer p.hitMapMu.Unlock()
	p.hitMap = hitMap
}

// Inject 用于测试时直接注入输入到 Pump
// 注意：此方法仅用于测试，不应用于生产代码
func (p *Pump) Inject(raw platform.RawInput) {
	if atomic.LoadInt32(&p.running) == 0 {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[PUMP] Inject: pump not running!\n")
		}
		return
	}
	message := p.convertToMsg(raw)
	if message != nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[PUMP] Injecting message: Type=%v\n", message.Type())
		}
		// Safe to send because Stop() waits for all goroutines to exit first
		select {
		case p.messages <- message:
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[PUMP] Message sent to channel\n")
			}
		case <-p.quit:
		}
	}
}
