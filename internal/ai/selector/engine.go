package selector

import (
	"fmt"
	"strconv"
	"strings"

	snapshotpkg "github.com/wwsheng009/mint/internal/ai/snapshot"
)

// Find resolves a selector against the latest frame indexes and snapshot state.
func Find(frame *snapshotpkg.Frame, selector string) ([]snapshotpkg.NodeLocator, error) {
	if frame == nil || frame.Snapshot == nil {
		return nil, nil
	}

	selector = strings.TrimSpace(selector)
	if selector == "" || selector == "*" {
		return all(frame), nil
	}
	if strings.HasPrefix(selector, "#") {
		id := selector[1:]
		loc, ok := frame.ByComponentID[id]
		if !ok {
			return nil, fmt.Errorf("component not found: %s", id)
		}
		return []snapshotpkg.NodeLocator{loc}, nil
	}
	if strings.HasPrefix(selector, ".") {
		typ := selector[1:]
		if locs, ok := frame.ByType[typ]; ok {
			return locs, nil
		}
		return nil, nil
	}
	if strings.HasPrefix(selector, "@") {
		nodeID, err := strconv.ParseUint(selector[1:], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid node selector: %s", selector)
		}
		loc, ok := frame.ByNodeID[nodeID]
		if !ok {
			return nil, fmt.Errorf("node not found: %d", nodeID)
		}
		return []snapshotpkg.NodeLocator{loc}, nil
	}
	if strings.HasPrefix(selector, "path:") {
		path := strings.TrimPrefix(selector, "path:")
		loc, ok := frame.ByPath[path]
		if !ok {
			return nil, fmt.Errorf("path not found: %s", path)
		}
		return []snapshotpkg.NodeLocator{loc}, nil
	}
	if strings.HasPrefix(selector, "[") && strings.HasSuffix(selector, "]") {
		key, value, err := parseAttribute(selector[1 : len(selector)-1])
		if err != nil {
			return nil, err
		}
		return filter(frame, func(loc snapshotpkg.NodeLocator) bool {
			comp, ok := frame.Snapshot.GetComponent(loc.ComponentID)
			if !ok {
				return false
			}
			if prop, ok := comp.Props[key]; ok && fmt.Sprintf("%v", prop) == value {
				return true
			}
			if stateVal, ok := comp.State[key]; ok && fmt.Sprintf("%v", stateVal) == value {
				return true
			}
			return false
		}), nil
	}

	if loc, ok := frame.ByComponentID[selector]; ok {
		return []snapshotpkg.NodeLocator{loc}, nil
	}
	return nil, fmt.Errorf("invalid selector: %s", selector)
}

func all(frame *snapshotpkg.Frame) []snapshotpkg.NodeLocator {
	result := make([]snapshotpkg.NodeLocator, 0, len(frame.ByComponentID))
	for _, loc := range frame.ByComponentID {
		result = append(result, loc)
	}
	return result
}

func filter(frame *snapshotpkg.Frame, keep func(snapshotpkg.NodeLocator) bool) []snapshotpkg.NodeLocator {
	result := make([]snapshotpkg.NodeLocator, 0)
	for _, loc := range frame.ByComponentID {
		if keep(loc) {
			result = append(result, loc)
		}
	}
	return result
}

func parseAttribute(expr string) (string, string, error) {
	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid attribute selector: [%s]", expr)
	}
	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
	if key == "" {
		return "", "", fmt.Errorf("invalid attribute selector: [%s]", expr)
	}
	return key, value, nil
}
