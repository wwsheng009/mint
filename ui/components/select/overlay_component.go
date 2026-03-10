package selectcomp

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const overlayCallbacksProp = "_selectOverlayCallbacks"

type overlayComponentModel struct {
	key             string
	selectID        string
	componentID     string
	options         []Option
	selectStyle     style.Style
	width           int
	placeholder     string
	maxVisibleRows  int
	portalRoot      string
	closeOnOutside  bool
	changeIntent    intent.Intent
	selectionMode   SelectionMode
	disabled        bool
	formID          string
	selectedIndex   int
	selectedIndices []int
}

type overlayComponentState struct {
	selectedIndex     int
	selectedIndices   []int
	open              bool
	highlightedIndex  int
	scrollOffset      int
	externalSelection string
}

type overlayControllerState struct {
	selectedIndex    int
	selectedIndices  []int
	open             bool
	highlightedIndex int
	scrollOffset     int
}

type overlayCallbacks struct {
	setOpen      func(bool) overlayControllerState
	setHighlight func(int) overlayControllerState
	commit       func(int) overlayControllerState
}

func newOverlayComponent(node *VNode) rtui.VNode {
	if node == nil {
		return nil
	}

	props := node.Props().Clone()
	props["key"] = node.key
	props["selectID"] = node.ownerID()

	component := rtui.NewComponentWithProps("SelectOverlay", renderOverlayComponent)
	component.SetKey(firstNonEmpty(node.key, node.ownerID()))
	if id := node.ID(); id != "" {
		component.SetID(id)
	}
	component.SetProps(props)
	return component
}

func renderOverlayComponent(props rtui.Props) rtui.VNode {
	model := overlayComponentModelFromProps(props)
	stateRef, ctx := currentOverlayComponentState(model)
	callbacks := newOverlayCallbacks(ctx, stateRef, model)

	children := []rtui.VNode{
		newOverlayTriggerVNode(model, *stateRef, callbacks),
	}
	if stateRef.open && len(model.options) > 0 {
		children = append(children, newOverlayPopupVNode(model, *stateRef, callbacks))
	}
	return rtui.Fragment(children...)
}

func overlayComponentModelFromProps(props rtui.Props) overlayComponentModel {
	model := overlayComponentModel{
		key:             getStringProp(props, "key", ""),
		selectID:        getStringProp(props, "selectID", ""),
		componentID:     getStringProp(props, "componentID", ""),
		options:         getOptionsProp(props),
		selectStyle:     getStyleProp(props),
		width:           getIntProp(props, "width", 0),
		placeholder:     getStringProp(props, "placeholder", "..."),
		maxVisibleRows:  getIntProp(props, "maxVisibleRows", defaultMaxVisibleRows),
		portalRoot:      getPortalRootProp(props, rtui.DefaultOverlayPortalRootID),
		closeOnOutside:  getBoolProp(props, "closeOnOutside", true),
		changeIntent:    getIntentProp(props, "changeIntent"),
		selectionMode:   getSelectionModeProp(props, SelectionSingle),
		disabled:        getBoolProp(props, "disabled", false),
		formID:          getStringProp(props, "formID", ""),
		selectedIndex:   getIntProp(props, "selectedIndex", -1),
		selectedIndices: getIntsProp(props, "selectedIndices", nil),
	}
	if model.selectID == "" {
		model.selectID = firstNonEmpty(model.componentID, model.key)
	}
	model.selectedIndex, model.selectedIndices = normalizeOverlaySelection(
		model.selectionMode,
		model.selectedIndex,
		model.selectedIndices,
		len(model.options),
	)
	if model.maxVisibleRows <= 0 {
		model.maxVisibleRows = defaultMaxVisibleRows
	}
	return model
}

func currentOverlayComponentState(model overlayComponentModel) (*overlayComponentState, *rtui.ComponentContext) {
	ctx := rtui.GetCurrentContext()
	initial := initialOverlayComponentState(model)
	if ctx == nil {
		return &initial, nil
	}

	if err := ctx.Validator.ValidateHookCall(rtui.HookRef); err != nil {
		panic(err)
	}
	hook := ctx.GetOrCreateHook(rtui.HookRef)
	if hook.Value == nil {
		state := initial
		hook.Value = &state
	}
	stateRef, ok := hook.Value.(*overlayComponentState)
	if !ok || stateRef == nil {
		state := initial
		stateRef = &state
		hook.Value = stateRef
	}
	syncOverlayComponentState(stateRef, model)
	return stateRef, ctx
}

