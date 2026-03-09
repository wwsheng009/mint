package menu

import "strings"

type ShortcutBinding struct {
	Item     MenuItem
	Path     []int
	Shortcut Shortcut
}

func NormalizeItems(items []MenuItem) []MenuItem {
	return cloneItems(items)
}

func FirstSelectableIndex(items []MenuItem) int {
	for i, item := range items {
		if item.Normalize().IsSelectable() {
			return i
		}
	}
	return -1
}

func LastSelectableIndex(items []MenuItem) int {
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].Normalize().IsSelectable() {
			return i
		}
	}
	return -1
}

func NextSelectableIndex(items []MenuItem, current int) int {
	if len(items) == 0 {
		return -1
	}
	if current < 0 || current >= len(items) {
		return FirstSelectableIndex(items)
	}
	for step := 1; step <= len(items); step++ {
		idx := (current + step) % len(items)
		if items[idx].Normalize().IsSelectable() {
			return idx
		}
	}
	return current
}

func PrevSelectableIndex(items []MenuItem, current int) int {
	if len(items) == 0 {
		return -1
	}
	if current < 0 || current >= len(items) {
		return LastSelectableIndex(items)
	}
	for step := 1; step <= len(items); step++ {
		idx := (current - step + len(items)) % len(items)
		if items[idx].Normalize().IsSelectable() {
			return idx
		}
	}
	return current
}

func MatchTypeahead(items []MenuItem, prefix string, start int) int {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" || len(items) == 0 {
		return -1
	}
	for step := 1; step <= len(items); step++ {
		idx := (start + step + len(items)) % len(items)
		item := items[idx].Normalize()
		if !item.IsSelectable() {
			continue
		}
		if strings.HasPrefix(strings.ToLower(item.Label), prefix) {
			return idx
		}
	}
	return -1
}

func CollectShortcuts(items []MenuItem) []ShortcutBinding {
	bindings := make([]ShortcutBinding, 0)
	collectShortcuts(&bindings, cloneItems(items), nil)
	return bindings
}

func MatchShortcut(items []MenuItem, combo string) (ShortcutBinding, bool) {
	normalized := normalizeCombo(combo)
	for _, binding := range CollectShortcuts(items) {
		if normalizeCombo(binding.Shortcut.Combo) == normalized {
			return binding, true
		}
	}
	return ShortcutBinding{}, false
}

func ParentPath(path []int) []int {
	if len(path) == 0 {
		return nil
	}
	parent := append([]int(nil), path[:len(path)-1]...)
	if len(parent) == 0 {
		return nil
	}
	return parent
}

func ChildPath(path []int, index int) []int {
	child := append([]int(nil), path...)
	child = append(child, index)
	return child
}

func PathEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ItemAtPath(items []MenuItem, path []int) (MenuItem, bool) {
	if len(path) == 0 {
		return MenuItem{}, false
	}
	current := cloneItems(items)
	var item MenuItem
	for depth, index := range path {
		if index < 0 || index >= len(current) {
			return MenuItem{}, false
		}
		item = current[index].Normalize()
		if depth < len(path)-1 {
			current = item.Children
		}
	}
	return item, true
}

func ChildrenAtPath(items []MenuItem, path []int) ([]MenuItem, bool) {
	if len(path) == 0 {
		return cloneItems(items), true
	}
	item, ok := ItemAtPath(items, path)
	if !ok {
		return nil, false
	}
	if len(item.Children) == 0 {
		return nil, false
	}
	return cloneItems(item.Children), true
}

func FirstSelectablePath(items []MenuItem, basePath []int) ([]int, bool) {
	children, ok := ChildrenAtPath(items, basePath)
	if !ok {
		return nil, false
	}
	index := FirstSelectableIndex(children)
	if index < 0 {
		return nil, false
	}
	return ChildPath(basePath, index), true
}

func collectShortcuts(dst *[]ShortcutBinding, items []MenuItem, basePath []int) {
	for index, raw := range items {
		item := raw.Normalize()
		if !item.IsVisible() {
			continue
		}
		path := append(append([]int(nil), basePath...), index)
		if strings.TrimSpace(item.Shortcut.Combo) != "" {
			*dst = append(*dst, ShortcutBinding{Item: item, Path: path, Shortcut: item.Shortcut})
		}
		if len(item.Children) > 0 {
			collectShortcuts(dst, item.Children, path)
		}
	}
}

func normalizeCombo(combo string) string {
	combo = strings.ToLower(strings.TrimSpace(combo))
	combo = strings.ReplaceAll(combo, " ", "")
	combo = strings.ReplaceAll(combo, "cmd+", "ctrl+")
	combo = strings.ReplaceAll(combo, "command+", "ctrl+")
	combo = strings.ReplaceAll(combo, "option+", "alt+")
	combo = strings.ReplaceAll(combo, "control+", "ctrl+")
	return combo
}

func cloneItems(items []MenuItem) []MenuItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]MenuItem, len(items))
	for i, item := range items {
		normalized := item.Normalize()
		out[i] = normalized
		if normalized.Metadata != nil {
			metadata := make(map[string]any, len(normalized.Metadata))
			for key, value := range normalized.Metadata {
				metadata[key] = value
			}
			out[i].Metadata = metadata
		}
		out[i].Children = cloneItems(normalized.Children)
	}
	return out
}
