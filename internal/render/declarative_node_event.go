package render

import (
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/event"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// framework.Event Component interface implementation
// =============================================================================

// HandleEvent processes events by distributing them to child components
func (n *DeclarativeNode) HandleEvent(ev frameworkevent.Event) bool {
	log.RenderLogger.Debug("DeclarativeNode.HandleEvent: event type=%d", ev.Type())

	n.mu.RLock()
	root := n.root
	focusMgr := n.focusMgr
	useFiber := n.useFiber
	reconciler := n.reconciler
	n.mu.RUnlock()

	if root == nil {
		return false
	}

	// 0. Handle layer-specific events (ESC to close modal, etc.)
	// This takes priority over all other event handling
	if ev.Type() == frameworkevent.EventKeyPress {
		if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
			if keyEv.Key.Name == "escape" || keyEv.Key.Name == "esc" {
				if n.handleLayerKeyEvent(root) {
					// Modal was closed, trigger re-render
					n.requestRender(useFiber, reconciler)
					return true
				}
			}
		}
	}

	// 1. Let focus manager handle navigation (Tab, Shift+Tab)
	if focusMgr != nil {
		handled, shouldRender := focusMgr.HandleEvent(ev)
		if handled {
			log.RenderLogger.Debug("DeclarativeNode.HandleEvent: focus manager handled event, shouldRender=%v", shouldRender)

			// Request a re-render when focus changes
			// In Fiber mode, use the reconciler; in non-Fiber mode, mark as dirty
			if shouldRender {
				if useFiber && reconciler != nil {
					// Fiber mode: schedule reconciler update
					if r, ok := reconciler.(*fiberReconcilerAdapter); ok {
						r.r.ScheduleUpdate(rtui.LaneSyncLane)
					}
				} else {
					// Non-Fiber mode: mark framework app as dirty to trigger re-render
					// This ensures the updated focus state is painted
					if fwApp := n.getFrameworkApp(); fwApp != nil {
						fwApp.MarkDirty()
					}
				}
			}
			return true
		}

		// 2. Keyboard events are handled by App.handleMsg via ActionBridge
		// Focus navigation (Tab/Shift+Tab) is handled above
	}

	// 1.5. Handle mouse clicks - switch focus before dispatching event
	// This ensures that clicking a button focuses it before triggering its action
	if ev.Type().IsMouse() {
		if mouseEv, ok := ev.(*frameworkevent.MouseEvent); ok {
			// Handle mouse press and click events
			if ev.Type() == frameworkevent.EventMousePress || ev.Type() == frameworkevent.EventClick {
				if n.handleMouseFocus(mouseEv) {
					log.RenderLogger.Debug("DeclarativeNode.HandleEvent: mouse click switched focus")

					// Focus was switched, trigger re-render
					if useFiber && reconciler != nil {
						if r, ok := reconciler.(*fiberReconcilerAdapter); ok {
							r.r.ScheduleUpdate(rtui.LaneSyncLane)
						}
					} else {
						if fwApp := n.getFrameworkApp(); fwApp != nil {
							fwApp.MarkDirty()
						}
					}
					// Continue to dispatch the event to the newly focused element
				}
			}
		}
	}

	// 3. Fall back to global event distribution
	// Try to distribute to root tree
	handled := n.distributeEventToVNode(root, ev)
	if handled {
		// Event was handled by a component (e.g., button click)
		// Trigger re-render to update the UI with new state
		n.requestRender(useFiber, reconciler)
	}
	return handled
}

