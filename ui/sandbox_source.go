// ui/sandbox_source.go - Sandbox 到 Pump 的桥接适配器
package ui

import (
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox/mock"
)

// SandboxEventSource 将 MockSandbox 适配为 Pump 的 EventSource
//
// 这个适配器连接了两个独立设计的系统：
// - sandbox/mock: 使用 EventHandler 回调模式
// - framework/event/Pump: 使用通道读取模式
type SandboxEventSource struct {
	sandbox    *mock.MockSandbox
	rawInputs  chan platform.RawInput
}

// NewSandboxEventSource 创建 Sandbox 事件源
func NewSandboxEventSource(sb *mock.MockSandbox) *SandboxEventSource {
	return &SandboxEventSource{
		sandbox:   sb,
		rawInputs: make(chan platform.RawInput, 100),
	}
}

// Start 启动事件源
// 设置 EventHandler 将注入的事件转发到通道
func (s *SandboxEventSource) Start() (<-chan platform.RawInput, error) {
	// 设置事件处理器：当 Sandbox.Inject() 被调用时，事件会发送到通道
	s.sandbox.SetEventHandler(func(raw platform.RawInput) error {
		select {
		case s.rawInputs <- raw:
		default:
			// 通道满时非阻塞丢弃（测试场景下通常不会发生）
		}
		return nil
	})

	return s.rawInputs, nil
}

// Stop 停止事件源
func (s *SandboxEventSource) Stop() error {
	close(s.rawInputs)
	return nil
}

// Ensure SandboxEventSource implements event.EventSource
var _ event.EventSource = (*SandboxEventSource)(nil)