func initialOverlayComponentState(model overlayComponentModel) overlayComponentState {
	state := overlayComponentState{
		selectedIndex:   model.selectedIndex,
		selectedIndices: append([]int(nil), model.selectedIndices...),
		open:            false,
		scrollOffset:    0,
	}
	state.externalSelection = overlaySelectionSignature(model.selectionMode, model.selectedIndex, model.selectedIndices)
	state.highlightedIndex = defaultOverlayHighlight(state.selectedIndex, state.selectedIndices, model.selectionMode, len(model.options))
	normalizeOverlayComponentState(&state, model)
	return state
}

func syncOverlayComponentState(state *overlayComponentState, model overlayComponentModel) {
	if state == nil {
		return
	}

	externalSelection := overlaySelectionSignature(model.selectionMode, model.selectedIndex, model.selectedIndices)
	if state.externalSelection != externalSelection {
		state.selectedIndex = model.selectedIndex
		state.selectedIndices = append([]int(nil), model.selectedIndices...)
		state.externalSelection = externalSelection
		if !state.open {
			state.highlightedIndex = defaultOverlayHighlight(state.selectedIndex, state.selectedIndices, model.selectionMode, len(model.options))
		}
	}

	normalizeOverlayComponentState(state, model)
}

func normalizeOverlayComponentState(state *overlayComponentState, model overlayComponentModel) {
	if state == nil {
		return
	}

	state.selectedIndex, state.selectedIndices = normalizeOverlaySelection(
		model.selectionMode,
		state.selectedIndex,
		state.selectedIndices,
		len(model.options),
	)

	if len(model.options) == 0 {
		state.open = false
		state.highlightedIndex = -1
		state.scrollOffset = 0
		return
	}

	if state.highlightedIndex < 0 || state.highlightedIndex >= len(model.options) {
		state.highlightedIndex = defaultOverlayHighlight(state.selectedIndex, state.selectedIndices, model.selectionMode, len(model.options))
	}
	state.highlightedIndex = clampInt(state.highlightedIndex, 0, len(model.options)-1)

	visibleRows := minInt(len(model.options), maxInt(1, model.maxVisibleRows))
	maxOffset := maxInt(0, len(model.options)-visibleRows)
	state.scrollOffset = clampInt(state.scrollOffset, 0, maxOffset)
	if state.highlightedIndex < state.scrollOffset {
		state.scrollOffset = state.highlightedIndex
	}
	if state.highlightedIndex >= state.scrollOffset+visibleRows {
		state.scrollOffset = state.highlightedIndex - visibleRows + 1
	}
	state.scrollOffset = clampInt(state.scrollOffset, 0, maxOffset)
}

func overlayControllerSnapshot(state *overlayComponentState) overlayControllerState {
	if state == nil {
		return overlayControllerState{}
	}
	return overlayControllerState{
		selectedIndex:    state.selectedIndex,
		selectedIndices:  append([]int(nil), state.selectedIndices...),
		open:             state.open,
		highlightedIndex: state.highlightedIndex,
		scrollOffset:     state.scrollOffset,
	}
}

func updateOverlayComponentState(
	ctx *rtui.ComponentContext,
	state *overlayComponentState,
	model overlayComponentModel,
	update func(*overlayComponentState),
) overlayControllerState {
	if state == nil {
		initial := initialOverlayComponentState(model)
		state = &initial
	}
	if update != nil {
		update(state)
	}
	syncOverlayComponentState(state, model)
	if ctx != nil {
		ctx.ScheduleUpdate()
	}
	return overlayControllerSnapshot(state)
}

