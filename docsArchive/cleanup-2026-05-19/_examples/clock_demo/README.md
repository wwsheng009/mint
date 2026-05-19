# Clock Demo

Runnable demo for the `clock` component.

## Run

```bash
go run ./examples/clock_demo/
```

## What it shows

- A live clock that updates in realtime
- A fixed snapshot clock rendered from `StaticTime(...)`
- Circle / ellipse dial switching
- Independent width / height radius controls for ellipse mode
- Locked width / height radii in circle mode
- Adjustable terminal cell-aspect compensation so Circle renders as a visually round dial
- Second-hand and smooth-second toggles
- ASCII / Unicode hand-style toggle
- Distinct dial / tick / center / digital styles
- Distinct hour / minute / second hand colors
- Built-in visual presets: Default / Classic / Neon / Minimal / Alert
- Theme composition from `ThemePreset(...).Merge(...)` with partial recoloring
- Timezone switching for the live preview
- A timezone comparison block for the same instant

## Shortcuts

- `F1`: cycle dial shape
- `F2`: toggle second hand
- `F3`: toggle smooth second
- `F4`: toggle digital label
- `F5` / `F6`: width radius down / up
- `F7`: next timezone
- `F8`: reset demo state
- `F9`: toggle ASCII / Unicode hands
- `F10`: cycle preset
- `F11` / `F12`: height radius down / up

The demo also includes on-screen buttons for `CellAspectX`, so you can tune the
horizontal compensation when a particular terminal font makes the circle look
too narrow or too wide.
