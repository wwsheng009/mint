# Inspector TreeView Click Debug Guide

## Problem
TreeView clicks in Inspector overlay are not working.

## Debug Steps

### 1. Enable Verbose Output
```bash
export TUI_INSPECTOR_VERBOSE=true
export TUI_DEBUG_UI=true
go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go
```

### 2. Check Event Routing
Look for these messages in console:
- `[APP] Routing mouse (X,Y) to Inspector` - App is routing to Inspector
- `[Inspector] TreeView click: localY=...` - Click is detected in overlay
- `[Inspector] TreeView clicked at line ...` - Click was processed

### 3. Verify TreeView State
If click is detected but not working:
- Check if `lineIndex` is valid (>= 0 and < lineCount)
- Check if `SetFocusIndex` is being called
- Check if `HandleAction` returns true

### 4. Common Issues

#### Issue 1: Click Not Detected
**Symptom**: No console output about TreeView click
**Cause**: Mouse event not routed to Inspector or overlay
**Fix**: Check inspector visibility and overlay bounds

#### Issue 2: Click Detected But Wrong Line
**Symptom**: `lineIndex` is negative or too large
**Cause**: Incorrect `treeViewStartY` calculation
**Fix**: Adjust overhead calculation in `handleOverlayClick`

#### Issue 3: Action Not Handled
**Symptom**: `handled=false` from HandleAction
**Cause**: TreeView doesn't handle ActionClick properly
**Fix**: Check TreeView.HandleAction implementation

## Manual Testing

1. Run demo with verbose output
2. Press F12 to open Inspector
3. Click on Elements tab
4. Click on different TreeView lines
5. Observe console output

## Expected Output

```
[APP] Routing mouse (45,15) to Inspector (type=0x1)
[Inspector] TreeView click: localY=18, lineIndex=5, lineCount=42
[Inspector] TreeView clicked at line 5, handled=true
```

## If Still Not Working

1. Check git status to ensure fixes are applied
2. Rebuild the project
3. Check for compilation errors
4. Verify TreeView component is properly initialized
