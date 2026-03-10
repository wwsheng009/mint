package snapshot

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/wwsheng009/mint/framework/component"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/state"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type Input struct {
	Root         component.Node
	FocusManager *rtui.FiberFocusManager
	HitMap       *runtimeevent.HitMap
	RenderSeq    uint64
	RenderedAt   time.Time
}

type Builder struct{}

type NodeLocator struct {
	ComponentID    string
	NodeID         uint64
	Path           string
	ActionTargetID string
	Type           string
	Tag            string
}

type Frame struct {
	Snapshot      *state.Snapshot
	ByComponentID map[string]NodeLocator
	ByNodeID      map[uint64]NodeLocator
	ByPath        map[string]NodeLocator
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Build(in Input) *state.Snapshot {
	frame := b.BuildFrame(in)
	if frame == nil || frame.Snapshot == nil {
		return state.NewSnapshot()
	}
	return frame.Snapshot
}

func (b *Builder) BuildFrame(in Input) *Frame {
	snap := state.NewSnapshot()
	if !in.RenderedAt.IsZero() {
		snap.Timestamp = in.RenderedAt
	}
	snap.Metadata["render_seq"] = in.RenderSeq
	if in.HitMap != nil {
		snap.Metadata["hitmap_size"] = in.HitMap.Size()
	}

	rects := buildRectIndex(in.Root, in.HitMap)
	rootFiber := getFiberRoot(in.Root)
	if rootFiber == nil {
		return &Frame{
			Snapshot:      snap,
			ByComponentID: map[string]NodeLocator{},
			ByNodeID:      map[uint64]NodeLocator{},
			ByPath:        map[string]NodeLocator{},
		}
	}

	frame := &Frame{
		Snapshot:      snap,
		ByComponentID: make(map[string]NodeLocator),
		ByNodeID:      make(map[uint64]NodeLocator),
		ByPath:        make(map[string]NodeLocator),
	}
	focused := currentFocusedFiber(in.FocusManager)
	seenIDs := make(map[string]struct{})
	rtui.WalkFiberDepthFirst(rootFiber, func(fiber *rtui.Fiber) bool {
		if fiber == nil || !includeFiberInSnapshot(fiber) {
			return true
		}

		componentID := resolveComponentID(fiber, seenIDs)
		props := extractProps(fiber)
		dynState := extractState(fiber)
		rect := rects[fiber.NodeID]
		visible := rect.Width > 0 && rect.Height > 0
		disabled := componentDisabled(fiber, props, dynState)

		snap.Components[componentID] = state.ComponentState{
			ID:       componentID,
			Type:     componentType(fiber),
			Props:    props,
			State:    dynState,
			Rect:     rect,
			Visible:  visible,
			Disabled: disabled,
		}

		locator := NodeLocator{
			ComponentID:    componentID,
			NodeID:         fiber.NodeID,
			Path:           fiber.Path,
			ActionTargetID: fiber.ActionTargetID,
			Type:           componentType(fiber),
			Tag:            fiber.Tag,
		}
		frame.ByComponentID[componentID] = locator
		frame.ByNodeID[fiber.NodeID] = locator
		if fiber.Path != "" {
			frame.ByPath[fiber.Path] = locator
		}

		if focused != nil && focused.NodeID == fiber.NodeID {
			snap.FocusPath = state.FocusPath{componentID}
		}
		return true
	})

	return frame
}

func includeFiberInSnapshot(fiber *rtui.Fiber) bool {
	if fiber.Instance != nil {
		return true
	}
	if fiber.ID != "" {
		return true
	}
	return fiber.ActionTargetID != ""
}

func getFiberRoot(root component.Node) *rtui.Fiber {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetFiberRoot() *rtui.Fiber }); ok {
		return provider.GetFiberRoot()
	}
	return nil
}

func getLayoutRoot(root component.Node) *layout.LayoutBox {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetLayoutRoot() *layout.LayoutBox }); ok {
		return provider.GetLayoutRoot()
	}
	return nil
}

