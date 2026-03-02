package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/ui"
)

// TestNewLayoutDiagnostics tests creating a new diagnostics instance
func TestNewLayoutDiagnostics(t *testing.T) {
	ld := NewLayoutDiagnostics()

	if ld == nil {
		t.Fatal("Expected non-nil LayoutDiagnostics")
	}

	if !ld.enabled {
		t.Error("New diagnostics should be enabled by default")
	}

	if ld.maxProblems != 100 {
		t.Errorf("Expected maxProblems 100, got %d", ld.maxProblems)
	}
}

// TestEnableDisable tests enabling and disabling
func TestEnableDisable(t *testing.T) {
	ld := NewLayoutDiagnostics()

	ld.Disable()
	if ld.enabled {
		t.Error("Should be disabled after Disable()")
	}

	ld.Enable()
	if !ld.enabled {
		t.Error("Should be enabled after Enable()")
	}
}

// TestAnalyze_EmptyTree tests analyzing empty tree
func TestAnalyze_EmptyTree(t *testing.T) {
	ld := NewLayoutDiagnostics()

	problems := ld.Analyze(nil)

	if len(problems) != 0 {
		t.Errorf("Expected no problems for nil tree, got %d", len(problems))
	}
}

// TestAnalyze_SimpleTree tests analyzing a simple tree
func TestAnalyze_SimpleTree(t *testing.T) {
	ld := NewLayoutDiagnostics()

	button := ui.NewButtonBuilder("Test").Build()

	// Set some bounds
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 18, 1)
	}

	problems := ld.Analyze(button)

	// Should have some problems (flex without growth, etc.)
	if len(problems) == 0 {
		t.Log("No problems detected (might be OK)")
	}
}

// TestImpossibleConstraints tests detection of impossible constraints
func TestImpossibleConstraints(t *testing.T) {
	ld := NewLayoutDiagnostics()

	button := ui.NewButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	// Set impossible constraints
	info.Constraints = runtime.BoxConstraints{
		MinWidth: 100,
		MaxWidth: 50, // Max < Min
	}

	ld.checkConstraints(info, "test")

	problems := ld.GetProblems()

	if len(problems) == 0 {
		t.Error("Should detect impossible constraints")
	}

	found := false
	for _, p := range problems {
		if p.Type == "Impossible Constraints" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should have Impossible Constraints problem")
	}
}

// TestZeroSize tests detection of zero size
func TestZeroSize(t *testing.T) {
	ld := NewLayoutDiagnostics()

	button := ui.NewButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	// Set zero size but non-zero natural width
	info.Size.Width = 0
	info.Layout.NaturalWidth = 10

	ld.checkSizeConsistency(info, "test")

	problems := ld.GetProblems()

	if len(problems) == 0 {
		t.Error("Should detect zero size")
	}

	found := false
	for _, p := range problems {
		if p.Type == "Zero Size" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should have Zero Size problem")
	}
}

// TestNegativePosition tests detection of negative positions
func TestNegativePosition(t *testing.T) {
	ld := NewLayoutDiagnostics()

	button := ui.NewButtonBuilder("Test").Build()
	info := ExtractElementInfo(button)

	// Set negative position
	info.Position.X = -5

	ld.checkOverflow(info, "test")

	problems := ld.GetProblems()

	if len(problems) == 0 {
		t.Error("Should detect negative position")
	}

	found := false
	for _, p := range problems {
		if p.Type == "Negative Position" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should have Negative Position problem")
	}
}

// TestGetProblemsBySeverity tests filtering by severity
func TestGetProblemsBySeverity(t *testing.T) {
	ld := NewLayoutDiagnostics()

	// Add some test problems
	ld.problems = append(ld.problems,
		LayoutProblem{Severity: SeverityError, Type: "Test Error"},
		LayoutProblem{Severity: SeverityWarning, Type: "Test Warning"},
		LayoutProblem{Severity: SeverityError, Type: "Another Error"},
	)

	errors := ld.GetProblemsBySeverity(SeverityError)

	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	warnings := ld.GetProblemsBySeverity(SeverityWarning)

	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}
}

// TestGetProblemsByType tests filtering by type
func TestGetProblemsByType(t *testing.T) {
	ld := NewLayoutDiagnostics()

	// Add some test problems
	ld.problems = append(ld.problems,
		LayoutProblem{Type: "Zero Size"},
		LayoutProblem{Type: "Overflow"},
		LayoutProblem{Type: "Zero Size"},
	)

	zeroSizeProblems := ld.GetProblemsByType("Zero Size")

	if len(zeroSizeProblems) != 2 {
		t.Errorf("Expected 2 Zero Size problems, got %d", len(zeroSizeProblems))
	}

	overflowProblems := ld.GetProblemsByType("Overflow")

	if len(overflowProblems) != 1 {
		t.Errorf("Expected 1 Overflow problem, got %d", len(overflowProblems))
	}
}

