# 组件 Intent 设置检查报告

> ⚠️ **DEPRECATED** - 本文档已归档
>
> **新架构**: **Store + Reducer**
> **实现审查**: [`/docs/ui/store/reviews/IMPLEMENTATION_REVIEW.md`](../../ui/store/reviews/IMPLEMENTATION_REVIEW.md)
> **新 Intent 文档**: [`/docs/ui/store/api/API_REFERENCE.md#intent-系统`](../../ui/store/api/API_REFERENCE.md)
>
> 本文档记录了 Phase 10 期间的组件 Intent 检查。现在 Store + Reducer 架构已完整实现，所有组件的 Intent 处理已标准化。

---

**生成时间**: 2026-02-26
**归档时间**: 2026-03-08
**最后更新**: 2026-02-26 (Phase 10 完成)

---

## 一、检查标准

### MVP 架构 Intent 模式

| 标准 | 说明 |
|------|------|
| **状态携带** | Intent 应该携带运行时值（如文本、索引、布尔值） |
| **FieldChangeIntent** | 表单组件应支持 `ForField()` 和发射 FieldChangeIntent |
| **类型安全** | 使用 `StateKey[T]` 定义字段键（可选） |
| **单一入口** | 所有状态变更通过 Intent 入口，不允许直接修改 |

---

## 二、组件检查结果

### ✅ 符合 MVP 模式的组件

#### 1. Button (`ui/components/button/`)

| 项目 | 状态 | 说明 |
|------|------|------|
| Builder API | ✅ | `OnPress(pressIntent intent.Intent)` |
| Instance 行为 | ✅ | PressableBehavior 发射 pressIntent |
| 状态携带 | ✅ | 不需要（点击事件无值） |
| ForField | N/A | 按钮不适合字段绑定模式 |

**代码示例**:
```go
button.NewBuilder("Submit").
    OnPress(SubmitFormIntent{}).
    Build()
```

---

#### 2. Checkbox (`ui/components/checkbox/`)

| 项目 | 状态 | 说明 |
|------|------|------|
| Builder API | ✅ | `OnToggle(toggleIntent intent.Intent)` |
| Instance 行为 | ✅ | `Toggle()` 发射 `FieldChangeIntent` |
| 状态携带 | ✅ | 值转换为字符串 `"true"`/`"false"` |
| ForField | ✅ | `ForField(intent.ForField(key))` |

**代码示例**:
```go
var agree = intent.StateKey[bool]("agree")

checkbox.NewBuilder().
    OnToggle(ToggleAgreeIntent{}).
    ForField(intent.ForField(agree)).
    Build()
```

---

#### 3. Input (`ui/components/input/`)

| 项目 | 状态 | 说明 |
|------|------|------|
| Builder API | ✅ | `OnChange(changeIntent intent.Intent)` |
| Instance 行为 | ✅ | `InsertText()`/`DeleteText()` 发射 `FieldChangeIntent` |
| 状态携带 | ✅ | 携带运行时文本值 |
| ForField | ✅ | `ForField(intent.ForField(key))` |

**代码示例**:
```go
var email = intent.StateKey[string]("email")

input.NewBuilder().
    Placeholder("Enter email").
    ForField(intent.ForField(email)).
    Build()
```

---

### ✅ 已修复的组件 (Phase 9)

#### 4. Textarea (`ui/components/textarea/`) ✅

| 项目 | 状态 | 说明 |
|------|------|------|
| Builder API | ✅ | `OnChange(i intent.Intent)` + **`ForField()`** |
| Instance 行为 | ✅ | **`InsertText()` 发射 `FieldChangeIntent`** |
| 状态携带 | ✅ | **携带运行时文本值** |
| ForField | ✅ | `ForField(intent.ForField(key))` |

**修改内容** (`instance.go`):
```go
// 添加 changeIntentField 字段
type Instance struct {
    changeIntentField intent.FieldIntent  // For FieldChangeIntent extraction
    // ...
}

// 修改 InsertText 发射 FieldChangeIntent
func (inst *Instance) InsertText(text string) bool {
    inst.value += text
    inst.dirty = true

    if inst.intentEmitter != nil {
        if inst.changeIntentField != nil {
            // ✅ 使用 FieldChangeIntent 模式
            changeIntent := intent.FieldChangeIntent{
                Field: inst.changeIntentField.GetField(),
                Value: inst.value,
            }
            inst.intentEmitter(changeIntent)
        } else if inst.changeIntent != nil {
            // fallback 到原始 Intent 模式
            inst.intentEmitter(inst.changeIntent)
        }
    }

    return true
}
```

**Builder API** (`builder.go`):
```go
// ✅ 新增 ForField 方法
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
    b.node.SetProps(rtui.Props{
        "changeIntent": binding,
    })
    return b
}
```

---

#### 5. Select (`ui/components/select/`) ✅

| 项目 | 状态 | 说明 |
|------|------|------|
| Builder API | ✅ | `OnChange(changeIntent intent.Intent)` + **`ForField()`** |
| Instance 行为 | ✅ | **`emitChange()` 发射 `FieldChangeIntent`** |
| 状态携带 | ✅ | **携带选中索引值** |
| ForField | ✅ | `ForField(intent.ForField(key))` |

**修改内容** (`instance.go`):
```go
// 添加 changeIntentField 字段
type Instance struct {
    changeIntentField intent.FieldIntent  // For FieldChangeIntent extraction
    // ...
}

// 修改 emitChange 发射 FieldChangeIntent
func (inst *Instance) emitChange() {
    if inst.intentEmitter != nil {
        if inst.changeIntentField != nil {
            // ✅ 使用 FieldChangeIntent 模式，携带 selectedIndex
            changeIntent := intent.FieldChangeIntent{
                Field: inst.changeIntentField.GetField(),
                Value: fmt.Sprintf("%d", inst.selectedIndex),
            }
            inst.intentEmitter(changeIntent)
        } else if inst.changeIntent != nil {
            // fallback
            inst.intentEmitter(inst.changeIntent)
        }
    }
}
```

