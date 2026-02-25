package action

import (
	"fmt"
	"strings"
	"sync"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// KeyMap 键盘映射表
//
// KeyMap 用于将键盘消息映射为语义化的 Action
// 支持全局映射和上下文相关映射
//
// Usage:
//   km := NewKeyMap()
//   km.Bind("ctrl+c", ActionCopy)
//   km.BindWithContext("input", "ctrl+v", ActionPaste)
//   action := km.LookupKeyMsg(keyMsg)
type KeyMap struct {
	mu             sync.RWMutex
	globalMappings map[string]*Action                 // Global mappings
	contextStack   []string                            // Context stack
	contextMaps    map[string]map[string]*Action       // Context-specific mappings
}

// KeyMapping is a simplified binding record
type KeyMapping struct {
	KeySpec string       // "ctrl+c", "enter", etc.
	Action *Action       // Target action
	Context string       // Optional context
}

// KeySignature 按键签名
type KeySignature struct {
	Key       rune   // Character key
	Special   string // Special key (e.g., "enter", "tab")
	Modifiers Modifier
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

// String returns modifier string representation
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

// NewKeyMap creates a new KeyMap
func NewKeyMap() *KeyMap {
	return &KeyMap{
		globalMappings: make(map[string]*Action),
		contextStack:   make([]string, 0),
		contextMaps:    make(map[string]map[string]*Action),
	}
}

// Bind binds a key spec to an Action
//
// keySpec format: "ctrl+c", "alt+f", "shift+tab", "enter", "space"
// Case insensitive
//
// Examples:
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

// BindWithContext binds a key spec in specific context
//
// context: context identifier (e.g., "input", "tree", "list")
// keySpec: key spec (same as Bind)
// action: action to execute
//
// Examples:
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

	key := signature.toString()
	km.contextMaps[context][key] = action

	return nil
}

// Add adds a key mapping
func (km *KeyMap) Add(mapping KeyMapping) {
	if mapping.Context != "" {
		km.BindWithContext(mapping.Context, mapping.KeySpec, mapping.Action)
	} else {
		km.Bind(mapping.KeySpec, mapping.Action)
	}
}

// AddAll adds multiple key mappings
func (km *KeyMap) AddAll(mappings []KeyMapping) {
	for _, mapping := range mappings {
		km.Add(mapping)
	}
}

// LookupKeyMsg looks up action for keyboard message
//
// Lookup order:
// 1. Current context mapping (if any)
// 2. Global mapping
// 3. Returns nil (not found)
func (km *KeyMap) LookupKeyMsg(keyMsg *runtimemsg.KeyMsg) *Action {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Build key signature
	signature := KeySignature{
		Key:       keyMsg.Rune,
		Special:   specialKeyToString(keyMsg.Special),
		Modifiers: parseKeyMsgModifiers(keyMsg),
	}

	key := signature.toString()

	// 1. Try current context mapping
	currentContext := km.getCurrentContext()
	if currentContext != "" {
		if ctxMap, ok := km.contextMaps[currentContext]; ok {
			if action, ok := ctxMap[key]; ok {
				return action.Clone().WithSource("keymap")
			}
		}
	}

	// 2. Try global mapping
	if action, ok := km.globalMappings[key]; ok {
		return action.Clone().WithSource("keymap")
	}

	// 3. Not found
	return nil
}

// LookupKeyEvent keeps backward compatibility (deprecated)
// Use LookupKeyMsg instead
func (km *KeyMap) LookupKeyEvent(key rune, special runtimeplatform.SpecialKey, modifiers runtimemsg.Modifiers) *Action {
	keyMsg := runtimemsg.NewKeyMsg(key, special, modifiers)
	return km.LookupKeyMsg(keyMsg)
}

// Lookup looks up action directly by key spec
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

// PushContext pushes a new context
func (km *KeyMap) PushContext(context string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.contextStack = append(km.contextStack, context)
}

// PopContext pops current context
func (km *KeyMap) PopContext() {
	km.mu.Lock()
	defer km.mu.Unlock()

	if len(km.contextStack) > 0 {
		km.contextStack = km.contextStack[:len(km.contextStack)-1]
	}
}

// SetCurrentContext sets current context (clears stack)
func (km *KeyMap) SetCurrentContext(context string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.contextStack = make([]string, 0)
	if context != "" {
		km.contextStack = append(km.contextStack, context)
	}
}

// GetCurrentContext returns current context
func (km *KeyMap) GetCurrentContext() string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return km.getCurrentContext()
}