func currentFocusedFiber(fm *rtui.FiberFocusManager) *rtui.Fiber {
	if fm == nil {
		return nil
	}
	return fm.GetCurrent()
}

func buildRectIndex(root component.Node, hitMap *runtimeevent.HitMap) map[uint64]state.Rect {
	rects := make(map[uint64]state.Rect)
	if hitMap != nil {
		for _, entry := range hitMap.AllEntries() {
			rects[entry.NodeID] = state.Rect{
				X:      entry.Bounds.X,
				Y:      entry.Bounds.Y,
				Width:  entry.Bounds.Width,
				Height: entry.Bounds.Height,
			}
		}
	}
	if len(rects) > 0 {
		return rects
	}

	layoutRoot := getLayoutRoot(root)
	if layoutRoot == nil {
		return rects
	}

	var walk func(box *layout.LayoutBox)
	walk = func(box *layout.LayoutBox) {
		if box == nil {
			return
		}
		nodeID := runtimeevent.StringToNodeID(box.ID)
		if nodeID != 0 {
			x := box.AbsX
			y := box.AbsY
			rects[nodeID] = state.Rect{
				X:      x,
				Y:      y,
				Width:  box.Width,
				Height: box.Height,
			}
		}
		for _, child := range box.Children {
			walk(child)
		}
	}
	walk(layoutRoot)
	return rects
}

func resolveComponentID(fiber *rtui.Fiber, seen map[string]struct{}) string {
	candidates := []string{
		fiber.ID,
		instanceKey(fiber),
		fiber.ActionTargetID,
		fiber.Path,
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		return candidate
	}

	fallback := fmt.Sprintf("node:%d", fiber.NodeID)
	seen[fallback] = struct{}{}
	return fallback
}

func instanceKey(fiber *rtui.Fiber) string {
	if fiber == nil || fiber.Instance == nil {
		return ""
	}
	return fiber.Instance.Key()
}

func componentType(fiber *rtui.Fiber) string {
	if fiber == nil {
		return "Unknown"
	}
	if fiber.Tag != "" {
		return fiber.Tag
	}
	return fiber.Type.String()
}

