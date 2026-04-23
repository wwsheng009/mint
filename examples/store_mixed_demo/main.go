// Store Mixed Mode Demo
//
// 演示混合模式状态管理：
//   1. useState - 组件内部局部状态
//   2. Store + Reducer - 全局状态管理
//   3. UseStoreSelector - 订阅派生状态
//
// 运行: go run ./examples/store_mixed_demo/

package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ==============================================================================
// AppState - 应用状态定义
// ==============================================================================

type AppState struct {
	// 计数器
	Count int

	// 表单字段
	Username string
	Email    string

	// 列表数据
	Items      []Item
	FilterText string
}

type Item struct {
	Name  string
	Price float64
}

// ==============================================================================
// 全局 Store
// ==============================================================================

var appStore = store.NewStore(AppState{
	Count:    0,
	Username: "",
	Email:    "",
	Items: []Item{
		{Name: "Apple", Price: 2.99},
		{Name: "Banana", Price: 1.49},
		{Name: "Orange", Price: 3.99},
		{Name: "Grape", Price: 4.99},
		{Name: "Mango", Price: 5.99},
	},
	FilterText: "",
})

// ==============================================================================
// Intents - 自定义 Intents
// ==============================================================================

type IncrementIntent struct{}
type DecrementIntent struct{}
type ClearTextIntent struct{}
type AddItemIntent struct{}
type ToggleExpandedIntent struct{}

func (IncrementIntent) IntentType() string      { return "MixedDemoIncrement" }
func (DecrementIntent) IntentType() string      { return "MixedDemoDecrement" }
func (ClearTextIntent) IntentType() string      { return "MixedDemoClearText" }
func (AddItemIntent) IntentType() string        { return "MixedDemoAddItem" }
func (ToggleExpandedIntent) IntentType() string { return "MixedDemoToggleExpanded" }

// ==============================================================================
// Reducer - 状态管理逻辑
// ==============================================================================

// Reducer 注册到 Intent Registry
// BuildAndRegister 有副作用（注册），这里不需要保存返回值
func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			return s
		}).
		On(ClearTextIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Username = ""
			s.Email = ""
			return s
		}).
		On(AddItemIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Items = append(s.Items, Item{
				Name:  fmt.Sprintf("Item %d", len(s.Items)+1),
				Price: float64(len(s.Items)+1) * 1.99,
			})
			return s
		}).
		// 字段绑定（通过 ForField 自动发送 FieldChangeIntent）
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			if fci, ok := i.(intent.FieldChangeIntent); ok {
				switch fci.Field {
				case "username":
					s.Username = fci.Value
				case "email":
					s.Email = fci.Value
				case "filterText":
					s.FilterText = fci.Value
				}
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

// ==============================================================================
// 组件：使用 useState 的组件（局部状态）
// ==============================================================================

func ExpanderComponent() ui.VNode {
	// 使用 useState - 组件内部局部状态
	// 注意：这个 expanded 状态只在组件内部存在，不会存储到 Store
	expanded := ui.UseStoreSelector(
		appStore,
		func(s AppState) int { return s.Count },
	)%2 == 0 // 简单地用 count 来模拟展开/折叠

	items := []ui.VNode{
		ui.NewTextBuilder("=== 1. Store 订阅状态 ===").
			FgColor("green").
			Build(),
		ui.Text(fmt.Sprintf("  折叠状态: %s (基于 Count) ", map[bool]string{
			true:  "展开",
			false: "折叠",
		}[expanded])),
	}

	if expanded {
		items = append(items,
			ui.Text("  这是一个基于 Store 的状态"),
			ui.Text("  可以跨组件访问"),
			ui.Text("  数据流：Intent → Reducer → Store → UI"),
		)
	}

	return ui.VStack(items...)
}

// ==============================================================================
// 组件：计数器（Store + UseStoreSelector）
// ==============================================================================

func CounterComponent() ui.VNode {
	// 使用 UseStoreSelector 订阅 Store 中的状态
	count := ui.UseStoreSelector(
		appStore,
		func(s AppState) int { return s.Count },
	)

	return ui.VStack(
		ui.NewTextBuilder("=== 2. Store 计数器 ===").
			FgColor("green").
			Build(),
		ui.HStack(
			ui.Text("计数: "),
			ui.NewTextBuilder(fmt.Sprintf("%d", count)).
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").
				OnPress(IncrementIntent{}).
				Variant(ui.ButtonVariantPrimary).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" - ").
				OnPress(DecrementIntent{}).
				Variant(ui.ButtonVariantSecondary).
				Build(),
		),
	)
}

