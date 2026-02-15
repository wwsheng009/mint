package reconciler

import (
	"github.com/wwsheng009/mint/runtime/event"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// CompleteWork Phase
// =============================================================================
// CompleteWork is where we finalize work on a Fiber node.
// For each Fiber, we:
// 1. Create/Update DOM node (or prepare for rendering)
// 2. Collect child effects
// 3. Return next Fiber to process
//
// This is the "completion" of processing a work unit.
// After CompleteWork, we move to the next work unit in traversal.
// =============================================================================

// CompleteWork completes processing of a Fiber node during render phase
// Returns next Fiber to process (usually nil, since we traverse in workLoop)
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
	_, ok := workInProgress.VNode.(*rtui.ComponentVNode)
	if !ok {
		return workInProgress
	}

	// Store element properties for rendering during commit
	workInProgress.MemoizedProps = workInProgress.Props

	return workInProgress
}

// =============================================================================
// Text CompleteWork
// =============================================================================
// completeWorkText finalizes a text Fiber
func completeWorkText(current, workInProgress *Fiber) *Fiber {
	textVNode, ok := workInProgress.VNode.(*rtui.TextVNode)
	if !ok {
		return workInProgress
	}

	// Store text content for rendering during commit
	workInProgress.MemoizedState = textVNode.Content()

	return workInProgress
}

// =============================================================================
// Element CompleteWork
// =============================================================================
// completeWorkElement finalizes an element Fiber
func completeWorkElement(current, workInProgress *Fiber) *Fiber {
	elementVNode, ok := workInProgress.VNode.(*rtui.ElementVNode)
	if !ok {
		return workInProgress
	}

	// Store element properties for rendering during commit
	workInProgress.MemoizedProps = workInProgress.Props

	// Store tag/component name for reference
	_ = elementVNode.Tag()

	// === Phase 1: Extract layout info to Fiber ===
	// This enables Fiber-first layout by avoiding VNode delegation
	// Use safe interface methods to check for layout properties
	extractLayoutInfoToFiber(workInProgress, elementVNode)

	// === Phase 2: Extract visual style to Fiber ===
	// This enables Fiber-first rendering by storing style info during completeWork
	extractVisualStyleToFiber(workInProgress, elementVNode)

	// === Phase 3: Extract event handlers to Fiber ===
	// This enables Fiber-first event handling by storing handlers on Fiber
	extractEventHandlersToFiber(workInProgress, elementVNode.Props())

	// === Phase 4: Extract ref to Fiber ===
	// This enables Fiber-first ref handling by storing ref on Fiber
	extractRefToFiber(workInProgress, elementVNode.Props())

	return workInProgress
}

// =============================================================================
// Fragment CompleteWork
// =============================================================================
// completeWorkFragment finalizes a fragment Fiber
func completeWorkFragment(current, workInProgress *Fiber) *Fiber {
	// Fragments don't render anything themselves
	// They just group children

	return workInProgress
}

// =============================================================================
// Event Handler Extraction (Phase 3)
// =============================================================================
// extractEventHandlers extracts event handlers from props and stores them on Fiber
// This enables Fiber-first event handling by storing handlers on Fiber
func extractEventHandlersToFiber(fiber *rtui.Fiber, props rtui.Props) {
	if fiber == nil || props == nil {
		return
	}

	// Common event handler keys in props
	// These follow React-style naming: onClick, onKeyDown, onMouseEnter, etc.
	eventKeys := []string{
		"onClick", "onDoubleClick", "onContextMenu",
		"onKeyDown", "onKeyUp", "onKeyPress", "onKeyRepeat",
		"onMouseDown", "onMouseUp", "onMouseMove", "onMouseWheel",
		"onMouseEnter", "onMouseLeave",
		"onFocus", "onBlur", "onChange", "onSubmit", "onCancel",
	}

	handlers := make(map[event.EventType]func(event.Event) bool)

	for _, key := range eventKeys {
		if handler, ok := props[key]; ok {
			// Determine event type based on prop name
			var eventType event.EventType
			switch key {
			case "onClick":
				eventType = event.EventClick
			case "onDoubleClick":
				eventType = event.EventDoubleClick
			case "onContextMenu":
				eventType = event.EventContextMenu
			case "onKeyDown":
				eventType = event.EventKeyPress
			case "onKeyUp":
				eventType = event.EventKeyRelease
			case "onKeyPress":
				eventType = event.EventKeyPress
			case "onKeyRepeat":
				eventType = event.EventKeyRepeat
			case "onMouseDown":
				eventType = event.EventMousePress
			case "onMouseUp":
				eventType = event.EventMouseRelease
			case "onMouseMove":
				eventType = event.EventMouseMove
			case "onMouseWheel":
				eventType = event.EventMouseWheel
			case "onMouseEnter":
				eventType = event.EventMouseEnter
			case "onMouseLeave":
				eventType = event.EventMouseLeave
			case "onFocus":
				eventType = event.EventFocus
			case "onBlur":
				eventType = event.EventBlur
			case "onChange":
				eventType = event.EventChange
			case "onSubmit":
				eventType = event.EventSubmit
			case "onCancel":
				eventType = event.EventCancel
			default:
				continue
			}

			// Type assert handler to func(event.Event) bool
			if handlerFunc, ok := handler.(func(event.Event) bool); ok {
				handlers[eventType] = handlerFunc
			}
		}
	}

	if len(handlers) > 0 {
		fiber.EventHandlers = handlers
	}
}

