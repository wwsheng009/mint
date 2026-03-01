// Package cmd provides command types for the Elm-style architecture.
//
// Cmd represents side effects or asynchronous operations, such as:
// - I/O operations
// - Delayed execution
// - Timers
// - Batch execution of multiple commands
//
// Cmd is lazy and only executes when returned.
package cmd

import (
	"time"
)

// Cmd is the command interface.
//
// Cmd represents potential side effects or async operations.
type Cmd interface {
	// Type returns the command type (for debugging).
	Type() CmdType
}

// CmdType represents the command type.
type CmdType string

const (
	// CmdTypeNone is an empty command (does nothing).
	CmdTypeNone CmdType = "none"

	// CmdTypeBatch is a batch command.
	CmdTypeBatch CmdType = "batch"

	// CmdTypeAfter is a delayed command.
	CmdTypeAfter CmdType = "after"

	// CmdTypeTick is a timer command.
	CmdTypeTick CmdType = "tick"

	// CmdTypeIO is an I/O operation command.
	CmdTypeIO CmdType = "io"
)

// NoneCmd represents an empty command (does nothing).
type NoneCmd struct{}

// Type returns the command type.
func (n *NoneCmd) Type() CmdType {
	return CmdTypeNone
}

// None returns an empty command.
func None() Cmd {
	return &NoneCmd{}
}

// BatchCmd represents executing multiple commands in sequence.
type BatchCmd struct {
	cmds []Cmd
}

// Type returns the command type.
func (b *BatchCmd) Type() CmdType {
	return CmdTypeBatch
}

// Batch creates a batch command that executes multiple commands in sequence.
func Batch(cmds ...Cmd) Cmd {
	// Filter out None commands
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

// AfterCmd represents a delayed execution command.
type AfterCmd struct {
	Duration time.Duration
	Cmd      Cmd
}

// Type returns the command type.
func (a *AfterCmd) Type() CmdType {
	return CmdTypeAfter
}

// After creates a delayed command that executes after the specified time.
func After(duration time.Duration, c Cmd) Cmd {
	if c.Type() == CmdTypeNone {
		return None()
	}
	return &AfterCmd{
		Duration: duration,
		Cmd:      c,
	}
}

// TickCmd represents a timer command.
type TickCmd struct {
	Duration time.Duration
	Msg      interface{} // msg.Msg
}

// Type returns the command type.
func (t *TickCmd) Type() CmdType {
	return CmdTypeTick
}

// Tick creates a timer command that sends a message every specified interval.
func Tick(duration time.Duration, m interface{}) Cmd {
	if m == nil {
		return None()
	}
	return &TickCmd{
		Duration: duration,
		Msg:      m,
	}
}

// IOCmd represents an I/O operation command.
type IOCmd struct {
	Operation func() interface{} // (msg.Msg)
}

// Type returns the command type.
func (i *IOCmd) Type() CmdType {
	return CmdTypeIO
}

// IO creates an I/O operation command.
func IO(operation func() interface{}) Cmd {
	if operation == nil {
		return None()
	}
	return &IOCmd{
		Operation: operation,
	}
}

// Execute executes a command and returns any messages produced.
// This is a helper function for executing commands at runtime.
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
		// Delayed commands need async execution
		// Returns nil here; actual execution should be handled by runtime
		go func() {
			time.Sleep(cmd.Duration)
			Execute(cmd.Cmd)
		}()
		return nil

	case *TickCmd:
		// Timer commands need async execution
		// Returns nil here; actual execution should be handled by runtime
		go func() {
			ticker := time.NewTicker(cmd.Duration)
			defer ticker.Stop()
			for range ticker.C {
				// Send cmd.Msg to app
				// Actual implementation needs channel communication
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
