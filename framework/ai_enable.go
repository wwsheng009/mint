package framework

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	aiservice "github.com/wwsheng009/mint/internal/ai/service"
)

// MCPConfig configures the embedded AI transport surface.
// The MCP server is implemented in internal/ai/mcp and can be enabled per-app.
type MCPConfig struct {
	Enabled     bool
	Transport   string
	Host        string
	Port        int
	ReadOnly    bool
	AuthToken   string
	ExposeTrees *bool
	ExposeWrite *bool
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
	AuthToken    string
}

// EnableAI attaches the embedded AI service to the app.
// Calling EnableAI multiple times is currently not supported.
func (a *App) EnableAI(cfg AIConfig) error {
	if a == nil {
		return errors.New("nil app")
	}

	normalized := normalizeAIConfig(cfg)
	service := aiservice.New(a, toInternalAIConfig(normalized))

	a.aiMu.Lock()
	if a.aiService != nil {
		a.aiMu.Unlock()
		return errors.New("AI service already enabled")
	}
	a.aiService = service
	a.aiMu.Unlock()

	if a.IsRunning() && normalized.AutoStart {
		return service.Start()
	}
	return nil
}

// AIStatus returns the current embedded AI service state.
func (a *App) AIStatus() AIStatus {
	service := a.getAIService()
	if service == nil {
		return AIStatus{}
	}

	status := service.Status()
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
		AuthToken:    status.AuthToken,
	}
}

func (a *App) getAIService() *aiservice.Service {
	if a == nil {
		return nil
	}
	a.aiMu.RLock()
	defer a.aiMu.RUnlock()
	return a.aiService
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
		if cfg.MCP.Transport == "http" && strings.TrimSpace(cfg.MCP.Host) == "" {
			cfg.MCP.Host = "127.0.0.1"
		}
		if strings.TrimSpace(cfg.MCP.AuthToken) == "" {
			cfg.MCP.AuthToken = generateRandomToken()
		}
	}
	return cfg
}

func toInternalAIConfig(cfg AIConfig) aiservice.Config {
	exposeTrees := boolValue(cfg.MCP.ExposeTrees, true)
	exposeWrite := boolValue(cfg.MCP.ExposeWrite, !cfg.ReadOnly && !cfg.MCP.ReadOnly)
	if !cfg.MCP.Enabled {
		exposeTrees = false
		exposeWrite = false
	}
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
			ExposeTrees: exposeTrees,
			ExposeWrite: exposeWrite,
		},
	}
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

// generateRandomToken creates a cryptographically random 32-byte hex token.
func generateRandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
