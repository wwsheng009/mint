package inspector

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// ProblemSeverity indicates how severe a layout problem is
type ProblemSeverity int

const (
	SeverityInfo     ProblemSeverity = iota // Informational
	SeverityWarning                          // Warning (might cause issues)
	SeverityError                            // Error (definitely causes issues)
	SeverityCritical                         // Critical (severe problems)
)

// LayoutProblem represents a detected layout issue
type LayoutProblem struct {
	Severity  ProblemSeverity // Severity level
	Type      string          // Problem type
	Location  string          // Path to the problematic element
	Message   string          // Human-readable description
	Element   ElementInfo     // Element information
	Suggestion string         // Suggested fix
}

// LayoutDiagnostics detects and reports layout problems
type LayoutDiagnostics struct {
	problems     []LayoutProblem
	enabled      bool
	maxProblems  int // Maximum problems to track
}

// NewLayoutDiagnostics creates a new layout diagnostics instance
func NewLayoutDiagnostics() *LayoutDiagnostics {
	return &LayoutDiagnostics{
		problems:    make([]LayoutProblem, 0),
		enabled:     true,
		maxProblems: 100,
	}
}

// Enable enables problem detection
func (ld *LayoutDiagnostics) Enable() {
	ld.enabled = true
}

// Disable disables problem detection
func (ld *LayoutDiagnostics) Disable() {
	ld.enabled = false
}

// IsEnabled returns whether detection is enabled
func (ld *LayoutDiagnostics) IsEnabled() bool {
	return ld.enabled
}

// Analyze analyzes a VNode tree for layout problems
func (ld *LayoutDiagnostics) Analyze(root ui.VNode) []LayoutProblem {
	if !ld.enabled || root == nil {
		return ld.problems
	}

	// Clear previous problems
	ld.problems = make([]LayoutProblem, 0, ld.maxProblems)

	// Analyze tree
	ld.analyzeRecursive(root, "", 0)

	return ld.problems
}

// analyzeRecursive recursively analyzes nodes
func (ld *LayoutDiagnostics) analyzeRecursive(vnode ui.VNode, path string, depth int) {
	if vnode == nil {
		return
	}

	// Stop if we've reached max problems
	if len(ld.problems) >= ld.maxProblems {
		return
	}

	// Generate path for this node
	nodePath := path
	if nodePath == "" {
		nodePath = getSimpleType(vnode)
	} else {
		nodePath = path + "." + getSimpleType(vnode)
	}

	// Extract element info
	info := ExtractElementInfo(vnode)

	// Run checks
	ld.checkConstraints(info, nodePath)
	ld.checkSizeConsistency(info, nodePath)
	ld.checkFlexValues(info, nodePath)
	ld.checkOverflow(info, nodePath)

	// Recursively analyze children
	for _, child := range vnode.Children() {
		ld.analyzeRecursive(child, nodePath, depth+1)
	}
}

// checkConstraints checks for constraint-related problems
func (ld *LayoutDiagnostics) checkConstraints(info ElementInfo, path string) {
	// Check for impossible constraints
	if info.Constraints.MinWidth > info.Constraints.MaxWidth {
		ld.addProblem(LayoutProblem{
			Severity: SeverityError,
			Type:     "Impossible Constraints",
			Location: path,
			Message:  fmt.Sprintf("MinWidth (%d) > MaxWidth (%d)", info.Constraints.MinWidth, info.Constraints.MaxWidth),
			Element:  info,
			Suggestion: "Fix MinWidth to be <= MaxWidth",
		})
	}

	if info.Constraints.MinHeight > info.Constraints.MaxHeight {
		ld.addProblem(LayoutProblem{
			Severity: SeverityError,
			Type:     "Impossible Constraints",
			Location: path,
			Message:  fmt.Sprintf("MinHeight (%d) > MaxHeight (%d)", info.Constraints.MinHeight, info.Constraints.MaxHeight),
			Element:  info,
			Suggestion: "Fix MinHeight to be <= MaxHeight",
		})
	}

	// Check for unbounded width
	if info.Constraints.MaxWidth >= 100000 && info.Layout.NaturalWidth > 100 {
		ld.addProblem(LayoutProblem{
			Severity: SeverityWarning,
			Type:     "Unbounded Width",
			Location: path,
			Message:  fmt.Sprintf("Element has large natural width (%d) but unbounded max width", info.Layout.NaturalWidth),
			Element:  info,
			Suggestion: "Consider setting MaxWidth constraint",
		})
	}

	// Check for tight constraints
	if info.Constraints.MinWidth == info.Constraints.MaxWidth && info.Constraints.MinWidth > 0 {
		// Element is tightly constrained - check if layout width matches
		if info.Layout.LayoutWidth != info.Constraints.MinWidth {
			ld.addProblem(LayoutProblem{
				Severity: SeverityWarning,
				Type:     "Tight Constraint Mismatch",
				Location: path,
				Message:  fmt.Sprintf("Tight width constraint (%d) but layout width is %d",
					info.Constraints.MinWidth, info.Layout.LayoutWidth),
				Element:  info,
				Suggestion: "Layout width should match tight constraints",
			})
		}
	}
}

