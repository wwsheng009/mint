# Inspector Event Routing Fix - WORKING! ✅

## Test Results

The automatic event routing is now **fully functional**!

```
=== RUN   TestAutomaticEventRouting
    Inspector registered with framework app (visible=true)
    Initial focus index: 0

    === Testing Automatic Event Routing (Down Arrow) ===
[APP] Routing key 'down' to Inspector (visible=true)
[APP] Inspector handled key 'down'
[APP] Routing key 'down' to Inspector (visible=true)
[APP] Inspector handled key 'down'
[APP] Routing key 'down' to Inspector (visible=true)
[APP] Inspector handled key 'down'
    After 3 Down arrows, focus index: 3  ✅

    === Testing Automatic Event Routing (Up Arrow) ===
[APP] Routing key 'up' to Inspector (visible=true)
[APP] Inspector handled key 'up'
[APP] Routing key 'up' to Inspector (visible=true)
[APP] Inspector handled key 'up'
    After 2 Up arrows, focus index: 1  ✅

    === Testing Automatic Event Routing (PageDown) ===
[APP] Routing key 'pagedown' to Inspector (visible=true)
[APP] Inspector handled key 'pagedown'
    After PageDown, focus index: 6  ✅

    === Testing Automatic Event Routing (Home) ===
[APP] Routing key 'home' to Inspector (visible=true)
[APP] Inspector handled key 'home'
    After Home, focus index: 0  ✅

    === Testing Automatic Event Routing (End) ===
[APP] Routing key 'end' to Inspector (visible=true)
[APP] Inspector handled key 'end'
    After End, focus index: 6  ✅

--- PASS: TestAutomaticEventRouting (1.67s)
```

---

## The Critical Missing Piece

### Why It Wasn't Working Initially

The test was creating an Inspector but **NOT registering it** with the framework app!

