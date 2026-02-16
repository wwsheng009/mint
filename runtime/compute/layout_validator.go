package compute

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime"
)

// LayoutValidator 布局验证器，检查布局的有效性
type LayoutValidator struct {
	StrictMode bool // 严格模式，任何问题都报错
}

// NewLayoutValidator 创建布局验证器
func NewLayoutValidator() *LayoutValidator {
	return &LayoutValidator{
		StrictMode: false,
	}
}

// ValidationIssue 表示一个验证问题
type ValidationIssue struct {
	Type        string        // 问题类型
	Severity    string        // "error", "warning", "info"
	Path        string        // 节点路径
	Message     string        // 问题描述
	NodeID      uint64        // 节点ID
	Box         runtime.Box   // 节点边界
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid    bool              // 是否有效
	Issues   []ValidationIssue // 所有问题
	Summary  map[string]int    // 问题统计
}

// Validate 全面验证布局
func (v *LayoutValidator) Validate(layout *ComputedLayout) *ValidationResult {
	result := &ValidationResult{
		Valid:   true,
		Issues:  make([]ValidationIssue, 0),
		Summary: make(map[string]int),
	}

	if layout == nil || layout.Root == nil {
		result.addIssue("error", "LayoutNil", "", "Layout or root is nil", 0, runtime.Box{})
		return result
	}

	// 1. 验证根节点
	v.validateRoot(layout.Root, result)

	// 2. 递归验证所有节点
	v.validateNode(layout.Root, "ROOT", result)

	// 3. 验证父子关系
	v.validateParentChildRelations(layout.Root, "ROOT", result)

	// 4. 验证重叠检测
	v.validateOverlaps(layout.Root, result)

	// 5. 验证边界约束
	v.validateBoundaryConstraints(layout.Root, result)

	return result
}

// validateRoot 验证根节点
func (v *LayoutValidator) validateRoot(root *ComputedBox, result *ValidationResult) {
	if root.Box.Width <= 0 || root.Box.Height <= 0 {
		result.addIssue("error", "InvalidRootSize", "ROOT", 
			fmt.Sprintf("Root has invalid size: %dx%d", root.Box.Width, root.Box.Height),
			root.NodeID, root.Box)
	}

	if root.Box.X != 0 || root.Box.Y != 0 {
		result.addIssue("warning", "RootNotAtOrigin", "ROOT",
			fmt.Sprintf("Root not at origin: (%d,%d)", root.Box.X, root.Box.Y),
			root.NodeID, root.Box)
	}
}

// validateNode 验证单个节点
func (v *LayoutValidator) validateNode(box *ComputedBox, path string, result *ValidationResult) {
	if box == nil {
		return
	}

	// 检查节点ID
	if box.NodeID == 0 {
		result.addIssue("warning", "ZeroNodeID", path,
			"Node has zero NodeID",
			box.NodeID, box.Box)
	}

	// 检查尺寸有效性
	if box.Box.Width < 0 {
		result.addIssue("error", "NegativeWidth", path,
			fmt.Sprintf("Negative width: %d", box.Box.Width),
			box.NodeID, box.Box)
	}
	if box.Box.Height < 0 {
		result.addIssue("error", "NegativeHeight", path,
			fmt.Sprintf("Negative height: %d", box.Box.Height),
			box.NodeID, box.Box)
	}

	// 检查位置是否合理
	if box.Box.X < 0 || box.Box.Y < 0 {
		result.addIssue("warning", "NegativePosition", path,
			fmt.Sprintf("Negative position: (%d,%d)", box.Box.X, box.Box.Y),
			box.NodeID, box.Box)
	}

	// 检查零尺寸节点
	if box.Box.Width == 0 || box.Box.Height == 0 {
		result.addIssue("info", "ZeroSize", path,
			fmt.Sprintf("Node has zero size: %dx%d", box.Box.Width, box.Box.Height),
			box.NodeID, box.Box)
	}

	// 检查VNode是否存在
	if box.VNode == nil {
		result.addIssue("error", "NilVNode", path,
			"ComputedBox has nil VNode",
			box.NodeID, box.Box)
	}

	// 递归验证子节点
	for i, child := range box.Children {
		childPath := fmt.Sprintf("%s/child[%d]", path, i)
		v.validateNode(child, childPath, result)
	}
}

