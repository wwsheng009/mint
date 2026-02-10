package action

import (
	"fmt"
	"strings"
	"sync"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
)

// KeyMap 键盘映射表
//
// KeyMap 用于将键盘事件映射为语义化的 Action
// 支持全局映射和上下文相关映射
//
// 使用示例：
//   km := NewKeyMap()
//   km.Bind("ctrl+c", ActionCopy)
//   km.BindWithContext("input", "ctrl+v", ActionPaste)
//   action := km.LookupKeyEvent(keyEvent)
type KeyMap struct {
	mu             sync.RWMutex
	globalMappings map[string]*Action  // 全局映射
	contextStack   []string             // 上下文栈
	contextMaps    map[string]map[string]*Action // 上下文相关映射
}

// KeySignature 按键签名
type KeySignature struct {
	Key      rune   // 字符键
	Special  string // 特殊键（如 "enter", "tab"）
	Modifiers Modifier // 修饰键
	Context  string // 上下文（可选）
}

// Modifier 键盘修饰键
type Modifier int

const (
	ModNone Modifier = iota
	ModCtrl
	ModAlt
	ModShift
	ModMeta
)

// String 返回修饰键的字符串表示
func (m Modifier) String() string {
	var parts []string

	if m&ModCtrl != 0 {
		parts = append(parts, "ctrl")
	}
	if m&ModAlt != 0 {
		parts = append(parts, "alt")
	}
	if m&ModShift != 0 {
		parts = append(parts, "shift")
	}
	if m&ModMeta != 0 {
		parts = append(parts, "meta")
	}

	if len(parts) == 0 {
		return "none"
	}
	return strings.ToLower(strings.Join(parts, "+"))
}

// NewKeyMap 创建新的键盘映射
func NewKeyMap() *KeyMap {
	return &KeyMap{
		globalMappings: make(map[string]*Action),
		contextStack:   make([]string, 0),
		contextMaps:    make(map[string]map[string]*Action),
	}
}

// Bind 绑定按键到 Action
//
// keySpec 格式: "ctrl+c", "alt+f", "shift+tab", "enter", "space"
// 大小写不敏感
//
// 示例:
//   km.Bind("ctrl+c", ActionCopy)
//   km.Bind("enter", ActionSelect)
func (km *KeyMap) Bind(keySpec string, action *Action) error {
	signature, err := parseKeySpec(keySpec)
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	key := signature.toString()
	km.globalMappings[key] = action

	return nil
}

// BindWithContext 在特定上下文中绑定按键
//
// context: 上下文标识符（如 "input", "tree", "list"）
// keySpec: 按键规格（同 Bind）
// action: 要执行的动作
//
// 示例:
//   km.BindWithContext("input", "ctrl+v", ActionPaste)
func (km *KeyMap) BindWithContext(context, keySpec string, action *Action) error {
	signature, err := parseKeySpec(keySpec)
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	if km.contextMaps[context] == nil {
		km.contextMaps[context] = make(map[string]*Action)
	}

	// 不把上下文包含在键中，而是分别存储
	key := signature.toString()
	km.contextMaps[context][key] = action

	return nil
}

// LookupKeyEvent 查找键盘事件对应的 Action
//
// 查找顺序：
// 1. 当前上下文映射（如果有）
// 2. 全局映射
// 3. 返回 nil（未找到）
func (km *KeyMap) LookupKeyEvent(ev *frameworkevent.KeyEvent) *Action {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// 构建按键签名
	signature := KeySignature{
		Key:      ev.Key.Rune,
		Special:  ev.Special.String(), // 使用 String() 方法获取字符串
		Modifiers: parseKeyModifiers(ev),
	}

	key := signature.toString()

	// 1. 尝试当前上下文映射
	currentContext := km.getCurrentContext()
	if currentContext != "" {
		if ctxMap, ok := km.contextMaps[currentContext]; ok {
			if action, ok := ctxMap[key]; ok {
				return action.Clone().WithSource("keymap")
			}
		}
	}

	// 2. 尝试全局映射
	if action, ok := km.globalMappings[key]; ok {
		return action.Clone().WithSource("keymap")
	}

	// 3. 未找到映射
	return nil
}

