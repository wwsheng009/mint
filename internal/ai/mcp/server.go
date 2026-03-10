package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wwsheng009/mint/runtime/state"
)

type Config struct {
	Host      string
	Port      int
	ReadOnly  bool
	AuthToken string
}

type API struct {
	Inspect      func() (*state.Snapshot, error)
	Find         func(selector string) (interface{}, error)
	Query        func(componentID, componentType, stateKey string, value interface{}) (map[string]interface{}, error)
	GetTree      func(kind string) (interface{}, error)
	GetNode      func(locator string) (interface{}, error)
	GetFormData  func(locator string) (map[string]interface{}, error)
	GetValue     func(locator string) (interface{}, error)
	WaitFor      func(condition map[string]any, timeout time.Duration) (interface{}, error)
	SetValue     func(locator string, value interface{}) error
	SetProp      func(locator, key string, value interface{}) error
	SetFormField func(locator, field string, value interface{}) error
	Select       func(locator string, value interface{}) error
	Click        func(locator string) error
	Input        func(locator, text string) error
	Dispatch     func(locator, actionType string, payload interface{}) error
	Navigate     func(direction string) error
	Batch        func(operations []BatchOperation, stopOnError bool) (*BatchResult, error)
}

type findInput struct {
	Selector string `json:"selector"`
}

type queryInput struct {
	ComponentID   string      `json:"component_id,omitempty"`
	ComponentType string      `json:"component_type,omitempty"`
	StateKey      string      `json:"state_key,omitempty"`
	Value         interface{} `json:"value,omitempty"`
}

type treeInput struct {
	Kind string `json:"kind"`
}

type locatorInput struct {
	Locator string `json:"locator"`
}

type setValueInput struct {
	Locator string      `json:"locator"`
	Value   interface{} `json:"value"`
}

type setPropInput struct {
	Locator string      `json:"locator"`
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
}

type setFormFieldInput struct {
	Locator string      `json:"locator"`
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
}

type selectInput struct {
	Locator string      `json:"locator"`
	Value   interface{} `json:"value"`
}

type inputTextInput struct {
	Locator string `json:"locator"`
	Text    string `json:"text"`
}

type dispatchInput struct {
	Locator    string      `json:"locator"`
	ActionType string      `json:"action_type"`
	Payload    interface{} `json:"payload,omitempty"`
}

type navigateInput struct {
	Direction string `json:"direction"`
}

