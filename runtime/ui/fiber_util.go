package ui

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/log"
	fcontext "github.com/wwsheng009/mint/runtime/context"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
)

// =============================================================================
// NodeID Generation
// =============================================================================
// Global ID allocator for NodeID generation
var nodeIDGenerator uint64 = 0

// generateNodeID generates a new unique NodeID (internal use)
// This provides stable runtime identity for Fiber nodes
// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md
func generateNodeID() uint64 {
	nodeIDGenerator++
	return nodeIDGenerator
}

// AllocateNodeID generates a new unique NodeID (exported for external use)
// This provides stable runtime identity for Fiber nodes
// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md
func AllocateNodeID() uint64 {
	return generateNodeID()
}

// =============================================================================
// Fiber Creation
// =============================================================================

// CreateFiber creates a new fiber from a VNode
// This function extracts ALL data from VNode during creation - NO VNode reference is kept
func CreateFiber(vnode VNode) *Fiber {
	if vnode == nil {
		return nil
	}

	// Automatically call Build() if VNode implements Buildable interface
	// This allows builders like LayoutBuilder to be finalized automatically
	// Users can write: ui.NewVStack().SingleBorder("Title") without manual Build()
	if buildable, ok := vnode.(Buildable); ok {
		vnode = buildable.Build()
	}

	vnodeType := vnode.Type()

	// Determine tag from VNode
	// Priority: Tag() method > Name() method > type-specific fallback
	var tag string
	var componentName string
	var componentFunc ComponentFunc
	var componentFuncWithProps ComponentFuncWithProps
	var errorBoundaryFunc ComponentFunc
	var errorBoundaryFallback *Fiber
	var errorBoundaryVNode *ErrorBoundaryVNode // Store reference for state sync
	var memoCompare PropsEqual

	switch n := vnode.(type) {
	case interface{ Tag() string }:
		tag = n.Tag()
	case interface{ Name() string }:
		tag = n.Name()
		componentName = n.Name()
	default:
		if vnodeType == VNodeText {
			tag = "text"
		} else {
			tag = "unknown"
		}
	}

	// Extract special VNode type data
	switch n := vnode.(type) {
	case *ComponentVNode:
		componentFunc = n.Fn()
		componentFuncWithProps = n.FnWithProps()
		if componentName == "" {
			componentName = n.Name()
		}
	case *ErrorBoundaryVNode:
		errorBoundaryFunc = n.Component()
		if fallback := n.Fallback(); fallback != nil {
			errorBoundaryFallback = CreateFiber(fallback)
		}
		errorBoundaryVNode = n // Store reference for state sync
	case *MemoVNode:
		memoCompare = n.GetCompare()
	}

	// Debug logging to understand VNode types
	log.HitMapLogger.IfEnabled().Debug("[CREATEFIBER] Type=%s Key=%s Tag=%s", vnodeType.String(), vnode.Key(), tag)

	// ✨ DiffKey: Copy from VNode.Key() without any modification
	diffKey := vnode.Key()

	// ✨ ID: Copy from VNode.ID() for business reference/positioning
	businessID := vnode.ID()

	// Get props and ensure children are stored for elements/fragments
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	// For elements and fragments, store children in Props so beginWork can access them
	children := vnode.Children()
	if vnodeType == VNodeElement || vnodeType == VNodeFragment {
		if len(children) > 0 {
			// Copy props to avoid mutating the original
			newProps := make(Props)
			for k, v := range props {
				newProps[k] = v
			}
			newProps["children"] = children
			props = newProps
		}
	}

	// For text nodes, store content in MemoizedState for easier access
	var memoizedState interface{}
	if vnodeType == VNodeText {
		// Try Content() method first (TextVNode stores content in struct field)
		if contentProvider, ok := vnode.(interface{ Content() string }); ok {
			memoizedState = contentProvider.Content()
		} else if content, ok := props["content"]; ok {
			memoizedState = content
		}
	}

	// Extract FocusableVNode if the VNode implements it (for buttons, inputs, etc.)
	// DEPRECATED: Use Instance instead for Fiber-first approach
	// var focusableVNode FocusableVNode
	// if f, ok := vnode.(FocusableVNode); ok && f.IsFocusable() {
	// 	focusableVNode = f
	// }

	// === Fiber-first: Instance Creation ===
	// ✅ CORRECT APPROACH: Create instances here in CreateFiber from VNode.InstanceFactory
	//
	// Architecture:
	// - VNode with InstanceFactory (Button, VStack, Text) → creates specific Instance
	// - VNode without InstanceFactory (ComponentFunc, Element) → instance stays nil
	// - beginWorkComponent() then:
	//   1. Reuses existing instance from Fiber (from previous render via CloneFiber)
	//   2. Or creates BaseComponentInstance via InstanceManager for nil instances
	//
	// This ensures:
	// 1. Button gets ButtonInstance (with ResetPressed, Paint, etc.)
	// 2. Stack gets StackInstance (with layout caching)
	// 3. ComponentFunc gets BaseComponentInstance
	// 4. Instance persists across renders (CloneFiber reuses Fiber.Instance)
	var instance ComponentInstance

	// Check if VNode implements InstanceFactory
	if factory, ok := vnode.(InstanceFactory); ok {
		// VNode defines its own instance type
		instance = factory.CreateInstance()

		// IntentEmitter routing strategy:
		// - Global Intent (IsGlobal() = true or not implemented): Send to global Runtime
		// - Local Intent (IsGlobal() = false): Bubble locally through Parent()
		//
		// This design enables:
		// 1. Backward compatibility: All existing intents default to global
		// 2. Optional Intent Bubble: Components can use local intents for communication
		// 3. Clear separation: State management vs component internal logic
		//
		// See: docs/ui/mint2/INTENT_ARCHITECTURE_ANALYSIS.md
		if setter, ok := instance.(interface{ SetIntentEmitter(func(i intent.Intent)) }); ok {
			setter.SetIntentEmitter(func(i intent.Intent) {
				// Check if this is a local intent (implements GlobalIntent with IsGlobal() = false)
				isLocal := false
				if global, ok := i.(intent.GlobalIntent); ok {
					isLocal = !global.IsGlobal()
				}

				if isLocal {
					// Local Intent: Bubble locally through Parent() chain
					// This enables component internal communication and event delegation
					if component, ok := instance.(intent.TreeComponent); ok {
						intent.Emit(component, i)
					}
				} else {
					// Global Intent: Send to global Intent Runtime
					// This goes through Reducer/FieldMap handlers for state management
					if runtime := GetGlobalIntentRuntime(); runtime != nil {
						result := runtime.Emit(i)
						if result.Error != nil {
							// Log intent emission errors
							if log.UILogger.Enabled() {
								log.UILogger.Debug("[IntentEmitter] Failed to emit intent %s: %v",
									i.IntentType(), result.Error)
							} else {
								fmt.Printf("[IntentEmitter] Failed to emit intent %s: %v\n",
									i.IntentType(), result.Error)
							}
						}
					}
				}
			})
		}
	}
	// Otherwise, instance stays nil (will be created in beginWorkComponent)

	// === Extract Layout Properties from VNode ===
	// These are used by the layout engine to determine how to position children.
	var layoutDirection Direction
	var layoutAlign Align
	var layoutCrossAlign Align
	var layoutGap int
	var layoutPadding [4]int
	var layoutMargin [4]int
	var layoutFlex int

	// ✨ Border Properties (方案 A - 边框作为容器属性)
	var borderStyle string
	var borderLabel string

	// Determine direction from tag (hstack = Row, vstack = Column)
	// This works for any VNode that has Tag() method returning "hstack"/"vstack"
	switch tag {
	case "hstack", "row":
		layoutDirection = DirectionRow
	case "vstack", "column":
		layoutDirection = DirectionColumn
	case "bordered":
		// 边框组件默认无边框
		borderStyle = "none"
	case "modal":
		// Modal 默认双线边框
		borderStyle = "double"
	}

	// Extract other layout properties from Props
	if props != nil {
		// direction may be stored as int (rtui.Direction) or custom type
		if dir, ok := props["direction"].(int); ok {
			layoutDirection = Direction(dir)
		}
		// Also check for rtui.Direction type
		if dir, ok := props["direction"].(Direction); ok {
			layoutDirection = dir
		}
		// Extract align - rtui.Align type (or int for backward compatibility)
		if a, ok := props["align"].(Align); ok {
			layoutAlign = a
		} else if a, ok := props["align"].(int); ok {
			layoutAlign = Align(a)
		}
		// Extract crossAlign - rtui.Align type (or int for backward compatibility)
		if ca, ok := props["crossAlign"].(Align); ok {
			layoutCrossAlign = ca
		} else if ca, ok := props["crossAlign"].(int); ok {
			layoutCrossAlign = Align(ca)
		}
		if g, ok := props["gap"].(int); ok {
			layoutGap = g
		}
		if p, ok := props["padding"].([4]int); ok {
			layoutPadding = p
		}
		if m, ok := props["margin"].([4]int); ok {
			layoutMargin = m
		}
		if f, ok := props["flex"].(int); ok {
			layoutFlex = f
		}

		// ✨ Extract border properties from Props
		// 支持多种边框样式名称
		if bs, ok := props["borderStyle"].(string); ok {
			borderStyle = bs
		}
		if label, ok := props["label"].(string); ok {
			borderLabel = label
		}
		// Modal 的 title 属性映射到边框标签
		if tag == "modal" {
			if title, ok := props["title"].(string); ok {
				borderLabel = title
			}
		}
		// 边框组件的 style 属性映射到边框样式
		if tag == "bordered" {
			if style, ok := props["style"].(string); ok {
				borderStyle = style
			}
		}
	}
	nodeId := generateNodeID()
	return &Fiber{
		Type:                       vnodeType,
		Tag:                        tag,
		Props:                      props,
		MemoizedProps:              props,
		MemoizedState:              memoizedState,
		DiffKey:                    diffKey,
		Key:                        diffKey,
		ID:                         businessID, // ✨ Business identifier for positioning/reference
		NodeID:                     nodeId,
		Layer:                      vnode.GetLayer(),
		Style:                      vnode.Style(),
		Lanes:                      LaneNoLane,
		ChildLanes:                 LaneNoLane,
		Flags:                      EffectNoEffect,
		SubtreeFlags:               EffectNoEffect,
		ComponentFunc:              componentFunc,
		ComponentFuncWithProps:     componentFuncWithProps,
		ComponentName:              componentName,
		ErrorBoundaryFunc:          errorBoundaryFunc,
		ErrorBoundaryFallbackFiber: errorBoundaryFallback,
		ErrorBoundary:              errorBoundaryVNode, // Store reference for state sync
		MemoCompare:                memoCompare,
		// FocusableVNode:             focusableVNode,
		ActionTargetID: fmt.Sprintf("%d", nodeId),
		// Fiber-first Architecture
		Instance: instance,
		// Layout Properties (extracted from VNode)
		LayoutDirection:  layoutDirection,
		LayoutAlign:      layoutAlign,
		LayoutCrossAlign: layoutCrossAlign,
		LayoutGap:        layoutGap,
		LayoutPadding:    layoutPadding,
		LayoutMargin:     layoutMargin,
		LayoutFlex:       layoutFlex,
		// ✨ Border Properties (方案 A)
		BorderStyle: borderStyle,
		BorderLabel: borderLabel,
	}
}

