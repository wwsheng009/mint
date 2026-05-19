package input

import (
	"time"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// InputSnapshot 是输入状态的快照
// 包含鼠标、键盘等所有输入的当前状态
//
// Based on the archived pressed-state design:
// docsArchive/cleanup-2026-05-19/docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md
// - 输入不是事件流，而是"状态快照流"
// - 通过比较前后两个快照推断边缘事件（Press/Release/Move）
type InputSnapshot struct {
	// 鼠标状态
	MouseX      int
	MouseY      int
	MouseButton runtimemsg.MouseButton
	MouseAction runtimemsg.MouseAction

	// 键盘状态
	KeyboardKey rune // 可打印字符
	SpecialKey  runtimeplatform.SpecialKey
	Modifiers   runtimemsg.Modifiers

	// 时间戳（用于事件顺序）
	Timestamp int64
}

// Clone 创建快照的副本
func (s *InputSnapshot) Clone() *InputSnapshot {
	if s == nil {
		return &InputSnapshot{
			Timestamp: time.Now().UnixNano(),
		}
	}
	return &InputSnapshot{
		MouseX:      s.MouseX,
		MouseY:      s.MouseY,
		MouseButton: s.MouseButton,
		MouseAction: s.MouseAction,
		KeyboardKey: s.KeyboardKey,
		SpecialKey:  s.SpecialKey,
		Modifiers:   s.Modifiers,
		Timestamp:   s.Timestamp,
	}
}

// IsEmpty 检查快照是否为空
func (s *InputSnapshot) IsEmpty() bool {
	if s == nil {
		return true
	}
	return s.MouseButton == runtimemsg.MouseButtonUnknown &&
		s.MouseAction == runtimemsg.MouseActionUnknown &&
		s.KeyboardKey == 0 &&
		s.SpecialKey == runtimeplatform.KeyUnknown
}
