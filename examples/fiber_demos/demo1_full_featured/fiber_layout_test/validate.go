package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Automated Layout Constraint Validator
// =============================================================================

// Validator performs automated validation of layout constraints
type Validator struct {
	PassCount   int
	FailCount   int
	SkipCount   int
	Errors      []ValidationError
	Warnings    []string
	Constraints runtime.BoxConstraints
}

// ValidationError represents a specific validation failure
type ValidationError struct {
	Category string // e.g., "RootConstraints", "ParentChildBoundaries"
	NodeID   uint64 // NodeID where error occurred
	Message  string
	Details  string
}

// NewValidator creates a new validator
func NewValidator(constraints runtime.BoxConstraints) *Validator {
	return &Validator{
		Errors:      make([]ValidationError, 0),
		Warnings:    make([]string, 0),
		Constraints: constraints,
	}
}

// Pass records a passing check
func (v *Validator) Pass() {
	v.PassCount++
}

// Fail records a failing check with details
func (v *Validator) Fail(category string, nodeID uint64, message, details string) {
	v.FailCount++
	v.Errors = append(v.Errors, ValidationError{
		Category: category,
		NodeID:   nodeID,
		Message:  message,
		Details:  details,
	})
}

// Skip records a skipped check
func (v *Validator) Skip(reason string) {
	v.SkipCount++
}

// Warn adds a warning (not a failure, but worth noting)
func (v *Validator) Warn(format string, args ...interface{}) {
	v.Warnings = append(v.Warnings, fmt.Sprintf(format, args...))
}

// =============================================================================
// Validation Rules
// =============================================================================

// Rule 1: Root box must fit within constraints
func (v *Validator) validateRootConstraints(root *compute.ComputedBox) {
	if root == nil {
		v.Fail("RootConstraints", 0, "Root box is nil", "The layout root should never be nil")
		return
	}

	// Check max width constraint
	if root.Width > v.Constraints.MaxWidth {
		v.Fail("RootConstraints", root.NodeID,
			fmt.Sprintf("Root width %d exceeds MaxWidth %d", root.Width, v.Constraints.MaxWidth),
			fmt.Sprintf("Root box must not exceed maximum width constraint"))
	} else if root.Width < v.Constraints.MinWidth {
		v.Fail("RootConstraints", root.NodeID,
			fmt.Sprintf("Root width %d is less than MinWidth %d", root.Width, v.Constraints.MinWidth),
			"Root box must respect minimum width constraint")
	} else {
		v.Pass()
	}

	// Check max height constraint
	if root.Height > v.Constraints.MaxHeight {
		v.Fail("RootConstraints", root.NodeID,
			fmt.Sprintf("Root height %d exceeds MaxHeight %d", root.Height, v.Constraints.MaxHeight),
			"Root box must not exceed maximum height constraint")
	} else if root.Height < v.Constraints.MinHeight {
		v.Fail("RootConstraints", root.NodeID,
			fmt.Sprintf("Root height %d is less than MinHeight %d", root.Height, v.Constraints.MinHeight),
			"Root box must respect minimum height constraint")
	} else {
		v.Pass()
	}

	// Check root position (should start at origin)
	if root.X != 0 || root.Y != 0 {
		v.Warn("Root position (%d, %d) is not at origin (0, 0)", root.X, root.Y)
	}
	v.Pass() // Position check passes regardless
}

