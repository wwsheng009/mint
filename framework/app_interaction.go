package framework

import (
	"time"

	"github.com/wwsheng009/mint/runtime/input"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// msgToSnapshot 将 Msg 转换为 InputSnapshot
// 用于 InputTracker 推断输入边缘事件
func (a *App) msgToSnapshot(msg runtimemsg.Msg) *input.InputSnapshot {
	if msg == nil {
		return nil
	}

	snapshot := &input.InputSnapshot{
		Timestamp: time.Now().UnixNano(),
	}

	switch m := msg.(type) {
	case *runtimemsg.MouseMsg:
		snapshot.MouseX = m.X
		snapshot.MouseY = m.Y

		// 如果是 Press 或 Release 事件，记录按钮状态
		if m.Action == runtimemsg.MouseActionPress {
			snapshot.MouseButton = m.Button
			snapshot.MouseAction = runtimemsg.MouseActionPress
		} else if m.Action == runtimemsg.MouseActionRelease {
			// Release 时设置 MouseButton 为 Unknown（表示按钮释放）
			snapshot.MouseButton = runtimemsg.MouseButtonUnknown
			snapshot.MouseAction = runtimemsg.MouseActionRelease
		} else {
			snapshot.MouseButton = runtimemsg.MouseButtonUnknown
			snapshot.MouseAction = runtimemsg.MouseActionUnknown
		}

	case *runtimemsg.KeyMsg:
		snapshot.KeyboardKey = m.Rune
		snapshot.SpecialKey = m.Special
		snapshot.Modifiers = runtimemsg.Modifiers{
			Alt:   m.Mod.Alt,
			Ctrl:  m.Mod.Ctrl,
			Shift: m.Mod.Shift,
		}

	default:
		// 其他消息类型（Resize, Quit 等）不生成 InputSnapshot
		return nil
	}

	return snapshot
}

// hitTest 命中测试：查找鼠标位置所在的组件
// 返回组件 ID（使用 Fiber.NodeID），如果没有找到则返回 0
func (a *App) hitTest(x, y int) int {
	if a.hitMap == nil {
		return 0
	}

	// 使用 HitMap 查找命中目标
	// FindAllAt 返回所有在位置  的条目
	entries := a.hitMap.FindAllAt(x, y)
	if len(entries) > 0 {
		// 返回最顶层的组件（最后一个条目）
		return int(entries[len(entries)-1].NodeID)
	}
	return 0
}

// updateInteractionInstances 更新 InteractionContext 的组件注册表
// 从 DeclarativeNode 获取所有实现了交互接口的实例
func (a *App) updateInteractionInstances() {
	if a.interactionCtx == nil {
		return
	}

	// 尝试从 root 获取所有实现了交互接口的实例（Fiber-first 模式）
	if root, ok := a.root.(interface{ GetAllInteractionInstances() map[int]interface{} }); ok {
		// 清空注册表，重新注册（简化处理，防止重复）
		a.interactionCtx.Instances = make(map[int]interface{})

		// 获取所有实现了交互接口的实例
		allInstances := root.GetAllInteractionInstances()

		// 注册到 InteractionContext
		for nodeID, inst := range allInstances {
			a.interactionCtx.RegisterInstance(nodeID, inst)
		}
	}
}

