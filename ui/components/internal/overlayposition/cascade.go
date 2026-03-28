package overlayposition

// CascadeDirection describes which horizontal side a cascading overlay prefers.
type CascadeDirection int

const (
	CascadeRight CascadeDirection = iota + 1
	CascadeLeft
)

// CascadeConfig describes a cascading popup placement request.
type CascadeConfig struct {
	Parent             Rect
	Overlay            Size
	Viewport           Size
	Top                int
	PreferredDirection CascadeDirection
}

// CascadeResult is the resolved absolute box position for one cascade surface.
type CascadeResult struct {
	X         int
	Y         int
	Direction CascadeDirection
	ClampedX  bool
	ClampedY  bool
}

// ResolveCascade resolves one cascading overlay box relative to its parent.
// It first tries the preferred side, then the mirrored side, and finally clamps
// the preferred-side candidate into the viewport.
func ResolveCascade(cfg CascadeConfig) CascadeResult {
	preferredDirection := cfg.PreferredDirection
	if preferredDirection != CascadeLeft {
		preferredDirection = CascadeRight
	}

	rightStartX := cfg.Parent.X + cfg.Parent.Width
	leftStartX := cfg.Parent.X - cfg.Overlay.Width
	preferredX := rightStartX
	mirroredX := leftStartX
	mirroredDirection := CascadeLeft
	if preferredDirection == CascadeLeft {
		preferredX = leftStartX
		mirroredX = rightStartX
		mirroredDirection = CascadeRight
	}

	switch {
	case fitsCascadeViewport(preferredX, cfg.Overlay.Width, cfg.Viewport.Width):
		return CascadeResult{
			X:         preferredX,
			Y:         clampCascadeTop(cfg.Top, cfg.Overlay.Height, cfg.Viewport.Height),
			Direction: preferredDirection,
			ClampedY:  cfg.Top != clampCascadeTop(cfg.Top, cfg.Overlay.Height, cfg.Viewport.Height),
		}
	case fitsCascadeViewport(mirroredX, cfg.Overlay.Width, cfg.Viewport.Width):
		return CascadeResult{
			X:         mirroredX,
			Y:         clampCascadeTop(cfg.Top, cfg.Overlay.Height, cfg.Viewport.Height),
			Direction: mirroredDirection,
			ClampedY:  cfg.Top != clampCascadeTop(cfg.Top, cfg.Overlay.Height, cfg.Viewport.Height),
		}
	default:
		clampedX := clampCascadeLeft(preferredX, cfg.Overlay.Width, cfg.Viewport.Width)
		clampedY := clampCascadeTop(cfg.Top, cfg.Overlay.Height, cfg.Viewport.Height)
		return CascadeResult{
			X:         clampedX,
			Y:         clampedY,
			Direction: inferCascadeDirection(preferredDirection, cfg.Parent.X, clampedX),
			ClampedX:  clampedX != preferredX,
			ClampedY:  clampedY != cfg.Top,
		}
	}
}

func fitsCascadeViewport(left, width, viewportWidth int) bool {
	if viewportWidth <= 0 {
		return true
	}
	return left >= 0 && left+width <= viewportWidth
}

func clampCascadeLeft(left, width, viewportWidth int) int {
	if viewportWidth <= 0 {
		return left
	}
	return clamp(left, 0, maxInt(0, viewportWidth-width))
}

func clampCascadeTop(top, height, viewportHeight int) int {
	if viewportHeight <= 0 {
		return top
	}
	return clamp(top, 0, maxInt(0, viewportHeight-height))
}

func inferCascadeDirection(preferredDirection CascadeDirection, parentLeft, resolvedLeft int) CascadeDirection {
	switch {
	case resolvedLeft < parentLeft:
		return CascadeLeft
	case resolvedLeft > parentLeft:
		return CascadeRight
	default:
		return preferredDirection
	}
}
