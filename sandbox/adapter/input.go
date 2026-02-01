// adapter/input.go - 输入适配器 (桥接 platform.RawInput 和 sandbox)
package adapter

import (
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

// InputAdapter 输入适配器 - 桥接 platform.RawInput 和 sandbox
type InputAdapter struct {
	reader   platform.InputReader
	eventsCh chan platform.RawInput
	stopCh   chan struct{}
}

// NewInputAdapter 创建输入适配器
func NewInputAdapter() (*InputAdapter, error) {
	reader, err := platform.NewInputReader()
	if err != nil {
		return nil, err
	}
	return &InputAdapter{
		reader:   reader,
		eventsCh: make(chan platform.RawInput, 100),
		stopCh:   make(chan struct{}),
	}, nil
}

// Start 启动输入读取
func (a *InputAdapter) Start() error {
	return a.reader.Start(a.eventsCh)
}

// Stop 停止输入读取
func (a *InputAdapter) Stop() error {
	close(a.stopCh)
	return a.reader.Stop()
}

// Events 返回事件通道
func (a *InputAdapter) Events() <-chan platform.RawInput {
	return a.eventsCh
}

// ToSandboxEvent 转换为沙箱事件
func ToSandboxEvent(raw platform.RawInput, injected bool) sandbox.InputEvent {
	return sandbox.InputEvent{
		Raw:       raw,
		Injected:  injected,
		Timestamp: time.Now(),
	}
}

// BuildKeyEvent 构建按键事件
func BuildKeyEvent(key rune) platform.RawInput {
	return platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       key,
		Timestamp: time.Now(),
	}
}

// BuildSpecialKeyEvent 构建特殊按键事件
func BuildSpecialKeyEvent(key platform.SpecialKey, mods ...platform.KeyModifier) platform.RawInput {
	var mod platform.KeyModifier
	for _, m := range mods {
		mod |= m
	}
	return platform.RawInput{
		Type:      platform.InputKeyPress,
		Special:   key,
		Modifiers: mod,
		Timestamp: time.Now(),
	}
}

// BuildMouseEvent 构建鼠标事件
func BuildMouseEvent(x, y int, button platform.MouseButton, action platform.MouseAction) platform.RawInput {
	return platform.RawInput{
		Type:        platform.InputMouse,
		MouseX:      x,
		MouseY:      y,
		MouseButton: button,
		MouseAction: action,
		Timestamp:   time.Now(),
	}
}

// BuildResizeEvent 构建窗口调整事件
func BuildResizeEvent(width, height int) platform.RawInput {
	return platform.RawInput{
		Type:      platform.InputResize,
		Width:     width,
		Height:    height,
		Timestamp: time.Now(),
	}
}

// BuildPasteEvent 构建粘贴事件
func BuildPasteEvent(text string) platform.RawInput {
	return platform.RawInput{
		Type:      platform.InputPaste,
		Data:      []byte(text),
		Timestamp: time.Now(),
	}
}
