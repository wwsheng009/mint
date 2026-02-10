package cmd

import (
	"time"
)

// Cmd 是命令的接口
//
// Cmd 表示可能产生的副作用或异步操作，例如：
// - 执行 I/O 操作
// - 延迟执行某个命令
// - 定时器
// - 批量执行多个命令
//
// Cmd 是惰性的，只有被返回时才会执行。
type Cmd interface {
	// Type 返回命令的类型（用于调试）
	Type() CmdType
}

// CmdType 表示命令的类型
type CmdType string

const (
	// CmdTypeNone 空命令（什么都不做）
	CmdTypeNone CmdType = "none"

	// CmdTypeBatch 批量命令
	CmdTypeBatch CmdType = "batch"

	// CmdTypeAfter 延迟命令
	CmdTypeAfter CmdType = "after"

	// CmdTypeTick 定时器命令
	CmdTypeTick CmdType = "tick"

	// CmdTypeIO I/O 操作命令
	CmdTypeIO CmdType = "io"
)

// NoneCmd 表示空命令（什么都不做）
type NoneCmd struct{}

// Type 返回命令类型
func (n *NoneCmd) Type() CmdType {
	return CmdTypeNone
}

// None 返回一个空命令
func None() Cmd {
	return &NoneCmd{}
}

// BatchCmd 表示批量执行多个命令
type BatchCmd struct {
	cmds []Cmd
}

// Type 返回命令类型
func (b *BatchCmd) Type() CmdType {
	return CmdTypeBatch
}

// Batch 创建一个批量命令，按顺序执行多个命令
func Batch(cmds ...Cmd) Cmd {
	// 过滤掉 None 命令
	validCmds := make([]Cmd, 0, len(cmds))
	for _, c := range cmds {
		if c.Type() != CmdTypeNone {
			validCmds = append(validCmds, c)
		}
	}

	if len(validCmds) == 0 {
		return None()
	}
	if len(validCmds) == 1 {
		return validCmds[0]
	}

	return &BatchCmd{cmds: validCmds}
}

// AfterCmd 表示延迟执行的命令
type AfterCmd struct {
	Duration time.Duration
	Cmd      Cmd
}

// Type 返回命令类型
func (a *AfterCmd) Type() CmdType {
	return CmdTypeAfter
}

// After 创建一个延迟命令，在指定时间后执行
func After(duration time.Duration, c Cmd) Cmd {
	if c.Type() == CmdTypeNone {
		return None()
	}
	return &AfterCmd{
		Duration: duration,
		Cmd:      c,
	}
}

// TickCmd 表示定时器命令
type TickCmd struct {
	Duration time.Duration
	Msg      interface{} // msg.Msg
}

// Type 返回命令类型
func (t *TickCmd) Type() CmdType {
	return CmdTypeTick
}

// Tick 创建一个定时器命令，每隔指定时间发送一条消息
func Tick(duration time.Duration, m interface{}) Cmd {
	if m == nil {
		return None()
	}
	return &TickCmd{
		Duration: duration,
		Msg:      m,
	}
}

// IOCmd 表示 I/O 操作命令
type IOCmd struct {
	Operation func() interface{} // msg.Msg
}

// Type 返回命令类型
func (i *IOCmd) Type() CmdType {
	return CmdTypeIO
}

// IO 创建一个 I/O 操作命令
func IO(operation func() interface{}) Cmd {
	if operation == nil {
		return None()
	}
	return &IOCmd{
		Operation: operation,
	}
}

// Execute 执行命令并返回产生的消息（如果有的话）
// 这是一个辅助函数，用于在运行时执行命令
func Execute(c Cmd) []interface{} {
	if c == nil {
		return nil
	}

	switch cmd := c.(type) {
	case *NoneCmd:
		return nil

	case *BatchCmd:
		var msgs []interface{}
		for _, subCmd := range cmd.cmds {
			subMsgs := Execute(subCmd)
			msgs = append(msgs, subMsgs...)
		}
		return msgs

	case *AfterCmd:
		// 延迟命令需要异步执行
		// 这里返回 nil，实际执行应该由运行时处理
		go func() {
			time.Sleep(cmd.Duration)
			Execute(cmd.Cmd)
		}()
		return nil

	case *TickCmd:
		// 定时器命令需要异步执行
		// 这里返回 nil，实际执行应该由运行时处理
		go func() {
			ticker := time.NewTicker(cmd.Duration)
			defer ticker.Stop()
			for range ticker.C {
				// 将 cmd.Msg 发送到应用
				// 实际实现需要通过消息通道
				_ = cmd.Msg
			}
		}()
		return nil

	case *IOCmd:
		if cmd.Operation != nil {
			m := cmd.Operation()
			if m != nil {
				return []interface{}{m}
			}
		}
		return nil

	default:
		return nil
	}
}
