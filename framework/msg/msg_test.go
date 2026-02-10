package msg

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework/cmd"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// TestMsg_BaseMsg 测试基础消息
func TestMsg_BaseMsg(t *testing.T) {
	msg := NewBaseMsg(MsgTypeKey)

	if msg.Type() != MsgTypeKey {
		t.Errorf("Expected type %s, got %s", MsgTypeKey, msg.Type())
	}

	if msg.Timestamp().IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

// TestKeyMsg_NewKeyMsg 测试创建键盘消息
func TestKeyMsg_NewKeyMsg(t *testing.T) {
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})

	if keyMsg.Type() != MsgTypeKey {
		t.Errorf("Expected type %s, got %s", MsgTypeKey, keyMsg.Type())
	}

	if keyMsg.Rune != 'A' {
		t.Errorf("Expected rune 'A', got %c", keyMsg.Rune)
	}
}

// TestKeyMsg_IsPrintable 测试可打印字符检查
func TestKeyMsg_IsPrintable(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   *KeyMsg
		expected bool
	}{
		{
			name:     "printable character",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{}),
			expected: true,
		},
		{
			name:     "space",
			keyMsg:   NewKeyMsg(' ', runtimeplatform.KeyUnknown, Modifiers{}),
			expected: true,
		},
		{
			name:     "enter key",
			keyMsg:   NewKeyMsg(0, runtimeplatform.KeyEnter, Modifiers{}),
			expected: false,
		},
		{
			name:     "tab key",
			keyMsg:   NewKeyMsg(0, runtimeplatform.KeyTab, Modifiers{}),
			expected: false,
		},
		{
			name:     "escape key",
			keyMsg:   NewKeyMsg(0, runtimeplatform.KeyEscape, Modifiers{}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.keyMsg.IsPrintable(); got != tt.expected {
				t.Errorf("IsPrintable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestKeyMsg_HasModifier 测试修饰键检查
func TestKeyMsg_HasModifier(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   *KeyMsg
		expected bool
	}{
		{
			name:     "no modifier",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{}),
			expected: false,
		},
		{
			name:     "ctrl only",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{Ctrl: true}),
			expected: true,
		},
		{
			name:     "alt only",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{Alt: true}),
			expected: true,
		},
		{
			name:     "shift only",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{Shift: true}),
			expected: true,
		},
		{
			name:     "ctrl+alt",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{Ctrl: true, Alt: true}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.keyMsg.HasModifier(); got != tt.expected {
				t.Errorf("HasModifier() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestKeyMsg_IsNavigation 测试导航键检查
func TestKeyMsg_IsNavigation(t *testing.T) {
	navigationKeys := []runtimeplatform.SpecialKey{
		runtimeplatform.KeyUp,
		runtimeplatform.KeyDown,
		runtimeplatform.KeyLeft,
		runtimeplatform.KeyRight,
		runtimeplatform.KeyHome,
		runtimeplatform.KeyEnd,
		runtimeplatform.KeyPageUp,
		runtimeplatform.KeyPageDown,
	}

	for _, key := range navigationKeys {
		t.Run(key.String(), func(t *testing.T) {
			keyMsg := NewKeyMsg(0, key, Modifiers{})
			if !keyMsg.IsNavigation() {
				t.Errorf("IsNavigation() should return true for %s", key)
			}
		})
	}

	// 非导航键
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})
	if keyMsg.IsNavigation() {
		t.Error("IsNavigation() should return false for regular character")
	}
}

// TestKeyMsg_String 测试字符串表示
func TestKeyMsg_String(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   *KeyMsg
		expected string
	}{
		{
			name:     "simple character",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{}),
			expected: "KeyMsg{'A'}",
		},
		{
			name:     "ctrl+A",
			keyMsg:   NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{Ctrl: true}),
			expected: "KeyMsg{Ctrl+'A'}",
		},
		{
			name:     "enter key",
			keyMsg:   NewKeyMsg(0, runtimeplatform.KeyEnter, Modifiers{}),
			expected: "KeyMsg{Enter}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.keyMsg.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestMouseMsg_NewMouseMsg 测试创建鼠标消息
func TestMouseMsg_NewMouseMsg(t *testing.T) {
	mouseMsg := NewMouseMsg(100, 200, MouseLeft, MouseActionPress)

	if mouseMsg.Type() != MsgTypeMouse {
		t.Errorf("Expected type %s, got %s", MsgTypeMouse, mouseMsg.Type())
	}

	if mouseMsg.X != 100 || mouseMsg.Y != 200 {
		t.Errorf("Expected position (100, 200), got (%d, %d)", mouseMsg.X, mouseMsg.Y)
	}
}

// TestMouseMsg_IsClick 测试点击检查
func TestMouseMsg_IsClick(t *testing.T) {
	clickMsg := NewMouseMsg(0, 0, MouseLeft, MouseActionPress)
	if !clickMsg.IsClick() {
		t.Error("IsClick() should return true for left button press")
	}

	rightClickMsg := NewMouseMsg(0, 0, MouseRight, MouseActionPress)
	if rightClickMsg.IsClick() {
		t.Error("IsClick() should return false for right button")
	}

	releaseMsg := NewMouseMsg(0, 0, MouseLeft, MouseActionRelease)
	if releaseMsg.IsClick() {
		t.Error("IsClick() should return false for release")
	}
}

// TestMouseMsg_IsScroll 测试滚轮检查
func TestMouseMsg_IsScroll(t *testing.T) {
	scrollMsg := NewMouseMsg(0, 0, MouseButtonUnknown, MouseActionWheel)
	if !scrollMsg.IsScroll() {
		t.Error("IsScroll() should return true for wheel action")
	}

	clickMsg := NewMouseMsg(0, 0, MouseLeft, MouseActionPress)
	if clickMsg.IsScroll() {
		t.Error("IsScroll() should return false for click")
	}
}

// TestSandboxMsg 测试沙箱消息
func TestSandboxMsg(t *testing.T) {
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})
	sandboxMsg := NewSandboxKeyMsg(keyMsg)

	if sandboxMsg.Type() != MsgTypeSandbox {
		t.Errorf("Expected type %s, got %s", MsgTypeSandbox, sandboxMsg.Type())
	}

	if !sandboxMsg.IsInput() {
		t.Error("IsInput() should return true for key injection")
	}

	if sandboxMsg.IsStateMutation() {
		t.Error("IsStateMutation() should return false for key injection")
	}
}

