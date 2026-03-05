# 字段绑定优化指南 - 消除 Reducer 硬编码

**创建日期**: 2026-03-05
**目标**: 消除 FieldChangeIntent 处理器中的 switch-case 硬编码

---

## 问题：硬编码的 switch-case

在 Store + Reducer 架构中，处理 `FieldChangeIntent` 时通常需要硬编码字段名：

```go
// ❌ 问题：硬编码的字段映射
appReducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
    fci := i.(intent.FieldChangeIntent)
    switch fci.Field {
    case "Username":
        s.Username = fci.Value
    case "Email":
        s.Email = fci.Value
    case "Age":
        var ageVal int
        if _, err := fmt.Sscanf(fci.Value, "%d", &ageVal); err == nil {
            s.Age = ageVal
        }
    case "AcceptTerms":
        s.AcceptTerms = fci.Value == "true"
    case "Subscribe":
        s.Subscribe = fci.Value == "true"
    }
    return s
})
```

**问题**:
1. 每次添加字段都要在 switch 中添加 case
2. 类型转换逻辑分散在各处
3. 代码重复，不易维护
4. 字段名硬编码，容易出错

---

## 解决方案 1: FieldMap（推荐）

使用 `BindFieldMap` 创建字段映射表，所有字段更新逻辑集中定义。

```go
// ✅ 解决方案：使用 FieldMap
import (
    "strconv"
    "github.com/wwsheng009/mint/runtime/reducer"
)

// 定义字段映射表
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        // 字符串字段
        "Username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "Email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
        // 整型字段 - 自动转换字符串到 int
        "Age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
        // 布尔字段 - 自动转换字符串到 bool
        "AcceptTerms": func(s AppState, val string) AppState {
            s.AcceptTerms = val == "true"
            return s
        },
        "Subscribe": func(s AppState, val string) AppState {
            s.Subscribe = val == "true"
            return s
        },
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        // 其他 Intent 处理
        return s
    })
```

**优点**:
- ✅ 单一 `FieldChangeIntent` 处理器，性能更好
- ✅ 字段映射表集中定义，易于维护
- ✅ 每个字段有独立的更新函数，类型安全
- ✅ 类型转换逻辑内联处理，代码简洁
- ✅ 支持 Go 1.18+ 泛型，编译期类型检查

---

## 解决方案 2: 类型化字段绑定

使用 `BindStringField`, `BindIntField`, `BindBoolField` 等 API，自动处理类型转换。

```go
// ✅ 解决方案：使用类型化字段绑定
import "github.com/wwsheng009/mint/runtime/reducer"

// 使用类型化 API 绑定字段
var appReducer = func() reducer.Reducer[AppState] {
    builder := reducer.NewBuilder[AppState]()

    binder := reducer.BindField(builder).
        // 字符串字段
        BindStringField("Username", func(s AppState, val string) AppState {
            s.Username = val
            return s
        }).
        BindStringField("Email", func(s AppState, val string) AppState {
            s.Email = val
            return s
        }).
        // 整型字段 - BindIntField 自动处理字符串到 int 的转换
        BindIntField("Age", func(s AppState, val int) AppState {
            s.Age = val
            return s
        }).
        // 布尔字段 - BindBoolField 自动处理字符串到 bool 的转换
        BindBoolField("AcceptTerms", func(s AppState, val bool) AppState {
            s.AcceptTerms = val
            return s
        }).
        BindBoolField("Subscribe", func(s AppState, val bool) AppState {
            s.Subscribe = val
            return s
        })

    return binder.GetBuilder().
        On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
            return s
        }).
        Build()
}()
```

**优点**:
- ✅ 类型安全：`BindIntField` 接收 `int` 值，不需要手动转换
- ✅ 自动处理类型转换：字符串 ↔ int/bool 自动转换
- ✅ 链式调用：类似 Builder 模式
- ✅ 每个 `Bind*Field` 注册独立的处理器（灵活但性能略低）

**缺点**:
- ❌ 每个 `Bind*Field` 注册一个处理器（比 FieldMap 性能略低）
- ❌ 需要调用 `.GetBuilder()` 并最后调用 `.Build()`

---

## 解决方案 3: 自定义映射表（旧方式）

使用全局切片和查找函数（简单但不推荐）。

```go
// ⚠️ 不推荐：自定义映射表
type FieldType int

const (
    TypeString FieldType = iota
    TypeBool
    TypeInt
)

type FieldUpdate struct {
    Name  string
    Type  FieldType
    Update func(*AppState, string) // 指针更新
}

var fieldUpdates = []FieldUpdate{
    {"Username", TypeString, func(s *AppState, val string) { s.Username = val }},
    {"Age", TypeInt, func(s *AppState, val string) {
        var ageVal int
        if _, err := fmt.Sscanf(val, "%d", &ageVal); err == nil {
            s.Age = ageVal
        }
    }},
    // ...
}

func findFieldUpdate(fieldName string) *FieldUpdate {
    // ... 查找逻辑
}

// 使用
appReducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
    fci := i.(intent.FieldChangeIntent)
    if update := findFieldUpdate(fci.Field); update != nil {
        update.Update(&s, fci.Value)
    }
    return s
})
```