// Rule 2: Children must stay within parent boundaries
func (v *Validator) validateParentChildBoundaries(box *compute.ComputedBox) {
	if box == nil || len(box.Children) == 0 {
		return
	}

	for i, child := range box.Children {
		childRight := child.X + child.Width
		childBottom := child.Y + child.Height
		parentRight := box.X + box.Width
		parentBottom := box.Y + box.Height

		// Check horizontal bounds (only if parent has non-zero width)
		if box.Width > 0 {
			if childRight > parentRight {
				v.Fail("ParentChildBoundaries", child.NodeID,
					fmt.Sprintf("Child[%d] right edge (%d) exceeds parent right edge (%d)", i, childRight, parentRight),
					fmt.Sprintf("Parent NodeID=%d, Width=%d, X=%d", box.NodeID, box.Width, box.X))
			} else {
				v.Pass()
			}

			if child.X < box.X {
				v.Fail("ParentChildBoundaries", child.NodeID,
					fmt.Sprintf("Child[%d] X (%d) is before parent X (%d)", i, child.X, box.X),
					fmt.Sprintf("Child should not start before parent's left edge"))
			} else {
				v.Pass()
			}
		}

		// Check vertical bounds (only if parent has non-zero height)
		if box.Height > 0 {
			if childBottom > parentBottom {
				v.Fail("ParentChildBoundaries", child.NodeID,
					fmt.Sprintf("Child[%d] bottom edge (%d) exceeds parent bottom edge (%d)", i, childBottom, parentBottom),
					fmt.Sprintf("Parent NodeID=%d, Height=%d, Y=%d", box.NodeID, box.Height, box.Y))
			} else {
				v.Pass()
			}

			if child.Y < box.Y {
				v.Fail("ParentChildBoundaries", child.NodeID,
					fmt.Sprintf("Child[%d] Y (%d) is before parent Y (%d)", i, child.Y, box.Y),
					"Child should not start before parent's top edge")
			} else {
				v.Pass()
			}
		}

		// Recurse into children
		v.validateParentChildBoundaries(child)
	}
}

// Rule 3: All NodeIDs must be unique
func (v *Validator) validateNodeIDUniqueness(root *compute.ComputedBox) {
	nodeIDs := make(map[uint64]int) // NodeID -> count
	var collect func(box *compute.ComputedBox)
	collect = func(box *compute.ComputedBox) {
		if box == nil {
			return
		}
		nodeIDs[box.NodeID]++
		for _, child := range box.Children {
			collect(child)
		}
	}
	collect(root)

	duplicates := 0
	for id, count := range nodeIDs {
		if count > 1 {
			v.Fail("NodeIDUniqueness", id,
				fmt.Sprintf("NodeID %d appears %d times (must be unique)", id, count),
				"Each NodeID should appear exactly once in the layout tree")
			duplicates++
		}
	}

	if duplicates == 0 {
		v.Pass()
	}
}

// Rule 4: All dimensions must be non-negative
func (v *Validator) validateNonNegativeDimensions(box *compute.ComputedBox) {
	if box == nil {
		return
	}

	if box.Width < 0 {
		v.Fail("NonNegativeDimensions", box.NodeID,
			fmt.Sprintf("Width is negative: %d", box.Width),
			"Width must be >= 0")
	} else {
		v.Pass()
	}

	if box.Height < 0 {
		v.Fail("NonNegativeDimensions", box.NodeID,
			fmt.Sprintf("Height is negative: %d", box.Height),
			"Height must be >= 0")
	} else {
		v.Pass()
	}

	for _, child := range box.Children {
		v.validateNonNegativeDimensions(child)
	}
}

// Rule 5: Check for zero-size boxes that should have content
func (v *Validator) validateZeroSizeBoxes(box *compute.ComputedBox, depth int) {
	if box == nil {
		return
	}

	// Check for zero-size boxes (may indicate measurement issues)
	if box.Width == 0 || box.Height == 0 {
		// Root with zero size is a problem
		if depth == 0 {
			v.Fail("ZeroSizeCheck", box.NodeID,
				fmt.Sprintf("Root box has zero size: %dx%d", box.Width, box.Height),
				"Root should have non-zero dimensions after layout")
		} else {
			// Non-root zero-size might be intentional (e.g., empty containers)
			v.Warn("NodeID=%d has zero size (%dx%d) at depth %d", box.NodeID, box.Width, box.Height, depth)
			v.Pass() // Not a failure, just a warning
		}
	} else {
		v.Pass()
	}

	for _, child := range box.Children {
		v.validateZeroSizeBoxes(child, depth+1)
	}
}

// Rule 6: Validate tree structure consistency
func (v *Validator) validateTreeStructure(root *compute.ComputedBox) {
	if root == nil {
		return
	}

	var validate func(box *compute.ComputedBox, parent *compute.ComputedBox)
	validate = func(box *compute.ComputedBox, parent *compute.ComputedBox) {
		// Check parent reference consistency
		if parent != nil && box.Parent != parent {
			v.Fail("TreeStructure", box.NodeID,
				"Parent reference is inconsistent",
				fmt.Sprintf("Expected parent NodeID=%d, got NodeID=%d", parent.NodeID, box.Parent.NodeID))
		} else {
			v.Pass()
		}

		for _, child := range box.Children {
			validate(child, box)
		}
	}

	validate(root, nil)
}

