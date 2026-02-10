package component

import (
	"github.com/wwsheng009/mint/framework/cmd"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// Updater 是支持 Msg 更新的组件接口
//
// 实现了 Updater 接口的组件可以选择性地处理 Msg 而不是原始的 Event。
// 这提供了更高级别的抽象，使组件更容易测试和维护。
type Updater interface {
	// Update 处理消息并返回命令
	//
	// Update 方法接收一个 Msg，根据消息更新组件状态，
	// 并可选地返回一个 Cmd 用于副作用（如 I/O、定时器等）。
	//
	// 返回 nil 表示没有命令需要执行。
	//
	// 使用示例：
	//   func (b *Button) Update(message runtimemsg.Msg) cmd.Cmd {
	//       switch m := message.(type) {
	//       case *runtimemsg.KeyMsg:
	//           if m.IsEnter() {
	//               b.onClick()
	//           }
	//       case *runtimemsg.MouseMsg:
	//           if m.IsClick() {
	//               b.onClick()
	//           }
	//       }
	//       return nil
	//   }
	Update(message runtimemsg.Msg) cmd.Cmd
}

// UpdateWithModel 是带模型的组件更新接口
//
// 这个接口类似于 Elm Architecture 的 update 函数：
//   update(msg, model) -> (model, cmd)
type UpdateWithModel interface {
	// UpdateWithModel 更新模型并返回命令
	//
	// 这个方法返回新的模型和命令。
	// 如果返回的模型与输入模型不同，应该替换旧模型。
	UpdateWithModel(message runtimemsg.Msg, model interface{}) (newModel interface{}, command cmd.Cmd)
}

// CanUpdate 检查组件是否实现了 Updater 接口
func CanUpdate(component interface{}) bool {
	_, ok := component.(Updater)
	return ok
}

// TryUpdate 尝试调用组件的 Update 方法
//
// 如果组件实现了 Updater 接口，调用其 Update 方法并返回结果。
// 否则返回 nil。
func TryUpdate(component interface{}, message runtimemsg.Msg) cmd.Cmd {
	if updater, ok := component.(Updater); ok {
		return updater.Update(message)
	}
	return nil
}

// TryUpdateWithModel 尝试调用组件的 UpdateWithModel 方法
//
// 如果组件实现了 UpdateWithModel 接口，调用其方法并返回结果。
// 否则返回 (nil, nil)。
func TryUpdateWithModel(component interface{}, message runtimemsg.Msg, model interface{}) (interface{}, cmd.Cmd) {
	if updater, ok := component.(UpdateWithModel); ok {
		return updater.UpdateWithModel(message, model)
	}
	return nil, nil
}