// getCurrentContext returns current context (internal, no lock)
func (km *KeyMap) getCurrentContext() string {
	if len(km.contextStack) == 0 {
		return ""
	}
	return km.contextStack[len(km.contextStack)-1]
}

// Unbind unbinds a key
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

// UnbindWithContext unbinds a key in specific context
func (km *KeyMap) UnbindWithContext(context, keySpec string) error {
	signature, err := parseKeySpec(keySpec)
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	key := signature.toString()

	if ctxMap, ok := km.contextMaps[context]; ok {
		delete(ctxMap, key)
	}

	return nil
}

// Clear clears all mappings
func (km *KeyMap) Clear() {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.globalMappings = make(map[string]*Action)
	km.contextMaps = make(map[string]map[string]*Action)
	km.contextStack = make([]string, 0)
}

// Size returns global mapping count
func (km *KeyMap) Size() int {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return len(km.globalMappings)
}

// ============================================================================
// KeySignature Methods
// ============================================================================

// toString converts key signature to string key
func (ks KeySignature) toString() string {
	var parts []string

	// Modifiers
	if ks.Modifiers != ModNone {
		parts = append(parts, ks.Modifiers.String())
	}

	// Special key or character key
	if ks.Special != "" {
		parts = append(parts, strings.ToLower(ks.Special))
	} else if ks.Key != 0 {
		parts = append(parts, strings.ToLower(string(ks.Key)))
	}

	return strings.Join(parts, "+")
}

// ============================================================================
// Key Spec Parsing
// ============================================================================

