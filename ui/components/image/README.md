# Image

`image` provides a generic raster image component for terminal graphics-capable
Mint applications.

## Features

- Accepts raw RGBA pixels with explicit pixel width and height.
- Accepts Go `image.Image` values.
- Accepts PNG, JPEG and GIF data URIs through Go standard library decoders.
- Emits `paint.ImageLayer` through `SceneLayers()` for the existing Mint
  graphics presenter pipeline.
- Renders a text fallback with `Alt` when no raster payload is available or the
  terminal graphics presenter is disabled.
- Keeps SVG data URIs as fallback-only input. SVG rasterization is intentionally
  not built into the initial component because Go standard library image
  decoders do not support SVG.

## Usage

```go
ui.NewImageBuilder().
    ID("captcha").
    Alt("captcha").
    SourceDataURI(dataURI).
    Size(24, 8).
    Build()
```

For raw pixels:

```go
ui.NewImageBuilder().
    Alt("preview").
    SourceRGBA(rgba, pixelWidth, pixelHeight).
    Size(32, 12).
    Build()
```

## Graphics Runtime

The component uses Mint's existing scene image pipeline. Inline rendering still
depends on terminal graphics support and the active runtime presenter, such as:

- `MINT_GRAPHICS=kitty`
- `MINT_GRAPHICS=sixel`
- `MINT_GRAPHICS=inline-image`

When no presenter is available, the component remains usable as an accessible
fallback text placeholder.