func newOverlayCallbacks(ctx *rtui.ComponentContext, state *overlayComponentState, model overlayComponentModel) *overlayCallbacks {
	return &overlayCallbacks{
		setOpen: func(open bool) overlayControllerState {
			return updateOverlayComponentState(ctx, state, model, func(state *overlayComponentState) {
				if len(model.options) == 0 {
					state.open = false
					return
				}
				state.open = open
				if open && (state.highlightedIndex < 0 || state.highlightedIndex >= len(model.options)) {
					state.highlightedIndex = defaultOverlayHighlight(
						state.selectedIndex,
						state.selectedIndices,
						model.selectionMode,
						len(model.options),
					)
				}
			})
		},
		setHighlight: func(index int) overlayControllerState {
			return updateOverlayComponentState(ctx, state, model, func(state *overlayComponentState) {
				if len(model.options) == 0 {
					state.open = false
					state.highlightedIndex = -1
					state.scrollOffset = 0
					return
				}
				state.open = true
				state.highlightedIndex = clampInt(index, 0, len(model.options)-1)
			})
		},
		commit: func(index int) overlayControllerState {
			return updateOverlayComponentState(ctx, state, model, func(state *overlayComponentState) {
				nextIndex, nextIndices, _, shouldClose := applyOverlayCommit(
					model.selectionMode,
					len(model.options),
					state.selectedIndex,
					state.selectedIndices,
					index,
				)
				state.selectedIndex = nextIndex
				state.selectedIndices = nextIndices
				state.highlightedIndex = clampIndexForOptions(index, len(model.options))
				if shouldClose {
					state.open = false
				} else if len(model.options) > 0 {
					state.open = true
				}
			})
		},
	}
}

func newOverlayTriggerVNode(
	model overlayComponentModel,
	state overlayComponentState,
	callbacks *overlayCallbacks,
) rtui.VNode {
	trigger := New()
	trigger.key = firstNonEmpty(model.key, model.selectID)
	trigger.componentID = model.componentID
	trigger.options = append([]Option(nil), model.options...)
	trigger.style = model.selectStyle
	trigger.width = model.width
	trigger.placeholder = model.placeholder
	trigger.maxVisibleRows = model.maxVisibleRows
	trigger.overlayPopup = true
	trigger.portalRoot = model.portalRoot
	trigger.closeOnOutside = model.closeOnOutside
	trigger.changeIntent = model.changeIntent
	trigger.selectedIndex = state.selectedIndex
	trigger.selectedIndices = append([]int(nil), state.selectedIndices...)
	trigger.selectionMode = model.selectionMode
	trigger.disabled = model.disabled
	trigger.formID = model.formID
	trigger.open = state.open
	trigger.highlightedIndex = state.highlightedIndex
	trigger.scrollOffset = state.scrollOffset
	trigger.selectID = model.selectID
	trigger.overlayCallbacks = callbacks
	trigger.SetID(model.selectID)
	return trigger
}

func newOverlayPopupVNode(
	model overlayComponentModel,
	state overlayComponentState,
	callbacks *overlayCallbacks,
) rtui.VNode {
	surface := &popupVNode{ElementVNode: rtui.NewElement("select-popup")}
	surface.SetKey(firstNonEmpty(model.key, model.selectID) + "-popup")
	surface.SetID(model.selectID + "-popup")
	surface.SetLayer(rtui.LayerOverlay)
	surface.SetProps(rtui.Props{
		"selectID":           model.selectID,
		"componentID":        resolvedSelectComponentID(model.componentID, model.selectID),
		"options":            append([]Option(nil), model.options...),
		"style":              model.selectStyle,
		"selectionMode":      model.selectionMode,
		"selectedIndex":      state.selectedIndex,
		"selectedIndices":    append([]int(nil), state.selectedIndices...),
		"highlightedIndex":   state.highlightedIndex,
		"scrollOffset":       state.scrollOffset,
		"maxVisibleRows":     model.maxVisibleRows,
		"minWidth":           triggerWidthForModel(model, state.selectedIndex, state.selectedIndices),
		"disabled":           model.disabled,
		"closeOnOutside":     model.closeOnOutside,
		"changeIntent":       model.changeIntent,
		"formID":             model.formID,
		overlayCallbacksProp: callbacks,
	})

	portal := rtui.NewElement("box")
	portal.SetKey(firstNonEmpty(model.key, model.selectID) + "-popup-portal")
	portal.SetID(model.selectID + "-popup-portal")
	portal.SetLayer(rtui.LayerOverlay)
	portal.SetProps(rtui.Props{
		"position": "absolute",
		"left":     0,
		"top":      0,
		"width":    1,
		"height":   1,
	})
	portal.SetPortalRoot(model.portalRoot)
	portal.SetAnchorTo(model.selectID, rttypes.AnchorBottomLeft)
	portal.SetPortalPosition(rttypes.PositionAbsolute)
	portal.SetChildren([]rtui.VNode{surface})
	return portal
}