type BatchOperation struct {
	Operation  string      `json:"operation"`
	Locator    string      `json:"locator,omitempty"`
	Key        string      `json:"key,omitempty"`
	Field      string      `json:"field,omitempty"`
	Value      interface{} `json:"value,omitempty"`
	Text       string      `json:"text,omitempty"`
	ActionType string      `json:"action_type,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
	Direction  string      `json:"direction,omitempty"`
}

type BatchOperationResult struct {
	Index     int    `json:"index"`
	Operation string `json:"operation"`
	Locator   string `json:"locator,omitempty"`
	Direction string `json:"direction,omitempty"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

type BatchResult struct {
	OK             bool                   `json:"ok"`
	Applied        int                    `json:"applied"`
	Total          int                    `json:"total"`
	StopOnError    bool                   `json:"stop_on_error"`
	StoppedOnError bool                   `json:"stopped_on_error"`
	Results        []BatchOperationResult `json:"results"`
}

type batchInput struct {
	Operations  []BatchOperation `json:"operations"`
	StopOnError *bool            `json:"stop_on_error,omitempty"`
}

type waitInput struct {
	Locator          string      `json:"locator,omitempty"`
	Selector         string      `json:"selector,omitempty"`
	ComponentID      string      `json:"component_id,omitempty"`
	ComponentType    string      `json:"component_type,omitempty"`
	Key              string      `json:"key,omitempty"`
	Equals           interface{} `json:"equals,omitempty"`
	Exists           *bool       `json:"exists,omitempty"`
	Visible          *bool       `json:"visible,omitempty"`
	Disabled         *bool       `json:"disabled,omitempty"`
	RenderSeqAtLeast uint64      `json:"render_seq_at_least,omitempty"`
	Any              []waitInput `json:"any,omitempty"`
	All              []waitInput `json:"all,omitempty"`
	Not              *waitInput  `json:"not,omitempty"`
	TimeoutMS        int         `json:"timeout_ms,omitempty"`
}

type Server struct {
	cfg           Config
	api           API
	listener      net.Listener
	server        *http.Server
	baseEndpoint  string
	mcpEndpoint   string
	sdkServer     *mcpsdk.Server
	sdkHandler    http.Handler
	sdkHandlerRaw interface{}
}

func New(cfg Config, api API) *Server {
	return &Server{cfg: cfg, api: api}
}

func (s *Server) Start() error {
	host := s.cfg.Host
	if strings.TrimSpace(host) == "" {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(s.cfg.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if err := s.initSDK(); err != nil {
		_ = ln.Close()
		return err
	}

	mux := http.NewServeMux()
	if s.sdkHandler != nil {
		mux.Handle("/mcp", s.sdkHandler)
	}
	mux.HandleFunc("/ai/inspect", s.handleInspect)
	mux.HandleFunc("/ai/find", s.handleFind)
	mux.HandleFunc("/ai/query", s.handleQuery)
	mux.HandleFunc("/ai/tree", s.handleTree)
	mux.HandleFunc("/ai/node", s.handleNode)
	mux.HandleFunc("/ai/value/get", s.handleGetValue)
	mux.HandleFunc("/ai/value/set", s.handleSetValue)
	mux.HandleFunc("/ai/prop/set", s.handleSetProp)
	mux.HandleFunc("/ai/form/get", s.handleGetFormData)
	mux.HandleFunc("/ai/form/set", s.handleSetFormField)
	mux.HandleFunc("/ai/select", s.handleSelect)
	mux.HandleFunc("/ai/click", s.handleClick)
	mux.HandleFunc("/ai/input", s.handleInput)
	mux.HandleFunc("/ai/navigate", s.handleNavigate)
	mux.HandleFunc("/ai/batch", s.handleBatch)

	s.listener = ln
	s.baseEndpoint = "http://" + ln.Addr().String()
	s.mcpEndpoint = s.baseEndpoint + "/mcp"
	s.server = &http.Server{
		Handler:           s.withMiddleware(mux),
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		_ = s.server.Serve(ln)
	}()
	return nil
}

func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if closer, ok := s.sdkHandlerRaw.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
	s.server = nil
	s.listener = nil
	s.sdkHandler = nil
	s.sdkHandlerRaw = nil
	s.sdkServer = nil
	s.baseEndpoint = ""
	s.mcpEndpoint = ""
	return err
}

func (s *Server) Endpoint() string {
	return s.mcpEndpoint
}

func (s *Server) BaseEndpoint() string {
	return s.baseEndpoint
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AuthToken != "" {
			token := r.Header.Get("Authorization")
			if strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimPrefix(token, "Bearer ")
			}
			if token != s.cfg.AuthToken && r.Header.Get("X-Mint-AI-Token") != s.cfg.AuthToken {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) initSDK() error {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "mint-ai",
		Version: "0.1.0",
	}, &mcpsdk.ServerOptions{
		Instructions: "Mint interactive UI inspection and control server.",
	})

	s.registerTools(server)
	s.registerResources(server)

	handler := mcpsdk.NewStreamableHTTPHandler(func(_ *http.Request) *mcpsdk.Server {
		return server
	}, &mcpsdk.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})

	s.sdkServer = server
	s.sdkHandler = handler
	s.sdkHandlerRaw = handler
	return nil
}

