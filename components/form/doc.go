// Package form 提供表单输入组件。
//
// 组件列表:
//   - Input/TextInput: 单行文本输入
//   - TextArea: 多行文本输入
//   - Checkbox: 复选框
//   - Select: 下拉选择
//   - Switch: 开关
//   - Slider: 滑块
//   - Field: 表单字段包装器
//
// 使用示例:
//
//	import "github.com/wwsheng009/mint/components/form"
//
//	input := form.Input("Enter name").
//	    Placeholder("Name").
//	    Value(name).
//	    OnChange(func(s string) { name = s })
package form
