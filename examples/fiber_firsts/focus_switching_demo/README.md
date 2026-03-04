# Focus Switching Demo

This demo demonstrates focus management in the Mint UI framework using the **Store + Reducer Architecture**.

## Architecture: Store + Reducer (Single Source of Truth)

The demo follows the recommended architecture for Mint UI applications:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Store + Reducer Architecture               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User Input → FieldChangeIntent → Reducer → Store → View        │
│                                                                 │
│  Components:                                                    │
│    - Store[T]: Single source of truth                          │
│    - Reducer[T]: Pure function for state transformations        │
│    - ForField(): Automatic Intent emission                      │
│    - Value(state): Read latest state from Store               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Principles

| Principle | Description |
|-----------|-------------|
| **Single Source of Truth** | All state stored in one `Store[T]` |
| **Pure State Updates** | Only Reducer can modify state |
| **Automatic Re-render** | Store updates trigger view re-render |
| **No Closure Dependencies** | View reads from Store, not from closures |

### Data Flow

```
┌──────────┐   User Input    ┌─────────────┐   Intent      ┌─────────┐
│  Input   │ ───────────────► │  Instance   │ ───────────► │Dispatch │
└──────────┘                 └─────────────┘               └─────────┘
                                    │                       │
                                    ▼                       │
                               ForField() API              │
                                    │                       │
                                    ▼                       │
                            FieldChangeIntent              │
                                    │                       │
                                    ▼                       │
┌─────────┐   Reducer     ┌──────────┐   New State   ┌─────────┐
│  View   │ ◄───────────── │  Store   │ ◄──────────── │ Handler │
└─────────┘               └──────────┘               └─────────┘
```

---

## Features

### Focusable Components

| Component | Count | Focus ID Format | Intent |
|-----------|-------|-----------------|--------|
| Button | 3 | `button:{key}` | `ClickButtonIntent{}` |
| Input | 3 | `input:{key}` | `FieldChangeIntent{Field, Value}` |
| Checkbox | 3 | `checkbox:{key}` | `FieldChangeIntent{Field, Value}` |

### Navigation

- **TAB** - Move to next focusable element
- **SHIFT+TAB** - Move to previous focusable element
- **ENTER** - Activate focused button/checkbox
- **SPACE** - Toggle focused checkbox
- **ESC / CTRL+C** - Exit the app

---

## Implementation

### 1. Define Application State

```go
// AppState is the single source of truth
type AppState struct {
    Input1     string
    Input2     string
    Input3     string
    Checked1   string
    Checked2   string
    Checked3   string
    ClickCount int
    ActiveTab  int
}
```

### 2. Create Global Store

```go
var appStore *store.Store[AppState]

func initStore() {
    appStore = store.NewStore(AppState{
        Input1: "", Input2: "", Input3: "Disabled Input",
        Checked1: "false", Checked2: "false", Checked3: "false",
        ClickCount: 0, ActiveTab: 0,
    })
}
```

### 3. Define Reducer

```go
var appReducer = reducer.NewBuilder[AppState]().
    On(ClickButtonIntent{}, func(s AppState, i intent.Intent) AppState {
        s.ClickCount++
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            switch fieldChange.Field {
            case "input1-value", "input2-value":
                // Update state directly
                if fieldChange.Field == "input1-value" {
                    s.Input1 = fieldChange.Value
                } else {
                    s.Input2 = fieldChange.Value
                }
            case "checked1", "checked2":
                // Checkbox boolean stored as string
                if fieldChange.Field == "checked1" {
                    s.Checked1 = fieldChange.Value  // "true" or "false"
                } else {
                    s.Checked2 = fieldChange.Value
                }
            }
        }
        return s
    })
```

### 4. Register Handlers

```go
func registerHandlers() {
    // BuildAndRegister automatically registers handlers to global registry
    // Each handler will:
    //   1. Run Reducer to compute new state
    //   2. Update Store
    //   3. Call ScheduleUpdate() to trigger re-render
    appReducer.RegisterToGlobal(appStore)
}
```

### 5. View - Read from Store

```go
func FocusApp() ui.VNode {
    // Get current state from Store (always latest value)
    state := appStore.Get()

    return ui.VStack(
        // Button - emits ClickButtonIntent
        buttonComp.NewBuilder("Button 1 - First").
            OnPress(ClickButtonIntent{}).
            Build().
            SetKey("btn1"),

        // Input - reads from Store
        inputComp.NewBuilder().
            ForField(intent.BindField("input1-value")).  // Auto emit FieldChangeIntent
            Value(state.Input1).                               // Display latest value
            Placeholder("Enter name...").
            Build().
            SetKey("input1"),

        // Checkbox - reads from Store
        checkboxComp.NewBuilder().
            Label("Option A").
            ForField(intent.BindField("checked1")).   // Auto emit FieldChangeIntent
            Checked(state.Checked1 == "true").           // Display boolean
            Build().
            SetKey("chk1"),
        // ...
    )
}
```

---

## Complete Data Flow Example

### Button Click