func (s *Server) registerTools(server *mcpsdk.Server) {
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.inspect",
		Description: "Return the latest semantic UI snapshot.",
		InputSchema: schemaObject(),
	}, func(_ context.Context, _ struct{}) (any, error) {
		return s.api.Inspect()
	})
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.find",
		Description: "Find components by selector.",
		InputSchema: schemaObject(schemaReqProp("selector", "string", "component selector")),
	}, func(_ context.Context, in findInput) (any, error) {
		return s.api.Find(in.Selector)
	})
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.query",
		Description: "Query component state and props.",
		InputSchema: schemaObject(
			schemaProp("component_id", "string", "component id"),
			schemaProp("component_type", "string", "component type"),
			schemaProp("state_key", "string", "state or prop key"),
		),
	}, func(_ context.Context, in queryInput) (any, error) {
		return s.api.Query(in.ComponentID, in.ComponentType, in.StateKey, in.Value)
	})
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.get_tree",
		Description: "Get a structural tree snapshot.",
		InputSchema: schemaObject(schemaEnumProp("kind", "tree kind", "vnode", "fiber", "layout", "paintable", "hitmap")),
	}, func(_ context.Context, in treeInput) (any, error) {
		return s.api.GetTree(in.Kind)
	})
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.get_node",
		Description: "Get a node bundle by locator.",
		InputSchema: schemaObject(schemaReqProp("locator", "string", "component id, selector, path, or @nodeid")),
	}, func(_ context.Context, in locatorInput) (any, error) {
		return s.api.GetNode(in.Locator)
	})
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.get_value",
		Description: "Get a component value.",
		InputSchema: schemaObject(schemaReqProp("locator", "string", "component locator")),
	}, func(_ context.Context, in locatorInput) (any, error) {
		value, err := s.api.GetValue(in.Locator)
		if err != nil {
			return nil, err
		}
		return map[string]any{"value": value}, nil
	})
	addTypedTool(server, &mcpsdk.Tool{
		Name:        "mint.get_form_data",
		Description: "Get a form's field data.",
		InputSchema: schemaObject(schemaReqProp("locator", "string", "form locator")),
	}, func(_ context.Context, in locatorInput) (any, error) {
		return s.api.GetFormData(in.Locator)
	})
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "mint.wait_until",
		Description: "Wait for a structured condition to become true.",
		InputSchema: schemaObject(
			schemaProp("locator", "string", "component locator"),
			schemaProp("selector", "string", "component selector"),
			schemaProp("component_id", "string", "component id"),
			schemaProp("component_type", "string", "component type"),
			schemaProp("key", "string", "state or prop key"),
			schemaProp("any", "array", "match any nested conditions"),
			schemaProp("all", "array", "match all nested conditions"),
			schemaProp("not", "object", "negated nested condition"),
			schemaProp("timeout_ms", "number", "timeout in milliseconds"),
		),
	}, func(_ context.Context, req *mcpsdk.CallToolRequest, in waitInput) (*mcpsdk.CallToolResult, map[string]any, error) {
		timeout := 5 * time.Second
		if in.TimeoutMS > 0 {
			timeout = time.Duration(in.TimeoutMS) * time.Millisecond
		}
		raw := map[string]any{}
		if len(req.Params.Arguments) > 0 {
			_ = json.Unmarshal(req.Params.Arguments, &raw)
		}
		if len(raw) == 0 {
			raw = waitInputToMap(in)
		}
		result, err := s.api.WaitFor(raw, timeout)
		if err != nil {
			return jsonToolError(err), nil, nil
		}
		return nil, map[string]any{"result": result}, nil
	})

	if !s.cfg.ReadOnly {
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.set_value",
			Description: "Set a component value.",
			InputSchema: schemaObject(schemaReqProp("locator", "string", "component locator")),
		}, func(_ context.Context, in setValueInput) (any, error) {
			err := s.api.SetValue(in.Locator, in.Value)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.set_prop",
			Description: "Set a component prop override.",
			InputSchema: schemaObject(
				schemaReqProp("locator", "string", "component locator"),
				schemaReqProp("key", "string", "property key"),
			),
		}, func(_ context.Context, in setPropInput) (any, error) {
			err := s.api.SetProp(in.Locator, in.Key, in.Value)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.set_form_field",
			Description: "Set a form field value.",
			InputSchema: schemaObject(
				schemaReqProp("locator", "string", "form locator"),
				schemaReqProp("field", "string", "field name"),
			),
		}, func(_ context.Context, in setFormFieldInput) (any, error) {
			err := s.api.SetFormField(in.Locator, in.Field, in.Value)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.select",
			Description: "Select an option or item.",
			InputSchema: schemaObject(schemaReqProp("locator", "string", "component locator")),
		}, func(_ context.Context, in selectInput) (any, error) {
			err := s.api.Select(in.Locator, in.Value)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.click",
			Description: "Trigger a click action.",
			InputSchema: schemaObject(schemaReqProp("locator", "string", "component locator")),
		}, func(_ context.Context, in locatorInput) (any, error) {
			err := s.api.Click(in.Locator)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.input",
			Description: "Send semantic text input.",
			InputSchema: schemaObject(
				schemaReqProp("locator", "string", "component locator"),
				schemaReqProp("text", "string", "input text"),
			),
		}, func(_ context.Context, in inputTextInput) (any, error) {
			err := s.api.Input(in.Locator, in.Text)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.dispatch",
			Description: "Dispatch an arbitrary runtime action to a component.",
			InputSchema: schemaObject(
				schemaReqProp("locator", "string", "component locator"),
				schemaReqProp("action_type", "string", "runtime action type"),
			),
		}, func(_ context.Context, in dispatchInput) (any, error) {
			err := s.api.Dispatch(in.Locator, in.ActionType, in.Payload)
			return map[string]any{"ok": err == nil}, err
		})
		addTypedTool(server, &mcpsdk.Tool{
			Name:        "mint.navigate",
			Description: "Move focus.",
			InputSchema: schemaObject(schemaEnumProp("direction", "focus direction", "next", "prev", "first", "last", "up", "down", "left", "right")),
		}, func(_ context.Context, in navigateInput) (any, error) {
			err := s.api.Navigate(in.Direction)
			return map[string]any{"ok": err == nil}, err
		})
		mcpsdk.AddTool(server, &mcpsdk.Tool{
			Name:        "mint.batch",
			Description: "Execute multiple Mint write operations sequentially in one tool call. Supports set_value, set_prop, set_form_field, select, click, input, dispatch, and navigate. No rollback is performed.",
			InputSchema: batchToolInputSchema(),
		}, func(_ context.Context, _ *mcpsdk.CallToolRequest, in batchInput) (*mcpsdk.CallToolResult, map[string]any, error) {
			if len(in.Operations) == 0 {
				return jsonToolError(fmt.Errorf("batch requires at least one operation")), nil, nil
			}

			stopOnError := true
			if in.StopOnError != nil {
				stopOnError = *in.StopOnError
			}

			out, err := s.runBatch(in.Operations, stopOnError)
			if err != nil && out == nil {
				return jsonToolError(err), nil, nil
			}
			if err != nil && stopOnError {
				return &mcpsdk.CallToolResult{
					IsError:           true,
					Content:           []mcpsdk.Content{&mcpsdk.TextContent{Text: err.Error()}},
					StructuredContent: out,
				}, nil, nil
			}
			return nil, map[string]any{"result": out}, nil
		})
	}
}