// CreateFiberFromVNode creates a fiber tree from a VNode tree
func CreateFiberFromVNode(vnode VNode) *Fiber {
	root := CreateFiber(vnode)
	if root == nil {
		return nil
	}

	// Initialize root Context (Phase 2)
	root.Context = fcontext.NewContext(nil)

	// Build fiber tree for children
	buildFiberTree(root, vnode)
	return root
}

// buildFiberTree recursively builds fiber tree for children
func buildFiberTree(parentFiber *Fiber, parentVNode VNode) {
	children := parentVNode.Children()
	if len(children) == 0 {
		return
	}

	var previousChild *Fiber
	for i, childVNode := range children {
		childFiber := CreateFiber(childVNode)

		// Link to parent
		childFiber.Return = parentFiber

		// Inherit Context from parent (Phase 2)
		childFiber.Context = fcontext.NewContext(parentFiber.Context)

		// Process Provider components: inject Context values (Phase 2)
		if childFiber.Tag == "provider" && childFiber.Context != nil {
			props := childFiber.Props
			if props != nil {
				if key, ok := props["contextKey"].(fcontext.ContextKey); ok {
					if value, ok := props["contextValue"]; ok {
						childFiber.Context.Provide(key, value)
					}
				}
			}
		}

		// Link siblings
		if i == 0 {
			parentFiber.Child = childFiber
		} else {
			previousChild.Sibling = childFiber
		}

		previousChild = childFiber

		// Recursively build for this child's children
		buildFiberTree(childFiber, childVNode)
	}
}