// Lookup 直接通过按键规格查找 Action
func (km *KeyMap) Lookup(keySpec string) *Action {
	signature, err := parseKeySpec(keySpec)
	if err != nil {
		return nil
	}

	km.mu.RLock()
	defer km.mu.RUnlock()

	key := signature.toString()
	if action, ok := km.globalMappings[key]; ok {
		return action.Clone()
	}

	return nil
}

// PushContext 推入新的上下文
func (km *KeyMap) PushContext(context string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.contextStack = append(km.contextStack, context)
}

// PopContext 弹出当前上下文
func (km *KeyMap) PopContext() {
	km.mu.Lock()
	defer km.mu.Unlock()

	if len(km.contextStack) > 0 {
		km.contextStack = km.contextStack[:len(km.contextStack)-1]
	}
}

// SetCurrentContext 设置当前上下文（清空栈）
func (km *KeyMap) SetCurrentContext(context string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.contextStack = make([]string, 0)
	if context != "" {
		km.contextStack = append(km.contextStack, context)
	}
}

// GetCurrentContext 获取当前上下文
func (km *KeyMap) GetCurrentContext() string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return km.getCurrentContext()
}

// getCurrentContext 获取当前上下文（内部方法，不加锁）
func (km *KeyMap) getCurrentContext() string {
	if len(km.contextStack) == 0 {
		return ""
	}
	return km.contextStack[len(km.contextStack)-1]
}

// Unbind 解除按键绑定
func (km *KeyMap) Unbind(keySpec string) error {
	signature, err := parseKeySpec(keySpec)
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	key := signature.toString()
	delete(km.globalMappings, key)

	return nil
}

// UnbindWithContext 解除特定上下文的按键绑定
func (km *KeyMap) UnbindWithContext(context, keySpec string) error {
	signature, err := parseKeySpec(keySpec)
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	// 不把上下文包含在键中
	key := signature.toString()

	if ctxMap, ok := km.contextMaps[context]; ok {
		delete(ctxMap, key)
	}

	return nil
}

// Clear 清空所有映射
func (km *KeyMap) Clear() {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.globalMappings = make(map[string]*Action)
	km.contextMaps = make(map[string]map[string]*Action)
	km.contextStack = make([]string, 0)
}

// Size 返回全局映射的数量
func (km *KeyMap) Size() int {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return len(km.globalMappings)
}

// ============================================================================
// KeySignature 方法
// ============================================================================

// toString 将按键签名转换为字符串键
func (ks KeySignature) toString() string {
	var parts []string

	// 修饰键
	if ks.Modifiers != ModNone {
		parts = append(parts, ks.Modifiers.String())
	}

	// 特殊键或字符键
	if ks.Special != "" {
		parts = append(parts, strings.ToLower(ks.Special))
	} else if ks.Key != 0 {
		parts = append(parts, strings.ToLower(string(ks.Key)))
	}

	// 不包含上下文在键中（上下文用于查找，不在键本身）
	return strings.Join(parts, "+")
}

// ============================================================================
// 按键规格解析
// ============================================================================

// parseKeySpec 解析按键规格字符串
//
// 格式: [modifiers+]+key
// 例如: "ctrl+c", "alt+shift+f5", "enter", "space"
func parseKeySpec(keySpec string) (KeySignature, error) {
	spec := strings.ToLower(strings.TrimSpace(keySpec))

	if spec == "" {
		return KeySignature{}, fmt.Errorf("empty key spec")
	}

	signature := KeySignature{
		Modifiers: ModNone,
	}

	parts := strings.Split(spec, "+")

	// 解析每个部分
	for i, part := range parts {
		part = strings.TrimSpace(part)

		// 最后一个部分是键本身
		if i == len(parts)-1 {
			if isSpecialKey(part) {
				signature.Special = part
			} else if len(part) == 1 {
				signature.Key = rune(part[0])
			} else {
				return KeySignature{}, fmt.Errorf("invalid key: %s", part)
			}
		} else {
			// 修饰键
			switch part {
			case "ctrl", "control":
				signature.Modifiers |= ModCtrl
			case "alt":
				signature.Modifiers |= ModAlt
			case "shift":
				signature.Modifiers |= ModShift
			case "meta", "cmd", "win":
				signature.Modifiers |= ModMeta
			default:
				return KeySignature{}, fmt.Errorf("unknown modifier: %s", part)
			}
		}
	}

	return signature, nil
}