// checkSizeConsistency checks for size inconsistencies
func (ld *LayoutDiagnostics) checkSizeConsistency(info ElementInfo, path string) {
	// Check if bounds width matches layout width
	if info.Size.Width > 0 && info.Layout.LayoutWidth > 0 {
		if info.Size.Width != info.Layout.LayoutWidth {
			ld.addProblem(LayoutProblem{
				Severity: SeverityWarning,
				Type:     "Size Inconsistency",
				Location: path,
				Message:  fmt.Sprintf("Bounds width (%d) != Layout width (%d)",
					info.Size.Width, info.Layout.LayoutWidth),
				Element:  info,
				Suggestion: "Ensure bounds are set from layout results",
			})
		}
	}

	// Check for zero width with non-zero natural size
	if info.Size.Width == 0 && info.Layout.NaturalWidth > 0 {
		ld.addProblem(LayoutProblem{
			Severity: SeverityError,
			Type:     "Zero Size",
			Location: path,
			Message:  fmt.Sprintf("Element has zero width but natural width is %d",
				info.Layout.NaturalWidth),
			Element:  info,
			Suggestion: "Check layout constraints and parent sizing",
		})
	}
}

// checkFlexValues checks for flex-related problems
func (ld *LayoutDiagnostics) checkFlexValues(info ElementInfo, path string) {
	// Check for flex elements with no room to grow
	if info.Layout.Flex > 0 {
		if info.Layout.LayoutWidth <= info.Layout.NaturalWidth {
			ld.addProblem(LayoutProblem{
				Severity: SeverityInfo,
				Type:     "Flex Without Growth",
				Location: path,
				Message:  fmt.Sprintf("Element has flex=%d but layout width (%d) <= natural width (%d)",
					info.Layout.Flex, info.Layout.LayoutWidth, info.Layout.NaturalWidth),
				Element:  info,
				Suggestion: "Flex elements should grow beyond their natural width",
			})
		}
	}

	// Check for large flex values that might cause issues
	if info.Layout.Flex > 100 {
		ld.addProblem(LayoutProblem{
			Severity: SeverityWarning,
			Type:     "Large Flex Value",
			Location: path,
			Message:  fmt.Sprintf("Element has very large flex value (%d)", info.Layout.Flex),
			Element:  info,
			Suggestion: "Consider using smaller flex values (1-10)",
		})
	}
}

// checkOverflow checks for overflow issues
func (ld *LayoutDiagnostics) checkOverflow(info ElementInfo, path string) {
	// Check if content overflows bounds
	if info.Size.Width > 0 && info.Layout.NaturalWidth > info.Size.Width {
		ld.addProblem(LayoutProblem{
			Severity: SeverityWarning,
			Type:     "Content Overflow",
			Location: path,
			Message:  fmt.Sprintf("Natural width (%d) > bounds width (%d) - content may overflow",
				info.Layout.NaturalWidth, info.Size.Width),
			Element:  info,
			Suggestion: "Increase width or enable text wrapping/clipping",
		})
	}

	// Check for negative positions
	if info.Position.X < 0 {
		ld.addProblem(LayoutProblem{
			Severity: SeverityError,
			Type:     "Negative Position",
			Location: path,
			Message:  fmt.Sprintf("Element has negative X position: %d", info.Position.X),
			Element:  info,
			Suggestion: "Check layout calculations",
		})
	}

	if info.Position.Y < 0 {
		ld.addProblem(LayoutProblem{
			Severity: SeverityError,
			Type:     "Negative Position",
			Location: path,
			Message:  fmt.Sprintf("Element has negative Y position: %d", info.Position.Y),
			Element:  info,
			Suggestion: "Check layout calculations",
		})
	}
}

