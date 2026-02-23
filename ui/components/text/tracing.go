// Package text addon - Constraint tracing support
// This file extends Text instances with constraint tracing capability.
package text

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
)

// measureWithTrace wraps the original Measure method with constraint tracing.
func (inst *Instance) measureWithTrace(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Build text ID for tracing
	textID := "text"
	if inst.key != "" {
		textID = fmt.Sprintf("text(%s)", inst.key)
	}

	// Build path
	path := inst.path
	if path == "" {
		path = textID
	}

	// Trace constraint propagation at the beginning of measurement
	wrapMode := "single-line"
	if inst.wrap {
		wrapMode = "wrap"
	}

	// Calculate what the constraints actually mean for this text
	effectiveConstraints := constraints
	if inst.maxWidth > 0 {
		// User-specified maxWidth acts as a constraint
		if effectiveConstraints.MaxWidth <= 0 || inst.maxWidth < effectiveConstraints.MaxWidth {
			effectiveConstraints.MaxWidth = inst.maxWidth
		}
	}

	layout.TraceMeasuring(
		"text",
		textID,
		path,
		constraints,
		effectiveConstraints,
		layout.Size{}, // Size will be updated after measurement
		fmt.Sprintf("Text: wrap=%s, maxWidth=%d, contentLen=%d",
			wrapMode, inst.maxWidth, len(inst.content)),
	)

	// Perform original measurement
	result := inst.measureOriginal(constraints)

	// Update trace with actual size
	layout.TraceMeasuring(
		"text",
		textID,
		path,
		constraints,
		effectiveConstraints,
		result,
		fmt.Sprintf("Text measured: %dx%d, lines=%d", result.Width, result.Height, len(inst.wrapLines)),
	)

	return result
}

// measureOriginal is the original measurement logic from instance.go
func (inst *Instance) measureOriginal(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Calculate text width
	content := inst.content
	if content == "" {
		content = " "
	}

	// Apply user-specified padding
	horizontalPadding := inst.padding[1] + inst.padding[3]
	verticalPadding := inst.padding[0] + inst.padding[2]

	var contentWidth, contentHeight int

	// Handle wrap mode
	if inst.wrap {
		availableWidth := constraints.MaxWidth
		if inst.maxWidth > 0 && (availableWidth == 0 || inst.maxWidth < availableWidth) {
			availableWidth = inst.maxWidth
		}
		availableWidth -= horizontalPadding
		if availableWidth < 1 {
			availableWidth = 1
		}

		inst.wrapLines = wordWrap(content, availableWidth)
		contentHeight = len(inst.wrapLines)
		if contentHeight == 0 {
			contentHeight = 1
		}

		maxLineWidth := 0
		for _, line := range inst.wrapLines {
			lineWidth := utf8RuneCountInString(line)
			if lineWidth > maxLineWidth {
				maxLineWidth = lineWidth
			}
		}
		contentWidth = maxLineWidth
	} else {
		contentWidth = utf8RuneCountInString(content)
		contentHeight = 1
	}

	width := contentWidth + horizontalPadding
	height := contentHeight + verticalPadding

	if !inst.wrap && inst.maxWidth > 0 && width > inst.maxWidth {
		width = inst.maxWidth
	}

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// Helper functions copied from instance.go
func utf8RuneCountInString(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}

// =============================================================================
// Tracing extensions for Instance
// =============================================================================

// SetPath sets the constraint tracing path for this instance.
func (inst *Instance) SetPath(path string) *Instance {
	inst.path = path
	return inst
}

// GetPath returns the constraint tracing path.
func (inst *Instance) GetPath() string {
	return inst.path
}