// =============================================================================
// Fiber Tree Traversal
// =============================================================================

// WalkFiberDepthFirst walks fiber tree in depth-first order
// Uses iterative approach to avoid stack overflow on very deep trees
func WalkFiberDepthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	if root == nil {
		return true
	}

	// Use explicit stack for iterative traversal
	// This avoids stack overflow for very deep trees (e.g., deeply nested lists)
	type frame struct {
		fiber    *Fiber
		state    int  // 0 = visit self, 1 = visit children, 2 = visit siblings, 3 = done
		children bool // whether children were visited
		siblings bool // whether siblings were visited
	}

	stack := make([]frame, 0, 32)
	stack = append(stack, frame{fiber: root, state: 0})

	for len(stack) > 0 {
		// Peek at top of stack
		top := &stack[len(stack)-1]

		switch top.state {
		case 0: // Visit current node
			if !callback(top.fiber) {
				return false // Stop traversal
			}
			top.state = 1

		case 1: // Visit children
			if !top.children && top.fiber.Child != nil {
				top.children = true
				// Push child onto stack
				stack = append(stack, frame{fiber: top.fiber.Child, state: 0})
			} else {
				top.state = 2
			}

		case 2: // Visit siblings
			if !top.siblings && top.fiber.Sibling != nil {
				top.siblings = true
				// Push sibling onto stack
				stack = append(stack, frame{fiber: top.fiber.Sibling, state: 0})
			}
			top.state = 3

		case 3: // Done with this frame
			stack = stack[:len(stack)-1]
		}
	}

	return true
}