**Wrong** (doesn't work):
```go
insp := inspector.NewStandaloneInspector()
insp.Enable()
insp.ToggleVisibility()

testApp, err := ui.RunTest(func() ui.VNode {
    return insp.RenderOverlay()
}, ...)
// Inspector exists but framework doesn't know about it!
```

**Correct** (works!):
```go
insp := inspector.NewStandaloneInspector()
insp.Enable()
insp.ToggleVisibility()

testApp, err := ui.RunTest(func() ui.VNode {
    return insp.RenderOverlay()
}, ...)

// CRITICAL: Register Inspector with framework app!
fwApp := testApp.GetFrameworkApp()
fwApp.SetInspector(insp)
fwApp.SetupInspectorShortcut()
```

---

## Complete Working Example

```go
package main

import (
    "testing"
    "time"

    "github.com/wwsheng009/mint/internal/inspector"
    "github.com/wwsheng009/mint/runtime/platform"
    ui "github.com/wwsheng009/mint/ui"
)

func TestAutomaticEventRouting(t *testing.T) {
    // Step 1: Create Inspector
    insp := inspector.NewStandaloneInspector()
    insp.Enable()
    insp.ToggleVisibility()

    // Step 2: Create test content
    testRoot := ui.VStack(
        ui.Text("Root Node"),
        ui.Text("Node 1"),
        ui.Text("Node 2"),
    )
    insp.AttachToApp(testRoot)

    // Step 3: Create testable app
    testApp, err := ui.RunTest(func() ui.VNode {
        return insp.RenderOverlay()
    }, ui.WithWidth(120), ui.WithHeight(40))
    if err != nil {
        t.Fatalf("Failed to create test app: %v", err)
    }
    defer testApp.Close()

    // Step 4: CRITICAL - Register Inspector with framework!
    fwApp := testApp.GetFrameworkApp()
    fwApp.SetInspector(insp)           // ← REQUIRED!
    fwApp.SetupInspectorShortcut()     // ← Enables F12 toggle

    time.Sleep(100 * time.Millisecond)

    // Step 5: Test automatic routing - NO MANUAL HandleKeyEvent CALLS!
    treeView := insp.GetTreeViewComponent()
    initialFocus := treeView.GetFocusIndex()
    t.Logf("Initial focus: %d", initialFocus)

    // Inject keys - Inspector automatically receives them!
    testApp.InjectSpecialKey(platform.KeyDown)
    time.Sleep(150 * time.Millisecond)

    newFocus := treeView.GetFocusIndex()
    t.Logf("After KeyDown: focus = %d", newFocus)

    if newFocus <= initialFocus {
        t.Errorf("Focus should have moved from %d to > 0", initialFocus)
    }

    t.Log("✅ Automatic event routing works!")
}
```

---

## Event Flow (Working)

```
1. testApp.InjectSpecialKey(platform.KeyDown)
   ↓
2. Creates platform.RawInput{Type: InputKeyPress, Special: KeyDown}
   ↓
3. Calls fwApp.InjectEvent(raw)
   ↓
4. Pump.Inject(raw) → convertToEvent() → KeyEvent{Key.Name: "down"}
   ↓
5. Pump sends event to events channel
   ↓
6. Framework handleEvent() receives KeyEvent
   ↓
7. Check keyMap shortcuts → Not found
   ↓
8. Check if Inspector visible → YES!
   ↓
9. Inspector.HandleKeyEvent("down", false, false)
   ↓
10. TreeView.HandleKey(KeyDown, 0)
    ↓
11. Focus moves from 0 → 1
    ↓
12. Returns true (handled)
    ↓
13. Event NOT sent to VNode tree
```

**Debug output confirms this**:
```
[APP] Routing key 'down' to Inspector (visible=true)
[APP] Inspector handled key 'down'
```

---

## Production Apps Need This Too!

### For Production Apps:

**File**: `examples/your_app/main.go`

```go
package main

import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/framework/theme"
    "github.com/wwsheng009/mint/internal/inspector"
    "github.com/wwsheng009/mint/internal/render"
    ui "github.com/wwsheng009/mint/ui"
)

func main() {
    _ = theme.SetTheme("nord")

    // Create Inspector
    insp := inspector.NewStandaloneInspector()
    insp.Enable()

    // Create framework app
    fwApp := framework.NewApp()
    fwApp.Resize(100, 35)
    fwApp.InitTheme("dark")

    // CRITICAL: Register Inspector with framework
    fwApp.SetInspector(insp)           // ← REQUIRED for automatic routing!
    fwApp.SetupInspectorShortcut()     // ← Enables F12/Ctrl+D toggle

    // Set root
    declarativeRoot := render.NewDeclarativeNodeFromFunc(App)
    declarativeRoot.SetFrameworkApp(fwApp)
    fwApp.SetRoot(declarativeRoot)

    // Run
    fwApp.Run()
}

func App() ui.VNode {
    return ui.VStack(
        ui.Text("Press F12 to toggle Inspector"),
        ui.Text("Arrow keys navigate tree when Inspector is visible"),
    )
}
```

**Without `fwApp.SetInspector(insp)`**, the framework won't route events to Inspector!

---

## Why ui.Run() Doesn't Work Yet

The current `ui.Run()` helper doesn't auto-create Inspector:

```go
// Current implementation in ui/app.go
func Run(app ComponentFunc, opts ...Option) error {
    fwApp := framework.NewApp()
    fwApp.Resize(options.Width, options.Height)

    // NO Inspector creation here!
    // NO fwApp.SetInspector() call!

    declarativeNode := render.NewDeclarativeNodeFromFunc(app)
    fwApp.SetRoot(declarativeNode)
    return fwApp.Run()
}
```

**Workaround**: Use framework app directly instead of `ui.Run()`:

```go
// Instead of:
err := ui.Run(App, ui.WithWidth(100), ui.WithHeight(35))

// Use:
insp := inspector.NewStandaloneInspector()
insp.Enable()

fwApp := framework.NewApp()
fwApp.Resize(100, 35)
fwApp.SetInspector(insp)           // ← Manual setup required
fwApp.SetupInspectorShortcut()

declarativeRoot := render.NewDeclarativeNodeFromFunc(App)
declarativeRoot.SetFrameworkApp(fwApp)
fwApp.SetRoot(declarativeRoot)
fwApp.Run()
```

---

## Summary

### ✅ What's Working

1. **Framework event routing** - Automatically routes keyboard events to Inspector when visible
2. **Inspector registration** - `SetInspector()` makes framework aware of Inspector
3. **Test injection** - TestableApp.InjectSpecialKey() routes to Inspector automatically
4. **All navigation keys** - Arrow keys, PageUp/PageDown, Home/End all work
5. **PageUp/PageDown strings** - Fixed mismatch, accepts both "pageup"/"pgup"

### ⚠️ Critical Requirements

**For Tests**:
```go
fwApp := testApp.GetFrameworkApp()
fwApp.SetInspector(insp)           // ← REQUIRED!
```

**For Production Apps**:
```go
fwApp := framework.NewApp()
fwApp.SetInspector(insp)           // ← REQUIRED!
fwApp.SetupInspectorShortcut()
```

### 📋 Next Steps

1. ✅ Framework event routing - WORKING
2. ✅ Tests demonstrate automatic routing - WORKING
3. ⚠️ Need to update `ui.Run()` to auto-create Inspector (optional convenience)
4. ⚠️ Need to fix demo2/main.go to create Inspector (optional)

---

## Test Command

```bash
# Run with debug output
TUI_DEBUG_UI=true go test -v ./examples/ui_demos/demo2_runtime_internals -run TestAutomaticEventRouting

# Run without debug
go test -v ./examples/ui_demos/demo2_runtime_internals -run TestAutomaticEventRouting
```

**Expected result**: All tests pass, focus moves correctly, automatic event routing confirmed! 🎉
