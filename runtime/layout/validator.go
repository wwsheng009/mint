package layout

import (
	"fmt"
)

// ==============================================================================
// BoundsValidator (V3)
// ==============================================================================
// BoundsValidator 用于验证布局结果的正确性

// Severity 验证问题严重性
type Severity int

const (
	SeverityInfo    Severity = iota
	SeverityWarning Severity = iota
	SeverityError   Severity = iota
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "Info"
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ValidationProblem 验证问题
type ValidationProblem struct {
	Box        *LayoutBox
	Problem    string
	Description string
	Severity   Severity
}

// BoundsValidator 边界验证器
// 用于检查布局结果中的各种问题
type BoundsValidator struct {
	maxOverlaps int
	maxDepth    int
	strict      bool
}

// NewBoundsValidator 创建新的边界验证器
func NewBoundsValidator() *BoundsValidator {
	return &BoundsValidator{
		maxOverlaps: 100, // 允许最多100个重叠对
		maxDepth:    1000, // 允许最多1000层深度
		strict:      false,
	}
}

// SetMaxOverlaps 设置允许的最大重叠数量
func (v *BoundsValidator) SetMaxOverlaps(max int) {
	v.maxOverlaps = max
}

// SetMaxDepth 设置允许的最大深度
func (v *BoundsValidator) SetMaxDepth(max int) {
	v.maxDepth = max
}

// SetStrict 设置严格模式
// 在严格模式下，更多的检查会被执行
func (v *BoundsValidator) SetStrict(strict bool) {
	v.strict = strict
}

// ValidateBox 验证单个盒子
func (v *BoundsValidator) ValidateBox(box *LayoutBox) []ValidationProblem {
	if box == nil {
		return []ValidationProblem{}
	}

	var problems []ValidationProblem

	// 检查尺寸是否为正
	if box.Width <= 0 {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "NonPositiveWidth",
			Description: fmt.Sprintf("Width is not positive: %d", box.Width),
			Severity:   SeverityError,
		})
	}

	if box.Height <= 0 {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "NonPositiveHeight",
			Description: fmt.Sprintf("Height is not positive: %d", box.Height),
			Severity:   SeverityError,
		})
	}

	// 在严格模式下，检查位置是否为负
	if v.strict {
		if box.X < 0 {
			problems = append(problems, ValidationProblem{
				Box:        box,
				Problem:    "NegativePosition",
				Description: fmt.Sprintf("X position is negative: %d", box.X),
				Severity:   SeverityWarning,
			})
		}

		if box.Y < 0 {
			problems = append(problems, ValidationProblem{
				Box:        box,
				Problem:    "NegativePosition",
				Description: fmt.Sprintf("Y position is negative: %d", box.Y),
				Severity:   SeverityWarning,
			})
		}
	}

	return problems
}

// ValidateTree 验证整个布局树
func (v *BoundsValidator) ValidateTree(root *LayoutBox) []ValidationProblem {
	if root == nil {
		return []ValidationProblem{}
	}

	var problems []ValidationProblem

	// 收集所有盒子
	boxes := v.collectBoxes(root)

	// 验证每个盒子
	for _, box := range boxes {
		boxProblems := v.ValidateBox(box)
		problems = append(problems, boxProblems...)
	}

	// 检查重叠
	overlapProblems := v.checkOverlaps(boxes)
	problems = append(problems, overlapProblems...)

	// 检查深度
	depthProblems := v.checkDepth(root, 0)
	problems = append(problems, depthProblems...)

	// 在严格模式下，检查树结构完整性
	if v.strict {
		structureProblems := v.validateTreeStructure(root)
		problems = append(problems, structureProblems...)
	}

	return problems
}

// collectBoxes 收集布局树中的所有盒子
func (v *BoundsValidator) collectBoxes(root *LayoutBox) []*LayoutBox {
	if root == nil {
		return []*LayoutBox{}
	}

	boxes := []*LayoutBox{root}
	for _, child := range root.Children {
		boxes = append(boxes, v.collectBoxes(child)...)
	}

	return boxes
}

