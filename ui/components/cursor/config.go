package cursor

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
)

// Shape controls how the cursor is painted.
type Shape int

const (
	ShapeBlock Shape = iota
	ShapeUnderline
	ShapeBar
)

// ThemeRole selects which theme color family drives the cursor.
type ThemeRole int

const (
	ThemeCaret ThemeRole = iota
	ThemeFocus
	ThemeText
	ThemeMuted
	ThemeAccent
)

const (
	FastBlinkInterval   = 250 * time.Millisecond
	NormalBlinkInterval = 500 * time.Millisecond
	SlowBlinkInterval   = 800 * time.Millisecond
)

// Config contains the declarative cursor configuration shared by standalone and embedded cursors.
type Config struct {
	Blink         bool
	BlinkInterval time.Duration
	Shape         Shape
	Theme         ThemeRole
	Style         style.Style
	Glyph         string
}

// DefaultConfig returns the baseline cursor behavior.
func DefaultConfig() Config {
	return Config{
		Blink:         true,
		BlinkInterval: NormalBlinkInterval,
		Shape:         ShapeBlock,
		Theme:         ThemeCaret,
	}
}

// NormalizeConfig applies default values for zero-value fields.
func NormalizeConfig(cfg Config) Config {
	def := DefaultConfig()
	if cfg == (Config{}) {
		return def
	}

	// Keep default blink behavior when caller only overrides non-blink fields.
	// To disable blink explicitly, provide Blink=false with a positive BlinkInterval.
	if !cfg.Blink && cfg.BlinkInterval <= 0 {
		cfg.Blink = def.Blink
	}

	if cfg.BlinkInterval <= 0 {
		cfg.BlinkInterval = def.BlinkInterval
	}

	return cfg
}
