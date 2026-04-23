package mcp

import "fmt"

type WarningCode string

const (
	WarningCodePreviewBeforeMissing WarningCode = "preview_before_missing"
)

type warningCodeInfo struct {
	code        WarningCode
	description string
}

var warningCodeInfos = []warningCodeInfo{
	{
		code:        WarningCodePreviewBeforeMissing,
		description: "Preview `before` value missing because the snapshot did not include it.",
	},
}

func warningCodeDocLines() []string {
	lines := make([]string, 0, len(warningCodeInfos))
	for _, info := range warningCodeInfos {
		lines = append(lines, fmt.Sprintf("%s: %s", info.code, info.description))
	}
	return lines
}