// ==============================================================================
// 组件：功能性更新计数器（UseStoreFieldFunctional）
// ==============================================================================

// FunctionalCounterIntent - 直接更新
type FunctionalCounterDirectIntent struct {
	Value int
}
type FunctionalCounterStepIntent struct {
	Step int
}

func (FunctionalCounterDirectIntent) IntentType() string { return "FuncDirect" }
func (FunctionalCounterStepIntent) IntentType() string   { return "FuncStep" }

func init() {
	reducer.NewBuilder[AppState]().
		On(FunctionalCounterDirectIntent{}, func(s AppState, i intent.Intent) AppState {
			if d, ok := i.(FunctionalCounterDirectIntent); ok {
				s.Count = d.Value
			}
			return s
		}).
		On(FunctionalCounterStepIntent{}, func(s AppState, i intent.Intent) AppState {
			if st, ok := i.(FunctionalCounterStepIntent); ok {
				s.Count += st.Step
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

func FunctionalCounterComponent() ui.VNode {
	// 使用 UseStoreFieldFunctional 支持函数式更新
	// 这里演示订阅，实际更新通过 Intent（上面的 init 注册）
	count, _ := ui.UseStoreFieldFunctional(
		appStore,
		func(s AppState) int { return s.Count },
		func(s AppState, v int) AppState { s.Count = v; return s },
	)

	return ui.VStack(
		ui.NewTextBuilder("=== 5. 函数式更新计数器 ===").
			FgColor("purple").
			Build(),
		ui.HStack(
			ui.Text("计数: "),
			ui.NewTextBuilder(fmt.Sprintf("%d", count)).
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(" "),
			// 方式1: 发送 Intent (Step)
			ui.NewButtonBuilder(" +10 ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(FunctionalCounterStepIntent{Step: 10}).
				Build(),
			ui.Text(" "),
			// 方式2: 发送 Intent (Step)
			ui.NewButtonBuilder(" +1 ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(FunctionalCounterStepIntent{Step: 1}).
				Build(),
			ui.Text(" "),
			// 方式3: 发送 Intent (Direct)
			ui.NewButtonBuilder(" 设置为 0 ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(FunctionalCounterDirectIntent{Value: 0}).
				Build(),
		),
		ui.Text("  +10/+1: 使用 FunctionalCounterStepIntent (Step)"),
		ui.Text("  设置为 0: 使用 FunctionalCounterDirectIntent (Direct)"),
		ui.Text(""),
		ui.Text("  UseStoreFieldFunctional 支持:"),
		ui.Text("    1. setField(newValue) - 直接设置值"),
		ui.Text("    2. setField(func(old) old+1) - 函数式更新"),
		ui.Text("  函数式更新避免闭包捕获旧值问题"),
	)
}

// ==============================================================================
// 组件：表单（ForField + Reducer）
// ==============================================================================

func FormComponent() ui.VNode {
	// 从 Store 读取状态
	state := appStore.Get()

	return ui.VStack(
		ui.NewTextBuilder("=== 3. 表单（ForField + Reducer）===").
			FgColor("green").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("用户名:").FgColor("blue").Build(),
			ui.NewInputBuilder().
				Placeholder("用户名").
				Value(state.Username).
				InsertCursor().
				ForField(intent.BindField("username")).
				Width(20).
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("邮箱:").FgColor("blue").Build(),
			ui.NewInputBuilder().
				Placeholder("邮箱").
				Value(state.Email).
				InsertCursor().
				ForField(intent.BindField("email")).
				Width(20).
				Build(),
		),
		ui.HStack(
			ui.Text("当前值: "),
			ui.NewTextBuilder(fmt.Sprintf("[%s, %s]", state.Username, state.Email)).
				FgColor("gray").
				Build(),
		),
		ui.HStack(
			ui.NewButtonBuilder("清空").
				OnPress(ClearTextIntent{}).
				Variant(ui.ButtonVariantSecondary).
				Build(),
		),
	)
}

// ==============================================================================
// 组件：列表 + UseStoreSelector
// ==============================================================================

func ListComponent() ui.VNode {
	// 使用 UseStoreSelector 计算派生状态
	total := ui.UseStoreSelector(
		appStore,
		func(s AppState) int { return len(s.Items) },
	)

	filtered := ui.UseStoreSelector(
		appStore,
		func(s AppState) []Item {
			if s.FilterText == "" {
				return s.Items
			}
			filtered := make([]Item, 0)
			for _, item := range s.Items {
				if strings.Contains(strings.ToLower(item.Name), strings.ToLower(s.FilterText)) {
					filtered = append(filtered, item)
				}
			}
			return filtered
		},
	)

	// 从 Store 读取筛选文本
	state := appStore.Get()

	var items []ui.VNode
	for _, item := range filtered {
		items = append(items,
			ui.HStack(
				ui.Text("  - "),
				ui.NewTextBuilder(item.Name).
					FgColor("cyan").
					Build(),
				ui.Text(" "),
				ui.NewTextBuilder(fmt.Sprintf("$%.2f", item.Price)).
					FgColor("yellow").
					Build(),
			),
		)
	}

	return ui.VStack(
		ui.NewTextBuilder("=== 4. 列表 (UseStoreSelector) ===").
			FgColor("green").
			Build(),
		ui.HStack(
			ui.Text("筛选: "),
			ui.NewInputBuilder().
				Placeholder("输入关键词").
				Value(state.FilterText).
				InsertCursor().
				ForField(intent.BindField("filterText")).
				Width(15).
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("(共 %d 件)", total)).
				FgColor("gray").
				Build(),
		),
		ui.VStack(items...),
		ui.HStack(
			ui.NewButtonBuilder("添加商品").
				OnPress(AddItemIntent{}).
				Variant(ui.ButtonVariantPrimary).
				Build(),
		),
	)
}

// ==============================================================================
// 主应用
// ==============================================================================

func App() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("Mint UI 混合模式状态管理示例").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.NewTextBuilder("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━").
			FgColor("gray").
			Build(),

		// 1. useState 组件（演示局部状态）
		ExpanderComponent(),

		ui.Text(""),

		// 2. Store + UseStoreSelector 组件（演示全局状态订阅）
		CounterComponent(),

		ui.Text(""),

		// 2.5. 功能性更新计数器（UseStoreFieldFunctional）
		FunctionalCounterComponent(),

		ui.Text(""),

		// 3. 表单（ForField + Reducer）（演示字段绑定）
		FormComponent(),

		ui.Text(""),

		// 4. 列表 + UseStoreSelector（演示派生状态）
		ListComponent(),

		ui.NewTextBuilder("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("按 ESC 或 Ctrl+Q 退出").
			FgColor("gray").
			Build(),
	)
}

// ==============================================================================
// Main
// ==============================================================================

func main() {
	fmt.Println("===============================================")
	fmt.Println("Mint UI 混合模式状态管理演示")
	fmt.Println("=")
	fmt.Println("本示例演示混合模式状态管理：")
	fmt.Println("")
	fmt.Println("1. useState - 组件内部局部状态")
	fmt.Println("   - 用于：展开/折叠等 UI 临时状态")
	fmt.Println("   - 特点：简单、不共享、不持久化")
	fmt.Println("")
	fmt.Println("2. UseStoreSelector - 订阅派生状态")
	fmt.Println("   - 用于：计数器、列表等共享状态")
	fmt.Println("   - 特点：自动订阅、类型安全、触发重渲染")
	fmt.Println("")
	fmt.Println("3. ForField + Reducer - 字段绑定")
	fmt.Println("   - 用于：表单输入、列表筛选等")
	fmt.Println("   - 特点：声明式绑定、统一状态更新")
	fmt.Println("")
	fmt.Println("===============================================")
	fmt.Println()

	err := ui.Run(App, ui.WithTitle("Store Mixed Mode Demo"))
	if err != nil {
		panic(err)
	}
}