func extractProps(fiber *rtui.Fiber) map[string]interface{} {
	if fiber == nil {
		return nil
	}
	merged := make(map[string]interface{})
	for k, v := range sanitizeMap(map[string]interface{}(fiber.Props)) {
		merged[k] = v
	}
	if fiber.Instance != nil {
		for k, v := range sanitizeMap(map[string]interface{}(fiber.Instance.GetProps())) {
			merged[k] = v
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func extractState(fiber *rtui.Fiber) map[string]interface{} {
	if fiber == nil || fiber.Instance == nil {
		return nil
	}

	result := make(map[string]interface{})

	if focusable, ok := fiber.Instance.(interface {
		HasFocus() bool
		IsDisabled() bool
	}); ok {
		result["focused"] = focusable.HasFocus()
		result["disabled"] = focusable.IsDisabled()
	}

	if valueGetter, ok := fiber.Instance.(interface{ GetValue() string }); ok {
		result["value"] = valueGetter.GetValue()
	}

	if valuesGetter, ok := fiber.Instance.(interface{ GetValues() map[string]interface{} }); ok {
		result["values"] = sanitizeMap(valuesGetter.GetValues())
	}
	if selectedValueGetter, ok := fiber.Instance.(interface{ SelectedValue() string }); ok {
		result["selectedValue"] = selectedValueGetter.SelectedValue()
	}
	if selectedValuesGetter, ok := fiber.Instance.(interface{ SelectedValues() []string }); ok {
		result["selectedValues"] = selectedValuesGetter.SelectedValues()
	}
	if selectedIndexGetter, ok := fiber.Instance.(interface{ GetSelectedIndex() int }); ok {
		result["selectedIndex"] = selectedIndexGetter.GetSelectedIndex()
	}
	if checkedGetter, ok := fiber.Instance.(interface{ GetCheckedIndices() []int }); ok {
		result["checkedIndices"] = checkedGetter.GetCheckedIndices()
	}
	if rowGetter, ok := fiber.Instance.(interface {
		GetSelectedRow() (string, bool)
	}); ok {
		if row, ok := rowGetter.GetSelectedRow(); ok {
			result["selectedRow"] = row
		}
	}
	if rowGetter, ok := fiber.Instance.(interface {
		GetSelectedRow() ([]string, bool)
	}); ok {
		if row, ok := rowGetter.GetSelectedRow(); ok {
			result["selectedRow"] = row
		}
	}
	if node, ok := callSelectedNode(fiber.Instance); ok {
		result["selectedNode"] = sanitizeValue(node, 0)
	}

	if propGetter, ok := fiber.Instance.(interface {
		GetProp(key string) (interface{}, bool)
	}); ok {
		for _, key := range []string{"value", "checked", "selected", "selecteds", "selectedIndex", "disabled", "placeholder"} {
			if _, exists := result[key]; exists {
				continue
			}
			if value, ok := propGetter.GetProp(key); ok {
				result[key] = sanitizeValue(value, 0)
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func componentDisabled(fiber *rtui.Fiber, props map[string]interface{}, dynState map[string]interface{}) bool {
	if fiber != nil {
		if focusable := fiber.GetFocusableInstance(); focusable != nil {
			return focusable.IsDisabled()
		}
	}
	if dynState != nil {
		if v, ok := dynState["disabled"].(bool); ok {
			return v
		}
	}
	if props != nil {
		if v, ok := props["disabled"].(bool); ok {
			return v
		}
	}
	return false
}

func sanitizeMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = sanitizeValue(v, 0)
	}
	return out
}

func sanitizeValue(v interface{}, depth int) interface{} {
	if v == nil {
		return nil
	}
	if depth >= 4 {
		return "<max-depth>"
	}

	switch val := v.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return val
	case fmt.Stringer:
		return val.String()
	case time.Time:
		return val.Format(time.RFC3339Nano)
	case []string:
		out := make([]string, len(val))
		copy(out, val)
		return out
	case []int:
		out := make([]int, len(val))
		copy(out, val)
		return out
	case map[string]interface{}:
		return sanitizeMap(val)
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Func:
		return "<func>"
	case reflect.Pointer:
		if rv.IsNil() {
			return nil
		}
		return sanitizeValue(rv.Elem().Interface(), depth+1)
	case reflect.Slice, reflect.Array:
		n := rv.Len()
		if n > 16 {
			n = 16
		}
		out := make([]interface{}, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, sanitizeValue(rv.Index(i).Interface(), depth+1))
		}
		return out
	case reflect.Map:
		out := make(map[string]interface{})
		iter := rv.MapRange()
		count := 0
		for iter.Next() {
			if count >= 32 {
				out["<truncated>"] = true
				break
			}
			key := stringifyMapKey(iter.Key())
			out[key] = sanitizeValue(iter.Value().Interface(), depth+1)
			count++
		}
		return out
	case reflect.Struct:
		out := make(map[string]interface{})
		rt := rv.Type()
		exported := 0
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.PkgPath != "" {
				continue
			}
			exported++
			if exported > 16 {
				out["<truncated>"] = true
				break
			}
			out[field.Name] = sanitizeValue(rv.Field(i).Interface(), depth+1)
		}
		if len(out) == 0 {
			return fmt.Sprintf("<struct:%s>", rv.Type().String())
		}
		return out
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stringifyMapKey(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	if v.CanInt() {
		return strconv.FormatInt(v.Int(), 10)
	}
	if v.CanUint() {
		return strconv.FormatUint(v.Uint(), 10)
	}
	return fmt.Sprintf("%v", v.Interface())
}

func callSelectedNode(inst interface{}) (interface{}, bool) {
	if inst == nil {
		return nil, false
	}
	method := reflect.ValueOf(inst).MethodByName("GetSelectedNode")
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 2 {
		return nil, false
	}
	results := method.Call(nil)
	if len(results) != 2 || results[1].Kind() != reflect.Bool || !results[1].Bool() {
		return nil, false
	}
	return results[0].Interface(), true
}