func (s *Server) registerResources(server *mcpsdk.Server) {
	server.AddResource(&mcpsdk.Resource{
		URI:         "mint://frame/current",
		Name:        "current-frame",
		Title:       "Current Frame Snapshot",
		Description: "Latest semantic snapshot for the running Mint app.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		result, err := s.api.Inspect()
		if err != nil {
			return nil, err
		}
		return jsonResource(req.Params.URI, result)
	})

	server.AddResource(&mcpsdk.Resource{
		URI:         "mint://schema/component",
		Name:        "component-schema",
		Title:       "Component Resource Schema",
		Description: "JSON Schema for mint://component/{id} resource responses.",
		MIMEType:    "application/schema+json",
		Meta:        schemaResourceMeta("mint://component/{id}"),
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return jsonSchemaResource(req.Params.URI, componentResourceSchema(), schemaResourceMeta("mint://component/{id}"))
	})

	server.AddResource(&mcpsdk.Resource{
		URI:         "mint://schema/node",
		Name:        "node-schema",
		Title:       "Node Resource Schema",
		Description: "JSON Schema for mint://node/{id} resource responses.",
		MIMEType:    "application/schema+json",
		Meta:        schemaResourceMeta("mint://node/{id}"),
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return jsonSchemaResource(req.Params.URI, nodeBundleResourceSchema(), schemaResourceMeta("mint://node/{id}"))
	})

	server.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		URITemplate: "mint://tree/{kind}",
		Name:        "tree",
		Title:       "Tree Snapshot",
		Description: "Structured vnode/fiber/layout/paintable/hitmap tree snapshot.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		kind := strings.TrimPrefix(req.Params.URI, "mint://tree/")
		result, err := s.api.GetTree(kind)
		if err != nil {
			return nil, err
		}
		return jsonResource(req.Params.URI, result)
	})

	server.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		URITemplate: "mint://component/{id}",
		Name:        "component",
		Title:       "Component Snapshot By ID",
		Description: "Read the semantic snapshot for one component by component ID. Returns a single JSON object with id, type, props, state, rect, visible, and disabled. Output schema: mint://schema/component.",
		MIMEType:    "application/json",
		Meta:        resourceTemplateMeta("component", "mint://schema/component", componentResourceSchema()),
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		id := strings.TrimPrefix(req.Params.URI, "mint://component/")
		components, err := s.api.Find("#" + id)
		if err != nil {
			return nil, err
		}
		component, err := componentResourceResult(id, components)
		if err != nil {
			return nil, err
		}
		return jsonResourceWithMeta(req.Params.URI, component, resourceReadMeta("component", "mint://schema/component", componentResourceSchema()))
	})

	server.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		URITemplate: "mint://node/{id}",
		Name:        "node",
		Title:       "Node Bundle By ID",
		Description: "Read the cross-layer bundle for one numeric node ID. Returns node_id plus any available fiber, layout, paintable, and hit_entries sections. Output schema: mint://schema/node.",
		MIMEType:    "application/json",
		Meta:        resourceTemplateMeta("node", "mint://schema/node", nodeBundleResourceSchema()),
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		id := strings.TrimPrefix(req.Params.URI, "mint://node/")
		result, err := s.api.GetNode("@" + id)
		if err != nil {
			return nil, err
		}
		return jsonResourceWithMeta(req.Params.URI, result, resourceReadMeta("node", "mint://schema/node", nodeBundleResourceSchema()))
	})
}

