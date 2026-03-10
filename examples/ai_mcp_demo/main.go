package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

var (
	setUsernameState   func(string)
	setEmailState      func(string)
	setThemeIndexState func(interface{})
	setSubscribedState func(bool)
)

func registerIntentHandlers() {
	ui.RegisterIntent(func(_ *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
		switch i.Field {
		case "username":
			if setUsernameState != nil {
				setUsernameState(i.Value)
			}
		case "email":
			if setEmailState != nil {
				setEmailState(i.Value)
			}
		case "theme":
			if setThemeIndexState != nil {
				if idx, err := strconv.Atoi(i.Value); err == nil {
					setThemeIndexState(idx)
				}
			}
		case "subscribe":
			if setSubscribedState != nil {
				setSubscribedState(i.Value == "true")
			}
		}
		return intent.HandledResult()
	})
}

func App() ui.VNode {
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	themeIndex, setThemeIndex, _ := ui.UseStateInt(0)
	subscribed, setSubscribed := ui.UseStateBool(false)
	setUsernameState = setUsername
	setEmailState = setEmail
	setThemeIndexState = setThemeIndex
	setSubscribedState = setSubscribed

	return ui.VStack(
		ui.NewTextBuilder("Mint AI MCP Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Use MCP to inspect and control this UI.").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Username:").
			FgColor("blue").
			Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				Key("username").
				Value(username).
				Placeholder("Type username").
				ForField(intent.BindField("username")).
				Width(32).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Email:").
			FgColor("blue").
			Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				Key("email").
				Value(email).
				Placeholder("Type email").
				ForField(intent.BindField("email")).
				Width(32).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Theme:").
			FgColor("blue").
			Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewSelectBuilder().
				Key("theme").
				ForField(intent.BindField("theme")).
				AddOption("dark", "Dark").
				AddOption("light", "Light").
				AddOption("dracula", "Dracula").
				Selected(themeIndex).
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("  "),
			ui.NewCheckboxBuilder().
				Key("subscribe").
				Label("Subscribe to updates").
				Checked(subscribed).
				ForField(intent.BindField("subscribe")).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Tip: try mint.inspect / mint.set_value / mint.select").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("Quit: press q").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	mcpConfig := ui.MCPConfig{
		Transport:   envString("MINT_MCP_TRANSPORT", "http"),
		Host:        envString("MINT_MCP_HOST", ""),
		Port:        envInt("MINT_MCP_PORT", 0),
		AuthToken:   os.Getenv("MINT_AI_TOKEN"),
		ExposeTrees: envBoolPtr("MINT_MCP_EXPOSE_TREES"),
		ExposeWrite: envBoolPtr("MINT_MCP_EXPOSE_WRITE"),
	}

	err := ui.Run(App,
		ui.WithWidth(68),
		ui.WithHeight(22),
		ui.WithTitle("AI MCP Demo"),
		ui.WithNoAlternateScreen(),
		ui.WithMCP(mcpConfig),
		ui.WithInit(registerIntentHandlers),
		ui.WithPluginSetup(func(app *framework.App) {
			selectcomp.Install(app)
			go writeEndpointsOnce(app, endpointFilePath())
		}),
	)
	if err != nil {
		panic(err)
	}
}

func endpointFilePath() string {
	dir := "tmp"
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "ai_mcp_demo_endpoints.txt")
}

func writeEndpointsOnce(app *framework.App, path string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		status := app.AIStatus()
		if status.Running && status.HTTPEndpoint != "" {
			payload := fmt.Sprintf("MCP=%s\nHTTP=%s\n", status.MCPEndpoint, status.HTTPEndpoint)
			_ = os.WriteFile(path, []byte(payload), 0o644)
			return
		}
	}
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envBoolPtr(key string) *bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &value
}
