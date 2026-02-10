package msg

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/framework/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// SandboxMsg 表示测试沙箱注入的消息
//
// SandboxMsg 用于测试环境中的消息注入，支持：
// - 注入键盘输入
// - 注入鼠标事件
// - 注入语义化的 Action
// - 直接修改组件状态
type SandboxMsg struct {
	runtimemsg.BaseMsg // 嵌入 runtime/msg 的 BaseMsg

	// InjectType 表示注入的类型
	InjectType SandboxInjectType

	// KeyData 是键盘注入数据（仅当 InjectType = SandboxInjectKey 时有效）
	KeyData *runtimemsg.KeyMsg

	// MouseData 是鼠标注入数据（仅当 InjectType = SandboxInjectMouse 时有效）
	MouseData *runtimemsg.MouseMsg

	// ActionData 是 Action 注入数据（仅当 InjectType = SandboxInjectAction 时有效）
	ActionData *action.Action

	// StateMutation 是状态修改数据（仅当 InjectType = SandboxInjectState 时有效）
	StateMutation *StateMutation
}

// SandboxInjectType 表示沙箱注入的类型
type SandboxInjectType string

const (
	// SandboxInjectKey 注入键盘输入
	SandboxInjectKey SandboxInjectType = "key"

	// SandboxInjectMouse 注入鼠标事件
	SandboxInjectMouse SandboxInjectType = "mouse"

	// SandboxInjectAction 注入语义化 Action
	SandboxInjectAction SandboxInjectType = "action"

	// SandboxInjectState 直接修改状态
	SandboxInjectState SandboxInjectType = "state"
)

// StateMutation 表示状态修改
type StateMutation struct {
	TargetID string                 // 目标组件 ID
	Path     string                 // 状态路径（如 "value", "items.0"）
	Value    interface{}            // 新值
	Metadata map[string]interface{} // 额外元数据
}

// NewSandboxKeyMsg 创建键盘注入消息
func NewSandboxKeyMsg(keyMsg *runtimemsg.KeyMsg) *SandboxMsg {
	return &SandboxMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue:      runtimemsg.MsgTypeSandbox,
			TimestampValue: time.Now(),
		},
		InjectType: SandboxInjectKey,
		KeyData:    keyMsg,
	}
}

// NewSandboxMouseMsg 创建鼠标注入消息
func NewSandboxMouseMsg(mouseMsg *runtimemsg.MouseMsg) *SandboxMsg {
	return &SandboxMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue:      runtimemsg.MsgTypeSandbox,
			TimestampValue: time.Now(),
		},
		InjectType: SandboxInjectMouse,
		MouseData:  mouseMsg,
	}
}

// NewSandboxActionMsg 创建 Action 注入消息
func NewSandboxActionMsg(act *action.Action) *SandboxMsg {
	return &SandboxMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue:      runtimemsg.MsgTypeSandbox,
			TimestampValue: time.Now(),
		},
		InjectType: SandboxInjectAction,
		ActionData: act,
	}
}

// NewSandboxStateMsg 创建状态修改消息
func NewSandboxStateMsg(targetID, path string, value interface{}) *SandboxMsg {
	return &SandboxMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue:      runtimemsg.MsgTypeSandbox,
			TimestampValue: time.Now(),
		},
		InjectType: SandboxInjectState,
		StateMutation: &StateMutation{
			TargetID: targetID,
			Path:     path,
			Value:    value,
		},
	}
}

// Type 实现 runtimemsg.Msg 接口
func (s *SandboxMsg) Type() runtimemsg.MsgType {
	return s.BaseMsg.Type()
}

// Timestamp 实现 runtimemsg.Msg 接口
func (s *SandboxMsg) Timestamp() time.Time {
	return s.BaseMsg.Timestamp()
}

// String 实现 runtimemsg.Msg 接口
func (s *SandboxMsg) String() string {
	switch s.InjectType {
	case SandboxInjectKey:
		if s.KeyData != nil {
			return fmt.Sprintf("SandboxMsg{Key: %s}", s.KeyData.String())
		}
		return "SandboxMsg{Key: <nil>}"

	case SandboxInjectMouse:
		if s.MouseData != nil {
			return fmt.Sprintf("SandboxMsg{Mouse: %s}", s.MouseData.String())
		}
		return "SandboxMsg{Mouse: <nil>}"

	case SandboxInjectAction:
		if s.ActionData != nil {
			return fmt.Sprintf("SandboxMsg{Action: %s}", s.ActionData.String())
		}
		return "SandboxMsg{Action: <nil>}"

	case SandboxInjectState:
		if s.StateMutation != nil {
			return fmt.Sprintf("SandboxMsg{State: %s.%s = %v}",
				s.StateMutation.TargetID,
				s.StateMutation.Path,
				s.StateMutation.Value)
		}
		return "SandboxMsg{State: <nil>}"

	default:
		return "SandboxMsg{<unknown>}"
	}
}

// IsInput 检查是否为输入类注入（键盘或鼠标）
func (s *SandboxMsg) IsInput() bool {
	return s.InjectType == SandboxInjectKey || s.InjectType == SandboxInjectMouse
}

// IsDirectAction 检查是否为直接 Action 注入
func (s *SandboxMsg) IsDirectAction() bool {
	return s.InjectType == SandboxInjectAction
}

// IsStateMutation 检查是否为状态修改
func (s *SandboxMsg) IsStateMutation() bool {
	return s.InjectType == SandboxInjectState
}
