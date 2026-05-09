# Store + Reducer + Intent 文档

本目录保存 Mint 应用级状态管理相关文档。当前推荐的应用级状态流是：

```text
Component interaction
  -> runtime/intent.Intent
  -> intent.Registry / Dispatcher
  -> runtime/reducer.Reducer[T]
  -> runtime/store.Store[T]
  -> ui.UseStoreSelector or runtime/statemachine.AppRuntime[T]
  -> Fiber update / render
```

局部组件状态仍可使用 Hooks；跨组件、业务状态和可测试状态流优先使用 Store/Reducer/Intent。

## 当前源码入口

| 关注点 | 源码 |
|---|---|
| Store | `../../../runtime/store/store.go` |
| Reducer | `../../../runtime/reducer/reducer.go` |
| Intent runtime | `../../../runtime/intent` |
| AppRuntime / RunApp | `../../../runtime/statemachine`, `../../../ui/app.go` |
| Store selector hook | `../../../ui/hooks.go` |
| 示例 | `../../../examples/store_reducer_demo`, `../../../examples/store_mixed_demo`, `../../../examples/runapp_demo`, `../../../examples/typed_intent_demo`, `../../../examples/typesafe_form_demo_runapp` |

## 文档目录

当前实际目录：

```text
docs/ui/store/
  api/
  features/
  fixes/
  guides/
  hybrid/
  issues/
  migration/
  optimization/
```

早期文档中提到的 `status/`、`plans/`、`reviews/`、`design/` 目录当前不在 `docs/ui/store/` 下。若需要历史资料，请在 `docsArchive/` 中查找。

## 推荐阅读路径

1. [guides/README.md](guides/README.md): Store/Reducer 使用指南入口。
2. [guides/DEVELOPMENT_GUIDE.md](guides/DEVELOPMENT_GUIDE.md): 开发指南。
3. [guides/RUNAPP_GUIDE.md](guides/RUNAPP_GUIDE.md): `ui.RunApp` 和 AppRuntime。
4. [guides/STORE_REDUCER_GUIDE.md](guides/STORE_REDUCER_GUIDE.md): Store + Reducer 模式。
5. [api/API_REFERENCE.md](api/API_REFERENCE.md): API 参考。
6. [features/TYPE_SAFE_INTENT.md](features/TYPE_SAFE_INTENT.md): 类型安全 Intent。
7. [optimization/FIELD_BINDING_OPTIMIZATION.md](optimization/FIELD_BINDING_OPTIMIZATION.md): 字段绑定优化。

注意：`guides/HOOK_USAGE_GUIDE.md`、`fixes/`、`migration/`、`optimization/` 下的部分文档包含历史问题复盘和旧 API 对比。凡是看到 `ui.Button(label, func)`、`OnClick(func)`、`OnPress(func)`、`input.ForField("name")` 这类片段，都应按当前 `ui.NewXBuilder` + `intent.Intent` / `intent.BindField` 方式改写。

## 子目录说明

### `guides/`

面向使用者的主要指南。

- [README.md](guides/README.md)
- [DEVELOPMENT_GUIDE.md](guides/DEVELOPMENT_GUIDE.md)
- [HOOK_USAGE_GUIDE.md](guides/HOOK_USAGE_GUIDE.md)
- [MIGRATION_GUIDE.md](guides/MIGRATION_GUIDE.md)
- [RUNAPP_GUIDE.md](guides/RUNAPP_GUIDE.md)
- [STORE_REDUCER_GUIDE.md](guides/STORE_REDUCER_GUIDE.md)

### `api/`

- [API_REFERENCE.md](api/API_REFERENCE.md)

### `features/`

- [TYPE_SAFE_INTENT.md](features/TYPE_SAFE_INTENT.md)
- [LOGGING_AND_ERROR_HANDLING_GUIDE.md](features/LOGGING_AND_ERROR_HANDLING_GUIDE.md)

### `optimization/`

- [FIELD_BINDING_OPTIMIZATION.md](optimization/FIELD_BINDING_OPTIMIZATION.md)
- [SYSTEM_ANALYSIS_OPTIMIZATION.md](optimization/SYSTEM_ANALYSIS_OPTIMIZATION.md)
- [APPVIEW_TYPE_OPTIMIZATION.md](optimization/APPVIEW_TYPE_OPTIMIZATION.md)

### `migration/`

- [FORM_FIELDMAP_MIGRATION.md](migration/FORM_FIELDMAP_MIGRATION.md)
- [INTENT_HANDLER_MIGRATION.md](migration/INTENT_HANDLER_MIGRATION.md)

### `hybrid/`

- [STATE_MANAGEMENT_GUIDE.md](hybrid/STATE_MANAGEMENT_GUIDE.md)
- [HYBRID_MODE_IMPLEMENTATION.md](hybrid/HYBRID_MODE_IMPLEMENTATION.md)

### `fixes/` 和 `issues/`

问题记录和修复说明：

- [fixes/RUN_VS_RUNAPP.md](fixes/RUN_VS_RUNAPP.md)
- [fixes/TIMETRAVEL_FIX.md](fixes/TIMETRAVEL_FIX.md)
- [issues/FMT_PRINT_ISSUE.md](issues/FMT_PRINT_ISSUE.md)

## 最小模式

典型结构：

```go
type AppState struct {
    Count int
}

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }

var appStore = store.NewStore(AppState{})

func init() {
    reducer.NewBuilder[AppState]().
        On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
            s.Count++
            return s
        }).
        BuildAndRegister(intent.DefaultRegistry(), appStore)
}
```

在视图中订阅状态切片：

```go
count := ui.UseStoreSelector(appStore, func(s AppState) int {
    return s.Count
})
```

组件发送 intent：

```go
button.NewBuilder("+1").
    OnPress(IncrementIntent{}).
    Build()
```

## 测试建议

```bash
go test ./runtime/store ./runtime/reducer ./runtime/intent -count=1
go test ./examples/store_reducer_demo ./examples/store_mixed_demo ./examples/runapp_demo -count=1
go test ./ui/e2e -run "Button|Form|Input|Select|Tabs" -count=1
```

全量：

```bash
go test ./... -count=1
```

资源受限环境下，优先用分层命令定位行为问题。

## 维护注意

- 不要把旧目录 `status/`、`plans/`、`reviews/`、`design/` 写成当前存在目录。
- 不要把 UseState 描述成应用级业务状态首选方案。
- 新交互组件应优先声明 Intent，状态变更集中到 Reducer 或明确的 instance 状态。