// TestSandboxMsg_StateMutation 测试状态修改
func TestSandboxMsg_StateMutation(t *testing.T) {
	stateMsg := NewSandboxStateMsg("button1", "value", "test")

	if stateMsg.InjectType != SandboxInjectState {
		t.Errorf("Expected inject type %s, got %s", SandboxInjectState, stateMsg.InjectType)
	}

	if !stateMsg.IsStateMutation() {
		t.Error("IsStateMutation() should return true")
	}

	if stateMsg.StateMutation == nil {
		t.Error("StateMutation should not be nil")
	}

	if stateMsg.StateMutation.TargetID != "button1" {
		t.Errorf("Expected target ID 'button1', got %s", stateMsg.StateMutation.TargetID)
	}
}

// TestCmd_None 测试空命令
func TestCmd_None(t *testing.T) {
	none := cmd.None()
	if none.Type() != cmd.CmdTypeNone {
		t.Errorf("Expected type %s, got %s", cmd.CmdTypeNone, none.Type())
	}

	msgs := cmd.Execute(none)
	if msgs != nil {
		t.Error("Execute(None) should return nil")
	}
}

// TestCmd_Batch 测试批量命令
func TestCmd_Batch(t *testing.T) {
	batch := cmd.Batch(
		cmd.None(),
		cmd.None(),
	)

	if batch.Type() != cmd.CmdTypeNone {
		t.Errorf("Batch of None commands should be None, got %s", batch.Type())
	}

	// Test with actual commands
	cmd1 := cmd.After(time.Millisecond, nil)
	cmd2 := cmd.After(time.Millisecond, nil)
	batch2 := cmd.Batch(cmd1, cmd2)

	if batch2.Type() != cmd.CmdTypeBatch {
		t.Errorf("Expected type %s, got %s", cmd.CmdTypeBatch, batch2.Type())
	}
}