// checkOverlaps 检查盒子之间的重叠
// 注意：父子节点的重叠是正常的，只检查非父子节点之间的重叠
func (v *BoundsValidator) checkOverlaps(boxes []*LayoutBox) []ValidationProblem {
	var problems []ValidationProblem
	overlapCount := 0

	for i, box1 := range boxes {
		for j, box2 := range boxes {
			if i >= j {
				continue
			}

			// 跳过父子关系
			if v.isAncestor(box1, box2) || v.isAncestor(box2, box1) {
				continue
			}

			// 检查是否重叠
			rect1 := Rect{X: box1.X, Y: box1.Y, Width: box1.Width, Height: box1.Height}
			rect2 := Rect{X: box2.X, Y: box2.Y, Width: box2.Width, Height: box2.Height}

			if rect1.Intersects(rect2) {
				overlapCount++

				if overlapCount <= v.maxOverlaps {
					problems = append(problems, ValidationProblem{
						Box:        box1,
						Problem:    "Overlap",
						Description: fmt.Sprintf("Box %s overlaps with box %s", box1.ID, box2.ID),
						Severity:   SeverityWarning,
					})
				}
			}
		}
	}

	// 如果重叠数量超过限制，添加一个汇总问题
	if overlapCount > v.maxOverlaps {
		problems = append(problems, ValidationProblem{
			Box:        boxes[0],
			Problem:    "TooManyOverlaps",
			Description: fmt.Sprintf("Found %d overlaps (max allowed: %d)", overlapCount, v.maxOverlaps),
			Severity:   SeverityError,
		})
	}

	return problems
}

// isAncestor 检查box1是否是box2的祖先
func (v *BoundsValidator) isAncestor(box1, box2 *LayoutBox) bool {
	if box1 == nil || box2 == nil {
		return false
	}

	// 递归检查box2的子节点
	for _, child := range box1.Children {
		if child == box2 {
			return true
		}
		if v.isAncestor(child, box2) {
			return true
		}
	}

	return false
}

// checkDepth 检查树的深度
func (v *BoundsValidator) checkDepth(box *LayoutBox, currentDepth int) []ValidationProblem {
	if box == nil {
		return []ValidationProblem{}
	}

	var problems []ValidationProblem

	// 检查当前深度是否超过限制
	if currentDepth > v.maxDepth {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "MaxDepthExceeded",
			Description: fmt.Sprintf("Tree depth %d exceeds maximum %d", currentDepth, v.maxDepth),
			Severity:   SeverityError,
		})
	}

	// 递归检查子节点
	for _, child := range box.Children {
		childProblems := v.checkDepth(child, currentDepth+1)
		problems = append(problems, childProblems...)
	}

	return problems
}

// validateTreeStructure 验证树结构完整性
func (v *BoundsValidator) validateTreeStructure(root *LayoutBox) []ValidationProblem {
	if root == nil {
		return []ValidationProblem{}
	}

	var problems []ValidationProblem

	// 检查循环引用
	visited := make(map[*LayoutBox]bool)
	if v.hasCycle(root, visited) {
		problems = append(problems, ValidationProblem{
			Box:        root,
			Problem:    "CycleDetected",
			Description: "Tree contains a cycle",
			Severity:   SeverityError,
		})
	}

	return problems
}

// hasCycle 检查树中是否存在循环
func (v *BoundsValidator) hasCycle(box *LayoutBox, visited map[*LayoutBox]bool) bool {
	if box == nil {
		return false
	}

	// 如果已经访问过，说明有循环
	if visited[box] {
		return true
	}

	// 标记为已访问
	visited[box] = true

	// 递归检查子节点
	for _, child := range box.Children {
		if v.hasCycle(child, visited) {
			return true
		}
	}

	return false
}

// ValidateWithinConstraints 验证盒子是否在约束范围内
func (v *BoundsValidator) ValidateWithinConstraints(box *LayoutBox, constraints Constraints) []ValidationProblem {
	if box == nil {
		return []ValidationProblem{}
	}

	var problems []ValidationProblem

	// 检查宽度是否违反约束
	if constraints.MaxWidth < MaxInt && box.Width > constraints.MaxWidth {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "MaxWidthViolation",
			Description: fmt.Sprintf("Width %d exceeds maximum %d", box.Width, constraints.MaxWidth),
			Severity:   SeverityError,
		})
	}

	if box.Width < constraints.MinWidth {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "MinWidthViolation",
			Description: fmt.Sprintf("Width %d is less than minimum %d", box.Width, constraints.MinWidth),
			Severity:   SeverityError,
		})
	}

	// 检查高度是否违反约束
	if constraints.MaxHeight < MaxInt && box.Height > constraints.MaxHeight {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "MaxHeightViolation",
			Description: fmt.Sprintf("Height %d exceeds maximum %d", box.Height, constraints.MaxHeight),
			Severity:   SeverityError,
		})
	}

	if box.Height < constraints.MinHeight {
		problems = append(problems, ValidationProblem{
			Box:        box,
			Problem:    "MinHeightViolation",
			Description: fmt.Sprintf("Height %d is less than minimum %d", box.Height, constraints.MinHeight),
			Severity:   SeverityError,
		})
	}

	return problems
}