// parseKeySpec parses key spec string
//
// Format: [modifiers+]+key
// Examples: "ctrl+c", "alt+shift+f5", "enter", "space"
func parseKeySpec(keySpec string) (KeySignature, error) {
	spec := strings.ToLower(strings.TrimSpace(keySpec))

	if spec == "" {
		return KeySignature{}, fmt.Errorf("empty key spec")
	}

	signature := KeySignature{
		Modifiers: ModNone,
	}

	parts := strings.Split(spec, "+")

	for i, part := range parts {
		part = strings.TrimSpace(part)

		// Last part is the key itself
		if i == len(parts)-1 {
			if isSpecialKey(part) {
				signature.Special = part
			} else if len(part) == 1 {
				signature.Key = rune(part[0])
			} else {
				return KeySignature{}, fmt.Errorf("invalid key: %s", part)
			}
		} else {
			// Modifier
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

// isSpecialKey checks if key is a special key
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

// parseKeyMsgModifiers parses modifiers from KeyMsg
func parseKeyMsgModifiers(keyMsg *runtimemsg.KeyMsg) Modifier {
	var mod Modifier

	if keyMsg.Mod.Ctrl {
		mod |= ModCtrl
	}
	if keyMsg.Mod.Alt {
		mod |= ModAlt
	}
	if keyMsg.Mod.Shift {
		mod |= ModShift
	}

	return mod
}

// specialKeyToString converts runtimeplatform.SpecialKey to string
func specialKeyToString(special runtimeplatform.SpecialKey) string {
	switch special {
	case runtimeplatform.KeyUp:
		return "up"
	case runtimeplatform.KeyDown:
		return "down"
	case runtimeplatform.KeyLeft:
		return "left"
	case runtimeplatform.KeyRight:
		return "right"
	case runtimeplatform.KeyEnter:
		return "enter"
	case runtimeplatform.KeyTab:
		return "tab"
	case runtimeplatform.KeyBackspace:
		return "backspace"
	case runtimeplatform.KeyDelete:
		return "delete"
	case runtimeplatform.KeyEscape:
		return "escape"
	case runtimeplatform.KeyHome:
		return "home"
	case runtimeplatform.KeyEnd:
		return "end"
	case runtimeplatform.KeyPageUp:
		return "page up"
	case runtimeplatform.KeyPageDown:
		return "page down"
	case runtimeplatform.KeyF1:
		return "f1"
	case runtimeplatform.KeyF2:
		return "f2"
	case runtimeplatform.KeyF3:
		return "f3"
	case runtimeplatform.KeyF4:
		return "f4"
	case runtimeplatform.KeyF5:
		return "f5"
	case runtimeplatform.KeyF6:
		return "f6"
	case runtimeplatform.KeyF7:
		return "f7"
	case runtimeplatform.KeyF8:
		return "f8"
	case runtimeplatform.KeyF9:
		return "f9"
	case runtimeplatform.KeyF10:
		return "f10"
	case runtimeplatform.KeyF11:
		return "f11"
	case runtimeplatform.KeyF12:
		return "f12"
	default:
		return ""
	}
}

// ============================================================================
// Predefined Default Mappings
// ============================================================================

// DefaultKeyMap returns default keyboard mappings
func DefaultKeyMap() *KeyMap {
	km := NewKeyMap()

	// Navigation
	km.Bind("up", NewAction(ActionNavigateUp))
	km.Bind("down", NewAction(ActionNavigateDown))
	km.Bind("left", NewAction(ActionNavigateLeft))
	km.Bind("right", NewAction(ActionNavigateRight))
	km.Bind("page up", NewAction(ActionNavigatePageUp))
	km.Bind("page down", NewAction(ActionNavigatePageDown))
	km.Bind("home", NewAction(ActionNavigateHome))
	km.Bind("end", NewAction(ActionNavigateEnd))

	// Editing
	km.Bind("enter", NewAction(ActionEnter))
	km.Bind("tab", NewAction(ActionNavigateNext))
	km.Bind("backspace", NewAction(ActionBackspace))
	km.Bind("delete", NewAction(ActionDeleteChar))
	km.Bind("escape", NewAction(ActionCancel))

	// Function keys
	km.Bind("f1", NewAction(ActionInspect))
	km.Bind("f5", NewAction(ActionRefresh))
	km.Bind("f10", NewAction(ActionQuit))

	// Ctrl combinations
	km.Bind("ctrl+c", NewAction(ActionCopy))
	km.Bind("ctrl+v", NewAction(ActionPaste))
	km.Bind("ctrl+x", NewAction(ActionCut))
	km.Bind("ctrl+f", NewAction(ActionSearch))
	km.Bind("ctrl+q", NewAction(ActionQuit))
	km.Bind("ctrl+s", NewAction(ActionSubmit))
	km.Bind("ctrl+a", NewAction(ActionNavigateHome))
	km.Bind("ctrl+e", NewAction(ActionNavigateEnd))

	// Alt combinations
	km.Bind("alt+space", NewAction(ActionToggle))

	// Shift+Tab for previous navigation
	km.Bind("shift+tab", NewAction(ActionNavigatePrev))

	return km
}

// ============================================================================
// Debugging and Diagnosis
// ============================================================================

// Dump exports all mappings (for debugging)
func (km *KeyMap) Dump() string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== KeyMap ===\n")

	// Global mappings
	sb.WriteString("\nGlobal Mappings:\n")
	for key, action := range km.globalMappings {
		fmt.Fprintf(&sb, "  %s -> %s\n", key, action.Type)
	}

	// Context mappings
	for context, ctxMap := range km.contextMaps {
		fmt.Fprintf(&sb, "\nContext '%s':\n", context)
		for key, action := range ctxMap {
			fmt.Fprintf(&sb, "  %s -> %s\n", key, action.Type)
		}
	}

	// Current context
	if currentContext := km.getCurrentContext(); currentContext != "" {
		fmt.Fprintf(&sb, "\nCurrent Context: %s\n", currentContext)
	}

	return sb.String()
}

// GetContexts returns all registered contexts
func (km *KeyMap) GetContexts() []string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	contexts := make([]string, 0, len(km.contextMaps))
	for ctx := range km.contextMaps {
		contexts = append(contexts, ctx)
	}
	return contexts
}

// GetMappingsForContext returns mappings for a specific context
func (km *KeyMap) GetMappingsForContext(context string) map[string]*Action {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if ctxMap, ok := km.contextMaps[context]; ok {
		// Copy the map
		result := make(map[string]*Action, len(ctxMap))
		for k, v := range ctxMap {
			result[k] = v
		}
		return result
	}
	return nil
}
