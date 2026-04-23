package ui

import (
	fcontext "github.com/wwsheng009/mint/runtime/context"
)

// Provider 组件用于提供 Context 值
// 类似 React 的 Context.Provider，让子组件可以通过 UseContext 访问值
type Provider struct {
	*ElementVNode
}

var (
	_ VNode           = (*Provider)(nil)
	_ InstanceFactory = (*Provider)(nil)
)

// NewProvider 创建一个新的 Provider 组件
func NewProvider(key fcontext.ContextKey, value any, child VNode) *Provider {
	provider := &Provider{
		ElementVNode: NewElement("provider"),
	}

	if child != nil {
		provider.SetChildren([]VNode{child})
	}

	// Store context key and value in props
	if provider.props == nil {
		provider.props = make(Props)
	}
	provider.props["contextKey"] = key
	provider.props["contextValue"] = value

	return provider
}

// GetContextKey 获取 Provider 的 Context Key
func (p *Provider) GetContextKey() fcontext.ContextKey {
	if key, ok := p.props["contextKey"].(fcontext.ContextKey); ok {
		return key
	}
	return ""
}

// GetContextValue 获取 Provider 的 Context Value
func (p *Provider) GetContextValue() any {
	return p.props["contextValue"]
}

// CreateInstance implements InstanceFactory
func (p *Provider) CreateInstance() ComponentInstance {
	return &ProviderInstance{}
}

// ProviderInstance Provider 的 Instance 实现
type ProviderInstance struct {
	BaseComponentInstance
}

var (
	_ ComponentInstance = (*ProviderInstance)(nil)
)
