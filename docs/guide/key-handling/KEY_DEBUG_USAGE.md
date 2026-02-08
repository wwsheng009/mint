# How to Use Key Debug Feature

## Step 1: Run the App
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go
```

## Step 2: Look at the Inspector Title Bar

When the app starts, you'll see:
```
╔═ INSPECTOR ═╗
F12:关闭 | Alt+H/J/K/L:移动 | Ctrl+D:按键调试
🔍 Last key: '' (无)
```

The third line shows the **last key that was pressed**.

## Step 3: Press F12

This shows/hides the Inspector. Make sure it's **visible**.

## Step 4: Press Some Keys

Try these keys while the Inspector is visible:

| Press | You Should See |
|-------|----------------|
| **k** | `🔍 Last key: 'k' (无)` |
| **Alt+k** | `🔍 Last key: 'k' (Alt+)` |
| **F12** | `🔍 Last key: 'f12' (无)` |
| **Ctrl+d** | `🔍 Last key: 'd' (Ctrl+)` |

## What It Means

- **'k'** - The key name or character
- **(无)** - No modifiers (Chinese for "none")
- **(Alt+)** - Alt was held down
- **(Ctrl+)** - Ctrl was held down
- **(Shift+)** - Shift was held down

## Example Session

```
1. App starts:  🔍 Last key: '' (无)
2. Press 'a':   🔍 Last key: 'a' (无)
3. Press 'k':   🔍 Last key: 'k' (无)
4. Hold Alt + press 'k':  🔍 Last key: 'k' (Alt+)
5. Press F12:   🔍 Last key: 'f12' (无)
6. Press Arrow Up:  🔍 Last key: 'up' (无)
7. Hold Alt + press Arrow Up:  🔍 Last key: 'up' (Alt+)
```

## Debug Output in Terminal

If you want to see debug logs:
```bash
set TUI_DEBUG_UI=true
set TUI_INSPECTOR_VERBOSE=true
go run main.go
```

Then when you press keys, you'll see:
```
[Inspector] Key received: key='k' modifiers=Alt+ showKeyDebug=false
```

## Important Notes

1. **Inspector must be visible** - The key debug only shows when Inspector is displayed
2. **Press F12 first** - Make Inspector visible before testing keys
3. **Real-time updates** - The display updates immediately when you press keys
4. **Shows LAST key** - It shows the previous key pressed, not current key being held

## Test Results

From the automated test:
```
✅ Inspector handled 'k' key
✅ Inspector handled Alt+K
✅ Inspector handled Ctrl+D
✅ Inspector handled F12
```

All keys are being received correctly!
