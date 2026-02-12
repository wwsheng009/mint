package reconciler

// =============================================================================
// List Detector - 动态列表检测器
// =============================================================================
// 检测父节点是否是动态列表组件
// 动态列表的子节点必须设置稳定的Key
// =============================================================================

import (
	"fmt"
	"os"
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 动态列表类型的标签
// 这些组件的子节点必须设置Key
var dynamicListTags = map[string]bool{
	"List":        true,
	"GridView":    true,
	"VirtualList": true,
	"ForEach":     true,
	"Map":         true,
}

// isDynamicList 检查父节点是否是动态列表
//
// 动态列表的判断依据:
// 1. 父节点的标签在预定义列表中
// 2. VNode实现了IsDynamicList()方法
// 3. 父节点的Props中标记了_dynamicList
//
// 返回: true表示是动态列表，子节点必须设置Key
func isDynamicList(parent *Fiber) bool {
	if parent == nil {
		return false
	}

	// 检查1: 标签匹配
	if dynamicListTags[parent.Tag] {
		return true
	}

	// 检查2: VNode类型检查（通过接口）
	if parent.VNode != nil {
		if vnode, ok := parent.VNode.(interface{ IsDynamicList() bool }); ok {
			return vnode.IsDynamicList()
		}
	}

	// 检查3: Props标记
	if parent.VNode != nil && parent.VNode.Props() != nil {
		if isDynamic, ok := parent.VNode.Props()["_dynamicList"].(bool); ok {
			return isDynamic
		}
	}

	return false
}

// requireKeyPanic 当动态列表缺少Key时panic
// 提供详细的错误信息和修复建议
func requireKeyPanic(parent *Fiber, vnode rtui.VNode, siblingIndex int) {
	// 获取组件类型信息
	childType := "unknown"
	childTag := ""
	if elem, ok := vnode.(*rtui.ElementVNode); ok {
		childTag = elem.Tag()
		childType = fmt.Sprintf("Element(%s)", elem.Tag())
	} else if comp, ok := vnode.(*rtui.ComponentVNode); ok {
		childTag = comp.Name()
		childType = fmt.Sprintf("Component(%s)", comp.Name())
	} else {
		childType = fmt.Sprintf("%T", vnode)
	}

	// 获取父节点标签（安全访问）
	parentTag := "unknown"
	if parent != nil {
		parentTag = parent.Tag
	}

	// 构建详细的panic消息
	panicMsg := buildKeyPanicMessage(parentTag, childType, childTag, siblingIndex)

	if os.Getenv("TUI_DEBUG_KEY") == "true" || os.Getenv("TUI_DEBUG") == "true" {
		// 调试模式：详细信息和调用栈
		panic(fmt.Sprintf(
			"%s\n\nParent Path: %s\nChild Type: %s\nChild Tag: %s\nSibling Index: %d",
			panicMsg,
			parent.Path,
			childType,
			childTag,
			siblingIndex,
		))
	} else {
		// 生产模式：简洁的错误信息
		panic(fmt.Sprintf(
			"Dynamic list '%s' requires key for child at index %d. "+
				"Use .Key(item.ID) to provide a stable identifier.",
			parent.Tag, siblingIndex,
		))
	}
}

// buildKeyPanicMessage 构建详细的panic消息
func buildKeyPanicMessage(parentTag, childType, childTag string, siblingIndex int) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔═══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           Dynamic List Requires Key - Missing Key               ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Problem: Child component at index %d is missing a required key.\n\n", siblingIndex))

	sb.WriteString("Details:\n")
	sb.WriteString("────────\n")
	sb.WriteString(fmt.Sprintf("  Parent Component: %s (dynamic list)\n", parentTag))
	sb.WriteString(fmt.Sprintf("  Child Component:  %s\n", childType))
	if childTag != "" && childTag != childType {
		sb.WriteString(fmt.Sprintf("  Child Tag:        %s\n", childTag))
	}
	sb.WriteString("\n")

	sb.WriteString("Why is a key required?\n")
	sb.WriteString("─────────────────────────\n")
	sb.WriteString("  Dynamic lists require each child to have a stable key for:\n")
	sb.WriteString("    • Preserving component state across renders\n")
	sb.WriteString("    • Correct event routing and handler binding\n")
	sb.WriteString("    • Proper reconciliation when list order changes\n")
	sb.WriteString("    • Avoiding state loss when items are added/removed\n")
	sb.WriteString("\n")

	sb.WriteString("How to Fix:\n")
	sb.WriteString("───────────\n")
	sb.WriteString("  ❌ Wrong (missing key):\n")
	sb.WriteString("     " + parentTag + "().Children(\n")
	sb.WriteString("       Item(item1).Build(),\n")
	sb.WriteString("       Item(item2).Build(),\n")
	sb.WriteString("     )\n")
	sb.WriteString("\n")
	sb.WriteString("  ✅ Correct (using data ID):\n")
	sb.WriteString("     " + parentTag + "().Children(\n")
	sb.WriteString("       Item(item1).Key(item1.ID).Build(),\n")
	sb.WriteString("       Item(item2).Key(item2.ID).Build(),\n")
	sb.WriteString("     )\n")
	sb.WriteString("\n")

	sb.WriteString("Recommended Key Sources:\n")
	sb.WriteString("────────────────────────\n")
	sb.WriteString("  • item.ID       (database ID)\n")
	sb.WriteString("  • item.UUID     (unique identifier)\n")
	sb.WriteString("  • item.Slug     (URL-friendly slug)\n")
	sb.WriteString("  • item.Email    (email address)\n")
	sb.WriteString("\n")
	sb.WriteString("⚠️  DO NOT use:\n")
	sb.WriteString("  • array index   (will change when list is modified)\n")
	sb.WriteString("  • random UUID   (will change on every render)\n")
	sb.WriteString("  • array position (unstable)\n")
	sb.WriteString("\n")

	sb.WriteString("Learn More:\n")
	sb.WriteString("──────────\n")
	sb.WriteString("  https://react.dev/learn/rendering-lists#why-does-react-need-keys\n")

	return sb.String()
}

// ValidateDynamicListKeys 验证动态列表的所有子节点都有Key
// 这是一个可选的预检查，可以在构建HitMap时调用
func ValidateDynamicListKeys(parent *Fiber, children []rtui.VNode) error {
	if !isDynamicList(parent) {
		return nil
	}

	var missingKeys []int
	for i, child := range children {
		if child.Key() == "" {
			missingKeys = append(missingKeys, i)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf(
			"dynamic list '%s' has %d children missing keys at indices: %v",
			parent.Tag, len(missingKeys), missingKeys,
		)
	}

	return nil
}
