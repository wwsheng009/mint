package ui

import "github.com/wwsheng009/mint/framework"

// AIConfig re-exports framework.AIConfig for ui.Run / ui.RunApp options.
type AIConfig = framework.AIConfig

// MCPConfig re-exports framework.MCPConfig for ui.Run / ui.RunApp options.
type MCPConfig = framework.MCPConfig

// WithAI enables the embedded AI service on the app host.
func WithAI(cfg AIConfig) Option {
	return func(o *Options) {
		cfg.Enabled = true
		o.AIConfig = &cfg
	}
}

// WithMCP enables MCP transport on top of the embedded AI service.
func WithMCP(cfg MCPConfig) Option {
	return func(o *Options) {
		if o.AIConfig == nil {
			o.AIConfig = &framework.AIConfig{}
		}
		o.AIConfig.Enabled = true
		o.AIConfig.MCP = cfg
		o.AIConfig.MCP.Enabled = true
	}
}
