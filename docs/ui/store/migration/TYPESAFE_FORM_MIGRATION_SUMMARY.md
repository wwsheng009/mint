# 表单示例迁移完成

## 迁移结果

参考 `examples/runapp_demo/main.go` 成功将 `typesafe_form_demo` 迁移到 `ui.RunApp[T]` 模式。

### 文件位置

- **原始版本**：`examples/typesafe_form_demo/main_optimized_test.go` (ui.Run + Store)
- **新版本**：`examples/typesafe_form_demo_runapp/main.go` (ui.RunApp[T] + AppRuntime)
- **文档**：`docs/architecture/store/FORM_FIELDMAP_MIGRATION.md` (详细对比)

---

## 核心改动

### 1. FieldMap 使用（完全保留）✅

```go
// 两个版本都使用相同的 FieldMap 模式
var fieldMapBuilder = reducer.BindField(reducer.NewBuilder[AppState]()).
	BindFieldMap(map[string]func(AppState, string) AppState{
		"Username": func(s AppState, val string) AppState {
			s.Username = val
			return s
		},
		"Email": func(s AppState, val string) AppState {
			s.Email = val
			return s
		},
		"Age": func(s AppState, val string) AppState {
			if v, err := strconv.Atoi(val); err == nil {
				s.Age = v
			}
			return s
		},
		"AcceptTerms": func(s AppState, val string) AppState {
			s.AcceptTerms = val == "true"
			return s
		},
		"Subscribe": func(s AppState, val string) AppState {
			s.Subscribe = val == "true"
			return s
		},
	})
```

### 2. 注册方式

```diff
- var appStore = store.NewStore(AppState{...})
- func main() {
-     fieldMap.RegisterToGlobal(appStore)
-     ui.Run(App, ...)
- }

+ func main() {
+     rt := statemachine.NewAppRuntime(initialState, AppView, AppReducer)
+     ui.RunApp(rt,
+         ui.WithInit(func() {
+             appReducerBuilder.RegisterToGlobal(rt.GetStore())
+         }),
+     )
+ }
```

### 3. View 函数

```diff
- func App() ui.VNode {
-     state := appStore.Get()
-     return renderForm(state)
- }

+ func AppView(state AppState) any {
+     return renderAppView(state)
+ }
+
+ func renderAppView(state AppState) ui.VNode {
+     return renderForm(state)
+ }
```

---

## FieldMap 优势

| 特性 | 传统的 switch-case | FieldMap |
|------|-----------------|----------|
| **代码行数** | ~40 行 | ~20 行 (-50%) |
| **类型断言** | 需要 | ❌ 不需要 |
| **维护性** | 难扩展 | 易扩展 |
| **类型安全** | 开发期 | 编译期 |

**示例对比**：

```go
// ❌ 传统模式 - switch-case
reducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
    if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
        switch fieldChange.Field {
        case "Username":
            s.Username = fieldChange.Value
        case "Email":
            s.Email = fieldChange.Value
        case "Age":
            if v, err := strconv.Atoi(fieldChange.Value); err == nil {
                s.Age = v
            }
        // ... 更多 case
        }
    }
    return s
})

// ✅ FieldMap 模式 - 映射表
reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "Username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "Email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
        "Age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
    })
```

---

## 编译验证

```bash
✅ examples/typesafe_form_demo_runapp/main.go - 编译通过
✅ examples/runapp_demo/main.go - 编译通过
✅ examples/store_reducer_demo/ - 编译通过
```

---

## 使用示例

### 运行新版本

```bash
go run ./examples/typesafe_form_demo_runapp/
```

### 特性演示

- ✅ **FieldMap 自动化**：所有字段自动处理，无需手动 switch-case
- ✅ **类型安全**：编译时检查字段类型
- ✅ **自动重新渲染**：状态变化时 UI 自动更新
- ✅ **时间旅行调试**：AppRuntime 内置历史记录

---

## 对比总结

| 方面 | 原 (ui.Run) | 新 (ui.RunApp[T]) |
|------|-------------|-------------------|
| **状态管理** | 全局 Store | AppRuntime |
| **启动方式** | ui.Run(App) | ui.RunApp(rt) |
| **View 返回类型** | ui.VNode | any (包装) |
| **FieldMap** | ✅ 支持 | ✅ 支持 |
| **自动重新渲染** | ❌ 否 | ✅ 是 |
| **时间旅行** | ❌ 否 | ✅ 是 |
| **代码行数** | ~184 行 | ~260 行 |

---

## 最佳实践

### 定义 FieldMap

```go
var fieldMapBuilder = reducer.BindField(reducer.NewBuilder[AppState]())

// 添加字段处理器
fieldMapBuilder.BindFieldMap(map[string]func(AppState, string) AppState{
    "Field1": func(s AppState, val string) AppState { ... },
    "Field2": func(s AppState, val string) AppState { ... },
})

// 获取 Builder 并添加其他 Intent
var appReducerBuilder = fieldMapBuilder.GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState { ... }).
    On(ResetIntent{}, func(s AppState, i intent.Intent) AppState { ... })
```

### 注册 handlers

```go
ui.RunApp(rt,
    ui.WithInit(func() {
        // 注册 FieldMap 的所有字段处理器 + 其他 Intent 处理器
        appReducerBuilder.RegisterToGlobal(rt.GetStore())
    }),
)
```

---

## 迁移检查清单

- [x] 移除全局 `appStore`
- [x] 创建 `AppRuntime`
- [x] 改 `App()` 为 `AppView(state AppState) any`
- [x] 添加类型安全的内部函数 `renderAppView()`
- [x] 将 `fieldMap.RegisterToGlobal(appStore)` 改为 `appReducerBuilder.RegisterToGlobal(rt.GetStore())`
- [x] 使用 `ui.WithInit` 注册 handlers
- [x] **保留 FieldMap 使用**（无需修改）

---

## 相关文档

- **完整对比**：`docs/architecture/store/FORM_FIELDMAP_MIGRATION.md`
- **FieldMap 优化**：`docs/architecture/store/FIELD_BINDING_OPTIMIZATION.md`
- **RunApp 指南**：`docs/architecture/store/RUNAPP_GUIDE.md`
- **类型优化**：`docs/architecture/store/APPVIEW_TYPE_OPTIMIZATION.md`

---

**迁移完成！** 🎉

FieldMap 模式在两个版本中都可以使用，迁移只改变了状态管理方式，不影响字段处理逻辑。
