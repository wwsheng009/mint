package snapshot

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type TreeNode struct {
	Kind     string                 `json:"kind"`
	NodeID   uint64                 `json:"node_id,omitempty"`
	Path     string                 `json:"path,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Tag      string                 `json:"tag,omitempty"`
	Key      string                 `json:"key,omitempty"`
	ID       string                 `json:"id,omitempty"`
	Rect     map[string]int         `json:"rect,omitempty"`
	Props    map[string]interface{} `json:"props,omitempty"`
	State    map[string]interface{} `json:"state,omitempty"`
	Children []*TreeNode            `json:"children,omitempty"`
}

type TreeOptions struct {
	Compact bool
}

type HitEntry struct {
	NodeID uint64         `json:"node_id"`
	Bounds map[string]int `json:"bounds"`
	ZOrder int            `json:"z_order"`
}

type NodeBundle struct {
	NodeID     uint64     `json:"node_id"`
	Fiber      *TreeNode  `json:"fiber,omitempty"`
	Layout     *TreeNode  `json:"layout,omitempty"`
	Paintable  *TreeNode  `json:"paintable,omitempty"`
	HitEntries []HitEntry `json:"hit_entries,omitempty"`
}

func BuildTree(root component.Node, kind string) (interface{}, error) {
	return BuildTreeWithOptions(root, kind, TreeOptions{})
}

func BuildTreeWithOptions(root component.Node, kind string, opts TreeOptions) (interface{}, error) {
	switch kind {
	case "", "snapshot":
		return nil, fmt.Errorf("unsupported tree kind: %s", kind)
	case "fiber":
		return buildFiberTree(GetFiberRootFromComponent(root), opts), nil
	case "layout":
		return buildLayoutTree(getLayoutRootForTree(root), opts), nil
	case "paintable":
		return buildPaintableTree(getPaintableRootForTree(root), opts), nil
	case "vnode":
		return buildVNodeTree(getRenderedRootForTree(root), opts), nil
	case "hitmap":
		return buildHitEntries(getHitMapForTree(root)), nil
	default:
		return nil, fmt.Errorf("unsupported tree kind: %s", kind)
	}
}

func BuildNodeBundle(root component.Node, nodeID uint64) (*NodeBundle, error) {
	return BuildNodeBundleWithOptions(root, nodeID, TreeOptions{})
}

func BuildNodeBundleWithOptions(root component.Node, nodeID uint64, opts TreeOptions) (*NodeBundle, error) {
	if nodeID == 0 {
		return nil, fmt.Errorf("invalid node id")
	}
	bundle := &NodeBundle{NodeID: nodeID}
	if fiberRoot := GetFiberRootFromComponent(root); fiberRoot != nil {
		if fiber := rtui.FindFiberByID(fiberRoot, nodeID); fiber != nil {
			bundle.Fiber = buildFiberTree(fiber, opts)
		}
	}
	if layoutRoot := getLayoutRootForTree(root); layoutRoot != nil {
		if box := findLayoutByNodeID(layoutRoot, nodeID); box != nil {
			bundle.Layout = buildLayoutTree(box, opts)
		}
	}
	if paintRoot := getPaintableRootForTree(root); paintRoot != nil {
		if box := paintRoot.FindByID(nodeID); box != nil {
			bundle.Paintable = buildPaintableTree(box, opts)
		}
	}
	if hitMap := getHitMapForTree(root); hitMap != nil {
		for _, entry := range hitMap.AllEntries() {
			if entry.NodeID == nodeID {
				bundle.HitEntries = append(bundle.HitEntries, HitEntry{
					NodeID: entry.NodeID,
					Bounds: rectMap(entry.Bounds.X, entry.Bounds.Y, entry.Bounds.Width, entry.Bounds.Height),
					ZOrder: entry.ZOrder,
				})
			}
		}
	}
	if bundle.Fiber == nil && bundle.Layout == nil && bundle.Paintable == nil && len(bundle.HitEntries) == 0 {
		return nil, fmt.Errorf("node not found: %d", nodeID)
	}
	return bundle, nil
}

func GetFiberRootFromComponent(root component.Node) *rtui.Fiber {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetFiberRoot() *rtui.Fiber }); ok {
		return provider.GetFiberRoot()
	}
	return nil
}

func getLayoutRootForTree(root component.Node) *layout.LayoutBox {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetLayoutRoot() *layout.LayoutBox }); ok {
		return provider.GetLayoutRoot()
	}
	return nil
}

func getPaintableRootForTree(root component.Node) *paint.PaintableBox {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetPaintableRoot() *paint.PaintableBox }); ok {
		return provider.GetPaintableRoot()
	}
	return nil
}

func getRenderedRootForTree(root component.Node) rtui.VNode {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetRenderedRoot() rtui.VNode }); ok {
		return provider.GetRenderedRoot()
	}
	return nil
}

func getHitMapForTree(root component.Node) *event.HitMap {
	if root == nil {
		return nil
	}
	if provider, ok := root.(interface{ GetHitMap() *event.HitMap }); ok {
		return provider.GetHitMap()
	}
	return nil
}

func buildFiberTree(fiber *rtui.Fiber, opts TreeOptions) *TreeNode {
	if fiber == nil {
		return nil
	}
	props := extractProps(fiber)
	state := extractState(fiber)
	if opts.Compact {
		props = compactMap(props)
		state = compactMap(state)
	}
	node := &TreeNode{
		Kind:   "fiber",
		NodeID: fiber.NodeID,
		Path:   fiber.Path,
		Type:   fiber.Type.String(),
		Tag:    fiber.Tag,
		Key:    fiber.Key,
		ID:     fiber.ID,
		Props:  props,
		State:  state,
	}
	if child := fiber.Child; child != nil {
		for curr := child; curr != nil; curr = curr.Sibling {
			node.Children = append(node.Children, buildFiberTree(curr, opts))
		}
	}
	return node
}

func buildLayoutTree(box *layout.LayoutBox, opts TreeOptions) *TreeNode {
	if box == nil {
		return nil
	}
	nodeID := event.StringToNodeID(box.ID)
	node := &TreeNode{
		Kind:   "layout",
		NodeID: nodeID,
		Type:   box.Tag,
		ID:     box.PropsID,
		Rect:   rectMap(box.AbsX, box.AbsY, box.Width, box.Height),
	}
	for _, child := range box.Children {
		node.Children = append(node.Children, buildLayoutTree(child, opts))
	}
	return node
}

func buildPaintableTree(box *paint.PaintableBox, opts TreeOptions) *TreeNode {
	if box == nil {
		return nil
	}
	node := &TreeNode{
		Kind:   "paintable",
		NodeID: box.NodeID,
		Type:   nodeTypeFromPaintable(box),
		ID:     box.DiffKey,
		Rect:   rectMap(box.X, box.Y, box.Width, box.Height),
	}
	for _, child := range box.Children {
		node.Children = append(node.Children, buildPaintableTree(child, opts))
	}
	return node
}

func buildVNodeTree(vnode rtui.VNode, opts TreeOptions) *TreeNode {
	if vnode == nil {
		return nil
	}
	props := sanitizeMap(map[string]interface{}(vnode.Props()))
	if opts.Compact {
		props = compactMap(props)
	}
	node := &TreeNode{
		Kind:  "vnode",
		Type:  vnode.Type().String(),
		Tag:   vnode.Tag(),
		Key:   vnode.Key(),
		Props: props,
	}
	for _, child := range vnode.Children() {
		node.Children = append(node.Children, buildVNodeTree(child, opts))
	}
	return node
}

func buildHitEntries(hitMap *event.HitMap) []HitEntry {
	if hitMap == nil {
		return nil
	}
	result := make([]HitEntry, 0, hitMap.Size())
	for _, entry := range hitMap.AllEntries() {
		result = append(result, HitEntry{
			NodeID: entry.NodeID,
			Bounds: rectMap(entry.Bounds.X, entry.Bounds.Y, entry.Bounds.Width, entry.Bounds.Height),
			ZOrder: entry.ZOrder,
		})
	}
	return result
}

func rectMap(x, y, width, height int) map[string]int {
	return map[string]int{
		"x":      x,
		"y":      y,
		"width":  width,
		"height": height,
	}
}

func compactMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		if compacted, ok := compactValue(value); ok {
			out[key] = compacted
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func compactValue(value interface{}) (interface{}, bool) {
	switch val := value.(type) {
	case nil:
		return nil, false
	case string:
		if val == "" {
			return nil, false
		}
		return val, true
	case bool:
		if !val {
			return nil, false
		}
		return val, true
	case int:
		if val == 0 {
			return nil, false
		}
		return val, true
	case int8:
		if val == 0 {
			return nil, false
		}
		return val, true
	case int16:
		if val == 0 {
			return nil, false
		}
		return val, true
	case int32:
		if val == 0 {
			return nil, false
		}
		return val, true
	case int64:
		if val == 0 {
			return nil, false
		}
		return val, true
	case uint:
		if val == 0 {
			return nil, false
		}
		return val, true
	case uint8:
		if val == 0 {
			return nil, false
		}
		return val, true
	case uint16:
		if val == 0 {
			return nil, false
		}
		return val, true
	case uint32:
		if val == 0 {
			return nil, false
		}
		return val, true
	case uint64:
		if val == 0 {
			return nil, false
		}
		return val, true
	case float32:
		if val == 0 {
			return nil, false
		}
		return val, true
	case float64:
		if val == 0 {
			return nil, false
		}
		return val, true
	case []string:
		if len(val) == 0 {
			return nil, false
		}
		return val, true
	case []int:
		if len(val) == 0 {
			return nil, false
		}
		return val, true
	case []interface{}:
		if len(val) == 0 {
			return nil, false
		}
		out := make([]interface{}, 0, len(val))
		for _, item := range val {
			if compacted, ok := compactValue(item); ok {
				out = append(out, compacted)
			}
		}
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	case map[string]interface{}:
		out := compactMap(val)
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	default:
		return value, true
	}
}

func findLayoutByNodeID(box *layout.LayoutBox, nodeID uint64) *layout.LayoutBox {
	if box == nil {
		return nil
	}
	if event.StringToNodeID(box.ID) == nodeID {
		return box
	}
	for _, child := range box.Children {
		if found := findLayoutByNodeID(child, nodeID); found != nil {
			return found
		}
	}
	return nil
}

func nodeTypeFromPaintable(box *paint.PaintableBox) string {
	if box == nil || box.Node == nil {
		return ""
	}
	if box.Node.Tag() != "" {
		return box.Node.Tag()
	}
	return fmt.Sprintf("%T", box.Node)
}
