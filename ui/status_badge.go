package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/statusbadge"
)

type StatusTone = statusbadge.Tone

const (
	StatusToneNormal  = statusbadge.ToneNormal
	StatusToneWarn    = statusbadge.ToneWarn
	StatusToneError   = statusbadge.ToneError
	StatusToneNeutral = statusbadge.ToneNeutral
	StatusToneInfo    = statusbadge.ToneInfo
)

type StatusBadgeOption = statusbadge.Option
type StatusBadgeConfig = statusbadge.Config

func StatusBadge(status string, opts ...StatusBadgeOption) rtui.VNode {
	return statusbadge.New(status, opts...)
}

func DefaultStatusTone(status string) StatusTone {
	return statusbadge.DefaultTone(status)
}

func StatusBadgeTone(tone StatusTone) StatusBadgeOption {
	return statusbadge.ForceTone(tone)
}

func StatusBadgeKey(key string) StatusBadgeOption {
	return statusbadge.Key(key)
}

func StatusBadgeLabel(label string) StatusBadgeOption {
	return statusbadge.Label(label)
}

func StatusBadgeText(text string) StatusBadgeOption {
	return statusbadge.Text(text)
}

func StatusBadgeDot() StatusBadgeOption {
	return statusbadge.Dot()
}

func StatusBadgeMapper(mapper func(string) StatusTone) StatusBadgeOption {
	return statusbadge.Mapper(mapper)
}

func StatusBadgeLabelStyle(labelStyle style.Style) StatusBadgeOption {
	return statusbadge.LabelStyle(labelStyle)
}

func StatusBadgeStyle(badgeStyle style.Style) StatusBadgeOption {
	return statusbadge.BadgeStyle(badgeStyle)
}