**Builder API** (`builder.go`):
```go
// ✅ 新增 ForField 方法
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
    b.node.SetProps(rtui.Props{
        "changeIntent": binding,
    })
    return b
}
```

---

#### 6. Tabs (`ui/components/tabs/`) ✅

| 项目 | 状态 | 说明 |
|------|------|------|
| Builder API | ✅ | `OnChange(intent intent.Intent)` |
| Instance 行为 | ✅ | **`emitChangeIntent()` 发射 `FieldChangeIntent`** |
| 状态携带 | ✅ | **携带 activeTab 索引值** |
| ForField | ℹ️ | 不提供（Tabs 主要是 UI 切换，非表单字段） |

**修改内容** (`instance.go`):
```go
// 添加 changeIntentField 字段
type Instance struct {
    changeIntentField intent.FieldIntent  // For FieldChangeIntent extraction
    // ...
}

// 修改 emitChangeIntent 发射 FieldChangeIntent
func (inst *Instance) emitChangeIntent(tabID string) {
    if inst.intentEmitter != nil {
        if inst.changeIntentField != nil {
            // ✅ 使用 FieldChangeIntent 模式，携带 activeTab
            changeIntent := intent.FieldChangeIntent{
                Field: inst.changeIntentField.GetField(),
                Value: fmt.Sprintf("%d", inst.activeTab),
            }
            inst.intentEmitter(changeIntent)
        } else if inst.changeIntent != nil {
            // fallback
            inst.intentEmitter(inst.changeIntent)
        }
    }
}
```

---

## 三、非表单组件

以下组件不需要 MVP 模式：

| 组件 | 原因 |
|------|------|
| `border` | 布局组件，无状态 |
| `divider` | 装饰组件，无状态 |
| `grid` | 布局组件，无状态 |
| `stack` | 布局组件，无状态 |
| `text` | 纯展示组件，无状态 |
| `absolute` | 定位组件，无状态 |
| `modal` | 控制组件，已有 Show/Hide Intent |
| `tooltip` | 显示组件，生命周期模式 |
| `list` | 数据展示组件，可扩展但暂不需要 |
| `treeview` | 数据展示组件，可扩展但暂不需要 |
| `virtuallist` | 数据展示组件，可扩展但暂不需要 |
| `scrollview` | 容器组件，滚动位置管理 |
| `wrap` | 布局组件，无状态 |
| `progress` | 展示组件，通过外部传入值控制 |
| `panel` | 容器组件，无状态 |

---

## 四、总体架构评分

| 组件 | Intent 支持 | MVP 合规性 | ForField | 值携带 | Phase 9 状态 |
|------|-------------|------------|----------|--------|--------------|
| Button | ✅ | ✅ | N/A | N/A | - |
| Checkbox | ✅ | ✅ | ✅ | ✅ | - |
| Input | ✅ | ✅ | ✅ | ✅ | - |
| **Textarea** | ✅ | ✅ | ✅ | ✅ | ✅ 已完成 |
| **Select** | ✅ | ✅ | ✅ | ✅ | ✅ 已完成 |
| **Tabs** | ✅ | ✅ | ℹ️ | ✅ | ✅ 已完成 |

**合规率**: **100%** (6/6 核心组件)

---

## 五、Phase 9 完成总结

### 修复记录

| 组件 | 修改文件 | 关键变更 |
|------|---------|---------|
| Textarea | `instance.go`, `builder.go` | 添加 `changeIntentField`, `ForField()` |
| Select | `instance.go`, `builder.go` | 添加 `changeIntentField`, `ForField()` |
| Tabs | `instance.go` | 添加 `changeIntentField`, 发射逻辑修复 |

### 统一模式

所有状态携带组件现在遵循统一的 MVP 模式：

```go
// 1. Instance 添加 changeIntentField 字段
type Instance struct {
    changeIntentField intent.FieldIntent
    // ...
}

// 2. NewInstance/SetProps 获取该字段
inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")

// 3. 发射时检查并创建 FieldChangeIntent
if inst.changeIntentField != nil {
    changeIntent := intent.FieldChangeIntent{
        Field: inst.changeIntentField.GetField(),
        Value: <运行时值>,
    }
    inst.intentEmitter(changeIntent)
} else {
    // fallback 模式支持旧代码
    inst.intentEmitter(inst.changeIntent)
}
```

---

## 六、下一步行动

### 短期（Phase 9 - 已完成 ✅）

1. ✅ 创建组件 Intent 检查报告
2. ✅ 修复 Textarea 组件
3. ✅ 修复 Select 组件
4. ✅ 修复 Tabs 组件

### 中期（Phase 10 - 已完成 ✅）

5. ✅ 创建 MVP 组件综合示例 (`examples/mvp_components_demo`)
   - Input: 用户名、电子邮件
   - Textarea: 个人简介
   - Select: 国家选择
   - Checkbox: 同意条款
   - 展示 StateKey[T] 类型安全
   - 展示完整的 FieldChangeIntent Handler

### 长期（Phase 11 - 可选）

6. 创建 MVP 模式单元测试
7. 扩展 List/TreeView 组件的 FieldChangeIntent 支持
8. 迁移指南文档更新

---

**审核人**: Qwen Code
**Phase 10 状态**: ✅ **已完成**
**整体 MVP 架构**: ✅ **100% 合规**
