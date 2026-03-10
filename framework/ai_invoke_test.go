package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	irender "github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/state"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/list"
	"github.com/wwsheng009/mint/ui/components/table"
	"github.com/wwsheng009/mint/ui/components/treeview"
)

func TestAppInvokeInlineWhenNotRunning(t *testing.T) {
	app := NewApp()

	got, err := app.Invoke(context.Background(), func() (any, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if got != "ok" {
		t.Fatalf("Invoke() = %v, want ok", got)
	}
}

func TestAppEnableAIBeforeRun(t *testing.T) {
	app := NewApp()

	if err := app.EnableAI(AIConfig{}); err != nil {
		t.Fatalf("EnableAI() error = %v", err)
	}

	status := app.AIStatus()
	if !status.Enabled {
		t.Fatal("AIStatus.Enabled = false, want true")
	}
	if status.Running {
		t.Fatal("AIStatus.Running = true before app start, want false")
	}
}

func TestAppInvokeThroughMainLoop(t *testing.T) {
	t.Setenv("MINT_ASYNC_RENDER", "false")
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	rawCh := make(chan platform.RawInput)
	app := NewAppWithSource(frameworkevent.NewChannelEventSource(rawCh))

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()

	waitForAppState(t, app, StateRunning)

	if err := app.EnableAI(AIConfig{}); err != nil {
		t.Fatalf("EnableAI() while running error = %v", err)
	}
	status := app.AIStatus()
	if !status.Running {
		t.Fatal("AIStatus.Running = false after enabling on running app, want true")
	}

	got, err := app.Invoke(context.Background(), func() (any, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("Invoke() = %v, want 42", got)
	}

	app.Quit()

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not stop after Quit()")
	}
}

func TestAIControllerInspectAfterRender(t *testing.T) {
	t.Setenv("MINT_ASYNC_RENDER", "false")
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	rawCh := make(chan platform.RawInput)
	app := NewAppWithSource(frameworkevent.NewChannelEventSource(rawCh))

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return input.NewBuilder().
			Key("username").
			Value("alice").
			Placeholder("name").
			Build()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	if err := app.EnableAI(AIConfig{}); err != nil {
		t.Fatalf("EnableAI() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()
	defer func() {
		app.Quit()
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
			t.Fatal("Run() did not stop")
		}
	}()

	waitForAppState(t, app, StateRunning)
	waitForRenderSeq(t, app, 1)

	ctrl := app.AIController()
	if ctrl == nil {
		t.Fatal("AIController() = nil")
	}

	snap, err := ctrl.Inspect()
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if snap == nil {
		t.Fatal("Inspect() returned nil snapshot")
	}

	comp, ok := snap.Components["username"]
	if !ok {
		t.Fatalf("snapshot missing component username: %+v", snap.Components)
	}
	if got, _ := comp.State["value"].(string); got != "alice" {
		t.Fatalf("component state value = %q, want alice", got)
	}
	if got, _ := comp.Props["placeholder"].(string); got != "name" {
		t.Fatalf("component props placeholder = %q, want name", got)
	}
	if !comp.Visible {
		t.Fatal("component should be visible after render")
	}
}

func TestAIControllerFindQuerySetValueAndTree(t *testing.T) {
	t.Setenv("MINT_ASYNC_RENDER", "false")
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	rawCh := make(chan platform.RawInput)
	app := NewAppWithSource(frameworkevent.NewChannelEventSource(rawCh))

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			input.NewBuilder().Key("username").Placeholder("name").Build(),
			input.NewBuilder().Key("city").Value("shanghai").Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	if err := app.EnableAI(AIConfig{}); err != nil {
		t.Fatalf("EnableAI() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()
	defer stopAppAndWait(t, app, runDone)

	waitForAppState(t, app, StateRunning)
	waitForRenderSeq(t, app, 1)

	ctrl := app.AIController()
	results, err := ctrl.Find("#username")
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(results) != 1 || results[0].ID != "username" {
		t.Fatalf("Find() = %+v, want username", results)
	}

	query, err := ctrl.Query(AIStateQuery{ComponentID: "username", StateKey: "placeholder"})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if got, _ := query["placeholder"].(string); got != "name" {
		t.Fatalf("Query placeholder = %q, want name", got)
	}

	prevSeq := app.AIStatus().RenderSeq
	if err := ctrl.SetValue("username", "bob"); err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}
	waitForRenderSeq(t, app, prevSeq+1)

	value, err := ctrl.GetValue("username")
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if got, _ := value.(string); got != "bob" {
		t.Fatalf("GetValue() = %q, want bob", got)
	}

	if err := ctrl.WaitUntil(func(s *state.Snapshot) bool {
		comp, ok := s.GetComponent("username")
		return ok && comp.State["value"] == "bob"
	}, time.Second); err != nil {
		t.Fatalf("WaitUntil() error = %v", err)
	}
	if _, err := ctrl.WaitFor(AIWaitCondition{
		ComponentID: "username",
		Key:         "value",
		Equals:      "bob",
	}, time.Second); err != nil {
		t.Fatalf("WaitFor() error = %v", err)
	}
	if _, err := ctrl.WaitFor(AIWaitCondition{
		All: []AIWaitCondition{
			{ComponentID: "username", Key: "value", Equals: "bob"},
			{ComponentID: "username", Visible: boolPtr(true)},
		},
		Not: &AIWaitCondition{ComponentID: "username", Disabled: boolPtr(true)},
	}, time.Second); err != nil {
		t.Fatalf("WaitFor(all/not) error = %v", err)
	}

	tree, err := ctrl.GetTree("fiber")
	if err != nil {
		t.Fatalf("GetTree() error = %v", err)
	}
	if tree == nil {
		t.Fatal("GetTree() returned nil")
	}

	node, err := ctrl.GetNode("username")
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}
	if node == nil {
		t.Fatal("GetNode() returned nil")
	}

	prevSeq = app.AIStatus().RenderSeq
	if err := ctrl.Dispatch("username", "input_text", "zz"); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	waitForRenderSeq(t, app, prevSeq+1)
	value, err = ctrl.GetValue("username")
	if err != nil {
		t.Fatalf("GetValue() after dispatch error = %v", err)
	}
	if got, _ := value.(string); got != "bobzz" {
		t.Fatalf("GetValue() after dispatch = %q, want bobzz", got)
	}
}

func TestAIHTTPInspectAndSetValue(t *testing.T) {
	t.Setenv("MINT_ASYNC_RENDER", "false")
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	rawCh := make(chan platform.RawInput)
	app := NewAppWithSource(frameworkevent.NewChannelEventSource(rawCh))

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return input.NewBuilder().Key("username").Build()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	if err := app.EnableAI(AIConfig{
		MCP: MCPConfig{
			Enabled:   true,
			Transport: "http",
			Host:      "127.0.0.1",
			Port:      0,
		},
	}); err != nil {
		t.Fatalf("EnableAI() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()
	defer stopAppAndWait(t, app, runDone)

	waitForAppState(t, app, StateRunning)
	waitForRenderSeq(t, app, 1)

	endpoint := app.AIStatus().HTTPEndpoint
	if endpoint == "" {
		t.Fatal("HTTPEndpoint is empty")
	}

	resp, err := http.Get(endpoint + "/ai/inspect")
	if err != nil {
		t.Fatalf("GET /ai/inspect error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /ai/inspect status = %d", resp.StatusCode)
	}

	body := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode inspect response error = %v", err)
	}
	if ok, _ := body["ok"].(bool); !ok {
		t.Fatalf("inspect response not ok: %+v", body)
	}

	prevSeq := app.AIStatus().RenderSeq
	payload, _ := json.Marshal(map[string]interface{}{
		"locator": "username",
		"value":   "carol",
	})
	resp, err = http.Post(endpoint+"/ai/value/set", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /ai/value/set error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /ai/value/set status = %d", resp.StatusCode)
	}
	waitForRenderSeq(t, app, prevSeq+1)

	value, err := app.AIController().GetValue("username")
	if err != nil {
		t.Fatalf("GetValue() after HTTP set error = %v", err)
	}
	if got, _ := value.(string); got != "carol" {
		t.Fatalf("value after HTTP set = %q, want carol", got)
	}
}

func TestAIControllerListTableTreeViewSelection(t *testing.T) {
	t.Setenv("MINT_ASYNC_RENDER", "false")
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	rawCh := make(chan platform.RawInput)
	app := NewAppWithSource(frameworkevent.NewChannelEventSource(rawCh))

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			list.NewBuilder().
				Key("cities").
				Rows([]string{"beijing", "shanghai", "shenzhen"}).
				Build(),
			table.NewBuilder().
				Key("users").
				Columns([]table.TableColumn{{Title: "name"}, {Title: "role"}}).
				Rows([][]string{{"alice", "dev"}, {"bob", "ops"}}).
				Build(),
			treeview.NewBuilder().
				Key("files").
				Nodes([]treeview.TreeNode{
					{Indent: 0, Content: "src", Path: "src", NodeType: "folder", NodeID: 1},
					{Indent: 4, Content: "main.go", Path: "src/main.go", NodeType: "file", NodeID: 2},
				}).
				Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	if err := app.EnableAI(AIConfig{}); err != nil {
		t.Fatalf("EnableAI() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()
	defer stopAppAndWait(t, app, runDone)

	waitForAppState(t, app, StateRunning)
	waitForRenderSeq(t, app, 1)

	ctrl := app.AIController()
	if err := ctrl.Select("cities", map[string]interface{}{"index": 1}); err != nil {
		t.Fatalf("Select(list) error = %v", err)
	}
	waitForRenderSeq(t, app, 2)
	value, err := ctrl.GetValue("cities")
	if err != nil {
		t.Fatalf("GetValue(list) error = %v", err)
	}
	if got, _ := value.(int); got != 1 {
		t.Fatalf("GetValue(list) = %v, want 1", value)
	}
	query, err := ctrl.Query(AIStateQuery{ComponentID: "cities", StateKey: "selectedRow"})
	if err != nil {
		t.Fatalf("Query(list selectedRow) error = %v", err)
	}
	if got, _ := query["selectedRow"].(string); got != "shanghai" {
		t.Fatalf("selectedRow = %q, want shanghai", got)
	}
	if err := ctrl.Select("cities", "shenzhen"); err != nil {
		t.Fatalf("Select(list by text) error = %v", err)
	}
	waitForRenderSeq(t, app, 3)
	query, err = ctrl.Query(AIStateQuery{ComponentID: "cities", StateKey: "selectedRow"})
	if err != nil {
		t.Fatalf("Query(list selectedRow by text) error = %v", err)
	}
	if got, _ := query["selectedRow"].(string); got != "shenzhen" {
		t.Fatalf("selectedRow after text select = %q, want shenzhen", got)
	}

	if err := ctrl.Select("users", "bob"); err != nil {
		t.Fatalf("Select(table by text) error = %v", err)
	}
	waitForRenderSeq(t, app, 4)
	query, err = ctrl.Query(AIStateQuery{ComponentID: "users", StateKey: "selectedRow"})
	if err != nil {
		t.Fatalf("Query(table selectedRow by text) error = %v", err)
	}
	row, ok := query["selectedRow"].([]string)
	if !ok || len(row) != 2 || row[0] != "bob" {
		t.Fatalf("table selectedRow after text select = %#v, want [bob ops]", query["selectedRow"])
	}

	if err := ctrl.Select("users", map[string]interface{}{"index": 1}); err != nil {
		t.Fatalf("Select(table) error = %v", err)
	}
	waitForRenderSeq(t, app, 5)
	query, err = ctrl.Query(AIStateQuery{ComponentID: "users", StateKey: "selectedRow"})
	if err != nil {
		t.Fatalf("Query(table selectedRow) error = %v", err)
	}
	row, ok = query["selectedRow"].([]string)
	if !ok || len(row) != 2 || row[0] != "bob" {
		t.Fatalf("table selectedRow = %#v, want [bob ops]", query["selectedRow"])
	}

	if err := ctrl.Select("files", "src/main.go"); err != nil {
		t.Fatalf("Select(treeview by path) error = %v", err)
	}
	waitForRenderSeq(t, app, 6)
	query, err = ctrl.Query(AIStateQuery{ComponentID: "files", StateKey: "selectedNode"})
	if err != nil {
		t.Fatalf("Query(tree selectedNode) error = %v", err)
	}
	node, ok := query["selectedNode"].(map[string]interface{})
	if !ok {
		t.Fatalf("selectedNode = %#v, want map", query["selectedNode"])
	}
	if got, _ := node["Path"].(string); got != "src/main.go" {
		t.Fatalf("selectedNode.Path = %q, want src/main.go", got)
	}

	if _, err := ctrl.WaitFor(AIWaitCondition{
		Any: []AIWaitCondition{
			{ComponentID: "files", Key: "selectedIndex", Equals: 1},
			{ComponentID: "files", Key: "selectedNode", Exists: boolPtr(true)},
		},
	}, time.Second); err != nil {
		t.Fatalf("WaitFor(any) error = %v", err)
	}
}

func TestMCPSDKListToolsAndReadResource(t *testing.T) {
	t.Setenv("MINT_ASYNC_RENDER", "false")
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	rawCh := make(chan platform.RawInput)
	app := NewAppWithSource(frameworkevent.NewChannelEventSource(rawCh))

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return input.NewBuilder().Key("username").Value("alice").Build()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	if err := app.EnableAI(AIConfig{
		MCP: MCPConfig{
			Enabled:   true,
			Transport: "http",
			Host:      "127.0.0.1",
			Port:      0,
		},
	}); err != nil {
		t.Fatalf("EnableAI() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()
	defer stopAppAndWait(t, app, runDone)

	waitForAppState(t, app, StateRunning)
	waitForRenderSeq(t, app, 1)

	mcpEndpoint := app.AIStatus().MCPEndpoint
	if mcpEndpoint == "" {
		t.Fatal("MCPEndpoint is empty")
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "mint-test-client", Version: "0.1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{Endpoint: mcpEndpoint}, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(tools.Tools) == 0 {
		t.Fatal("ListTools() returned no tools")
	}

	foundInspect := false
	for _, tool := range tools.Tools {
		if tool.Name == "mint.inspect" {
			foundInspect = true
			break
		}
	}
	if !foundInspect {
		t.Fatalf("mint.inspect not found in tools: %+v", tools.Tools)
	}

	templates, err := session.ListResourceTemplates(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListResourceTemplates() error = %v", err)
	}
	var componentTemplate *mcpsdk.ResourceTemplate
	var nodeTemplate *mcpsdk.ResourceTemplate
	for _, template := range templates.ResourceTemplates {
		switch template.URITemplate {
		case "mint://component/{id}":
			componentTemplate = template
		case "mint://node/{id}":
			nodeTemplate = template
		}
	}
	if componentTemplate == nil {
		t.Fatalf("mint://component/{id} template not found: %+v", templates.ResourceTemplates)
	}
	if nodeTemplate == nil {
		t.Fatalf("mint://node/{id} template not found: %+v", templates.ResourceTemplates)
	}
	if got, _ := componentTemplate.Meta["mint/outputSchemaURI"].(string); got != "mint://schema/component" {
		t.Fatalf("component template schema URI = %q, want mint://schema/component", got)
	}
	if got, _ := nodeTemplate.Meta["mint/outputSchemaURI"].(string); got != "mint://schema/node" {
		t.Fatalf("node template schema URI = %q, want mint://schema/node", got)
	}
	if _, ok := componentTemplate.Meta["mint/outputSchema"].(map[string]any); !ok {
		t.Fatalf("component template missing inline output schema: %+v", componentTemplate.Meta)
	}

	resources, err := session.ListResources(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	foundComponentSchema := false
	foundNodeSchema := false
	for _, resource := range resources.Resources {
		switch resource.URI {
		case "mint://schema/component":
			foundComponentSchema = true
		case "mint://schema/node":
			foundNodeSchema = true
		}
	}
	if !foundComponentSchema || !foundNodeSchema {
		t.Fatalf("schema resources missing: %+v", resources.Resources)
	}

	res, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{URI: "mint://frame/current"})
	if err != nil {
		t.Fatalf("ReadResource() error = %v", err)
	}
	if len(res.Contents) == 0 || res.Contents[0].Text == "" {
		t.Fatalf("ReadResource() returned empty contents: %+v", res.Contents)
	}

	componentRes, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{URI: "mint://component/username"})
	if err != nil {
		t.Fatalf("ReadResource(component) error = %v", err)
	}
	if len(componentRes.Contents) == 0 || componentRes.Contents[0].Text == "" {
		t.Fatalf("ReadResource(component) returned empty contents: %+v", componentRes.Contents)
	}
	if got, _ := componentRes.Contents[0].Meta["mint/outputSchemaURI"].(string); got != "mint://schema/component" {
		t.Fatalf("component resource schema URI = %q, want mint://schema/component", got)
	}
	var componentPayload map[string]any
	if err := json.Unmarshal([]byte(componentRes.Contents[0].Text), &componentPayload); err != nil {
		t.Fatalf("decode component resource error = %v", err)
	}
	if got, _ := componentPayload["id"].(string); got != "username" {
		t.Fatalf("component resource id = %q, want username", got)
	}

	componentSchemaRes, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{URI: "mint://schema/component"})
	if err != nil {
		t.Fatalf("ReadResource(component schema) error = %v", err)
	}
	if len(componentSchemaRes.Contents) == 0 || componentSchemaRes.Contents[0].Text == "" {
		t.Fatalf("ReadResource(component schema) returned empty contents: %+v", componentSchemaRes.Contents)
	}
	var componentSchema map[string]any
	if err := json.Unmarshal([]byte(componentSchemaRes.Contents[0].Text), &componentSchema); err != nil {
		t.Fatalf("decode component schema error = %v", err)
	}
	if got, _ := componentSchema["title"].(string); got != "Mint Component Snapshot" {
		t.Fatalf("component schema title = %q, want Mint Component Snapshot", got)
	}

	nodeSchemaRes, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{URI: "mint://schema/node"})
	if err != nil {
		t.Fatalf("ReadResource(node schema) error = %v", err)
	}
	if len(nodeSchemaRes.Contents) == 0 || nodeSchemaRes.Contents[0].Text == "" {
		t.Fatalf("ReadResource(node schema) returned empty contents: %+v", nodeSchemaRes.Contents)
	}
	var nodeSchema map[string]any
	if err := json.Unmarshal([]byte(nodeSchemaRes.Contents[0].Text), &nodeSchema); err != nil {
		t.Fatalf("decode node schema error = %v", err)
	}
	if got, _ := nodeSchema["title"].(string); got != "Mint Node Bundle" {
		t.Fatalf("node schema title = %q, want Mint Node Bundle", got)
	}

	toolRes, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "mint.inspect",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool(mint.inspect) error = %v", err)
	}
	if toolRes.IsError || len(toolRes.Content) == 0 {
		t.Fatalf("mint.inspect tool returned error content: %+v", toolRes)
	}

	toolRes, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "mint.set_value",
		Arguments: map[string]any{
			"locator": "username",
			"value":   "dave",
		},
	})
	if err != nil {
		t.Fatalf("CallTool(mint.set_value) error = %v", err)
	}
	if toolRes.IsError {
		t.Fatalf("mint.set_value returned tool error: %+v", toolRes)
	}
	waitForRenderSeq(t, app, 2)
	value, err := app.AIController().GetValue("username")
	if err != nil {
		t.Fatalf("GetValue() after MCP tool set error = %v", err)
	}
	if got, _ := value.(string); got != "dave" {
		t.Fatalf("value after MCP tool set = %q, want dave", got)
	}

	toolRes, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "mint.wait_until",
		Arguments: map[string]any{
			"component_id": "username",
			"key":          "value",
			"equals":       "dave",
			"timeout_ms":   1000,
		},
	})
	if err != nil {
		t.Fatalf("CallTool(mint.wait_until) error = %v", err)
	}
	if toolRes.IsError {
		t.Fatalf("mint.wait_until returned tool error: %+v", toolRes)
	}

	toolRes, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "mint.wait_until",
		Arguments: map[string]any{
			"any": []map[string]any{
				{"component_id": "username", "key": "value", "equals": "dave"},
				{"component_id": "username", "visible": true},
			},
			"not": map[string]any{
				"component_id": "username",
				"disabled":     true,
			},
			"timeout_ms": 1000,
		},
	})
	if err != nil {
		t.Fatalf("CallTool(mint.wait_until composite) error = %v", err)
	}
	if toolRes.IsError {
		t.Fatalf("mint.wait_until composite returned tool error: %+v", toolRes)
	}

	toolRes, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "mint.batch",
		Arguments: map[string]any{
			"operations": []map[string]any{
				{"operation": "set_value", "locator": "username", "value": "zoe"},
				{"operation": "dispatch", "locator": "username", "action_type": "input_text", "payload": "!"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool(mint.batch) error = %v", err)
	}
	if toolRes.IsError {
		t.Fatalf("mint.batch returned tool error: %+v", toolRes)
	}
	waitForRenderSeq(t, app, 3)
	value, err = app.AIController().GetValue("username")
	if err != nil {
		t.Fatalf("GetValue() after batch error = %v", err)
	}
	if got, _ := value.(string); got != "zoe!" {
		t.Fatalf("value after batch = %q, want zoe!", got)
	}

	toolRes, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "mint.batch",
		Arguments: map[string]any{
			"stop_on_error": false,
			"operations": []map[string]any{
				{"operation": "set_value", "locator": "missing", "value": "bad"},
				{"operation": "set_value", "locator": "username", "value": "ivy"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool(mint.batch continue) error = %v", err)
	}
	if toolRes.IsError {
		t.Fatalf("mint.batch continue returned tool error: %+v", toolRes)
	}
	waitForRenderSeq(t, app, 4)
	value, err = app.AIController().GetValue("username")
	if err != nil {
		t.Fatalf("GetValue() after continue batch error = %v", err)
	}
	if got, _ := value.(string); got != "ivy" {
		t.Fatalf("value after continue batch = %q, want ivy", got)
	}
}

func waitForAppState(t *testing.T, app *App, want AppState) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if app.GetState() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("app state = %v, want %v", app.GetState(), want)
}

func waitForRenderSeq(t *testing.T, app *App, min uint64) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if status := app.AIStatus(); status.RenderSeq >= min {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("AI render seq = %d, want >= %d", app.AIStatus().RenderSeq, min)
}

func stopAppAndWait(t *testing.T, app *App, runDone <-chan error) {
	t.Helper()
	app.Quit()
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not stop")
	}
}

func boolPtr(v bool) *bool {
	return &v
}