// TestCmd_After 测试延迟命令
func TestCmd_After(t *testing.T) {
	after := cmd.After(time.Second, cmd.None())

	if after.Type() != cmd.CmdTypeAfter {
		t.Errorf("Expected type %s, got %s", cmd.CmdTypeAfter, after.Type())
	}

	// After with None should return None
	after2 := cmd.After(time.Second, cmd.None())
	if after2.Type() != cmd.CmdTypeNone {
		t.Errorf("After(None) should be None, got %s", after2.Type())
	}
}

// TestCmd_Tick 测试定时器命令
func TestCmd_Tick(t *testing.T) {
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})
	tick := cmd.Tick(time.Second, keyMsg)

	if tick.Type() != cmd.CmdTypeTick {
		t.Errorf("Expected type %s, got %s", cmd.CmdTypeTick, tick.Type())
	}

	// Tick with nil msg should return None
	tick2 := cmd.Tick(time.Second, nil)
	if tick2.Type() != cmd.CmdTypeNone {
		t.Errorf("Tick(nil) should be None, got %s", tick2.Type())
	}
}

// TestCmd_IO 测试 I/O 命令
func TestCmd_IO(t *testing.T) {
	io := cmd.IO(func() Msg {
		return NewBaseMsg(MsgTypeQuit)
	})

	if io.Type() != cmd.CmdTypeIO {
		t.Errorf("Expected type %s, got %s", cmd.CmdTypeIO, io.Type())
	}

	// IO with nil function should return None
	io2 := cmd.IO(nil)
	if io2.Type() != cmd.CmdTypeNone {
		t.Errorf("IO(nil) should be None, got %s", io2.Type())
	}
}

// TestIsInputMsg 测试输入消息检查
func TestIsInputMsg(t *testing.T) {
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})
	mouseMsg := NewMouseMsg(0, 0, MouseLeft, MouseActionPress)
	resizeMsg := NewResizeMsg(100, 100)

	if !IsInputMsg(keyMsg) {
		t.Error("KeyMsg should be input message")
	}

	if !IsInputMsg(mouseMsg) {
		t.Error("MouseMsg should be input message")
	}

	if IsInputMsg(resizeMsg) {
		t.Error("ResizeMsg should not be input message")
	}
}

// TestIsSystemMsg 测试系统消息检查
func TestIsSystemMsg(t *testing.T) {
	resizeMsg := NewResizeMsg(100, 100)
	quitMsg := NewBaseMsg(MsgTypeQuit)
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})

	if !IsSystemMsg(resizeMsg) {
		t.Error("ResizeMsg should be system message")
	}

	if !IsSystemMsg(quitMsg) {
		t.Error("QuitMsg should be system message")
	}

	if IsSystemMsg(keyMsg) {
		t.Error("KeyMsg should not be system message")
	}
}

// TestFormatMsg 测试消息格式化
func TestFormatMsg(t *testing.T) {
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})
	formatted := FormatMsg(keyMsg)

	if formatted == "" {
		t.Error("FormatMsg should not return empty string")
	}

	// Check if it contains the type
	if formatted == "" || formatted[0] != '[' {
		t.Errorf("FormatMsg should start with '[', got: %s", formatted)
	}
}

// TestMsgAge 测试消息年龄
func TestMsgAge(t *testing.T) {
	keyMsg := NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{})

	age := MsgAge(keyMsg)
	if age < 0 {
		t.Error("MsgAge should not be negative")
	}

	// Wait a bit and check if age increased
	time.Sleep(10 * time.Millisecond)
	age2 := MsgAge(keyMsg)
	if age2 <= age {
		t.Error("MsgAge should increase over time")
	}
}
