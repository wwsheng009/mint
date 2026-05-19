package input

import (
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// InputTracker 跟踪输入状态变化，推断边缘事件
//
// Based on the archived pressed-state design:
// docsArchive/cleanup-2026-05-19/docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md
// - 不依赖底层事件完整性
// - 通过比较前一帧和当前帧状态推断 Press/Release
type InputTracker struct {
	// 上一帧状态
	lastSnapshot *InputSnapshot

	// 边缘事件推断结果
	lastIntent InputIntent
}

// InputIntent 表示输入意图（状态变化）
type InputIntent interface {
	isInputIntent()
}

// InputPressIntent 按钮按下意图
type InputPressIntent struct {
	X, Y   int
	Button runtimemsg.MouseButton
	Source string // "mouse" | "keyboard"
}

func (InputPressIntent) isInputIntent() {}

// InputReleaseIntent 按钮释放意图
type InputReleaseIntent struct {
	X, Y   int
	Button runtimemsg.MouseButton
	Source string // "mouse" | "keyboard"
}

func (InputReleaseIntent) isInputIntent() {}

// InputMoveIntent 鼠标移动意图
type InputMoveIntent struct {
	X, Y int
}

func (InputMoveIntent) isInputIntent() {}

// InputKeyboardIntent 键盘输入意图
type InputKeyboardIntent struct {
	Key     rune
	Special runtimeplatform.SpecialKey
	Mod     runtimemsg.Modifiers
}

func (InputKeyboardIntent) isInputIntent() {}

// NewInputTracker 创建新的输入跟踪器
func NewInputTracker() *InputTracker {
	return &InputTracker{
		lastSnapshot: &InputSnapshot{},
		lastIntent:   nil,
	}
}

// Update 更新输入状态并返回推断的意图
//
// 核心推断逻辑：
// - prev.Button == 0 && curr.Button != 0 → 推断 Press
// - prev.Button != 0 && curr.Button == 0 → 推断 Release
// - prev.{X,Y} != curr.{X,Y} → 推断 Move
func (t *InputTracker) Update(snapshot *InputSnapshot) []InputIntent {
	if snapshot == nil {
		return nil
	}

	var intents []InputIntent

	// 鼠标状态推断
	t.inferMouseState(t.lastSnapshot, snapshot, &intents)

	// 键盘状态推断
	t.inferKeyboardState(t.lastSnapshot, snapshot, &intents)

	// 更新状态
	t.lastSnapshot = snapshot.Clone()
	t.lastIntent = nil

	return intents
}

// inferMouseState 推断鼠标状态变化
func (t *InputTracker) inferMouseState(prev, curr *InputSnapshot, intents *[]InputIntent) {
	// 推断 Press：之前无按钮，现在有按钮
	if prev.MouseButton == runtimemsg.MouseButtonUnknown && curr.MouseButton != runtimemsg.MouseButtonUnknown {
		// 只在实际操作时才推断 Press（忽略状态同步）
		if curr.MouseAction == runtimemsg.MouseActionPress {
			*intents = append(*intents, InputPressIntent{
				X:      curr.MouseX,
				Y:      curr.MouseY,
				Button: curr.MouseButton,
				Source: "mouse",
			})
		}
	}

	// 推断 Release：之前有按钮，现在无按钮
	if prev.MouseButton != runtimemsg.MouseButtonUnknown && curr.MouseButton == runtimemsg.MouseButtonUnknown {
		if curr.MouseAction == runtimemsg.MouseActionRelease {
			*intents = append(*intents, InputReleaseIntent{
				X:      curr.MouseX,
				Y:      curr.MouseY,
				Button: prev.MouseButton,
				Source: "mouse",
			})
		}
	}

	// 推断 Move：位置变化
	if prev.MouseX != curr.MouseX || prev.MouseY != curr.MouseY {
		*intents = append(*intents, InputMoveIntent{
			X: curr.MouseX,
			Y: curr.MouseY,
		})
	}
}

// inferKeyboardState 推断键盘状态变化
func (t *InputTracker) inferKeyboardState(prev, curr *InputSnapshot, intents *[]InputIntent) {
	// 推断键盘输入：有新按键或特殊键
	if curr.KeyboardKey != 0 || curr.SpecialKey != runtimeplatform.KeyUnknown {
		// 只在实际按键时才触发（忽略状态同步）
		if !curr.IsEmpty() {
			*intents = append(*intents, InputKeyboardIntent{
				Key:     curr.KeyboardKey,
				Special: curr.SpecialKey,
				Mod:     curr.Modifiers,
			})
		}
	}
}
