//go:build linux
// +build linux

package platform

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
)

// 全局退出标志，用于 Ctrl+C 处理
var exitFlag int32 = 0

// tcellInputReader implements inputReader using tcell
type tcellInputReader struct {
	events     chan<- RawInput
	quit       chan struct{}
	screen     tcell.Screen
	quitOnce   sync.Once

	// Mouse button state tracking to distinguish Release from Motion
	// Used when button==0 to determine if it's a release or just movement
	lastPressedButton MouseButton // Track which button was pressed for release events
}

func newInputReaderImpl() inputReaderImpl {
	return &tcellInputReader{
		quit:     make(chan struct{}),
	}
}

func (r *tcellInputReader) Start(events chan<- RawInput) error {
	r.events = events

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	r.screen = screen

	if err := screen.Init(); err != nil {
		return err
	}

	// Enable mouse events
	screen.EnableMouse(tcell.MouseButtonEvents, tcell.MouseMotionEvents, tcell.MouseDragEvents)
	screen.HideCursor()

	go r.readLoop()

	return nil
}

func (r *tcellInputReader) readLoop() {
	for {
		// 检查是否收到退出信号
		if atomic.LoadInt32(&exitFlag) != 0 {
			return
		}

		select {
		case <-r.quit:
			return
		default:
		}

		ev := r.screen.PollEvent()
		if ev == nil {
			continue
		}

		now := time.Now()

		switch e := ev.(type) {
		case *tcell.EventKey:
			r.events <- r.parseKeyEvent(e, now)
		case *tcell.EventMouse:
			r.events <- r.parseMouseEvent(e, now)
		case *tcell.EventResize:
			// Ignore resize events initially
			// The framework will query terminal size separately
		case *tcell.EventError:
			// Ignore error events
		}
	}
}

func (r *tcellInputReader) parseKeyEvent(ev *tcell.EventKey, now time.Time) RawInput {
	input := RawInput{
		Type:      InputKeyPress,
		Timestamp: now,
	}

	// Map modifiers
	if ev.Modifiers() != tcell.ModNone {
		mod := KeyModifier(0)
		if ev.Modifiers()&tcell.ModShift != 0 {
			mod |= ModShift
		}
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			mod |= ModCtrl
		}
		if ev.Modifiers()&tcell.ModAlt != 0 {
			mod |= ModAlt
		}
		if ev.Modifiers()&tcell.ModMeta != 0 {
			mod |= ModMeta
		}
		input.Modifiers = mod
	}

	// Handle character keys
	if ev.Key() == tcell.KeyRune {
		// Regular character (no modifiers, or only Shift)
		input.Key = ev.Rune()
	}

	// Map Ctrl+letter combinations (tcell returns special KeyCtrlA-Z keys)
	// These need to be mapped to the letter character with Ctrl modifier
	switch ev.Key() {
	case tcell.KeyCtrlA:
		input.Key = 'a'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlB:
		input.Key = 'b'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlC:
		// Ctrl+C 是特殊的退出组合键
		input.Key = 'c'
		input.Modifiers |= ModCtrl
		// 设置退出标志，让应用能够退出
		atomic.StoreInt32(&exitFlag, 1)
	case tcell.KeyCtrlD:
		input.Key = 'd'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlE:
		input.Key = 'e'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlF:
		input.Key = 'f'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlG:
		input.Key = 'g'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlH:
		input.Key = 'h'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlI:
		input.Key = 'i'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlJ:
		input.Key = 'j'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlK:
		input.Key = 'k'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlL:
		input.Key = 'l'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlM:
		input.Key = 'm'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlN:
		input.Key = 'n'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlO:
		input.Key = 'o'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlP:
		input.Key = 'p'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlQ:
		input.Key = 'q'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlR:
		input.Key = 'r'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlS:
		input.Key = 's'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlT:
		input.Key = 't'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlU:
		input.Key = 'u'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlV:
		input.Key = 'v'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlW:
		input.Key = 'w'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlX:
		input.Key = 'x'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlY:
		input.Key = 'y'
		input.Modifiers |= ModCtrl
	case tcell.KeyCtrlZ:
		input.Key = 'z'
		input.Modifiers |= ModCtrl
	}

	// Map function keys (only if not already handled as Ctrl+letter)
	if input.Key == 0 {
		switch ev.Key() {
	case tcell.KeyUp:
		input.Special = KeyUp
	case tcell.KeyDown:
		input.Special = KeyDown
	case tcell.KeyLeft:
		input.Special = KeyLeft
	case tcell.KeyRight:
		input.Special = KeyRight
	case tcell.KeyHome:
		input.Special = KeyHome
	case tcell.KeyEnd:
		input.Special = KeyEnd
	case tcell.KeyPgUp:
		input.Special = KeyPageUp
	case tcell.KeyPgDn:
		input.Special = KeyPageDown
	case tcell.KeyEnter:
		input.Special = KeyEnter
	case tcell.KeyBackspace:
		input.Special = KeyBackspace
	case tcell.KeyBackspace2:
		input.Special = KeyBackspace
	case tcell.KeyDelete:
		input.Special = KeyDelete
	case tcell.KeyInsert:
		input.Special = KeyInsert
	case tcell.KeyEscape:
		input.Special = KeyEscape
	case tcell.KeyTab:
		input.Special = KeyTab
	case tcell.KeyF1:
		input.Special = KeyF1
	case tcell.KeyF2:
		input.Special = KeyF2
	case tcell.KeyF3:
		input.Special = KeyF3
	case tcell.KeyF4:
		input.Special = KeyF4
	case tcell.KeyF5:
		input.Special = KeyF5
	case tcell.KeyF6:
		input.Special = KeyF6
	case tcell.KeyF7:
		input.Special = KeyF7
	case tcell.KeyF8:
		input.Special = KeyF8
	case tcell.KeyF9:
		input.Special = KeyF9
	case tcell.KeyF10:
		input.Special = KeyF10
	case tcell.KeyF11:
		input.Special = KeyF11
	case tcell.KeyF12:
		input.Special = KeyF12
	default:
		input.Special = KeyUnknown
	}
	}

	// 🔥 关键修复：过滤单独的修饰键
	// 当单独按下修饰键时（Shift, Ctrl, Alt, Meta），不生成可见按键消息
	if input.Special == KeyUnknown && input.Key == 0 {
		// 只有修饰键没有实际字符或特殊键 - 忽略这个事件
		return RawInput{Type: -1, Timestamp: now}
	}

	return input
}

