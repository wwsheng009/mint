# Modal Button HitTest Position Issue

## Root Cause Found

### Problem Summary
- **Requested buffer size**: 80x24 (via `ui.WithWidth/Height`)
- **Actual terminal size**: 156x44
- **Modal centering calculation**: Uses actual terminal size (156x44)
- **Modal position**: Centered at Y=(44-9)/2 = 17
- **Button HitMap position**: Y=17+1+3 = 21

### The Issue
User reports:
- Buttons visually appear at Y=40
- Clicking Y=20 works (HitTest succeeds)
- **This indicates HitMap position (Y=20) ≠ Visual position (Y=40)**

### Possible Causes

#### Cause 1: Double Centering
Modal might be centered twice:
1. Once in buffer (Y=17)
2. Once when buffer is rendered to terminal (if buffer is smaller than terminal)

If buffer is 80x24 but terminal is 156x44:
- Modal centered in 80x24 buffer at Y=7
- Buffer centered in 156x44 terminal at Y=10
- Final visual position: Y=7+10 = 17
- But user sees Y=40, which doesn't match

#### Cause 2: Coordinate System Mismatch
- HitMap uses buffer coordinates (0-based, relative to buffer)
- Mouse events might use terminal coordinates (0-based, relative to terminal)
- If buffer is rendered at an offset in the terminal, coordinates won't match

#### Cause 3: Resize Event Timing
1. Initial: Buffer=80x24, modal centered at Y=7
2. ResizeEvent: Buffer→156x44, modal re-centered at Y=17
3. But HitMap from initial render still has Y=7
4. User clicks Y=17 (new position) but HitMap has Y=7 (old position)

### Verification Needed

Please run the demo and check:

```bash
cd /e/projects/yao/wwsheng009/mint/examples/ui_demos/demo1_full_featured
AUTO_OPEN_MODAL=true TUI_LAYER_DEBUG=true ./demo1.exe 2>&1 | grep -E "centerModal|Buffer size"
```

Look for:
1. `Buffer size: ?x?` - What is the buffer size?
2. `container=...` - What size is used for centering?
3. `after shift: modal=(..., ...)` - Where is modal positioned?
4. **Visually count**: Which line shows the modal title?

### Expected vs Actual

**If buffer=156x44:**
- Modal should be at Y=17 (centered)
- Buttons at Y=21
- This is correct for 44-line terminal

**If buffer=80x24 but terminal=156x44:**
- Modal at Y=7 (centered in 80x24)
- But where is the 80x24 buffer rendered in the 156x44 terminal?
- If at top: Y=7
- If centered: Y=7 + (44-24)/2 = Y=7+10 = Y=17

The user seeing Y=40 suggests neither of these calculations match.

### Most Likely Issue

**The ResizeEvent is changing the buffer size AFTER initial render, causing a mismatch between:**
1. Initial render (80x24) → Modal at Y=7
2. Resize event → Buffer becomes 156x44
3. Modal re-rendered at Y=17
4. But somehow visual rendering still shows old position?

Need to verify actual terminal size and buffer size at runtime.