// isSpecialKey 判断是否是特殊键
func isSpecialKey(key string) bool {
	specialKeys := map[string]bool{
		"enter": true, "return": true, "tab": true,
		"space": true, "escape": true, "esc": true,
		"up": true, "down": true, "left": true, "right": true,
		"home": true, "end": true,
		"page up": true, "page down": true, "pgup": true, "pgdn": true,
		"insert": true, "delete": true,
		"f1": true, "f2": true, "f3": true, "f4": true,
		"f5": true, "f6": true, "f7": true, "f8": true,
		"f9": true, "f10": true, "f11": true, "f12": true,
		"backspace": true,
	}

	return specialKeys[key]
}

// parseKeyModifiers 从 KeyEvent 解析修饰键
func parseKeyModifiers(ev *frameworkevent.KeyEvent) Modifier {
	var mod Modifier

	if ev.Key.Ctrl {
		mod |= ModCtrl
	}
	if ev.Key.Alt {
		mod |= ModAlt
	}
	if ev.Key.Shift {
		mod |= ModShift
	}

	return mod
}

// ============================================================================
// 预定义的默认映射
// ============================================================================

// DefaultKeyMap 返回默认的键盘映射
func DefaultKeyMap() *KeyMap {
	km := NewKeyMap()

	// 导航
	km.Bind("up", NewAction(ActionNavigateUp))
	km.Bind("down", NewAction(ActionNavigateDown))
	km.Bind("left", NewAction(ActionNavigateLeft))
	km.Bind("right", NewAction(ActionNavigateRight))
	km.Bind("page up", NewAction(ActionNavigatePageUp))
	km.Bind("page down", NewAction(ActionNavigatePageDown))
	km.Bind("home", NewAction(ActionNavigateHome))
	km.Bind("end", NewAction(ActionNavigateEnd))

	// 编辑
	km.Bind("enter", NewAction(ActionEnter))
	km.Bind("tab", NewAction(ActionNavigateNext))
	km.Bind("backspace", NewAction(ActionBackspace))
	km.Bind("delete", NewAction(ActionDeleteChar))
	km.Bind("escape", NewAction(ActionCancel))

	// 功能键
	km.Bind("f1", NewAction(ActionInspect))
	km.Bind("f5", NewAction(ActionRefresh))
	km.Bind("f10", NewAction(ActionQuit))

	// Ctrl 组合
	km.Bind("ctrl+c", NewAction(ActionCopy))
	km.Bind("ctrl+v", NewAction(ActionPaste))
	km.Bind("ctrl+x", NewAction(ActionCut))
	km.Bind("ctrl+f", NewAction(ActionSearch))
	km.Bind("ctrl+q", NewAction(ActionQuit))
	km.Bind("ctrl+s", NewAction(ActionSubmit))

	// Alt 组合
	km.Bind("alt+space", NewAction(ActionToggle))

	return km
}

// ============================================================================
// 调试和诊断
// ============================================================================

// Dump 导出所有映射（用于调试）
func (km *KeyMap) Dump() string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== KeyMap ===\n")

	// 全局映射
	sb.WriteString("\nGlobal Mappings:\n")
	for key, action := range km.globalMappings {
		fmt.Fprintf(&sb, "  %s -> %s\n", key, action.Type)
	}

	// 上下文映射
	for context, ctxMap := range km.contextMaps {
		fmt.Fprintf(&sb, "\nContext '%s':\n", context)
		for key, action := range ctxMap {
			fmt.Fprintf(&sb, "  %s -> %s\n", key, action.Type)
		}
	}

	// 当前上下文
	if currentContext := km.getCurrentContext(); currentContext != "" {
		fmt.Fprintf(&sb, "\nCurrent Context: %s\n", currentContext)
	}

	return sb.String()
}
