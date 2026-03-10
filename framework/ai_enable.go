package framework

import (
	"errors"
	"strings"
	"time"

	aiservice "github.com/wwsheng009/mint/internal/ai/service"
)

// MCPConfig configures the embedded AI transport surface.
// The actual MCP server is not implemented yet; this config establishes the host API.
type MCPConfig struct {
	Enabled     bool
	Transport   string
	Host        string
	Port        int
	ReadOnly    bool
	AuthToken   string
	ExposeTrees bool
	ExposeWrite bool
}

// AIConfig configures the embedded AI service for an app instance.
type AIConfig struct {
	Enabled   bool
	AutoStart bool
	ReadOnly  bool
	MCP       MCPConfig
}

// AIStatus reports lifecycle information for the embedded AI service.
type AIStatus struct {
	Enabled      bool
	Running      bool
	ReadOnly     bool
	StartedAt    time.Time
	StoppedAt    time.Time
	LastRenderAt time.Time
	RenderSeq    uint64
	MCPEnabled   bool
	MCPEndpoint  string
	HTTPEndpoint string
}

// EnableAI attaches the embedded AI service to the app.
// Calling EnableAI multiple times is currently not supported.
func (a *App) EnableAI(cfg AIConfig) error {
	if a == nil {
		return errors.New("nil app")
	}
	if a.aiService != nil {
		return errors.New("AI service already enabled")
	}

	normalized := normalizeAIConfig(cfg)
	a.aiService = aiservice.New(a, toInternalAIConfig(normalized))

	if a.state == StateRunning && normalized.AutoStart {
		return a.aiService.Start()
	}
	return nil
}

// AIStatus returns the current embedded AI service state.
func (a *App) AIStatus() AIStatus {
	if a == nil || a.aiService == nil {
		return AIStatus{}
	}

	status := a.aiService.Status()
	return AIStatus{
		Enabled:      status.Enabled,
		Running:      status.Running,
		ReadOnly:     status.ReadOnly,
		StartedAt:    status.StartedAt,
		StoppedAt:    status.StoppedAt,
		LastRenderAt: status.LastRenderAt,
		RenderSeq:    status.RenderSeq,
		MCPEnabled:   status.MCPEnabled,
		MCPEndpoint:  status.MCPEndpoint,
		HTTPEndpoint: status.HTTPEndpoint,
	}
}

func normalizeAIConfig(cfg AIConfig) AIConfig {
	cfg.Enabled = true
	if !cfg.AutoStart {
		cfg.AutoStart = true
	}
	if cfg.MCP.Enabled {
		if strings.TrimSpace(cfg.MCP.Transport) == "" {
			cfg.MCP.Transport = "http"
		}
		if strings.TrimSpace(cfg.MCP.Host) == "" {
			cfg.MCP.Host = "127.0.0.1"
		}
	}
	return cfg
}

func toInternalAIConfig(cfg AIConfig) aiservice.Config {
	return aiservice.Config{
		Enabled:   cfg.Enabled,
		AutoStart: cfg.AutoStart,
		ReadOnly:  cfg.ReadOnly,
		MCP: aiservice.MCPConfig{
			Enabled:     cfg.MCP.Enabled,
			Transport:   cfg.MCP.Transport,
			Host:        cfg.MCP.Host,
			Port:        cfg.MCP.Port,
			ReadOnly:    cfg.MCP.ReadOnly,
			AuthToken:   cfg.MCP.AuthToken,
			ExposeTrees: cfg.MCP.ExposeTrees,
			ExposeWrite: cfg.MCP.ExposeWrite,
		},
	}
}