// validateParentChildRelations 验证父子关系
func (v *LayoutValidator) validateParentChildRelations(box *ComputedBox, path string, result *ValidationResult) {
	if box == nil || len(box.Children) == 0 {
		return
	}

	for i, child := range box.Children {
		childPath := fmt.Sprintf("%s/child[%d]", path, i)

		// 检查子节点是否在父节点边界内
		if !v.isChildWithinParent(child.Box, box.Box) {
			result.addIssue("warning", "ChildOutOfBounds", childPath,
				fmt.Sprintf("Child at (%d,%d) size %dx%d extends outside parent at (%d,%d) size %dx%d",
					child.Box.X, child.Box.Y, child.Box.Width, child.Box.Height,
					box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height),
				child.NodeID, child.Box)
		}

		// 检查父子NodeID是否相同
		if box.NodeID == child.NodeID && box.NodeID != 0 {
			result.addIssue("error", "ParentChildSameID", childPath,
				fmt.Sprintf("Parent and child have same NodeID: %d", box.NodeID),
				child.NodeID, child.Box)
		}

		// 递归检查
		v.validateParentChildRelations(child, childPath, result)
	}
}

// validateOverlaps 验证兄弟节点重叠
func (v *LayoutValidator) validateOverlaps(root *ComputedBox, result *ValidationResult) {
	// 收集所有叶子节点的边界
	leaves := v.collectLeafBoxes(root, "ROOT")
	
	// 检查重叠
	for i := 0; i < len(leaves); i++ {
		for j := i + 1; j < len(leaves); j++ {
			if v.boxesOverlap(leaves[i].Box, leaves[j].Box) {
				result.addIssue("warning", "Overlap", leaves[i].Path,
					fmt.Sprintf("Node overlaps with %s", leaves[j].Path),
					leaves[i].NodeID, leaves[i].Box)
			}
		}
	}
}

// validateBoundaryConstraints 验证边界约束
func (v *LayoutValidator) validateBoundaryConstraints(root *ComputedBox, result *ValidationResult) {
	// 收集所有节点
	allNodes := v.collectAllBoxes(root, "ROOT")

	for _, node := range allNodes {
		// 检查极大尺寸（可能是错误）
		if node.Box.Width > 1000 || node.Box.Height > 1000 {
			result.addIssue("warning", "OversizedNode", node.Path,
				fmt.Sprintf("Node is unusually large: %dx%d", node.Box.Width, node.Box.Height),
				node.NodeID, node.Box)
		}

		// 检查位置是否超出根节点
		if node.Box.X+node.Box.Width > root.Box.Width ||
			node.Box.Y+node.Box.Height > root.Box.Height {
			result.addIssue("info", "ExtendsBeyondRoot", node.Path,
				fmt.Sprintf("Node extends beyond root boundary"),
				node.NodeID, node.Box)
		}
	}
}

// boxInfo 用于收集节点信息
type boxInfo struct {
	Box    runtime.Box
	Path   string
	NodeID uint64
}

// isChildWithinParent 检查子节点是否在父节点边界内
func (v *LayoutValidator) isChildWithinParent(child, parent runtime.Box) bool {
	childRight := child.X + child.Width
	childBottom := child.Y + child.Height
	parentRight := parent.X + parent.Width
	parentBottom := parent.Y + parent.Height

	return child.X >= parent.X && child.Y >= parent.Y &&
		childRight <= parentRight && childBottom <= parentBottom
}