// WalkFiberBreadthFirst walks fiber tree in breadth-first order
// Optimized to avoid slice allocation on each dequeue operation
func WalkFiberBreadthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	if root == nil {
		return true
	}

	// Pre-allocate queue with reasonable capacity to reduce allocations
	queue := make([]*Fiber, 0, 32)
	queue = append(queue, root)

	for i := 0; i < len(queue); i++ {
		// Dequeue by index - avoids slice allocation from queue[1:]
		current := queue[i]

		// Visit current node
		if !callback(current) {
			return false
		}

		// Enqueue children
		for child := current.Child; child != nil; child = child.Sibling {
			queue = append(queue, child)
		}
	}

	return true
}

// =============================================================================
// Fiber Utilities
// =============================================================================

// CloneFiber creates a shallow copy of a fiber
// ✨ Preserves NodeID for stable runtime identity
// ✨ Reuses Instance (Instance is NEVER cloned - it persists across renders)
func CloneFiber(fiber *Fiber) *Fiber {
	if fiber == nil {
		return nil
	}

	clone := &Fiber{
		Type:          fiber.Type,
		Tag:           fiber.Tag,
		DiffKey:       fiber.DiffKey,      // ✨ Preserve DiffKey for diffing
		Key:           fiber.Key,          // Backward compatibility
		ID:            fiber.ID,           // Preserve business identifier for anchors/positioning
		IsRoot:        fiber.IsRoot,       // ✨ Preserve IsRoot marker
		NodeID:        fiber.NodeID,       // ✨ Preserve NodeID for stable identity
		Layer:         fiber.Layer,        // ✨ Preserve Layer
		Path:          fiber.Path,         // ✨ Preserve Path for key generation
		PathSegment:   fiber.PathSegment,  // ✨ Preserve PathSegment
		SiblingIndex:  fiber.SiblingIndex, // ✨ Preserve SiblingIndex
		Style:         fiber.Style,        // ✨ Preserve Style (Fiber-first)
		Props:         fiber.Props,
		MemoizedProps: fiber.MemoizedProps,
		MemoizedState: fiber.MemoizedState,
		Return:        fiber.Return,
		Child:         fiber.Child,   // Shallow copy of tree structure
		Sibling:       fiber.Sibling, // Shallow copy of sibling link
		Alternate:     fiber.Alternate,
		// Don't share UpdateQueue - cloned fiber gets its own empty queue
		UpdateQueue:  nil,
		Flags:        fiber.Flags,
		SubtreeFlags: fiber.SubtreeFlags,
		Lanes:        fiber.Lanes,
		ChildLanes:   fiber.ChildLanes,
		// Layout fields
		LayoutDirection:  fiber.LayoutDirection,
		LayoutAlign:      fiber.LayoutAlign,
		LayoutCrossAlign: fiber.LayoutCrossAlign,
		LayoutGap:        fiber.LayoutGap,
		LayoutPadding:    fiber.LayoutPadding,
		LayoutFlex:       fiber.LayoutFlex,
		// ✨ Border Style (Phase 1.3)
		BorderStyle: fiber.BorderStyle,
		BorderLabel: fiber.BorderLabel,
		// ✨ Modal Centering (Phase 1.4)
		Centered: fiber.Centered,
		// ✨ Position Fixed (Phase 2.1)
		Position: fiber.Position,
		// ✨ Anchor (Phase 2.1)
		Anchor: fiber.Anchor,
		// ✨ Portal Root (Phase 3.1) - Copy reference to target fiber
		PortalRoot: fiber.PortalRoot,
		// Special VNode types support
		ComponentFunc:              fiber.ComponentFunc,
		ComponentFuncWithProps:     fiber.ComponentFuncWithProps,
		ComponentName:              fiber.ComponentName,
		ErrorBoundaryFunc:          fiber.ErrorBoundaryFunc,
		ErrorBoundaryFallbackFiber: fiber.ErrorBoundaryFallbackFiber,
		ErrorBoundary:              fiber.ErrorBoundary, // Preserve reference to original ErrorBoundaryVNode for state sync
		MemoCompare:                fiber.MemoCompare,
		MemoShouldUpdate:           fiber.MemoShouldUpdate,
		// Focusable support (DEPRECATED - use Instance instead)
		// FocusableVNode: fiber.FocusableVNode,
		// ActionTargetID (Fiber-first Action Architecture)
		ActionTargetID: fiber.ActionTargetID,
		// Focusable metadata (Fiber-first)
		// FocusableMeta: fiber.FocusableMeta,
		// ComponentInstance (Fiber-first) - REUSE, NEVER CLONE
		// Instance persists across renders
		Instance: fiber.Instance,
	}

	return clone
}

