# Focus Switching Demo

This demo demonstrates focus management in the Mint UI framework using the `ui.Run()` entry point.

## Features

### Focusable Components

| Component | Count | Focus ID Format | Intent Method |
|-----------|-------|-----------------|---------------|
| Button | 3 | `button:{key}` | `.SetIntent(intent.Click(targetID))` |
| Input | 2 | `input:{key}` | `.SetChangeIntent(intent.SetState(key, value))` |
| Checkbox | 2 | `checkbox:{key}` | `.SetIntent(intent.Toggle(key))` |

### Navigation

- **TAB** - Move to next focusable element
- **SHIFT+TAB** - Move to previous focusable element
- **ENTER** - Activate focused button/checkbox
- **SPACE** - Toggle focused checkbox
- **ESC / CTRL+C** - Exit the app

## Using ui.Run()

The demo uses `ui.Run()` to start the application:

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui"
    buttonComp "github.com/wwsheng009/mint/ui/components/button"
    checkboxComp "github.com/wwsheng009/mint/ui/components/checkbox"
    inputComp "github.com/wwsheng009/mint/ui/components/input"
    newstack "github.com/wwsheng009/mint/ui/components/stack"
)

func SimpleApp() ui.VNode {
    // Button
    btn := buttonComp.New("Click Me")
    btn.SetIntent(intent.Click("btn1"))
    btn.SetKey("btn1")

    // Input
    input := inputComp.New()
    input.SetPlaceholder("Enter text...")
    input.SetChangeIntent(intent.SetState("input-value", ""))
    input.SetKey("input1")

    // Checkbox
    chk := checkboxComp.New("Enable option")
    chk.SetIntent(intent.Toggle("checkbox-checked"))
    chk.SetKey("chk1")

    return newstack.New(newstack.Column).
        SetChildrenList([]ui.VNode{
            btn,
            input,
            chk,
        })
}

func main() {
    err := ui.Run(SimpleApp,
        ui.WithWidth(60),
        ui.WithHeight(35),
        ui.WithTitle("Focus Switching Demo"),
    )
    if err != nil {
        panic(err)
    }
}
```

## Intent Usage Summary

### Button - SetIntent()

```go
btn := buttonComp.New("Submit")
btn.SetIntent(intent.Click("btn1"))
```

### Checkbox - SetIntent()

```go
chk := checkboxComp.New("Remember me")
chk.SetIntent(intent.Toggle("remember"))
```

### Input - SetChangeIntent()

```go
input := inputComp.New()
input.SetChangeIntent(intent.SetState("user-name", ""))
```

### Available Intents

| Intent | Constructor | Priority |
|--------|-------------|----------|
| Click | `intent.Click(targetID)` | UserBlocking |
| Press | `intent.Press(targetID)` | UserBlocking |
| SetState | `intent.SetState(key, value)` | Normal |
| Toggle | `intent.Toggle(key)` | UserBlocking |
| Focus | `intent.Focus(targetID)` | Immediate |
| Blur | `intent.Blur(targetID)` | Immediate |
| Navigate | `intent.Navigate(path)` | UserBlocking |
| OpenModal | `intent.OpenModal(modalID)` | UserBlocking |
| CloseModal | `intent.CloseModal(modalID)` | UserBlocking |

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
