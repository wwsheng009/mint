# Pixel Rendering Notes

This directory contains the current retained notes for experimental pixel/image rendering in Mint.

The current stable renderer is still cell-based text rendering. Pixel/image rendering is an experimental direction for higher-fidelity chart output and terminal graphics protocols.

## Current Retained Documents

| Document | Purpose |
|---|---|
| `PIXEL_CHART_RENDERING_ARCHITECTURE.md` | Architecture sketch for chart image rendering |
| `FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md` | App render flow sketch for image layers |
| `RUNTIME_PAINT_SCENE_API_SKETCH.md` | Scene API sketch |
| `RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md` | Platform graphics API sketch |

## Archived Planning Documents

Detailed handoff notes, decision logs, group specs, task breakdowns, implementation plans, and benchmark artifact plans were archived to:

```text
../../../docsArchive/cleanup-2026-05-19/docs/render/pixel/
```

Use those files for historical context only. Current implementation work should first verify the source in:

```text
runtime/paint/
runtime/platform/
framework/app.go
ui/components/charts/
examples/charts_linechart_image_prototype/
```

## Current Recommendation

Keep the production chart path text-first unless a terminal graphics capability is explicitly detected and the rendering path has a tested fallback. Pixel rendering should remain opt-in or experimental until platform capability detection, image layer cleanup, resize behavior, and benchmark coverage are stable.