func addTypedTool[In any](server *mcpsdk.Server, tool *mcpsdk.Tool, handler func(context.Context, In) (any, error)) {
	mcpsdk.AddTool(server, tool, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (*mcpsdk.CallToolResult, map[string]any, error) {
		result, err := handler(ctx, in)
		if err != nil {
			return jsonToolError(err), nil, nil
		}
		return nil, map[string]any{"result": result}, nil
	})
}

func jsonToolError(err error) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		IsError: true,
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: err.Error()}},
	}
}

func jsonResource(uri string, result any) (*mcpsdk.ReadResourceResult, error) {
	return jsonResourceWithMeta(uri, result, nil)
}

func jsonResourceWithMeta(uri string, result any, meta mcpsdk.Meta) (*mcpsdk.ReadResourceResult, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &mcpsdk.ReadResourceResult{
		Meta: meta,
		Contents: []*mcpsdk.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
			Meta:     meta,
		}},
	}, nil
}

func jsonSchemaResource(uri string, schema map[string]any, meta mcpsdk.Meta) (*mcpsdk.ReadResourceResult, error) {
	data, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	return &mcpsdk.ReadResourceResult{
		Meta: meta,
		Contents: []*mcpsdk.ResourceContents{{
			URI:      uri,
			MIMEType: "application/schema+json",
			Text:     string(data),
			Meta:     meta,
		}},
	}, nil
}

func schemaObject(properties ...map[string]any) map[string]any {
	props := map[string]any{}
	required := []string{}
	for _, property := range properties {
		name, _ := property["name"].(string)
		if name == "" {
			continue
		}
		if req, _ := property["required"].(bool); req {
			required = append(required, name)
		}
		copy := map[string]any{}
		for k, v := range property {
			if k == "name" || k == "required" {
				continue
			}
			copy[k] = v
		}
		props[name] = copy
	}
	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func schemaProp(name, typ, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"type":        typ,
		"description": description,
	}
}

func schemaReqProp(name, typ, description string) map[string]any {
	prop := schemaProp(name, typ, description)
	prop["required"] = true
	return prop
}

func schemaEnumProp(name, description string, values ...string) map[string]any {
	enum := make([]any, 0, len(values))
	for _, value := range values {
		enum = append(enum, value)
	}
	return map[string]any{
		"name":        name,
		"type":        "string",
		"description": description,
		"required":    true,
		"enum":        enum,
	}
}

func batchToolInputSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"operations"},
		"properties": map[string]any{
			"operations": map[string]any{
				"type":        "array",
				"description": "Write operations to execute in order.",
				"minItems":    1,
				"items": map[string]any{
					"type":     "object",
					"required": []string{"operation"},
					"properties": map[string]any{
						"operation": map[string]any{
							"type": "string",
							"enum": []string{"set_value", "set_prop", "set_form_field", "select", "click", "input", "dispatch", "navigate"},
						},
						"locator":     map[string]any{"type": "string"},
						"key":         map[string]any{"type": "string"},
						"field":       map[string]any{"type": "string"},
						"value":       map[string]any{},
						"text":        map[string]any{"type": "string"},
						"action_type": map[string]any{"type": "string"},
						"payload":     map[string]any{},
						"direction": map[string]any{
							"type": "string",
							"enum": []string{"next", "prev", "first", "last", "up", "down", "left", "right"},
						},
					},
					"additionalProperties": false,
				},
			},
			"stop_on_error": map[string]any{
				"type":        "boolean",
				"description": "Stop at the first failed operation. Defaults to true.",
			},
		},
		"additionalProperties": false,
	}
}