// Rule 7: Validate constraints are valid
func (v *Validator) validateConstraintsValid() {
	// MinWidth should not be greater than MaxWidth
	if v.Constraints.MinWidth > v.Constraints.MaxWidth && v.Constraints.MaxWidth > 0 {
		v.Fail("ConstraintsValid", 0,
			fmt.Sprintf("MinWidth (%d) > MaxWidth (%d)", v.Constraints.MinWidth, v.Constraints.MaxWidth),
			"MinWidth must not exceed MaxWidth")
	} else {
		v.Pass()
	}

	// MinHeight should not be greater than MaxHeight
	if v.Constraints.MinHeight > v.Constraints.MaxHeight && v.Constraints.MaxHeight > 0 {
		v.Fail("ConstraintsValid", 0,
			fmt.Sprintf("MinHeight (%d) > MaxHeight (%d)", v.Constraints.MinHeight, v.Constraints.MaxHeight),
			"MinHeight must not exceed MaxHeight")
	} else {
		v.Pass()
	}

	// MaxWidth and MaxHeight should be positive (or -1 for unbounded)
	if v.Constraints.MaxWidth < -1 || v.Constraints.MaxWidth == 0 {
		v.Warn("MaxWidth is %d (typically should be positive or -1 for unbounded)", v.Constraints.MaxWidth)
	}
	if v.Constraints.MaxHeight < -1 || v.Constraints.MaxHeight == 0 {
		v.Warn("MaxHeight is %d (typically should be positive or -1 for unbounded)", v.Constraints.MaxHeight)
	}
}

// =============================================================================
// Main Validation Entry Point
// =============================================================================

// ValidateLayout runs all validation rules on a layout
func ValidateLayout(layout *compute.ComputedLayout, constraints runtime.BoxConstraints) *Validator {
	v := NewValidator(constraints)

	// Run all validation rules
	v.validateConstraintsValid()
	v.validateRootConstraints(layout.Root)
	v.validateNonNegativeDimensions(layout.Root)
	v.validateNodeIDUniqueness(layout.Root)
	v.validateParentChildBoundaries(layout.Root)
	v.validateTreeStructure(layout.Root)
	v.validateZeroSizeBoxes(layout.Root, 0)

	return v
}

// PrintReport prints a detailed validation report
func (v *Validator) PrintReport() {
	fmt.Println("\n" + "========================================")
	fmt.Println("       LAYOUT VALIDATION REPORT")
	fmt.Println("========================================")

	fmt.Printf("\nConstraints: Min(%d, %d) Max(%d, %d)\n",
		v.Constraints.MinWidth, v.Constraints.MinHeight,
		v.Constraints.MaxWidth, v.Constraints.MaxHeight)

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("  PASS:  %d\n", v.PassCount)
	fmt.Printf("  FAIL:  %d\n", v.FailCount)
	fmt.Printf("  SKIP:  %d\n", v.SkipCount)

	// Print failures
	if len(v.Errors) > 0 {
		fmt.Printf("\n--- Failures (%d) ---\n", len(v.Errors))
		for i, err := range v.Errors {
			fmt.Printf("\n[%d] Category: %s\n", i+1, err.Category)
			fmt.Printf("    NodeID: %d\n", err.NodeID)
			fmt.Printf("    Message: %s\n", err.Message)
			fmt.Printf("    Details: %s\n", err.Details)
		}
	}

	// Print warnings
	if len(v.Warnings) > 0 {
		fmt.Printf("\n--- Warnings (%d) ---\n", len(v.Warnings))
		for i, warn := range v.Warnings {
			fmt.Printf("  [%d] %s\n", i+1, warn)
		}
	}

	// Final verdict
	fmt.Println("\n--- Verdict ---")
	if v.FailCount == 0 {
		fmt.Println("  ✓ ALL VALIDATIONS PASSED")
	} else {
		fmt.Printf("  ✗ %d VALIDATION(S) FAILED\n", v.FailCount)
	}
	fmt.Println("========================================")
}

