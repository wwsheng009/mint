package context

import (
	"sync"
)

// ContextKey Context 键类型（类型安全）
type ContextKey string

// FiberContext Fiber 级别的 Context 存储
type FiberContext struct {
	mu     sync.RWMutex
	values map[ContextKey]any
	parent *FiberContext // 支持嵌套
}

// NewContext 创建新的 Context
// 如果 parent 为 nil，创建根 Context
func NewContext(parent *FiberContext) *FiberContext {
	return &FiberContext{
		values: make(map[ContextKey]any),
		parent: parent,
	}
}

// Provide 提供 Context 值
func (c *FiberContext) Provide(key ContextKey, value any) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[key] = value
}

// UseContext 从 Context 中获取值
// 返回第一个匹配的值（当前或父级）
func (c *FiberContext) UseContext(key ContextKey) any {
	if c == nil {
		return nil
	}

	// 当前层级查找
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()

	if ok {
		return v
	}

	// 向上查找
	if c.parent != nil {
		return c.parent.UseContext(key)
	}

	return nil
}

// UseContextValue 类型安全的 Context 访问（Go 1.18+ 泛型）
func UseContextValue[T any](c *FiberContext, key ContextKey) (T, bool) {
	v := c.UseContext(key)
	if v == nil {
		var zero T
		return zero, false
	}

	t, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}

	return t, true
}

// HasContext 检查当前层级是否有指定的 Context
func (c *FiberContext) HasContext(key ContextKey) bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	_, ok := c.values[key]
	c.mu.RUnlock()

	return ok
}

// HasContextInHierarchy 检查当前层级或父级是否有指定的 Context
func (c *FiberContext) HasContextInHierarchy(key ContextKey) bool {
	if c == nil {
		return false
	}

	if c.HasContext(key) {
		return true
	}

	if c.parent != nil {
		return c.parent.HasContextInHierarchy(key)
	}

	return false
}