// =============================================================================
// Fiber Tree Utilities
// =============================================================================

// FindFiberByKey searches for a fiber with a given key in subtree
func FindFiberByKey(root *Fiber, key string) *Fiber {
	var result *Fiber
	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		if fiber.Key == key {
			result = fiber
			return false // Stop traversal
		}
		return true
	})
	return result
}

// FindFiberByID searches for a fiber with a given NodeID in subtree
// This is used for HitMap enrichment to set TargetFiber references
func FindFiberByID(root *Fiber, nodeID uint64) *Fiber {
	var result *Fiber
	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		if fiber.NodeID == nodeID {
			result = fiber
			return false // Stop traversal
		}
		return true
	})
	return result
}

// CountFibers counts all fibers in tree
func CountFibers(root *Fiber) int {
	count := 0
	WalkFiberDepthFirst(root, func(_ *Fiber) bool {
		count++
		return true
	})
	return count
}

// GetFiberDepth returns depth of a fiber in tree
func GetFiberDepth(fiber *Fiber) int {
	depth := 0
	for p := fiber.Return; p != nil; p = p.Return {
		depth++
	}
	return depth
}

// CollectFibersWithFlags collects all fibers with specific flags
func CollectFibersWithFlags(root *Fiber, flags EffectFlag) []*Fiber {
	var result []*Fiber
	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		if (fiber.Flags & flags) != 0 {
			result = append(result, fiber)
		}
		return true
	})
	return result
}