// IsSuccess returns true if all validations passed
func (v *Validator) IsSuccess() bool {
	return v.FailCount == 0
}

// =============================================================================
// Standalone Validation Runner
// =============================================================================

// RunValidation is the main entry point for standalone validation
func RunValidation() bool {
	fmt.Println("=== Automated Layout Constraint Validation ===")
	fmt.Println("This program validates layout constraints without manual inspection.")

	// Build test layout
	vnode := component_fixtures.BuildDemo1App()
	fiber := ui.CreateFiberFromVNode(vnode)
	engine := compute.NewEngine()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MaxHeight: 24,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(fiber, constraints)
	if err != nil {
		fmt.Printf("[FATAL] Layout computation failed: %v\n", err)
		return false
	}

	// Run validation
	validator := ValidateLayout(layout, constraints)

	// Print report
	validator.PrintReport()

	return validator.IsSuccess()
}

// =============================================================================
// Batch Validation for Multiple Fixtures
// =============================================================================

// BatchValidationResult holds results for multiple fixtures
type BatchValidationResult struct {
	FixtureName string
	PassCount   int
	FailCount   int
	Success     bool
	Errors      []ValidationError
}

// RunBatchValidation validates all standard fixtures
func RunBatchValidation() {
	fmt.Println("=== Batch Layout Validation ===")
	fmt.Println("Validating all standard fixtures...")

	fixtures := component_fixtures.StandardFixtures()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	results := make([]BatchValidationResult, 0, len(fixtures))

	for _, f := range fixtures {
		vnode := f.Build()
		fiber := ui.CreateFiberFromVNode(vnode)
		engine := compute.NewEngine()

		layout, err := engine.BuildComputedBoxFiberOnly(fiber, constraints)
		if err != nil {
			results = append(results, BatchValidationResult{
				FixtureName: f.Name,
				PassCount:   0,
				FailCount:   1,
				Success:     false,
				Errors: []ValidationError{{
					Category: "LayoutComputation",
					Message:  err.Error(),
				}},
			})
			continue
		}

		v := ValidateLayout(layout, constraints)
		results = append(results, BatchValidationResult{
			FixtureName: f.Name,
			PassCount:   v.PassCount,
			FailCount:   v.FailCount,
			Success:     v.IsSuccess(),
			Errors:      v.Errors,
		})
	}

	// Print batch report
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    BATCH VALIDATION REPORT")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n%-20s %8s %8s %8s\n", "Fixture", "Pass", "Fail", "Status")
	fmt.Println(strings.Repeat("-", 50))

	totalPass := 0
	totalFail := 0
	allSuccess := true

	for _, r := range results {
		status := "✓ OK"
		if !r.Success {
			status = "✗ FAIL"
			allSuccess = false
		}
		fmt.Printf("%-20s %8d %8d %8s\n", r.FixtureName, r.PassCount, r.FailCount, status)
		totalPass += r.PassCount
		totalFail += r.FailCount
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("%-20s %8d %8d\n", "TOTAL", totalPass, totalFail)

	fmt.Println("\n--- Final Verdict ---")
	if allSuccess {
		fmt.Println("✓ All fixtures passed validation")
	} else {
		fmt.Println("✗ Some fixtures failed validation")
		// Print detailed errors
		for _, r := range results {
			if !r.Success && len(r.Errors) > 0 {
				fmt.Printf("\n%s errors:\n", r.FixtureName)
				for i, err := range r.Errors {
					if i >= 5 {
						fmt.Printf("  ... and %d more errors\n", len(r.Errors)-5)
						break
					}
					fmt.Printf("  - [%s] %s\n", err.Category, err.Message)
				}
			}
		}
	}
}