// TestCountBySeverity tests counting problems by severity
func TestCountBySeverity(t *testing.T) {
	ld := NewLayoutDiagnostics()

	// Add some test problems
	ld.problems = append(ld.problems,
		LayoutProblem{Severity: SeverityCritical},
		LayoutProblem{Severity: SeverityError},
		LayoutProblem{Severity: SeverityError},
		LayoutProblem{Severity: SeverityWarning},
		LayoutProblem{Severity: SeverityInfo},
	)

	counts := ld.CountBySeverity()

	if counts[SeverityCritical] != 1 {
		t.Errorf("Expected 1 critical, got %d", counts[SeverityCritical])
	}

	if counts[SeverityError] != 2 {
		t.Errorf("Expected 2 errors, got %d", counts[SeverityError])
	}

	if counts[SeverityWarning] != 1 {
		t.Errorf("Expected 1 warning, got %d", counts[SeverityWarning])
	}

	if counts[SeverityInfo] != 1 {
		t.Errorf("Expected 1 info, got %d", counts[SeverityInfo])
	}
}

// TestFormatProblems tests problem formatting
func TestFormatProblems(t *testing.T) {
	ld := NewLayoutDiagnostics()

	// Add some test problems
	ld.problems = append(ld.problems,
		LayoutProblem{
			Severity:  SeverityError,
			Type:      "Test Problem",
			Location:  "root.button",
			Message:   "This is a test problem",
			Suggestion: "Fix it",
		},
	)

	output := ld.FormatProblems()

	if output == "" {
		t.Error("Output should not be empty")
	}

	requiredStrings := []string{
		"Layout Diagnostics",
		"Summary:",
		"Test Problem",
		"root.button",
		"This is a test problem",
	}

	for _, s := range requiredStrings {
		if !contains(output, s) {
			t.Errorf("Output should contain '%s'", s)
		}
	}
}

// TestFormatProblems_NoProblems tests formatting with no problems
func TestFormatProblems_NoProblems(t *testing.T) {
	ld := NewLayoutDiagnostics()

	output := ld.FormatProblems()

	if output != "No layout problems detected" {
		t.Errorf("Expected 'No layout problems detected', got '%s'", output)
	}
}

// TestFormatCompact_Diagnostics tests compact formatting
func TestFormatCompact_Diagnostics(t *testing.T) {
	ld := NewLayoutDiagnostics()

	// Add some test problems
	ld.problems = append(ld.problems,
		LayoutProblem{Severity: SeverityError},
		LayoutProblem{Severity: SeverityWarning},
		LayoutProblem{Severity: SeverityInfo},
	)

	output := ld.FormatCompact()

	if !contains(output, "Problems:") {
		t.Error("Output should contain 'Problems:'")
	}

	if !contains(output, "Errors") {
		t.Error("Output should contain 'Errors'")
	}

	if !contains(output, "Warnings") {
		t.Error("Output should contain 'Warnings'")
	}
}

// TestFormatCompact_NoProblems tests compact formatting with no problems
func TestFormatCompact_NoProblems(t *testing.T) {
	ld := NewLayoutDiagnostics()

	output := ld.FormatCompact()

	if !contains(output, "✓") {
		t.Errorf("Expected success indicator, got '%s'", output)
	}
}

// TestClear tests clearing problems
func TestClear(t *testing.T) {
	ld := NewLayoutDiagnostics()

	// Add some problems
	ld.problems = append(ld.problems,
		LayoutProblem{Severity: SeverityError},
		LayoutProblem{Severity: SeverityWarning},
	)

	if len(ld.problems) != 2 {
		t.Errorf("Expected 2 problems before clear, got %d", len(ld.problems))
	}

	ld.Clear()

	if len(ld.problems) != 0 {
		t.Errorf("Expected 0 problems after clear, got %d", len(ld.problems))
	}
}

// TestMaxProblemsLimit tests that problem limit is respected
func TestMaxProblemsLimit(t *testing.T) {
	ld := NewLayoutDiagnostics()
	ld.maxProblems = 5

	// Try to add more than max
	for i := 0; i < 10; i++ {
		ld.addProblem(LayoutProblem{Severity: SeverityInfo})
	}

	if len(ld.problems) > 5 {
		t.Errorf("Problems should be limited to %d, got %d", ld.maxProblems, len(ld.problems))
	}
}

// TestSeverityToString tests severity string conversion
func TestSeverityToString(t *testing.T) {
	tests := []struct {
		severity ProblemSeverity
		expected string
	}{
		{SeverityCritical, "CRIT"},
		{SeverityError, "ERR"},
		{SeverityWarning, "WARN"},
		{SeverityInfo, "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := severityToString(tt.severity)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestTruncateString tests string truncation
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"No truncation", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Needs truncation", "hello world", 8, "hello..."},
		{"Empty string", "", 10, ""},
		{"Very short", "hi", 10, "hi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