// =============================================================================
// Fiber Layout Helper Methods (Phase 1)
// =============================================================================
// These methods enable Fiber-first layout by providing access to child fibers
// and layout properties directly from the Fiber struct.

// GetChildFibers returns all child fibers as a slice
// Converts the Child → Sibling linked list to an array for easier iteration
func (f *Fiber) GetChildFibers() []*Fiber {
	var children []*Fiber
	for child := f.Child; child != nil; child = child.Sibling {
		children = append(children, child)
	}
	return children
}

// GetChildCount returns the number of children
func (f *Fiber) GetChildCount() int {
	count := 0
	for child := f.Child; child != nil; child = child.Sibling {
		count++
	}
	return count
}

// GetDirection returns the layout direction
// Returns Fiber.LayoutDirection field (Fiber-first, no VNode fallback)
func (f *Fiber) GetDirection() Direction {
	if f.LayoutDirection != 0 {
		return f.LayoutDirection
	}
	return DirectionRow // default
}

// GetAlign returns the main axis alignment
// Returns Fiber.LayoutAlign field (Fiber-first, no VNode fallback)
func (f *Fiber) GetAlign() Align {
	if f.LayoutAlign != 0 {
		return f.LayoutAlign
	}
	return AlignStart // default
}

// GetCrossAlign returns the cross axis alignment
// Returns Fiber.LayoutCrossAlign field (Fiber-first, no VNode fallback)
func (f *Fiber) GetCrossAlign() Align {
	if f.LayoutCrossAlign != 0 {
		return f.LayoutCrossAlign
	}
	return AlignStart // default
}

// GetGap returns the gap spacing between children
// Returns Fiber.LayoutGap field (Fiber-first, no VNode fallback)
func (f *Fiber) GetGap() int {
	return f.LayoutGap
}

// GetPadding returns the padding [top, right, bottom, left]
// Returns Fiber.LayoutPadding field (Fiber-first, no VNode fallback)
func (f *Fiber) GetPadding() [4]int {
	return f.LayoutPadding
}

// GetMargin returns the margin as layout.Margin type
// Returns Fiber.LayoutMargin field converted to layout.Margin (Fiber-first)
func (f *Fiber) GetMargin() layout.Margin {
	return layout.Margin{
		Top:    f.LayoutMargin[0],
		Right:  f.LayoutMargin[1],
		Bottom: f.LayoutMargin[2],
		Left:   f.LayoutMargin[3],
	}
}

// GetFlex returns the flex factor
// Returns Fiber.LayoutFlex field (Fiber-first, no VNode fallback)
func (f *Fiber) GetFlex() int {
	return f.LayoutFlex
}