// ValidateAndExit runs validation and exits with appropriate code
func ValidateAndExit() {
	success := RunValidation()
	if success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// =============================================================================
// ASCII Layout Visualization
// =============================================================================

// ASCIICanvas represents a canvas for drawing ASCII layout
type ASCIICanvas struct {
	Width  int
	Height int
	Grid   [][]rune
}

// NewASCIICanvas creates a new ASCII canvas
func NewASCIICanvas(width, height int) *ASCIICanvas {
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	return &ASCIICanvas{Width: width, Height: height, Grid: grid}
}

// DrawBox draws a box border on the canvas
func (c *ASCIICanvas) DrawBox(x, y, w, h int, label string, style BoxStyle) {
	if w <= 0 || h <= 0 {
		return
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Adjust for boundaries
	right := x + w - 1
	bottom := y + h - 1
	if right >= c.Width {
		right = c.Width - 1
	}
	if bottom >= c.Height {
		bottom = c.Height - 1
	}

	// Draw corners
	c.setCell(x, y, style.TopLeft)
	c.setCell(right, y, style.TopRight)
	c.setCell(x, bottom, style.BottomLeft)
	c.setCell(right, bottom, style.BottomRight)

	// Draw horizontal lines
	for i := x + 1; i < right; i++ {
		c.setCell(i, y, style.Horizontal)
		c.setCell(i, bottom, style.Horizontal)
	}

	// Draw vertical lines
	for i := y + 1; i < bottom; i++ {
		c.setCell(x, i, style.Vertical)
		c.setCell(right, i, style.Vertical)
	}

	// Draw label (truncate if too long)
	if label != "" && w > 4 && h > 0 {
		maxLabelLen := w - 2
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen]
		}
		for i, ch := range label {
			if x+1+i < right {
				c.setCell(x+1+i, y+1, ch)
			}
		}
	}
}

// DrawCross draws a cross marker for child boundaries
func (c *ASCIICanvas) DrawCross(x, y int) {
	c.setCell(x, y, '+')
}

// setCell sets a cell if within bounds
func (c *ASCIICanvas) setCell(x, y int, ch rune) {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		c.Grid[y][x] = ch
	}
}

// BoxStyle defines the style for drawing boxes
type BoxStyle struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

// Standard box styles
var (
	DoubleBox = BoxStyle{'╔', '╗', '╚', '╝', '═', '║'}
	SingleBox = BoxStyle{'┌', '┐', '└', '┘', '─', '│'}
	DottedBox = BoxStyle{'.', '.', '\'', '\'', '.', ':'}
	BoldBox   = BoxStyle{'┏', '┓', '┗', '┛', '━', '┃'}
)

