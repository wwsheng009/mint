package statusbadge

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/badge"
)

// Tone describes semantic status severity.
type Tone string

const (
	ToneNormal  Tone = "normal"
	ToneWarn    Tone = "warn"
	ToneError   Tone = "error"
	ToneNeutral Tone = "neutral"
	ToneInfo    Tone = "info"
)

// Option configures StatusBadge.
type Option func(*Config)

// Config stores StatusBadge configuration.
type Config struct {
	Label         string
	Tone          Tone
	HasTone       bool
	Key           string
	Text          string
	Dot           bool
	LabelStyle    style.Style
	BadgeStyle    style.Style
	HasLabelStyle bool
	HasBadgeStyle bool
	Mapper        func(string) Tone
}

// New builds a semantic status badge.
func New(status string, opts ...Option) rtui.VNode {
	cfg := Config{Label: status, Text: status}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	tone := cfg.Tone
	if !cfg.HasTone {
		mapper := cfg.Mapper
		if mapper == nil {
			mapper = DefaultTone
		}
		tone = mapper(status)
	}

	builder := badge.NewBuilder(cfg.Label).Text(cfg.Text).Status(toneToBadgeStatus(tone))
	if cfg.Key != "" {
		builder.Key(cfg.Key)
	}
	if cfg.Dot {
		builder.Dot(true)
	}
	if cfg.HasLabelStyle {
		builder.LabelStyle(cfg.LabelStyle)
	}
	if cfg.HasBadgeStyle {
		builder.BadgeStyle(cfg.BadgeStyle)
	}
	return builder.Build()
}

// DefaultTone maps common operational status strings to semantic tones.
func DefaultTone(status string) Tone {
	normalized := strings.ToLower(strings.TrimSpace(status))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "healthy", "active", "available", "effective", "enabled", "success", "ok", "ready", "running":
		return ToneNormal
	case "degraded", "rate_limited", "limited", "pending_restart", "pending", "warning", "cooldown", "partial":
		return ToneWarn
	case "unhealthy", "disabled", "unauthorized", "unavailable", "failed", "failure", "error", "down", "blocked":
		return ToneError
	case "info", "processing", "loading", "syncing":
		return ToneInfo
	default:
		return ToneNeutral
	}
}

func toneToBadgeStatus(tone Tone) badge.Status {
	switch tone {
	case ToneNormal:
		return badge.StatusSuccess
	case ToneWarn:
		return badge.StatusWarning
	case ToneError:
		return badge.StatusError
	case ToneInfo:
		return badge.StatusProcessing
	default:
		return badge.StatusDefault
	}
}

func ForceTone(tone Tone) Option {
	return func(cfg *Config) {
		cfg.Tone = tone
		cfg.HasTone = true
	}
}

func Key(key string) Option {
	return func(cfg *Config) {
		cfg.Key = key
	}
}

func Label(label string) Option {
	return func(cfg *Config) {
		cfg.Label = label
	}
}

func Text(text string) Option {
	return func(cfg *Config) {
		cfg.Text = text
	}
}

func Dot() Option {
	return func(cfg *Config) {
		cfg.Dot = true
	}
}

func Mapper(mapper func(string) Tone) Option {
	return func(cfg *Config) {
		cfg.Mapper = mapper
	}
}

func LabelStyle(labelStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.LabelStyle = labelStyle
		cfg.HasLabelStyle = true
	}
}

func BadgeStyle(badgeStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.BadgeStyle = badgeStyle
		cfg.HasBadgeStyle = true
	}
}