```
1. User presses ENTER on focused button
        |
        ▼
2. Button Instance emits ClickButtonIntent
        |
        ▼
3. Handler receives ClickButtonIntent
        |
        ▼
4. Reducer runs: s.ClickCount++ → newState
        |
        ▼
5. Store updates: store.Set(newState)
        |
        ▼
6. All listeners notified → ScheduleUpdate()
        |
        ▼
7. Re-render → View reads store.Get() → UI updates
```

### Input Field Typing

```
1. User types 'a' in input
        |
        ▼
2. Input Instance buffers: inst.value = "a"
        |
        ▼
3. ForField config emits FieldChangeIntent{Field: "input1-value", Value: "a"}
        |
        ▼
4. Handler receives FieldChangeIntent
        |
        ▼
5. Reducer updates: s.Input1 = "a" → newState
        |
        ▼
6. Store updates: store.Set(newState)
        |
        ▼
7. Re-render → View reads store.Get() → Input displays "a"
```

---

## Using ui.Run()

The demo uses `ui.Run()` to start the application:

```go
func main() {
    initStore()  // Initialize global Store
    
    err := ui.Run(FocusApp,
        ui.WithWidth(60),
        ui.WithHeight(35),
        ui.WithTitle("Focus Switching Demo"),
        ui.WithInit(registerHandlers),  // Register Reducer handlers
    )
    if err != nil {
        panic(err)
    }
}
```

---

## Intent Usage

### Button - OnPress()

```go
// Use custom Intent (recommended)
buttonComp.NewBuilder("Button 1 - First").
    OnPress(ClickButtonIntent{}).
    Build().
    SetKey("btn1")
```

**Note**: Do NOT use `Click` intent - it has no built-in handler.

### Input - ForField() + Value()

```go
// Component automatically emits FieldChangeIntent
inputComp.NewBuilder().
    ForField(intent.BindField("field")).  // Auto emit Intent
    Value(state.Field).                      // Display from Store
    Placeholder("Enter text...").
    Build()
```

### Checkbox - ForField() + Checked()

```go
// Component automatically emits FieldChangeIntent
checkboxComp.NewBuilder().
    Label("Option A").
    ForField(intent.BindField("checked")).
    Checked(state.Checked == "true").  // Display boolean from Store
    Build()
```

---

## Available Intents with Handlers

| Intent | Handler | Purpose |
|--------|--------|---------|
| `FieldChange` | ✅ `handleFieldChange` | Form field input |
| `SetState` | ✅ `handleSetState` | Set global state value |
| `Toggle` | ✅ `handleToggle` | Toggle boolean state |
| `Increment` | ✅ `handleIncrement` | Increment numeric state |
| `Navigate` | ✅ `handleNavigate` | Page navigation |
| `Focus`/`Blur` | ✅ `handleFocus`/`handleBlur` | Focus management |
| `OpenModal`/`CloseModal` | ✅ `handleOpenModal`/`handleCloseModal` | Modal management |

**Note**: `Click` and `Press` intents DO NOT have built-in handlers. Use custom intents instead.

---

## Running the Demo

```bash
cd examples/fiber_firsts/focus_switching_demo
go run main.go
```

Or build and run:

```bash
go build -o focus_demo.exe main.go
./focus_demo.exe
```

---

## Key Benefits of Store + Reducer Architecture

### Before: UseState + GlobalState (Complex)

```go
// ❌ 5 steps + type assertions
value, setValue := ui.UseStateString("")
ctx.GlobalState["setter"] = setValue
ui.WithInit(func() {
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        if fn, ok := ctx.GetState("setter"); ok {
            if setter, ok := fn.(func(string)); ok {  // Type assertion
                setter(i.Value)
            }
        }
        return intent.HandledResult()
    })
}, ...)
input.ForField(...).Value(value).Build()
```

**Problems**:
- ❌ Complex setter marshaling
- ❌ Type assertions required
- ❌ Timing dependencies (WithInit before render)
- ❌ Multiple state sources (UseState + GlobalState + Instance)

### After: Store + Reducer (Simple)

```go
// ✅ 3 steps, no type assertions
// Store (initialized once)
appStore := store.NewStore(AppState{Field: ""})

// Reducer (defined once)
reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Field = i.(intent.FieldChangeIntent).Value
        return s
    }).RegisterToGlobal(appStore)

// View (pure function)
state := appStore.Get()
input.ForField(intent.BindField("field")).Value(state.Field).Build()
```

**Benefits**:
- ✅ Single state source (Store only)
- ✅ No type assertions
- ✅ No timing dependencies
- ✅ Predictable state updates
- ✅ Easy to test (Reducer is pure function)

---

## Comparison with Other Demos

| Demo | Architecture | Complexity |
|------|-------------|------------|
| `store_reducer_demo` | Store + Reducer | ✅ Simple |
| `focus_switching_demo` | Store + Reducer | ✅ Simple |
| `validation_demo` | UseState + GlobalState | ❌ Complex |
| `mvp_form_demo` | UseState + GlobalState | ❌ Complex |
| `typesafe_form_demo` | UseState + GlobalState | ❌ Complex |

**Recommendation**: Use **Store + Reducer** architecture for all new applications.