func resourceTemplateMeta(kind, schemaURI string, schema map[string]any) mcpsdk.Meta {
	return mcpsdk.Meta{
		"mint/resourceKind":    kind,
		"mint/outputSchemaURI": schemaURI,
		"mint/outputSchema":    schema,
	}
}

func resourceReadMeta(kind, schemaURI string, schema map[string]any) mcpsdk.Meta {
	return mcpsdk.Meta{
		"mint/resourceKind":    kind,
		"mint/outputSchemaURI": schemaURI,
		"mint/outputSchema":    schema,
	}
}

func schemaResourceMeta(target string) mcpsdk.Meta {
	return mcpsdk.Meta{
		"mint/schemaFor": target,
	}
}

func componentResourceResult(id string, raw any) (any, error) {
	switch value := raw.(type) {
	case []interface{}:
		if len(value) == 0 {
			return nil, fmt.Errorf("component not found: %s", id)
		}
		return value[0], nil
	default:
		rv := reflectValue(raw)
		if !rv.IsValid() || rv.Len() == 0 {
			return nil, fmt.Errorf("component not found: %s", id)
		}
		return rv.Index(0).Interface(), nil
	}
}

func reflectValue(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return rv
	}
	if rv.Kind() == reflect.Pointer && !rv.IsNil() {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return reflect.Value{}
	}
	return rv
}

func componentResourceSchema() map[string]any {
	return map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"title":       "Mint Component Snapshot",
		"description": "Semantic snapshot for one Mint component.",
		"type":        "object",
		"required":    []string{"id", "type", "rect", "visible", "disabled"},
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "Stable component ID used by AI locators.",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Component/runtime tag type.",
			},
			"props": map[string]any{
				"type":                 "object",
				"description":          "Sanitized static props snapshot.",
				"additionalProperties": true,
			},
			"state": map[string]any{
				"type":                 "object",
				"description":          "Sanitized dynamic state snapshot.",
				"additionalProperties": true,
			},
			"rect":      rectSchema(),
			"visible":   map[string]any{"type": "boolean"},
			"disabled":  map[string]any{"type": "boolean"},
			"parent_id": map[string]any{"type": "string"},
			"children": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		"additionalProperties": false,
	}
}

func nodeBundleResourceSchema() map[string]any {
	return map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"title":       "Mint Node Bundle",
		"description": "Cross-layer runtime bundle for one Mint node ID.",
		"type":        "object",
		"required":    []string{"node_id"},
		"$defs": map[string]any{
			"rect":     rectSchema(),
			"treeNode": treeNodeSchema(),
			"hitEntry": hitEntrySchema(),
		},
		"properties": map[string]any{
			"node_id": map[string]any{
				"type":        "integer",
				"description": "Numeric runtime node ID.",
			},
			"fiber": map[string]any{
				"$ref": "#/$defs/treeNode",
			},
			"layout": map[string]any{
				"$ref": "#/$defs/treeNode",
			},
			"paintable": map[string]any{
				"$ref": "#/$defs/treeNode",
			},
			"hit_entries": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/$defs/hitEntry"},
			},
		},
		"additionalProperties": false,
	}
}

func rectSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"x", "y", "width", "height"},
		"properties": map[string]any{
			"x":      map[string]any{"type": "integer"},
			"y":      map[string]any{"type": "integer"},
			"width":  map[string]any{"type": "integer"},
			"height": map[string]any{"type": "integer"},
		},
		"additionalProperties": false,
	}
}

func treeNodeSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"kind"},
		"properties": map[string]any{
			"kind":    map[string]any{"type": "string"},
			"node_id": map[string]any{"type": "integer"},
			"path":    map[string]any{"type": "string"},
			"type":    map[string]any{"type": "string"},
			"tag":     map[string]any{"type": "string"},
			"key":     map[string]any{"type": "string"},
			"id":      map[string]any{"type": "string"},
			"rect":    map[string]any{"$ref": "#/$defs/rect"},
			"props": map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
			"state": map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
			"children": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/$defs/treeNode"},
			},
		},
		"additionalProperties": false,
	}
}

func hitEntrySchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"node_id", "bounds", "z_order"},
		"properties": map[string]any{
			"node_id": map[string]any{"type": "integer"},
			"bounds":  map[string]any{"$ref": "#/$defs/rect"},
			"z_order": map[string]any{"type": "integer"},
		},
		"additionalProperties": false,
	}
}

func (s *Server) runBatch(operations []BatchOperation, stopOnError bool) (*BatchResult, error) {
	if s.api.Batch != nil {
		return s.api.Batch(operations, stopOnError)
	}

	result := &BatchResult{
		OK:          true,
		Total:       len(operations),
		StopOnError: stopOnError,
		Results:     make([]BatchOperationResult, 0, len(operations)),
	}
	var firstErr error

	for idx, op := range operations {
		err := s.executeBatchOperation(op)
		item := BatchOperationResult{
			Index:     idx,
			Operation: op.Operation,
			Locator:   op.Locator,
			Direction: op.Direction,
			OK:        err == nil,
		}
		if err != nil {
			item.Error = err.Error()
			if firstErr == nil {
				firstErr = err
			}
		} else {
			result.Applied++
		}
		result.Results = append(result.Results, item)
		if err != nil && stopOnError {
			result.StoppedOnError = true
			break
		}
	}
	result.OK = firstErr == nil
	if firstErr != nil && stopOnError {
		return result, firstErr
	}
	return result, nil
}

func (s *Server) executeBatchOperation(op BatchOperation) error {
	switch op.Operation {
	case "set_value":
		if s.api.SetValue == nil {
			return fmt.Errorf("set_value not available")
		}
		return s.api.SetValue(op.Locator, op.Value)
	case "set_prop":
		if s.api.SetProp == nil {
			return fmt.Errorf("set_prop not available")
		}
		return s.api.SetProp(op.Locator, op.Key, op.Value)
	case "set_form_field":
		if s.api.SetFormField == nil {
			return fmt.Errorf("set_form_field not available")
		}
		return s.api.SetFormField(op.Locator, op.Field, op.Value)
	case "select":
		if s.api.Select == nil {
			return fmt.Errorf("select not available")
		}
		return s.api.Select(op.Locator, op.Value)
	case "click":
		if s.api.Click == nil {
			return fmt.Errorf("click not available")
		}
		return s.api.Click(op.Locator)
	case "input":
		if s.api.Input == nil {
			return fmt.Errorf("input not available")
		}
		return s.api.Input(op.Locator, op.Text)
	case "dispatch":
		if s.api.Dispatch == nil {
			return fmt.Errorf("dispatch not available")
		}
		return s.api.Dispatch(op.Locator, op.ActionType, op.Payload)
	case "navigate":
		if s.api.Navigate == nil {
			return fmt.Errorf("navigate not available")
		}
		return s.api.Navigate(op.Direction)
	default:
		return fmt.Errorf("unsupported batch operation: %s", op.Operation)
	}
}

func asString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

func waitInputToMap(in waitInput) map[string]any {
	out := map[string]any{
		"locator":             in.Locator,
		"selector":            in.Selector,
		"component_id":        in.ComponentID,
		"component_type":      in.ComponentType,
		"key":                 in.Key,
		"equals":              in.Equals,
		"exists":              in.Exists,
		"visible":             in.Visible,
		"disabled":            in.Disabled,
		"render_seq_at_least": in.RenderSeqAtLeast,
	}
	if len(in.Any) > 0 {
		children := make([]map[string]any, 0, len(in.Any))
		for _, child := range in.Any {
			children = append(children, waitInputToMap(child))
		}
		out["any"] = children
	}
	if len(in.All) > 0 {
		children := make([]map[string]any, 0, len(in.All))
		for _, child := range in.All {
			children = append(children, waitInputToMap(child))
		}
		out["all"] = children
	}
	if in.Not != nil {
		out["not"] = waitInputToMap(*in.Not)
	}
	return out
}

func (s *Server) handleInspect(w http.ResponseWriter, _ *http.Request) {
	if s.api.Inspect == nil {
		writeError(w, http.StatusNotImplemented, "inspect not available")
		return
	}
	result, err := s.api.Inspect()
	writeJSONResult(w, result, err)
}

