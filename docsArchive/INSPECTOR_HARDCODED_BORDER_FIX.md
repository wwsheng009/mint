# Inspector Hardcoded Border Fix

## Problem

The Inspector panel used **hardcoded Unicode border characters** instead of the BorderedNode component's built-in border rendering.

### Before (BROKEN - Inconsistent)

```go
// Title bar with hardcoded border characters
titleBarParts := []rtui.VNode{
    app.NewTextBuilder("╔═ INSPECTOR ═╗").  // ❌ Hardcoded!
        Style(style.NewStyle().Bold(true)...).
        Build(),
    // ...
}

// Panel with BorderedNode (but wrong parameter!)
panel := rtui.Bordered().
    Style(string(theme.Border())).  // ❌ theme.Border() returns a Color, not a style string!
    Child(content).
    Build()
```

### Issues

1. ❌ **Inconsistent design**: Title uses hardcoded text, panel uses BorderedNode
2. ❌ **Wrong parameter**: `theme.Border()` returns a **Color**, but `Style()` expects a **style string**
3. ❌ **Hard to maintain**: Changing border style requires changing code in multiple places
4. ❌ **Theme unaware**: Hardcoded characters don't respect theme settings

## Solution

Use BorderedNode's `Label()` feature for the title, and use correct style string.

### After (FIXED - Consistent)

```go
// Title bar WITHOUT hardcoded border characters
titleBarParts := []rtui.VNode{
    app.NewTextBuilder("F12:关闭 | ...").  // ✅ No border chars
        Style(style.NewStyle()...).
        Build(),
    // ...
}

// Panel with BorderedNode using Label feature
panel := rtui.Bordered().
    Style("double").      // ✅ Correct border style string
    Label("INSPECTOR").   // ✅ Uses Label feature
    Child(content).
    Build()
```

## Changes Made

### File: `internal/inspector/standalone_inspector.go`

#### Change 1: Removed hardcoded border text (line 271)

```diff
- app.NewTextBuilder("╔═ INSPECTOR ═╗").
-     Style(style.NewStyle().Bold(true).Foreground(style.Blue).Background(style.Yellow).Reverse(true)).
-     Build(),
```

#### Change 2: Added Label to BorderedNode (line 329)

```diff
  panel := rtui.Bordered().
-     Style(string(theme.Border())).
+     Style("double").
+     Label("INSPECTOR").
      Child(content).
```

#### Change 3: Fixed buildOverlayContainer (line 355)

```diff
  overlay := rtui.Bordered().
-     Style(string(theme.Border())).
+     Style("double").
      Child(mainContent).
```

## Technical Details

### BorderedNode API

```go
// Style sets the border LINE STYLE
// Valid values: "single", "double", "rounded", "dashed", "none"
func (b *BorderedBuilder) Style(style string) *BorderedBuilder

// Color sets the border COLOR
// Valid values: "red", "blue", "green", etc.
func (b *BorderedBuilder) Color(c string) *BorderedBuilder

// Label sets a title shown on the top border
func (b *BorderedBuilder) Label(label string) *BorderedBuilder
```

### Common Mistake: theme.Border()

**WRONG**:
```go
Style(string(theme.Border()))  // ❌ theme.Border() returns a Color!
```

**CORRECT**:
```go
Style("double")   // ✅ Style expects a style string
Color("blue")     // ✅ Color expects a color string
```

## Result

Now the Inspector panel border is **consistently** rendered by the BorderedNode component:

```
╔═ INSPECTOR ═╗
│ content...   │
╚═════════════╝
```

- ✅ No hardcoded Unicode characters in title
- ✅ Border drawn by BorderedNode component
- ✅ Correct use of `Style("double")`
- ✅ Title set using `Label("INSPECTOR")`
- ✅ Consistent with the rest of the framework

## Testing

Run the test to verify the fix:
```bash
cd internal/inspector
go test -run TestHardcodedBorderFixed -v
```

Expected output:
```
✅ FIXED: Inspector title bar no longer uses hardcoded border characters
✅ FIXED: Border style now uses correct string instead of theme.Border()
✅ Bordered().Label('TEST') works correctly
```

## Future Improvements

1. **Make border style configurable**: Use a setting or theme constant instead of hardcoded "double"
2. **Add inspector-specific theme**: Create theme settings for inspector appearance
3. **Support custom borders**: Allow users to customize inspector border style