func normalizeOverlaySelection(
	mode SelectionMode,
	selectedIndex int,
	selectedIndices []int,
	optionCount int,
) (int, []int) {
	normalized := normalizeIndices(selectedIndices, optionCount)
	if mode == SelectionMultiple {
		if len(normalized) == 0 && selectedIndex >= 0 && selectedIndex < optionCount {
			normalized = []int{selectedIndex}
		}
		if len(normalized) == 0 {
			return -1, nil
		}
		if containsInt(normalized, selectedIndex) {
			return selectedIndex, normalized
		}
		return normalized[len(normalized)-1], normalized
	}

	if optionCount == 0 {
		return -1, nil
	}
	if selectedIndex >= optionCount {
		selectedIndex = optionCount - 1
	}
	if selectedIndex < -1 {
		selectedIndex = -1
	}
	if selectedIndex >= 0 {
		return selectedIndex, []int{selectedIndex}
	}
	return -1, nil
}

func defaultOverlayHighlight(selectedIndex int, selectedIndices []int, mode SelectionMode, optionCount int) int {
	if optionCount <= 0 {
		return -1
	}
	if mode == SelectionMultiple && len(selectedIndices) > 0 {
		return clampInt(selectedIndices[len(selectedIndices)-1], 0, optionCount-1)
	}
	if selectedIndex >= 0 {
		return clampInt(selectedIndex, 0, optionCount-1)
	}
	return 0
}

func overlaySelectionSignature(mode SelectionMode, selectedIndex int, selectedIndices []int) string {
	return fmt.Sprintf("%d:%d:%s", mode, selectedIndex, joinIndices(selectedIndices))
}

func applyOverlayCommit(
	mode SelectionMode,
	optionCount int,
	selectedIndex int,
	selectedIndices []int,
	index int,
) (int, []int, bool, bool) {
	if index < 0 || index >= optionCount {
		return selectedIndex, append([]int(nil), selectedIndices...), false, false
	}
	if mode == SelectionMultiple {
		next := append([]int(nil), selectedIndices...)
		pos := indexOfInt(next, index)
		if pos >= 0 {
			next = append(next[:pos], next[pos+1:]...)
		} else {
			next = append(next, index)
		}
		next = normalizeIndices(next, optionCount)
		nextIndex := selectedIndex
		if pos >= 0 {
			if selectedIndex == index {
				if len(next) > 0 {
					nextIndex = next[len(next)-1]
				} else {
					nextIndex = -1
				}
			}
		} else {
			nextIndex = index
		}
		changed := nextIndex != selectedIndex || !equalIntSlices(next, selectedIndices)
		return nextIndex, next, changed, false
	}

	next := []int{index}
	changed := selectedIndex != index || len(selectedIndices) != 1 || selectedIndices[0] != index
	return index, next, changed, true
}

func clampIndexForOptions(index, optionCount int) int {
	if optionCount <= 0 {
		return -1
	}
	return clampInt(index, 0, optionCount-1)
}

func resolvedSelectComponentID(componentID, selectID string) string {
	return firstNonEmpty(componentID, selectID)
}

func triggerWidthForModel(model overlayComponentModel, selectedIndex int, selectedIndices []int) int {
	inst := &Instance{
		options:         append([]Option(nil), model.options...),
		selectStyle:     model.selectStyle,
		width:           model.width,
		placeholder:     model.placeholder,
		selectionMode:   model.selectionMode,
		selectedIndex:   selectedIndex,
		selectedIndices: append([]int(nil), selectedIndices...),
	}
	inst.normalizeSelectionState()
	return inst.triggerWidth()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