// String returns the canvas as a string
func (c *ASCIICanvas) String() string {
	var sb strings.Builder
	for _, row := range c.Grid {
		sb.WriteString(string(row))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// LayoutVisualizer creates ASCII visualization of layout
type LayoutVisualizer struct {
	MaxWidth  int
	MaxHeight int
	ShowSizes bool
	ShowIDs   bool
	ShowTags  bool
}

// NewLayoutVisualizer creates a new visualizer
func NewLayoutVisualizer(maxWidth, maxHeight int) *LayoutVisualizer {
	return &LayoutVisualizer{
		MaxWidth:  maxWidth,
		MaxHeight: maxHeight,
		ShowSizes: true,
		ShowIDs:   true,
		ShowTags:  false,
	}
}

// Visualize creates an ASCII visualization of the layout
func (v *LayoutVisualizer) Visualize(root *compute.ComputedBox) string {
	if root == nil {
		return "(empty layout)"
	}

	// Determine canvas size
	canvasWidth := v.MaxWidth
	canvasHeight := v.MaxHeight

	// Use root size if larger
	if root.Width > canvasWidth {
		canvasWidth = root.Width
	}
	if root.Height > canvasHeight {
		canvasHeight = root.Height
	}

	// Add padding for labels
	canvasWidth += 20
	canvasHeight += 5

	canvas := NewASCIICanvas(canvasWidth, canvasHeight)

	// Draw all boxes
	v.drawBoxes(canvas, root, 0)

	return canvas.String()
}

// drawBoxes recursively draws boxes on the canvas
func (v *LayoutVisualizer) drawBoxes(canvas *ASCIICanvas, box *compute.ComputedBox, depth int) {
	if box == nil {
		return
	}

	// Choose style based on depth
	var style BoxStyle
	switch depth % 4 {
	case 0:
		style = DoubleBox
	case 1:
		style = BoldBox
	case 2:
		style = SingleBox
	default:
		style = DottedBox
	}

	// Build label
	var label string
	if v.ShowSizes {
		label = fmt.Sprintf("%dx%d", box.Width, box.Height)
	}
	if v.ShowIDs && box.NodeID > 0 {
		if label != "" {
			label = fmt.Sprintf("#%d %s", box.NodeID, label)
		} else {
			label = fmt.Sprintf("#%d", box.NodeID)
		}
	}

	// Draw the box
	canvas.DrawBox(box.X, box.Y, box.Width, box.Height, label, style)

	// Draw children
	for _, child := range box.Children {
		v.drawBoxes(canvas, child, depth+1)
	}
}

// VisualizeWithRuler creates a visualization with ruler markings
func (v *LayoutVisualizer) VisualizeWithRuler(root *compute.ComputedBox) string {
	vis := v.Visualize(root)

	// Add horizontal ruler
	var sb strings.Builder
	sb.WriteString("     ")

	// Determine width from root
	width := 80
	if root != nil && root.Width > width {
		width = root.Width
	}

	// Ruler: every 10 positions marked
	for i := 0; i < width; i++ {
		if i%10 == 0 {
			sb.WriteString("|")
		} else if i%5 == 0 {
			sb.WriteString("+")
		} else {
			sb.WriteString(".")
		}
	}
	sb.WriteByte('\n')

	// Add numbers
	sb.WriteString("     ")
	for i := 0; i < width; i += 10 {
		sb.WriteString(fmt.Sprintf("%-10d", i))
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	// Add visualization
	sb.WriteString(vis)

	return sb.String()
}

// =============================================================================
// Layout Statistics
// =============================================================================

// LayoutStats holds layout statistics
type LayoutStats struct {
	TotalBoxes    int
	TotalWidth    int
	TotalHeight   int
	MaxDepth      int
	AvgBoxWidth   float64
	AvgBoxHeight  float64
	ZeroSizeCount int
	OverFlowCount int
}

// CollectStats collects statistics from layout
func CollectStats(root *compute.ComputedBox) LayoutStats {
	stats := LayoutStats{}

	var collect func(box *compute.ComputedBox, depth int)
	collect = func(box *compute.ComputedBox, depth int) {
		if box == nil {
			return
		}

		stats.TotalBoxes++
		stats.TotalWidth += box.Width
		stats.TotalHeight += box.Height

		if depth > stats.MaxDepth {
			stats.MaxDepth = depth
		}

		if box.Width == 0 || box.Height == 0 {
			stats.ZeroSizeCount++
		}

		for _, child := range box.Children {
			collect(child, depth+1)
		}
	}
	collect(root, 0)

	if stats.TotalBoxes > 0 {
		stats.AvgBoxWidth = float64(stats.TotalWidth) / float64(stats.TotalBoxes)
		stats.AvgBoxHeight = float64(stats.TotalHeight) / float64(stats.TotalBoxes)
	}

	return stats
}

// PrintStats prints layout statistics
func PrintStats(root *compute.ComputedBox) {
	stats := CollectStats(root)

	fmt.Println("\n--- Layout Statistics ---")
	fmt.Printf("  Total Boxes:    %d\n", stats.TotalBoxes)
	fmt.Printf("  Max Depth:      %d\n", stats.MaxDepth)
	fmt.Printf("  Avg Width:      %.1f\n", stats.AvgBoxWidth)
	fmt.Printf("  Avg Height:     %.1f\n", stats.AvgBoxHeight)
	fmt.Printf("  Zero Size:      %d boxes\n", stats.ZeroSizeCount)
}

// =============================================================================
// Visualization Entry Points
// =============================================================================

// VisualizeLayout creates and prints ASCII visualization
func VisualizeLayout(root *compute.ComputedBox, maxWidth, maxHeight int) {
	visualizer := NewLayoutVisualizer(maxWidth, maxHeight)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("               LAYOUT VISUALIZATION")
	fmt.Println(strings.Repeat("=", 60))

	// Print statistics
	PrintStats(root)

	// Print simple box overview
	fmt.Println("\n--- Box Overview ---")
	printBoxOverview(root, 0)

	// Print ASCII visualization
	fmt.Println("\n--- ASCII Layout Diagram ---")
	fmt.Printf("(Legend: Double-line=root, Bold=depth1, Single=depth2, Dotted=depth3+)\n")
	fmt.Printf("(Labels: #NodeID WxH)\n\n")

	vis := visualizer.VisualizeWithRuler(root)
	fmt.Print(vis)
}

// printBoxOverview prints a compact box overview
func printBoxOverview(box *compute.ComputedBox, depth int) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	sizeInfo := fmt.Sprintf("%dx%d", box.Width, box.Height)
	posInfo := fmt.Sprintf("@(%d,%d)", box.X, box.Y)

	// Mark issues
	issue := ""
	if box.Width == 0 || box.Height == 0 {
		issue = " [ZERO-SIZE]"
	}

	fmt.Printf("%s#%-3d %-8s %-12s%s\n", indent, box.NodeID, sizeInfo, posInfo, issue)

	for _, child := range box.Children {
		printBoxOverview(child, depth+1)
	}
}

// =============================================================================
// Grid-Based Visualization
// =============================================================================

// GridVisualizer creates a grid-based visualization
type GridVisualizer struct {
	Grid       [][]string
	Width      int
	Height     int
	Constraints runtime.BoxConstraints
}

// NewGridVisualizer creates a new grid visualizer
func NewGridVisualizer(constraints runtime.BoxConstraints) *GridVisualizer {
	w := constraints.MaxWidth
	h := constraints.MaxHeight
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	grid := make([][]string, h+1)
	for i := range grid {
		grid[i] = make([]string, w+1)
		for j := range grid[i] {
			grid[i][j] = "·"
		}
	}

	return &GridVisualizer{
		Grid:       grid,
		Width:      w,
		Height:     h,
		Constraints: constraints,
	}
}

// DrawBox draws a box on the grid
func (g *GridVisualizer) DrawBox(box *compute.ComputedBox, depth int) {
	if box == nil || box.Width <= 0 || box.Height <= 0 {
		return
	}

	// Clamp to grid bounds
	x := box.X
	y := box.Y
	w := box.Width
	h := box.Height

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > g.Width {
		w = g.Width - x
	}
	if y+h > g.Height {
		h = g.Height - y
	}

	// Choose marker based on depth
	markers := []string{"█", "▓", "▒", "░", "·"}
	marker := markers[depth%len(markers)]

	// Draw filled box
	for row := y; row < y+h && row < g.Height; row++ {
		for col := x; col < x+w && col < g.Width; col++ {
			// Draw corners
			if (row == y || row == y+h-1) && (col == x || col == x+w-1) {
				g.Grid[row][col] = "╬"
			} else if row == y || row == y+h-1 {
				g.Grid[row][col] = "═"
			} else if col == x || col == x+w-1 {
				g.Grid[row][col] = "║"
			} else {
				g.Grid[row][col] = marker
			}
		}
	}
}

// DrawLayout draws all boxes on the grid
func (g *GridVisualizer) DrawLayout(root *compute.ComputedBox) {
	var draw func(box *compute.ComputedBox, depth int)
	draw = func(box *compute.ComputedBox, depth int) {
		if box == nil {
			return
		}
		g.DrawBox(box, depth)
		for _, child := range box.Children {
			draw(child, depth+1)
		}
	}
	draw(root, 0)
}

// String returns the grid as a string
func (g *GridVisualizer) String() string {
	var sb strings.Builder

	// Draw constraint boundary
	sb.WriteString(fmt.Sprintf("Constraints: %dx%d\n", g.Width, g.Height))
	sb.WriteString("┌" + strings.Repeat("─", g.Width) + "┐\n")

	for y := 0; y < g.Height; y++ {
		sb.WriteString("│")
		for x := 0; x < g.Width; x++ {
			sb.WriteString(g.Grid[y][x])
		}
		sb.WriteString("│")
		// Y-axis label
		if y%5 == 0 {
			sb.WriteString(fmt.Sprintf(" %d", y))
		}
		sb.WriteByte('\n')
	}

	sb.WriteString("└" + strings.Repeat("─", g.Width) + "┘\n")

	// X-axis labels
	sb.WriteString(" ")
	for x := 0; x < g.Width; x += 10 {
		sb.WriteString(fmt.Sprintf("%-10d", x))
	}
	sb.WriteByte('\n')

	return sb.String()
}

// VisualizeWithDetails prints detailed layout information
func VisualizeWithDetails(root *compute.ComputedBox, maxWidth, maxHeight int) {
	VisualizeLayout(root, maxWidth, maxHeight)

	// Print tree with details
	fmt.Println("\n--- Layout Tree Details ---")
	printDetailedTree(root, 0)
}

func printDetailedTree(box *compute.ComputedBox, depth int) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	prefix := "├─"
	if depth == 0 {
		prefix = "●"
	}
	fmt.Printf("%s%s #%d [%dx%d @ (%d,%d)]\n",
		indent, prefix, box.NodeID, box.Width, box.Height, box.X, box.Y)

	for i, child := range box.Children {
		if i == len(box.Children)-1 {
			fmt.Printf("%s  └─ ", indent)
			printDetailedTreeNode(child, depth+1, true)
		} else {
			fmt.Printf("%s  ├─ ", indent)
			printDetailedTreeNode(child, depth+1, false)
		}
	}
}

func printDetailedTreeNode(box *compute.ComputedBox, depth int, isLast bool) {
	if box == nil {
		return
	}

	fmt.Printf("#%d [%dx%d @ (%d,%d)]\n", box.NodeID, box.Width, box.Height, box.X, box.Y)

	indent := strings.Repeat("  ", depth)
	for i, child := range box.Children {
		if i == len(box.Children)-1 {
			fmt.Printf("%s  └─ ", indent)
			printDetailedTreeNode(child, depth+1, true)
		} else {
			fmt.Printf("%s  ├─ ", indent)
			printDetailedTreeNode(child, depth+1, false)
		}
	}
}

// RunVisualization runs visualization for demo1 app
func RunVisualization() {
	fmt.Println("=== Layout Visualization ===")

	vnode := component_fixtures.BuildDemo1App()
	fiber := ui.CreateFiberFromVNode(vnode)
	engine := compute.NewEngine()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MaxHeight: 24,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(fiber, constraints)
	if err != nil {
		fmt.Printf("[ERR] Layout failed: %v\n", err)
		return
	}

	// Print statistics and overview
	VisualizeLayout(layout.Root, 100, 50)

	// Print grid visualization
	fmt.Println("\n--- Grid Layout View ---")
	grid := NewGridVisualizer(constraints)
	grid.DrawLayout(layout.Root)
	fmt.Print(grid.String())
}

// RunDetailedVisualization runs visualization with detailed tree output
func RunDetailedVisualization() {
	fmt.Println("=== Detailed Layout Visualization ===")

	vnode := component_fixtures.BuildDemo1App()
	fiber := ui.CreateFiberFromVNode(vnode)
	engine := compute.NewEngine()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MaxHeight: 24,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(fiber, constraints)
	if err != nil {
		fmt.Printf("[ERR] Layout failed: %v\n", err)
		return
	}

	// Run validation first
	validator := ValidateLayout(layout, constraints)
	validator.PrintReport()

	// Then visualize
	VisualizeWithDetails(layout.Root, 100, 50)
}

// =============================================================================
// Fixture Visualization
// =============================================================================

// VisualizeFixture visualizes a specific fixture by name
func VisualizeFixture(name string) {
	fixture := component_fixtures.GetFixture(name)
	if fixture == nil {
		fmt.Printf("[ERR] Fixture '%s' not found\n", name)
		fmt.Println("Available fixtures:")
		for _, f := range component_fixtures.StandardFixtures() {
			fmt.Printf("  - %s\n", f.Name)
		}
		return
	}

	fmt.Printf("=== Visualizing Fixture: %s ===\n", name)

	vnode := fixture.Build()
	fiber := ui.CreateFiberFromVNode(vnode)
	engine := compute.NewEngine()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MaxHeight: 24,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(fiber, constraints)
	if err != nil {
		fmt.Printf("[ERR] Layout failed: %v\n", err)
		return
	}

	VisualizeWithDetails(layout.Root, 100, 50)
}
