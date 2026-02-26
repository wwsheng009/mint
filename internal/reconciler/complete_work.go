package reconciler

// =============================================================================
// CompleteWork Phase
// =============================================================================
// CompleteWork is where we finalize the work on a Fiber node.
// For each Fiber, we:
// 1. Create/Update the DOM node (or prepare for rendering)
// 2. Collect child effects
// 3. Extract event handlers and focusable metadata (Fiber-first)
// 4. Return the next Fiber to process
//
// This is the "completion" of processing a work unit.
// After CompleteWork, we move to the next work unit in the traversal.
//
// IMPORTANT: In Fiber-first architecture:
// - VNode only declares "what I want" (intent)
// - Fiber holds "what I am now" (runtime state)
// - Events and focus are runtime state, so they are extracted here
// =============================================================================

import (
	"fmt"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// CompleteWork completes processing of a Fiber node during the render phase
// Returns the next Fiber to process (usually nil, since we traverse in workLoop)
func CompleteWork(current, workInProgress *Fiber) *Fiber {
	if workInProgress == nil {
		return nil
	}

	// Dispatch based on Fiber type
	switch workInProgress.Type {
	case rtui.VNodeComponent:
		return completeWorkComponent(current, workInProgress)

	case rtui.VNodeText:
		return completeWorkText(current, workInProgress)

	case rtui.VNodeElement:
		return completeWorkElement(current, workInProgress)

	case rtui.VNodeFragment:
		return completeWorkFragment(current, workInProgress)

	default:
		// Unknown type, skip
		return workInProgress
	}
}

// =============================================================================
// Component CompleteWork
// =============================================================================

// completeWorkComponent finalizes a component Fiber
func completeWorkComponent(current, workInProgress *Fiber) *Fiber {
	// Store the rendered result for later use during commit
	workInProgress.MemoizedProps = workInProgress.Props

	// Components don't directly render to buffer
	// Their children are rendered recursively

	return workInProgress
}

// =============================================================================
// Text CompleteWork
// =============================================================================

// completeWorkText finalizes a text Fiber
func completeWorkText(current, workInProgress *Fiber) *Fiber {
	// Store the text content for rendering during commit
	// Text content is stored in MemoizedState
	// ALWAYS update from Props["content"] to ensure latest content is used
	if workInProgress.Props != nil {
		if content, ok := workInProgress.Props["content"].(string); ok {
			workInProgress.MemoizedState = content
		}
	}

	return workInProgress
}

// =============================================================================
// Element CompleteWork
// =============================================================================

// completeWorkElement finalizes an element Fiber
func completeWorkElement(current, workInProgress *Fiber) *Fiber {
	// Store element properties for rendering during commit
	workInProgress.MemoizedProps = workInProgress.Props

	// === Phase 1: Extract layout info to Fiber ===
	// Layout info is already extracted during Fiber creation in CreateFiber
	// No need to re-extract from VNode

	// === Phase 2: Event handlers are already extracted in CreateFiber ===
	// EventHandlers are extracted via interface detection during Fiber creation
	// This is the clean API - no Props-based extraction needed

	// === Phase 3: Extract focusable metadata to Fiber (Fiber-first) ===
	// Focusable is runtime capability, not declaration
	// workInProgress.FocusableMeta = extractFocusableMeta(workInProgress)

	// === Phase 4: Sync border properties from Props (方案 A - 边框作为容器属性) ===
	// 边框是容器的视觉装饰属性，通过 Props 传递到 Fiber
	syncBorderProperties(workInProgress)

	return workInProgress
}

// =============================================================================
// Focusable Metadata Extraction
// =============================================================================

// extractFocusableMeta extracts focusable metadata from Fiber
// This determines if the Fiber can receive focus at runtime
func extractFocusableMeta(fiber *Fiber) *rtui.FocusableMeta {
	if fiber == nil {
		return nil
	}

	props := fiber.Props
	if props == nil {
		props = make(rtui.Props)
	}

	// Check for explicit disabled state
	disabled := false
	if d, ok := props["disabled"].(bool); ok {
		disabled = d
	}

	// Check for explicit tabIndex
	tabIndex := -1
	if ti, ok := props["tabIndex"].(int); ok {
		tabIndex = ti
	}

	// Determine if focusable
	// Priority: explicit tabIndex > disabled check > tag-based defaults
	var focusableMeta *rtui.FocusableMeta

	// If explicitly set tabIndex, use it
	if tabIndex >= 0 {
		focusableMeta = &rtui.FocusableMeta{
			TabIndex: tabIndex,
			Disabled: disabled,
			FocusID:  fiber.Key,
		}
	} else if !disabled {
		// Check tag-based defaults for focusable elements
		switch fiber.Tag {
		case "button", "input", "textarea", "select", "checkbox":
			focusableMeta = &rtui.FocusableMeta{
				TabIndex: 0,
				Disabled: disabled,
				FocusID:  fiber.Key,
			}
		}
	}

	// Use Fiber.NodeID as FocusID (Fiber-first approach)
	// FocusableVNode is DEPRECATED, NodeID provides stable runtime identity
	if focusableMeta != nil && focusableMeta.FocusID == "" {
		focusableMeta.FocusID = fmt.Sprintf("node-%d", fiber.NodeID)
	}

	return focusableMeta
}

// =============================================================================
// Border Property Synchronization (方案 A - 边框作为容器属性)
// =============================================================================

// syncBorderProperties 同步边框属性从 Props 到 Fiber
// 边框是容器的视觉装饰属性，所有容器组件都支持边框
//
// 方案 A 设计：
// - 边框通过 Props["borderStyle"] 和 Props["label"] 传递
// - Modal 使用 Props["title"] 作为边框标签（向后兼容）
// - 属性同步到 Fiber.BorderStyle 和 Fiber.BorderLabel 字段
func syncBorderProperties(fiber *Fiber) {
	if fiber == nil {
		return
	}

	props := fiber.Props
	if props == nil {
		return
	}

	// 从 Props 中读取边框样式
	// 注意：当前实现使用 string 类型，未来可以迁移到 layout.BorderStyle
	if styleProp, ok := props["borderStyle"].(string); ok {
		fiber.BorderStyle = styleProp
	}

	// 从 Props 中读取边框标签
	// 支持 "label" 和 "borderLabel" 两种属性名（为了向后兼容）
	if labelProp, ok := props["label"].(string); ok {
		fiber.BorderLabel = labelProp
	} else if labelProp, ok := props["borderLabel"].(string); ok {
		fiber.BorderLabel = labelProp
	}

	// 特殊处理：Modal 的 title 属性映射到边框标签（向后兼容）
	if fiber.Tag == "modal" {
		if titleProp, ok := props["title"].(string); ok && titleProp != "" {
			fiber.BorderLabel = titleProp
		}
	}
}

// =============================================================================
// Fragment CompleteWork
// =============================================================================

// completeWorkFragment finalizes a fragment Fiber
func completeWorkFragment(current, workInProgress *Fiber) *Fiber {
	// Fragments don't have their own properties
	// They just pass through their children
	return workInProgress
}

// =============================================================================
// Effect Collection
// =============================================================================

// collectChildEffects collects effect flags from children
// This bubbles up effect flags from descendant fibers
//
// SubtreeFlags Propagation Algorithm:
// - Bottom-up aggregation: child flags propagate to parent during render
// - For each child, we OR both child.Flags and child.SubtreeFlags into parent
// - This creates a complete picture of all effects in the subtree
//
// When called:
// - During performUnitOfWork, after CompleteWork for each fiber
// - Ensures parents know about all descendant effects before commit
//
// Example propagation:
//
//	Tree before collection:
//	  Parent (SubtreeFlags: 0)
//	    ├── ChildA (Flags: 2, SubtreeFlags: 4)
//	    └── ChildB (Flags: 8, SubtreeFlags: 0)
//
//	After collection (Parent.SubtreeFlags = 2 | 4 | 8 = 14):
//	  Parent (SubtreeFlags: 14) ← OR of all descendant flags
//
// Note: SubtreeFlags is NOT automatically propagated upward when flags change.
// The entire tree must be re-rendered to update SubtreeFlags. This is acceptable
// because flag changes trigger re-renders anyway.
func collectChildEffects(workInProgress *Fiber) {
	if workInProgress == nil {
		return
	}

	// Collect flags from all children
	child := workInProgress.Child
	for child != nil {
		// Merge child's flags into parent's subtree flags
		workInProgress.SubtreeFlags |= child.Flags
		workInProgress.SubtreeFlags |= child.SubtreeFlags

		child = child.Sibling
	}
}