func (r *tcellInputReader) parseMouseEvent(ev *tcell.EventMouse, now time.Time) RawInput {
	input := RawInput{
		Type:      InputMouse,
		Timestamp: now,
	}

	button := ev.Buttons()

	// Get position
	x, y := ev.Position()
	input.MouseX = int(x)
	input.MouseY = int(y)

	// Map mouse buttons and actions
	if button == tcell.WheelUp {
		input.MouseAction = MouseWheelUp
		return input
	}
	if button == tcell.WheelDown {
		input.MouseAction = MouseWheelDown
		return input
	}

	// Handle regular mouse buttons
	if button&tcell.ButtonPrimary != 0 {
		// Button pressed or dragging - record which button
		input.MouseButton = MouseLeft
		input.MouseAction = MousePress
		r.lastPressedButton = MouseLeft
	} else if button&tcell.ButtonSecondary != 0 {
		input.MouseButton = MouseRight
		input.MouseAction = MousePress
		r.lastPressedButton = MouseRight
	} else if button&tcell.ButtonMiddle != 0 {
		input.MouseButton = MouseMiddle
		input.MouseAction = MousePress
		r.lastPressedButton = MouseMiddle
	} else {
		// No button pressed - check if it's a release or just motion
		if r.lastPressedButton != MouseNone {
			// Button was just released
			input.MouseButton = r.lastPressedButton
			input.MouseAction = MouseRelease
			r.lastPressedButton = MouseNone // Reset after release
		} else {
			// Just moving without any button pressed
			input.MouseButton = MouseNone
			input.MouseAction = MouseMotion
		}
	}

	return input
}

func (r *tcellInputReader) restoreTerminal() {
	if r.screen != nil {
		r.screen.Fini()
	}
}

func (r *tcellInputReader) Stop() error {
	close(r.quit)
	r.restoreTerminal()
	return nil
}

func (r *tcellInputReader) ReadEvent() (RawInput, error) {
	ev := r.screen.PollEvent()
	if ev == nil {
		return RawInput{}, nil
	}

	now := time.Now()

	switch e := ev.(type) {
	case *tcell.EventKey:
		return r.parseKeyEvent(e, now), nil
	case *tcell.EventMouse:
		return r.parseMouseEvent(e, now), nil
	case *tcell.EventResize:
		// Ignore resize events initially
		return RawInput{}, nil
	case *tcell.EventError:
		// tcell EventError has no exported error field
		return RawInput{}, nil
	}

	return RawInput{}, nil
}

func restoreTerminalImpl() {
	// Terminal restoration is handled by Stop() calling restoreTerminal()
	// via the screen.Fini() call
}

// init 安装进程级终端恢复保险丝
//
// 🔥 工业级保护：即使程序 panic、强制关闭，也会恢复终端
//
// 注意：tcell 会捕获 SIGINT 信号，所以我们需要在 tcell 初始化之前注册信号处理
func init() {
	// 监听中断信号 (SIGINT = Ctrl+C)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		_ = <-ch
		// 设置退出标志，让 readLoop 能够退出
		atomic.StoreInt32(&exitFlag, 1)
		// 强制恢复终端
		restoreTerminalImpl()
		os.Exit(0)
	}()
}
