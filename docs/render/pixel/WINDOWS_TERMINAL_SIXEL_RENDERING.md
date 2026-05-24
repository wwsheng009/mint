# Windows Terminal Sixel Rendering

## Background

Mint image components render raster payloads through the scene image pipeline:

- component instance exposes `SceneLayers() []paint.ImageLayer`
- `internal/render.DeclarativeNode` collects image layers from paintable boxes
- `framework.App` renders the text buffer and then calls the configured `GraphicsPresenter`
- `runtime/platform.SixelGraphicsPresenter` emits SIXEL escape sequences for Windows Terminal

Windows Terminal is different from overlay protocols such as Kitty graphics. SIXEL mutates the terminal frame itself. Later text redraws can cover or erase the previously emitted image, so Mint treats Windows Terminal SIXEL as a terminal-frame protocol.

## Default Detection

When `WT_SESSION` is present, Mint now enables SIXEL by default:

```text
mode=sixel reliable=true source=heuristic-windows-terminal
```

This default is intended for Windows Terminal versions with SIXEL support. Users can still disable terminal graphics explicitly:

```powershell
$env:MINT_GRAPHICS = "off"
```

For better sizing, applications or users should provide cell pixel metrics when they know them:

```powershell
$env:MINT_CELL_PIXELS = "8x16"
```

If `MINT_CELL_PIXELS` is absent, Mint still enables SIXEL and uses source pixel dimensions. The image can render at a less precise cell size, but the component remains functional.

## Placement Fix

Image placement must be based on layout-computed bounds. A regression was observed where an image was rendered at the terminal origin because `SceneLayers()` was collected before the instance received `SetBounds()`.

The scene collection path now pushes the current `PaintableBox` bounds into the instance before reading `SceneLayers()`. Component-specific internal offsets remain the component's responsibility. For example, a chart component can place the plot image one row below its title after it receives the component bounds.

## Repaint Strategy

For overlay graphics protocols, Mint avoids re-presenting unchanged images by comparing:

- image id
- cell bounds
- pixel dimensions
- raster content hash

For terminal-frame protocols such as SIXEL, Mint re-presents the image after each text frame because text rendering can erase the previous image. To reduce flicker, unchanged terminal-frame images do not force a full text repaint and do not mark the image area dirty on every frame.

The practical effect is:

- image remains visible while the user types in nearby inputs
- text diff rendering still handles normal UI changes
- SIXEL payload is emitted after the text frame so the image stays on top of the cells it occupies
- full clear/repaint is reserved for layout or geometry changes

## Operational Notes

For Windows Terminal applications that display small images such as captchas, the recommended configuration is:

```powershell
$env:MINT_CELL_PIXELS = "8x16"
```

`MINT_GRAPHICS=sixel` is no longer required on Windows Terminal because `WT_SESSION` enables SIXEL by default. It remains useful for forcing SIXEL in unrecognized environments.

Applications should provide a text fallback for important images. Terminal image protocols are not universally available, and SIXEL behavior can still vary by terminal version and font/cell geometry.

## Linux Terminal Compatibility

Linux terminal support is intentionally capability-based rather than OS-based:

- Kitty-compatible terminals can use the Kitty graphics protocol when detected or forced with `MINT_GRAPHICS=kitty`.
- WezTerm/iTerm2-style inline images are only enabled for verified terminal markers, or explicitly with `MINT_GRAPHICS=inline-image` plus `MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE=1`.
- SIXEL can be forced with `MINT_GRAPHICS=sixel` when the terminal is known to support it.
- Common Linux desktop terminals can behave differently by distribution, version, profile, and multiplexer. Applications must not assume that a generic Linux terminal supports inline images.

For Linux deployments, the recommended operational model is:

```bash
MINT_GRAPHICS=kitty app
MINT_GRAPHICS=sixel MINT_CELL_PIXELS=8x16 app
MINT_GRAPHICS=off app
```

Application-level fallbacks should open or save the original image instead of attempting lossy ASCII reconstruction for security-sensitive images such as captchas. ASCII previews lose color, anti-aliasing, distortion, and interference-line detail, so they can mislead operators.