// addProblem adds a problem to the list
func (ld *LayoutDiagnostics) addProblem(problem LayoutProblem) {
	if len(ld.problems) < ld.maxProblems {
		ld.problems = append(ld.problems, problem)
	}
}

// GetProblems returns all detected problems
func (ld *LayoutDiagnostics) GetProblems() []LayoutProblem {
	return ld.problems
}

// GetProblemsBySeverity returns problems filtered by severity
func (ld *LayoutDiagnostics) GetProblemsBySeverity(severity ProblemSeverity) []LayoutProblem {
	var result []LayoutProblem
	for _, problem := range ld.problems {
		if problem.Severity == severity {
			result = append(result, problem)
		}
	}
	return result
}

// GetProblemsByType returns problems filtered by type
func (ld *LayoutDiagnostics) GetProblemsByType(problemType string) []LayoutProblem {
	var result []LayoutProblem
	for _, problem := range ld.problems {
		if problem.Type == problemType {
			result = append(result, problem)
		}
	}
	return result
}

// CountBySeverity returns problem counts by severity
func (ld *LayoutDiagnostics) CountBySeverity() map[ProblemSeverity]int {
	counts := make(map[ProblemSeverity]int)
	for _, problem := range ld.problems {
		counts[problem.Severity]++
	}
	return counts
}

// FormatProblems formats problems as text
func (ld *LayoutDiagnostics) FormatProblems() string {
	if len(ld.problems) == 0 {
		return "No layout problems detected"
	}

	var lines []string
	lines = append(lines, "┌─ Layout Diagnostics ────────────────────────────┐")

	// Group by severity
	critical := ld.GetProblemsBySeverity(SeverityCritical)
	errors := ld.GetProblemsBySeverity(SeverityError)
	warnings := ld.GetProblemsBySeverity(SeverityWarning)
	info := ld.GetProblemsBySeverity(SeverityInfo)

	// Summary
	lines = append(lines, "│ Summary:                                        │")
	lines = append(lines, fmt.Sprintf("│   Critical: %-32d │", len(critical)))
	lines = append(lines, fmt.Sprintf("│   Errors: %-34d │", len(errors)))
	lines = append(lines, fmt.Sprintf("│   Warnings: %-32d │", len(warnings)))
	lines = append(lines, fmt.Sprintf("│   Info: %-35d │", len(info)))
	lines = append(lines, "│                                                 │")

	// Details (show all problems)
	lines = append(lines, "│ Details:                                        │")

	for i, problem := range ld.problems {
		if i >= 20 { // Limit details
			lines = append(lines, fmt.Sprintf("│ ... and %d more problems                       │",
				len(ld.problems)-20))
			break
		}

		severity := severityToString(problem.Severity)
		lines = append(lines, fmt.Sprintf("│ [%d] %s: %-30s │",
			i+1, severity, problem.Type))
		lines = append(lines, fmt.Sprintf("│     Location: %s%-28s │",
			truncateString(problem.Location, 28), ""))
		lines = append(lines, fmt.Sprintf("│     %s%-37s │",
			truncateString(problem.Message, 37), ""))
		if problem.Suggestion != "" {
			lines = append(lines, fmt.Sprintf("│     → %s%-35s │",
				truncateString(problem.Suggestion, 35), ""))
		}
	}

	lines = append(lines, "└─────────────────────────────────────────────────┘")

	return joinLines(lines)
}

// FormatCompact formats problems in a compact way
func (ld *LayoutDiagnostics) FormatCompact() string {
	if len(ld.problems) == 0 {
		return "✓ No problems"
	}

	counts := ld.CountBySeverity()
	return fmt.Sprintf("Problems: %d Critical, %d Errors, %d Warnings, %d Info",
		counts[SeverityCritical],
		counts[SeverityError],
		counts[SeverityWarning],
		counts[SeverityInfo])
}

// Clear clears all detected problems
func (ld *LayoutDiagnostics) Clear() {
	ld.problems = make([]LayoutProblem, 0)
}

// Helper functions

// severityToString converts severity to string
func severityToString(severity ProblemSeverity) string {
	switch severity {
	case SeverityCritical:
		return "CRIT"
	case SeverityError:
		return "ERR"
	case SeverityWarning:
		return "WARN"
	case SeverityInfo:
		return "INFO"
	default:
		return "????"
	}
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