func (s *Server) handleFind(w http.ResponseWriter, r *http.Request) {
	if s.api.Find == nil {
		writeError(w, http.StatusNotImplemented, "find not available")
		return
	}
	var req struct {
		Selector string `json:"selector"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.api.Find(req.Selector)
	writeJSONResult(w, result, err)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if s.api.Query == nil {
		writeError(w, http.StatusNotImplemented, "query not available")
		return
	}
	var req struct {
		ComponentID   string      `json:"component_id"`
		ComponentType string      `json:"component_type"`
		StateKey      string      `json:"state_key"`
		Value         interface{} `json:"value"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.api.Query(req.ComponentID, req.ComponentType, req.StateKey, req.Value)
	writeJSONResult(w, result, err)
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	if s.api.GetTree == nil {
		writeError(w, http.StatusNotImplemented, "tree not available")
		return
	}
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	result, err := s.api.GetTree(kind)
	writeJSONResult(w, result, err)
}

func (s *Server) handleNode(w http.ResponseWriter, r *http.Request) {
	if s.api.GetNode == nil {
		writeError(w, http.StatusNotImplemented, "node lookup not available")
		return
	}
	locator := strings.TrimSpace(r.URL.Query().Get("locator"))
	result, err := s.api.GetNode(locator)
	writeJSONResult(w, result, err)
}

func (s *Server) handleGetValue(w http.ResponseWriter, r *http.Request) {
	if s.api.GetValue == nil {
		writeError(w, http.StatusNotImplemented, "get value not available")
		return
	}
	var req struct {
		Locator string `json:"locator"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.api.GetValue(req.Locator)
	writeJSONResult(w, map[string]interface{}{"value": result}, err)
}

func (s *Server) handleSetValue(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	var req struct {
		Locator string      `json:"locator"`
		Value   interface{} `json:"value"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.SetValue(req.Locator, req.Value)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleSetProp(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	var req struct {
		Locator string      `json:"locator"`
		Key     string      `json:"key"`
		Value   interface{} `json:"value"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.SetProp(req.Locator, req.Key, req.Value)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleGetFormData(w http.ResponseWriter, r *http.Request) {
	if s.api.GetFormData == nil {
		writeError(w, http.StatusNotImplemented, "get form data not available")
		return
	}
	var req struct {
		Locator string `json:"locator"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.api.GetFormData(req.Locator)
	writeJSONResult(w, result, err)
}

func (s *Server) handleSetFormField(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	if s.api.SetFormField == nil {
		writeError(w, http.StatusNotImplemented, "set form field not available")
		return
	}
	var req struct {
		Locator string      `json:"locator"`
		Field   string      `json:"field"`
		Value   interface{} `json:"value"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.SetFormField(req.Locator, req.Field, req.Value)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleSelect(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	if s.api.Select == nil {
		writeError(w, http.StatusNotImplemented, "select not available")
		return
	}
	var req struct {
		Locator string      `json:"locator"`
		Value   interface{} `json:"value"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.Select(req.Locator, req.Value)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleClick(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	var req struct {
		Locator string `json:"locator"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.Click(req.Locator)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleInput(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	var req struct {
		Locator string `json:"locator"`
		Text    string `json:"text"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.Input(req.Locator, req.Text)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleNavigate(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	var req struct {
		Direction string `json:"direction"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.api.Navigate(req.Direction)
	writeJSONResult(w, map[string]bool{"ok": err == nil}, err)
}

func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ReadOnly {
		writeError(w, http.StatusForbidden, "read-only")
		return
	}
	if s.api.Batch == nil {
		writeError(w, http.StatusNotImplemented, "batch not available")
		return
	}
	var req batchInput
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	stopOnError := true
	if req.StopOnError != nil {
		stopOnError = *req.StopOnError
	}
	result, err := s.api.Batch(req.Operations, stopOnError)
	writeJSONResultWithOptionalError(w, result, err)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSONResult(w http.ResponseWriter, result interface{}, err error) {
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func writeJSONResultWithOptionalError(w http.ResponseWriter, result interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		payload := map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		}
		if result != nil {
			payload["result"] = result
		}
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":     true,
		"result": result,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    false,
		"error": message,
	})
}

func (cfg Config) String() string {
	return fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
}