// boxesOverlap 检查两个边界框是否重叠
func (v *LayoutValidator) boxesOverlap(a, b runtime.Box) bool {
	if a.Width == 0 || a.Height == 0 || b.Width == 0 || b.Height == 0 {
		return false
	}

	aRight := a.X + a.Width
	aBottom := a.Y + a.Height
	bRight := b.X + b.Width
	bBottom := b.Y + b.Height

	// 检查是否重叠（允许边框接触）
	return a.X < bRight && aRight > b.X && a.Y < bBottom && aBottom > b.Y
}

// collectLeafBoxes 收集所有叶子节点
func (v *LayoutValidator) collectLeafBoxes(box *ComputedBox, path string) []boxInfo {
	if box == nil {
		return nil
	}

	var result []boxInfo

	if len(box.Children) == 0 {
		result = append(result, boxInfo{
			Box:    box.Box,
			Path:   path,
			NodeID: box.NodeID,
		})
	}

	for i, child := range box.Children {
		childPath := fmt.Sprintf("%s/child[%d]", path, i)
		result = append(result, v.collectLeafBoxes(child, childPath)...)
	}

	return result
}

// collectAllBoxes 收集所有节点
func (v *LayoutValidator) collectAllBoxes(box *ComputedBox, path string) []boxInfo {
	if box == nil {
		return nil
	}

	result := []boxInfo{{
		Box:    box.Box,
		Path:   path,
		NodeID: box.NodeID,
	}}

	for i, child := range box.Children {
		childPath := fmt.Sprintf("%s/child[%d]", path, i)
		result = append(result, v.collectAllBoxes(child, childPath)...)
	}

	return result
}

// addIssue 添加验证问题
func (r *ValidationResult) addIssue(severity, issueType, path, message string, nodeID uint64, box runtime.Box) {
	issue := ValidationIssue{
		Type:     issueType,
		Severity: severity,
		Path:     path,
		Message:  message,
		NodeID:   nodeID,
		Box:      box,
	}
	r.Issues = append(r.Issues, issue)
	r.Summary[issueType]++

	if severity == "error" {
		r.Valid = false
	}
}

// HasErrors 检查是否有错误
func (r *ValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

// HasWarnings 检查是否有警告
func (r *ValidationResult) HasWarnings() bool {
	for _, issue := range r.Issues {
		if issue.Severity == "warning" {
			return true
		}
	}
	return false
}

// ErrorCount 错误数量
func (r *ValidationResult) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			count++
		}
	}
	return count
}

// WarningCount 警告数量
func (r *ValidationResult) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "warning" {
			count++
		}
	}
	return count
}

// PrintReport 打印验证报告
func (r *ValidationResult) PrintReport() {
	fmt.Println("=== Layout Validation Report ===")
	fmt.Printf("Valid: %v\n", r.Valid)
	fmt.Printf("Total Issues: %d\n", len(r.Issues))
	fmt.Printf("  Errors: %d\n", r.ErrorCount())
	fmt.Printf("  Warnings: %d\n", r.WarningCount())
	fmt.Println()

	if len(r.Issues) > 0 {
		fmt.Println("--- Issues by Type ---")
		for issueType, count := range r.Summary {
			fmt.Printf("  %s: %d\n", issueType, count)
		}
		fmt.Println()

		fmt.Println("--- Details ---")
		for i, issue := range r.Issues {
			severityIcon := "ℹ️"
			if issue.Severity == "error" {
				severityIcon = "❌"
			} else if issue.Severity == "warning" {
				severityIcon = "⚠️"
			}
			fmt.Printf("%d. %s [%s] %s\n", i+1, severityIcon, issue.Type, issue.Message)
			fmt.Printf("   Path: %s, NodeID: %d, Box: (%d,%d,%d,%d)\n",
				issue.Path, issue.NodeID,
				issue.Box.X, issue.Box.Y, issue.Box.Width, issue.Box.Height)
		}
	}
}
