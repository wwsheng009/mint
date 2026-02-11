package instance

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/internal/log"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Reconcile 算法 - O(n) VNode → Instance 同步
// =============================================================================
// 根据 fix1.md：
// > 输入：oldChildren []*Instance, newVChildren []*VNode
// > 输出：newChildren []*Instance (复用 + 新建后的结果)
// >
// > 时间复杂度 O(n)
// > 无递归回溯
// > 无双层遍历
//
// 前提约束（这让 O(n) 成立）：
// > VNode.ID 在同级 children 中唯一且稳定

// ReconcileChildren 协调子节点列表
//
// 算法步骤：
// Step 1: 建旧节点哈希表
// Step 2: 单次遍历新 VNode 列表
// Step 3: 回收未使用旧节点
func ReconcileChildren(oldChildren []*Instance, newVChildren []ui.VNode) []*Instance {
	// Step 1 — 旧 Instance 建索引
	oldMap := make(map[string]*Instance, len(oldChildren))
	for _, inst := range oldChildren {
		oldMap[inst.ID] = inst
		inst._used = false // 临时标记
	}

	// Step 2 — 遍历新 VNode（核心 O(n)）
	var newChildren []*Instance

	for _, vnode := range newVChildren {
		vnodeID := vnode.Key()
		if vnodeID == "" {
			// Fallback: 使用 Tag 作为 ID
			if tagger, ok := vnode.(interface{ Tag() string }); ok {
				vnodeID = tagger.Tag()
			}
		}

		vnodeType := getVNodeType(vnode)

		if oldInst, ok := oldMap[vnodeID]; ok && oldInst.Type == vnodeType {
			// 🔁 复用
			reuseInstance(oldInst, vnode)
			oldInst._used = true
			newChildren = append(newChildren, oldInst)
		} else {
			// 🆕 创建
			newInst := createInstance(vnode)
			newInst.Mount()
			newChildren = append(newChildren, newInst)
		}
	}

	// Step 3 — 卸载多余旧节点
	for _, inst := range oldChildren {
		if !inst._used {
			unmount(inst)
		}
	}

	return newChildren
}

// reuseInstance 复用现有实例
func reuseInstance(inst *Instance, vnode ui.VNode) {
	// 提取 Props
	props := extractProps(vnode)

	// 更新 Props
	inst.Update(props)

	// 提取并更新事件处理器
	handlers := extractHandlers(vnode)
	inst.Handlers = handlers

	// 递归处理子节点
	inst.Children = ReconcileChildren(inst.Children, vnode.Children())

	log.UILogger.Debug("Reconcile: Reused instance ID=%s Type=%s", inst.ID, inst.Type)
}

// createInstance 从 VNode 创建新实例
func createInstance(vnode ui.VNode) *Instance {
	id := vnode.Key()
	if id == "" {
		if tagger, ok := vnode.(interface{ Tag() string }); ok {
			id = tagger.Tag()
		}
	}

	vnodeType := getVNodeType(vnode)
	props := extractProps(vnode)
	handlers := extractHandlers(vnode)

	inst := NewInstance(id, vnodeType, props)
	inst.Handlers = handlers

	// 递归处理子节点
	inst.Children = ReconcileChildren(nil, vnode.Children())

	log.UILogger.Debug("Reconcile: Created instance ID=%s Type=%s", id, vnodeType)
	return inst
}

// unmount 卸载实例
func unmount(inst *Instance) {
	// 先卸载子节点
	for _, child := range inst.Children {
		unmount(child)
	}

	// 调用生命周期
	inst.Unmount()

	log.UILogger.Debug("Reconcile: Unmounted instance ID=%s Type=%s", inst.ID, inst.Type)
}

// extractProps 从 VNode 提取 Props
func extractProps(vnode ui.VNode) map[string]interface{} {
	props := make(map[string]interface{})

	// 从 VNode 的 Props 提取
	if vnode.Props() != nil {
		for k, v := range vnode.Props() {
			props[k] = v
		}
	}

	return props
}

// extractHandlers 从 VNode 提取事件处理器
//
// 根据 fix1.md：
// > 事件处理器不再在 VNode 上，而是在 Instance 上
//
// 但是 VNode 现在还带着处理器（过渡期），
// 我们需要从 VNode 中提取出来，转移到 Instance
func extractHandlers(vnode ui.VNode) Handlers {
	handlers := Handlers{}

	// 检查 VNode 是否实现了各种处理器接口
	// 注意：这是过渡期的兼容逻辑，将来 VNode 应该是纯数据

	// 尝试提取 onClick
	if clicker, ok := vnode.(interface{ OnClick() func() }); ok {
		handlers.OnClick = clicker.OnClick()
		log.UILogger.Debug("[extractHandlers] Found OnClick handler")
	}

	// 尝试提取 onMouseEnter
	if enterer, ok := vnode.(interface{ OnMouseEnter() func() }); ok {
		handlers.OnMouseEnter = enterer.OnMouseEnter()
	}

	// 尝试提取 onMouseLeave
	if leaver, ok := vnode.(interface{ OnMouseLeave() func() }); ok {
		handlers.OnMouseLeave = leaver.OnMouseLeave()
	}

	// 尝试提取 onUpdate (Msg/Cmd 架构)
	// 修复：匹配实际的 Update 方法签名，返回 cmd.Cmd
	if updater, ok := vnode.(interface {
		Update(runtimemsg.Msg) cmd.Cmd
	}); ok {
		handlers.OnUpdate = updater.Update
		log.UILogger.Debug("[extractHandlers] ✅ Found Update handler for vnode type=%T", vnode)
	} else {
		log.UILogger.Debug("[extractHandlers] ❌ No Update handler for vnode type=%T", vnode)
	}

	return handlers
}

// getVNodeType 获取 VNode 类型
// 使用反射获取实际的类型名称，而不是 Type().String()
// 这样可以自动支持所有新组件，无需维护硬编码的类型列表
func getVNodeType(vnode ui.VNode) string {
	// 使用反射获取实际的类型名称
	// 例如：*ui.ButtonVNode -> "ButtonVNode"
	typeName := ""

	// 获取完整的类型名称
	fullType := ""
	switch v := vnode.(type) {
	case nil:
		return ""
	default:
		// 使用 fmt.Sprintf 或反射获取类型名称
		// 但更简单的方法是：使用指针解引用获取实际类型
		fullType = fmt.Sprintf("%T", v)
	}

	// 从完整类型名称中提取类型名
	// 例如： "*ui.ButtonVNode" -> "ButtonVNode"
	parts := strings.Split(fullType, ".")
	if len(parts) > 0 {
		typeName = parts[len(parts)-1]
	}

	// 如果类型名称为空或只有指针符号，使用 Type().String() 作为后备
	if typeName == "" || typeName == "*" {
		typeName = vnode.Type().String()
	}

	return typeName
}