// distributeEventToVNode recursively distributes an event to VNode tree
func (n *DeclarativeNode) distributeEventToVNode(vnode rtui.VNode, ev frameworkevent.Event) bool {
	if vnode == nil {
		log.RenderLogger.Debug("distributeEventToVNode: vnode is nil")
		return false
	}

	log.RenderLogger.Debug("distributeEventToVNode: called with vnode type=%d, actual type=%T", vnode.Type(), vnode)

	// Phase 3: Event-centric distribution
	// If this is a MouseEvent with TargetID, only distribute to the target component
	if mouseEv, ok := ev.(*frameworkevent.MouseEvent); ok && mouseEv.TargetID != 0 {
		targetID := mouseEv.TargetID
		// Convert VNode key to uint64 for comparison
		nodeID := uint64(0)
		if key := vnode.Key(); key != "" {
			// Use the same hash conversion as HitMap building
			nodeID = event.StringToNodeID(key)
		}

		// Check if this node is the target
		if nodeID == targetID {
			// This is the target component, call HandleEvent
			if component, ok := vnode.(frameworkevent.Component); ok {
				// Debug: check if this is a button and print its label and pointer
				if button, ok := vnode.(interface{ Label() string }); ok {
					log.RenderLogger.Debug("distributeEventToVNode: Found target button key=%d, label='%s', pointer=%p, calling HandleEvent", targetID, button.Label(), vnode)
				} else {
					log.RenderLogger.Debug("distributeEventToVNode: Found target component with key=%d (not a button), calling HandleEvent", targetID)
				}
				if component.HandleEvent(ev) {
					return true
				}
			}
		}

		// Not the target, but children might be - recursively check children
		children := vnode.Children()
		for _, child := range children {
			if n.distributeEventToVNode(child, ev) {
				return true
			}
		}

		// This branch doesn't contain the target
		return false
	}

	// Legacy behavior: broadcast to all components (for KeyEvent or MouseEvent without TargetID)
	// Check if this VNode implements the Component interface
	if component, ok := vnode.(frameworkevent.Component); ok {
		log.RenderLogger.Debug("distributeEventToVNode: VNode type=%d implements frameworkevent.Component, calling HandleEvent", vnode.Type())

		if component.HandleEvent(ev) {
			// Event was handled by this component - stop propagation
			// This prevents keyboard events from triggering multiple buttons
			return true
		}
	}

	// Try to distribute to children
	children := vnode.Children()
	if len(children) > 0 {
		log.RenderLogger.Debug("distributeEventToVNode: VNode type=%d has %d children, distributing...", vnode.Type(), len(children))

		for _, child := range children {
			if n.distributeEventToVNode(child, ev) {
				// Event was handled by a child - stop propagation
				return true
			}
		}
	}

	return false
}

// handleMouseFocus handles mouse clicks by switching focus to the clicked focusable node.
// Returns true if focus was switched, false otherwise.
func (n *DeclarativeNode) handleMouseFocus(mouseEv *frameworkevent.MouseEvent) bool {
	if n.focusMgr == nil {
		return false
	}

	// Collect all focusable nodes from the tree
	var focusable []rtui.FocusableVNode
	hasModal := rtui.HasModalInTree(n.root)

	if hasModal {
		// Modal is open: only consider focusable nodes in modal layer
		focusable = rtui.CollectFocusableInLayer(n.root, rtui.LayerModal)
	} else {
		// No modal: consider all focusable nodes
		focusable = rtui.CollectFocusable(n.root)
	}

	if len(focusable) == 0 {
		return false
	}

	// Find the focusable node that was clicked
	for i, node := range focusable {
		if n.nodeWasClicked(node, mouseEv.X, mouseEv.Y) {
			// Found the clicked focusable node
			currentIndex := n.focusMgr.CurrentIndex()
			if i == currentIndex {
				// Already focused, no change
				return false
			}

			log.RenderLogger.Debug("handleMouseFocus: switching focus from index %d to %d",
				currentIndex, i)

			// Switch focus to the clicked node
			n.focusMgr.SetFocusByIndex(i)
			return true
		}
	}

	return false
}

// nodeWasClicked checks if a VNode was clicked based on mouse coordinates.
// This performs hit testing using the node's bounds if available.
func (n *DeclarativeNode) nodeWasClicked(node rtui.VNode, x, y int) bool {
	// Check if node has bounds information (from SetBounds during Paint)
	if boundsAware, ok := node.(interface {
		GetBounds() (x, y, width, height int)
	}); ok {
		bx, by, bw, bh := boundsAware.GetBounds()
		log.RenderLogger.Debug("node Was Clicked: bounds=(%d,%d,%d,%d), mouse=(%d,%d)",
			bx, by, bw, bh, x, y)

		// Check if mouse click is within bounds
		if x >= bx && x < bx+bw && y >= by && y < by+bh {
			log.RenderLogger.Debug("node Was Clicked: HIT!")

			return true
		}

		log.RenderLogger.Debug("node Was Clicked: MISS!")

		return false
	}

	// Fallback: check if node implements button-like interface
	if hasContainsPoint, ok := node.(interface{ ContainsPoint(x, y int) bool }); ok {
		return hasContainsPoint.ContainsPoint(x, y)
	}

	return false
}