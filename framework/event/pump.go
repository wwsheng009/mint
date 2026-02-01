package event

import (
	"fmt"
	"os"
	"time"

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

// Pump reads raw input from an EventSource and converts to events.
type Pump struct {
	source EventSource
	events chan Event
	quit   chan struct{}
	running bool
}

// NewPump creates a new event pump with a platform input reader.
func NewPump(reader platform.InputReader) *Pump {
	return &Pump{
		source: &PlatformEventSource{reader: reader},
		events: make(chan Event, 100),
		quit:   make(chan struct{}),
		running: false,
	}
}

// NewPumpWithSource creates a new event pump with a custom EventSource.
// This allows using MockSandbox or other test event sources.
func NewPumpWithSource(source EventSource) *Pump {
	return &Pump{
		source: source,
		events: make(chan Event, 100),
		quit:   make(chan struct{}),
		running: false,
	}
}

// Start starts the event pump.
func (p *Pump) Start() error {
	if p.running {
		return nil
	}

	// Start event source and get input channel
	rawInputs, err := p.source.Start()
	if err != nil {
		return err
	}

	p.running = true

	// DEBUG: 打印启动信息
	if os.Getenv("TUI_DEBUG_PUMP") == "true" {
		fmt.Fprintf(os.Stderr, "[PUMP] Started, convertLoop running...\n")
	}

	// Start conversion loop
	go p.convertLoop(rawInputs)

	return nil
}

// convertLoop converts raw inputs to events.
func (p *Pump) convertLoop(rawInputs <-chan platform.RawInput) {
	for {
		select {
		case <-p.quit:
			return

		case raw, ok := <-rawInputs:
			if !ok {
				return
			}
			ev := p.convertToEvent(raw)
			if ev != nil {
				select {
				case p.events <- ev:
				case <-p.quit:
					return
				}
			}
		}
	}
}

// convertToEvent converts raw input to framework event.
func (p *Pump) convertToEvent(raw platform.RawInput) Event {
	switch raw.Type {
	case platform.InputKeyPress:
		return p.convertKeyEvent(raw)

	case platform.InputResize:
		return p.convertResizeEvent(raw)

	case platform.InputMouse:
			return p.convertMouseEvent(raw)

	default:
		return nil
	}
}

// convertKeyEvent converts keyboard raw input to KeyEvent.
func (p *Pump) convertKeyEvent(raw platform.RawInput) Event {
	baseEv := NewBaseEvent(event.EventKeyPress)

	// Create key event
	ev := &KeyEvent{
		BaseEvent: baseEv,
	}

	// Set special key
	if raw.Special != platform.KeyUnknown {
		ev.Special = SpecialKey(raw.Special)
		ev.Key.Name = ev.Special.String()
	} else {
		// Character key
		ev.Key.Rune = raw.Key
	}

	// Set modifiers
	if raw.Modifiers&platform.ModAlt != 0 {
		ev.Key.Alt = true
		ev.Modifiers |= ModAlt
	}
	if raw.Modifiers&platform.ModCtrl != 0 {
		ev.Key.Ctrl = true
		ev.Modifiers |= ModCtrl
	}
	if raw.Modifiers&platform.ModShift != 0 {
		ev.Modifiers |= ModShift
	}

	return ev
}

// convertResizeEvent converts resize raw input to ResizeEvent.
func (p *Pump) convertResizeEvent(raw platform.RawInput) Event {
	return &ResizeEvent{
		BaseEvent: NewBaseEvent(event.EventResize),
		OldWidth:  0,
		OldHeight: 0,
		NewWidth:  raw.Width,
		NewHeight: raw.Height,
	}
}

// convertMouseEvent converts mouse raw input to MouseEvent.
func (p *Pump) convertMouseEvent(raw platform.RawInput) Event {
	var eventType event.EventType

	switch raw.MouseAction {
	case platform.MousePress:
		eventType = event.EventMousePress
	case platform.MouseRelease:
		eventType = event.EventMouseRelease
	case platform.MouseMotion:
		eventType = event.EventMouseMove
	case platform.MouseWheelUp:
		eventType = event.EventMouseWheel
	case platform.MouseWheelDown:
		eventType = event.EventMouseWheel
	default:
		eventType = event.EventMousePress
	}

	return &MouseEvent{
		BaseEvent: NewBaseEvent(eventType),
		X:         raw.MouseX,
		Y:         raw.MouseY,
		Button:    MouseButton(raw.MouseButton),
	}
}

// Stop stops the event pump.
func (p *Pump) Stop() {
	if !p.running {
		return
	}

	p.running = false

	// Send quit signal
	close(p.quit)

	// Stop event source
	if p.source != nil {
		p.source.Stop()
	}

	// Close events channel
	close(p.events)
}

// Events returns the event channel.
func (p *Pump) Events() <-chan Event {
	return p.events
}

// IsRunning checks if the pump is running.
func (p *Pump) IsRunning() bool {
	return p.running
}

// PumpWithTimeout gets an event with timeout.
func (p *Pump) PumpWithTimeout(timeout time.Duration) (Event, bool) {
	select {
	case ev := <-p.events:
		return ev, true
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

// Inject 用于测试时直接注入事件到 Pump
// 注意：此方法仅用于测试，不应用于生产代码
func (p *Pump) Inject(raw platform.RawInput) {
	if !p.running {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[PUMP] Inject: pump not running!\n")
		}
		return
	}
	ev := p.convertToEvent(raw)
	if ev != nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[PUMP] Injecting event: Type=%v\n", ev.Type())
		}
		select {
		case p.events <- ev:
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[PUMP] Event sent to channel\n")
			}
		case <-p.quit:
		}
	}
}