// extractRefToFiber extracts ref callback from props and stores it on Fiber
// This enables Fiber-first ref handling by storing ref on Fiber
func extractRefToFiber(fiber *rtui.Fiber, props rtui.Props) {
	if fiber == nil || props == nil {
		return
	}

	// Common ref keys in props
	// Ref can be: "ref" (function ref) or a callback function
	var ref interface{}

	if refValue, ok := props["ref"]; ok {
		if refFunc, ok := refValue.(func(interface{})); ok {
			ref = refFunc
		} else if refValue, ok := refValue.(interface{}); ok {
			ref = refValue
		}
	}

	if ref != nil {
		fiber.Ref = ref
	}
}

// =============================================================================
// Effect Collection
// =============================================================================
// collectChildEffects collects effect flags from children
// This bubbles up effect flags from descendant fibers
//
// When called:
// - During performUnitOfWork, after CompleteWork for each fiber
// - Ensures parents know about all descendant effects before commit
//
// Example propagation:
//   Tree before collection:
//     Parent (SubtreeFlags: 0)
//       ├── ChildA (Flags: 2, SubtreeFlags: 4)
//       └── ChildB (Flags: 8, SubtreeFlags: 0)
//
// After collection (Parent.SubtreeFlags = 2 | 4 | 8 = 14):
//     Parent (SubtreeFlags: 14) ← OR of all descendant flags
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

// =============================================================================
// Layout Info Extraction (Phase 1)
// =============================================================================
// extractLayoutInfoToFiber extracts layout properties from VNode to Fiber
// This enables Fiber-first layout by storing layout info during completeWork
func extractLayoutInfoToFiber(fiber *Fiber, vnode rtui.VNode) {
	if fiber == nil || vnode == nil {
		return
	}

	// Use safe interface methods to extract layout info
	// Check for Direction method
	if dirGetter, ok := vnode.(interface{ Direction() rtui.Direction }); ok {
		fiber.LayoutDirection = dirGetter.Direction()
	}
	// Check for Align method
	if alignGetter, ok := vnode.(interface{ Align() rtui.Align }); ok {
		fiber.LayoutAlign = alignGetter.Align()
	}
	// Check for CrossAlign method
	if crossAlignGetter, ok := vnode.(interface{ CrossAlign() rtui.Align }); ok {
		fiber.LayoutCrossAlign = crossAlignGetter.CrossAlign()
	}
	// Check for Gap method
	if gapGetter, ok := vnode.(interface{ Gap() int }); ok {
		fiber.LayoutGap = gapGetter.Gap()
	}
	// Check for Padding method
	if paddingGetter, ok := vnode.(interface{ Padding() [4]int }); ok {
		fiber.LayoutPadding = paddingGetter.Padding()
	}
	// Check for Flex method
	if flexGetter, ok := vnode.(interface{ Flex() int }); ok {
		fiber.LayoutFlex = flexGetter.Flex()
	}

	// === Visual Style Extraction (Phase 2) ===
	// This enables Fiber-first rendering by storing style info during completeWork
	extractVisualStyleToFiber(fiber, vnode)
}

// =============================================================================
// Visual Style Extraction (Phase 2)
// =============================================================================
// extractVisualStyleToFiber extracts visual styling properties from VNode to Fiber
// This enables Fiber-first rendering by storing style info during completeWork
func extractVisualStyleToFiber(fiber *Fiber, vnode rtui.VNode) {
	if fiber == nil || vnode == nil {
		return
	}

	// Get props to extract style from
	props := fiber.MemoizedProps
	if props == nil {
		// Try to get props from VNode
		if propsGetter, ok := vnode.(interface{ Props() rtui.Props }); ok {
			props = propsGetter.Props()
		}
		if props == nil {
			return
		}
	}

	// StyleWidth from "width" prop (pixels or percentage)
	if w, ok := props["width"].(int); ok {
		fiber.StyleWidth = w
	} else if wStr, ok := props["width"].(string); ok {
		// Handle percentage values like "100%", "50%"
		switch wStr {
		case "100%":
			fiber.StyleWidth = -1 // Special marker for 100%
		case "50%":
			fiber.StyleWidth = -2 // Special marker for 50%
		}
	}

	// StyleHeight from "height" prop (pixels or percentage)
	if h, ok := props["height"].(int); ok {
		fiber.StyleHeight = h
	} else if hStr, ok := props["height"].(string); ok {
		switch hStr {
		case "100%":
			fiber.StyleHeight = -1
		case "50%":
			fiber.StyleHeight = -2
		}
	}

	// StyleMargin from "margin" prop (array [top, right, bottom, left])
	if m, ok := props["margin"].([4]interface{}); ok && len(m) >= 4 {
		if top, ok := m[0].(int); ok {
			fiber.StyleMargin[0] = top
		}
		if right, ok := m[1].(int); ok {
			fiber.StyleMargin[1] = right
		}
		if bottom, ok := m[2].(int); ok {
			fiber.StyleMargin[2] = bottom
		}
		if left, ok := m[3].(int); ok {
			fiber.StyleMargin[3] = left
		}
	}

	// StyleBorder from "border" prop (int or bool)
	// For bool true, set all sides to 1; for int, set all sides to that value
	if b, ok := props["border"].(bool); ok && b {
		fiber.StyleBorder = [4]int{1, 1, 1, 1}
	} else if bInt, ok := props["border"].(int); ok && bInt > 0 {
		fiber.StyleBorder = [4]int{bInt, bInt, bInt, bInt}
	}

	// StyleDisplay from "display" prop
	if d, ok := props["display"].(string); ok {
		fiber.StyleDisplay = d
	}

	// StylePosition from "position" prop
	if p, ok := props["position"].(string); ok {
		fiber.StylePosition = p
	}

	// StyleZIndex from "z-index" or "zIndex" prop
	if z, ok := props["z-index"].(int); ok {
		fiber.StyleZIndex = z
	} else if z, ok := props["zIndex"].(int); ok {
		fiber.StyleZIndex = z
	}
}
