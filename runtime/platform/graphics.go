package platform

import (
	"fmt"
	"strings"
)

// GraphicsMode describes the terminal graphics protocol mode detected for the
// current session. Phase 1 only requires stable support for None and Kitty.
type GraphicsMode int

const (
	GraphicsModeNone GraphicsMode = iota
	GraphicsModeKitty
	GraphicsModeSixel
	GraphicsModeInlineImage
)

// String returns a stable lowercase identifier suitable for diagnostics.
func (m GraphicsMode) String() string {
	switch m {
	case GraphicsModeNone:
		return "none"
	case GraphicsModeKitty:
		return "kitty"
	case GraphicsModeSixel:
		return "sixel"
	case GraphicsModeInlineImage:
		return "inline-image"
	default:
		return fmt.Sprintf("graphics-mode(%d)", m)
	}
}

// ParseGraphicsMode parses a textual graphics mode identifier.
func ParseGraphicsMode(raw string) (GraphicsMode, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "":
		return GraphicsModeNone, true
	case "none", "off":
		return GraphicsModeNone, true
	case "kitty":
		return GraphicsModeKitty, true
	case "sixel":
		return GraphicsModeSixel, true
	case "inline-image", "inline_image", "inline", "iterm2", "osc1337", "imgcat":
		return GraphicsModeInlineImage, true
	default:
		return GraphicsModeNone, false
	}
}

// GraphicsPresentationModel describes how a terminal protocol interacts with
// the terminal text frame. Overlay protocols behave like independent image
// objects layered over text. Terminal-frame protocols mutate the terminal
// surface directly, so the renderer cannot assume its front buffer still
// matches what the user sees after image presentation.
type GraphicsPresentationModel int

const (
	GraphicsPresentationModelUnknown GraphicsPresentationModel = iota
	GraphicsPresentationModelOverlay
	GraphicsPresentationModelTerminalFrame
)

// String returns a stable diagnostics-oriented identifier.
func (m GraphicsPresentationModel) String() string {
	switch m {
	case GraphicsPresentationModelOverlay:
		return "overlay"
	case GraphicsPresentationModelTerminalFrame:
		return "terminal-frame"
	default:
		return "unknown"
	}
}

// GraphicsCapabilities captures the minimal graphics capability metadata needed
// by Phase 1. Unknown pixel metrics are represented as 0.
type GraphicsCapabilities struct {
	Mode              GraphicsMode
	PresentationModel GraphicsPresentationModel
	Reliable          bool
	CellPixelWidth    int
	CellPixelHeight   int
	SupportsPlacement bool
	SupportsReplace   bool
	SupportsDelete    bool
	SupportsCrop      bool
	SupportsZOrder    bool
	ProbeSource       string
	Notes             []string
}

// HasGraphics reports whether a graphics protocol is available at all.
func (c GraphicsCapabilities) HasGraphics() bool {
	return c.Mode != GraphicsModeNone
}

// HasReliableGraphics reports whether the current graphics mode is both
// available and considered safe for automatic image mode enablement.
func (c GraphicsCapabilities) HasReliableGraphics() bool {
	return c.HasGraphics() && c.Reliable
}

// CellPixelsKnown reports whether terminal cell pixel metrics are known.
func (c GraphicsCapabilities) CellPixelsKnown() bool {
	return c.CellPixelWidth > 0 && c.CellPixelHeight > 0
}

// EffectivePresentationModel resolves the terminal interaction model for the
// current capability set, defaulting by protocol mode when call sites have not
// filled the field explicitly.
func (c GraphicsCapabilities) EffectivePresentationModel() GraphicsPresentationModel {
	if c.PresentationModel != GraphicsPresentationModelUnknown {
		return c.PresentationModel
	}

	switch c.Mode {
	case GraphicsModeKitty, GraphicsModeInlineImage:
		return GraphicsPresentationModelOverlay
	case GraphicsModeSixel:
		return GraphicsPresentationModelTerminalFrame
	default:
		return GraphicsPresentationModelUnknown
	}
}

// UsesTerminalFramePresentation reports whether image presentation mutates the
// terminal surface directly enough that buffered text state must be treated as
// stale after image emission.
func (c GraphicsCapabilities) UsesTerminalFramePresentation() bool {
	return c.EffectivePresentationModel() == GraphicsPresentationModelTerminalFrame
}

// Summary returns a stable diagnostics-oriented capability summary string.
func (c GraphicsCapabilities) Summary() string {
	source := c.ProbeSource
	if source == "" {
		source = "-"
	}

	cell := "unknown"
	if c.CellPixelsKnown() {
		cell = fmt.Sprintf("%dx%d", c.CellPixelWidth, c.CellPixelHeight)
	}

	features := make([]string, 0, 5)
	if c.SupportsPlacement {
		features = append(features, "placement")
	}
	if c.SupportsReplace {
		features = append(features, "replace")
	}
	if c.SupportsDelete {
		features = append(features, "delete")
	}
	if c.SupportsCrop {
		features = append(features, "crop")
	}
	if c.SupportsZOrder {
		features = append(features, "zorder")
	}
	featureSummary := "-"
	if len(features) > 0 {
		featureSummary = strings.Join(features, ",")
	}

	notes := "-"
	if len(c.Notes) > 0 {
		notes = strings.Join(c.Notes, "; ")
	}

	return fmt.Sprintf(
		"mode=%s reliable=%t source=%s cell=%s features=%s notes=%s",
		c.Mode.String(),
		c.Reliable,
		source,
		cell,
		featureSummary,
		notes,
	)
}

// WithNotes returns a shallow copy with notes appended, preserving the
// original slice to keep call sites side-effect free.
func (c GraphicsCapabilities) WithNotes(notes ...string) GraphicsCapabilities {
	if len(notes) == 0 {
		return c
	}

	out := c
	out.Notes = append(append([]string(nil), c.Notes...), notes...)
	return out
}

// DrawImageRequest is the Phase 1 draw payload handed from upper layers to a
// presenter. It intentionally carries simple raster and placement metadata.
type DrawImageRequest struct {
	ID              string
	PixelWidth      int
	PixelHeight     int
	CellX           int
	CellY           int
	CellWidth       int
	CellHeight      int
	RGBA            []byte
	AltText         string
	ReplaceIfExists bool
}

// GraphicsCapabilityProvider is an optional capability interface for objects
// that can report terminal graphics support.
type GraphicsCapabilityProvider interface {
	GraphicsCapabilities() GraphicsCapabilities
}

// GraphicsPresenter manages image object lifecycle for a specific graphics
// protocol. Phase 1 uses it only as an optional capability seam.
type GraphicsPresenter interface {
	Capabilities() GraphicsCapabilities
	Present(req DrawImageRequest) (string, error)
	Replace(id string, req DrawImageRequest) error
	Delete(id string) error
	Clear() error
}

// GraphicsPresenterProvider is an optional capability interface for objects
// that can expose a graphics presenter.
type GraphicsPresenterProvider interface {
	GraphicsPresenter() GraphicsPresenter
}