**优点**:
- ✅ 简单易懂

**缺点**:
- ❌ 遍历查找，性能差
- ❌ 需要外部函数辅助
- ❌ 类型定义复杂

---

## 对比表

| 方案 | 性能 | 可维护性 | 类型安全 | 推荐场景 |
|------|------|---------|---------|----------|
| **FieldMap** | ✅ 高 | ✅ 高 | ✅ 是 | **推荐** - 通用场景 |
| **类型化绑定** | ⚠️ 中 | ✅ 高 | ✅ 是 | 复杂类型转换场景 |
| **自定义映射表** | ❌ 低 | ⚠️ 中 | ⚠️ 部分 | 简单场景（不推荐） |
| **switch-case** | ✅ 高 | ❌ 低 | ⚠️ 部分 | 不推荐 |

---

## 最佳实践

### 推荐：使用 FieldMap

```go
// ✅ 推荐：使用 FieldMap 处理所有字段
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
    }).
    GetBuilder().
    On(OtherIntent{}, func(s AppState, i intent.Intent) AppState {
        // 其他 Intent
        return s
    })
```

**原因**:
1. 单一 `FieldChangeIntent` 处理器（性能好）
2. 所有字段集中在映射表（易维护）
3. 支持复杂的自定义逻辑（灵活）
4. 类型安全（编译期检查）
5. 无需外部辅助函数（自包含）

---

## 示例对比

### 原始方式（硬编码）

```go
// ❌ 172 行
appReducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
    fci := i.(intent.FieldChangeIntent)
    switch fci.Field {
    case "Username":
        s.Username = fci.Value
    case "Email":
        s.Email = fci.Value
    case "Age":
        var ageVal int
        if _, err := fmt.Sscanf(fci.Value, "%d", &ageVal); err == nil {
            s.Age = ageVal
        }
    case "AcceptTerms":
        s.AcceptTerms = fci.Value == "true"
    case "Subscribe":
        s.Subscribe = fci.Value == "true"
    }
    return s
})
```

### 优化方式（FieldMap）

```go
// ✅ 156 行（-9%）
import "strconv"

var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
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
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        return s
    })
```

---

## 性能说明

| 方案 | FieldChangeIntent 处理器数量 |
|------|--------------------------|
| **FieldMap** | 1 个（最佳） |
| **类型化绑定** | N 个（N = 字段数） |
| **switch-case** | 1 个 |
| **自定义映射表** | 1 个 |

**推荐**: FieldMap（单一处理器，性能最佳）

---

## API 参考

### FieldMap

```go
func (fb *FieldBinder[T]) BindFieldMap(fieldMap FieldMap[T]) *FieldBinder[T]

// 类型定义
type FieldMap[T any] map[string]func(T, string) T
```

**示例**:
```go
fieldMap := map[string]func(AppState, string) AppState{
    "username": func(s AppState, val string) AppState {
        s.Username = val
        return s
    },
}

reducer.BindField(builder).BindFieldMap(fieldMap)
```

### 类型化绑定

```go
// 字符串字段
func (fb *FieldBinder[T]) BindStringField(fieldName string, update func(T, string) T) *FieldBinder[T]

// 整型字段（自动转换）
func (fb *FieldBinder[T]) BindIntField(fieldName string, update func(T, int) T) *FieldBinder[T]

// 布尔字段（自动转换）
func (fb *FieldBinder[T]) BindBoolField(fieldName string, update func(T, bool) T) *FieldBinder[T]

// 通用字段
func (fb *FieldBinder[T]) BindFieldGeneric(fieldName string, update func(T, string) T) *FieldBinder[T]
```

---

## 总结

| 问题 | 解决方案 |
|------|---------|
| **硬编码 switch-case** | ✅ 使用 `FieldMap` 替代 |
| **类型转换分散** | ✅ 字段映射表内联处理 |
| **代码重复** | ✅ 每个字段独立函数 |
| **性能问题** | ✅ 单一 FieldChangeIntent 处理器 |
| **类型安全** | ✅ 泛型支持，编译期检查 |

**推荐**:
1. **FieldMap** - 通用场景，性能最佳 ✅
2. **类型化绑定** - 复杂类型转换场景 ⚠️

---

**文档创建**: 2026-03-05
**状态**: 完成 ✅
