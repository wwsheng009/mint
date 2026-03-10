package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/framework/component"
	capabilitypkg "github.com/wwsheng009/mint/internal/ai/capability"
	mcppkg "github.com/wwsheng009/mint/internal/ai/mcp"
	selectorpkg "github.com/wwsheng009/mint/internal/ai/selector"
	snapshotpkg "github.com/wwsheng009/mint/internal/ai/snapshot"
	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	runtimerender "github.com/wwsheng009/mint/runtime/render"
	"github.com/wwsheng009/mint/runtime/state"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Host is the narrow interface the AI service needs from the app host.
// Keep this minimal to avoid back edges from internal/ai into framework.
type Host interface {
	IsRunning() bool
	GetRoot() component.Node
	GetFocusManager() *rtui.FiberFocusManager
	GetHitMap() *runtimeevent.HitMap
	Invoke(ctx context.Context, fn func() (any, error)) (any, error)
	MarkDirty()
}

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

type Config struct {
	Enabled   bool
	AutoStart bool
	ReadOnly  bool
	MCP       MCPConfig
}

type RenderInfo struct {
	RenderSeq  uint64
	RenderedAt time.Time
}

type Status struct {
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

type Direction string

const (
	DirectionNext  Direction = "next"
	DirectionPrev  Direction = "prev"
	DirectionFirst Direction = "first"
	DirectionLast  Direction = "last"
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

type BatchOperation = mcppkg.BatchOperation
type BatchResult = mcppkg.BatchResult

type ComponentInfo struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Props    map[string]interface{} `json:"props,omitempty"`
	State    map[string]interface{} `json:"state,omitempty"`
	Rect     state.Rect             `json:"rect"`
	Visible  bool                   `json:"visible"`
	Disabled bool                   `json:"disabled"`
	ParentID string                 `json:"parent_id,omitempty"`
	Children []string               `json:"children,omitempty"`
}

type StateQuery struct {
	ComponentID   string      `json:"component_id,omitempty"`
	ComponentType string      `json:"component_type,omitempty"`
	StateKey      string      `json:"state_key,omitempty"`
	Value         interface{} `json:"value,omitempty"`
}

type WaitCondition struct {
	Locator          string          `json:"locator,omitempty"`
	Selector         string          `json:"selector,omitempty"`
	ComponentID      string          `json:"component_id,omitempty"`
	ComponentType    string          `json:"component_type,omitempty"`
	Key              string          `json:"key,omitempty"`
	Equals           interface{}     `json:"equals,omitempty"`
	Exists           *bool           `json:"exists,omitempty"`
	Visible          *bool           `json:"visible,omitempty"`
	Disabled         *bool           `json:"disabled,omitempty"`
	RenderSeqAtLeast uint64          `json:"render_seq_at_least,omitempty"`
	Any              []WaitCondition `json:"any,omitempty"`
	All              []WaitCondition `json:"all,omitempty"`
	Not              *WaitCondition  `json:"not,omitempty"`
}

// Service is the internal host-facing AI subsystem skeleton.
// It currently tracks lifecycle and render heartbeat information.
type Service struct {
	mu            sync.RWMutex
	host          Host
	cfg           Config
	status        Status
	builder       *snapshotpkg.Builder
	latest        *state.Snapshot
	latestFrame   *snapshotpkg.Frame
	httpServer    *mcppkg.Server
	overrides     map[string]map[string]interface{}
	hookInstalled bool
	renderCh      chan struct{}
	batchWarnings []string
}

func New(host Host, cfg Config) *Service {
	svc := &Service{
		host:      host,
		cfg:       cfg,
		builder:   snapshotpkg.NewBuilder(),
		overrides: make(map[string]map[string]interface{}),
		renderCh:  make(chan struct{}),
	}
	svc.status.Enabled = cfg.Enabled
	svc.status.ReadOnly = cfg.ReadOnly
	svc.status.MCPEnabled = cfg.MCP.Enabled
	return svc
}

func (s *Service) ShouldAutoStart() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Enabled && s.cfg.AutoStart
}

func (s *Service) Start() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.cfg.Enabled || s.status.Running {
		return nil
	}

	s.installOverrideHookLocked()

	if s.cfg.MCP.Enabled {
		transport := strings.TrimSpace(strings.ToLower(s.cfg.MCP.Transport))
		if transport == "" {
			transport = "http"
		}
		if transport != "http" && transport != "pipe" {
			return fmt.Errorf("unsupported MCP transport: %s", s.cfg.MCP.Transport)
		}
		server := mcppkg.New(mcppkg.Config{
			Transport:   transport,
			Host:        s.cfg.MCP.Host,
			Port:        s.cfg.MCP.Port,
			ReadOnly:    s.cfg.ReadOnly || s.cfg.MCP.ReadOnly,
			AuthToken:   s.cfg.MCP.AuthToken,
			ExposeTrees: s.cfg.MCP.ExposeTrees,
			ExposeWrite: s.cfg.MCP.ExposeWrite,
		}, mcppkg.API{
			Inspect: s.Inspect,
			Find: func(selector string) (interface{}, error) {
				return s.Find(selector)
			},
			Query:          s.Query,
			GetTree:        s.GetTree,
			GetTreeCompact: s.GetTreeCompact,
			GetNode:        s.GetNode,
			GetFormData:    s.GetFormData,
			GetValue: func(locator string) (interface{}, error) {
				return s.GetValue(locator)
			},
			WaitFor: func(condition map[string]any, timeout time.Duration) (interface{}, error) {
				waitCond := waitConditionFromMap(condition)
				return s.WaitFor(waitCond, timeout)
			},
			SetValue:     s.SetValue,
			SetProp:      s.SetProp,
			SetFormField: s.SetFormField,
			Select:       s.Select,
			Click:        s.Click,
			Input:        s.Input,
			Dispatch:     s.Dispatch,
			Navigate: func(direction string) error {
				return s.Navigate(Direction(direction))
			},
			Batch: s.Batch,
		})
		if err := server.Start(); err != nil {
			return err
		}
		s.httpServer = server
		s.status.MCPEndpoint = server.Endpoint()
		s.status.HTTPEndpoint = server.BaseEndpoint()
	}

	s.status.Running = true
	s.status.StartedAt = time.Now()
	s.status.StoppedAt = time.Time{}
	return nil
}

func (s *Service) Stop() error {
	if s == nil {
		return nil
	}

	var notify chan struct{}

	s.mu.Lock()
	if !s.status.Running {
		s.mu.Unlock()
		return nil
	}

	if s.httpServer != nil {
		_ = s.httpServer.Stop()
		s.httpServer = nil
	}

	s.status.Running = false
	s.status.StoppedAt = time.Now()
	s.status.MCPEndpoint = ""
	s.status.HTTPEndpoint = ""
	notify = s.renderCh
	s.renderCh = make(chan struct{})
	s.mu.Unlock()

	if notify != nil {
		close(notify)
	}
	return nil
}

func (s *Service) OnAfterRender(info RenderInfo) {
	if s == nil {
		return
	}

	var notify chan struct{}

	s.mu.Lock()
	if !s.status.Running {
		s.mu.Unlock()
		return
	}

	if s.builder != nil && s.host != nil {
		s.latestFrame = s.builder.BuildFrame(snapshotpkg.Input{
			Root:         s.host.GetRoot(),
			FocusManager: s.host.GetFocusManager(),
			HitMap:       s.host.GetHitMap(),
			RenderSeq:    info.RenderSeq,
			RenderedAt:   info.RenderedAt,
		})
		if s.latestFrame != nil {
			s.latest = s.latestFrame.Snapshot
		}
	}

	s.status.RenderSeq = info.RenderSeq
	if !info.RenderedAt.IsZero() {
		s.status.LastRenderAt = info.RenderedAt
	} else {
		s.status.LastRenderAt = time.Now()
	}

	notify = s.renderCh
	s.renderCh = make(chan struct{})
	s.mu.Unlock()

	if notify != nil {
		close(notify)
	}
}

func (s *Service) Status() Status {
	if s == nil {
		return Status{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Service) Inspect() (*state.Snapshot, error) {
	if s == nil {
		return nil, errors.New("AI service is nil")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.cfg.Enabled {
		return nil, errors.New("AI service is disabled")
	}
	if s.latest == nil {
		return state.NewSnapshot(), nil
	}
	return s.latest.Clone(), nil
}

func (s *Service) Find(selector string) ([]ComponentInfo, error) {
	s.mu.RLock()
	frame := s.latestFrame
	s.mu.RUnlock()

	if !s.cfg.Enabled {
		return nil, errors.New("AI service is disabled")
	}
	if frame == nil || frame.Snapshot == nil {
		return nil, nil
	}

	locators, err := selectorpkg.Find(frame, selector)
	if err != nil {
		return nil, err
	}
	result := make([]ComponentInfo, 0, len(locators))
	for _, loc := range locators {
		if comp, ok := frame.Snapshot.GetComponent(loc.ComponentID); ok {
			result = append(result, componentInfoFromState(comp))
		}
	}
	return result, nil
}

func (s *Service) Query(componentID, componentType, stateKey string, value interface{}) (map[string]interface{}, error) {
	return s.QueryState(StateQuery{
		ComponentID:   componentID,
		ComponentType: componentType,
		StateKey:      stateKey,
		Value:         value,
	})
}

func (s *Service) QueryState(query StateQuery) (map[string]interface{}, error) {
	snap, err := s.Inspect()
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return map[string]interface{}{}, nil
	}

	if query.ComponentID != "" {
		comp, ok := snap.GetComponent(query.ComponentID)
		if !ok {
			return nil, fmt.Errorf("component not found: %s", query.ComponentID)
		}
		if query.StateKey != "" {
			if value, ok := comp.State[query.StateKey]; ok {
				return map[string]interface{}{query.StateKey: value}, nil
			}
			return map[string]interface{}{query.StateKey: comp.Props[query.StateKey]}, nil
		}
		result := cloneMap(comp.State)
		if result == nil {
			result = map[string]interface{}{}
		}
		for k, v := range comp.Props {
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}
		return result, nil
	}

	result := make(map[string]interface{})
	for id, comp := range snap.Components {
		if query.ComponentType != "" && comp.Type != query.ComponentType {
			continue
		}
		if query.StateKey != "" {
			value, ok := comp.State[query.StateKey]
			if !ok {
				value, ok = comp.Props[query.StateKey]
				if !ok {
					continue
				}
			}
			if query.Value != nil && value != query.Value {
				continue
			}
			result[id] = value
			continue
		}
		if query.Value != nil {
			matched := false
			for _, value := range comp.State {
				if value == query.Value {
					matched = true
					break
				}
			}
			if !matched {
				for _, value := range comp.Props {
					if value == query.Value {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}
		merged := cloneMap(comp.State)
		if merged == nil {
			merged = map[string]interface{}{}
		}
		for k, v := range comp.Props {
			if _, exists := merged[k]; !exists {
				merged[k] = v
			}
		}
		result[id] = merged
	}
	return result, nil
}

func (s *Service) WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error {
	if condition == nil {
		return errors.New("nil wait condition")
	}
	_, err := s.waitForSnapshot(timeout, condition)
	return err
}

func (s *Service) WaitFor(condition WaitCondition, timeout time.Duration) (*state.Snapshot, error) {
	return s.waitForSnapshot(timeout, func(snap *state.Snapshot) bool {
		return s.matchWaitCondition(snap, condition)
	})
}

func (s *Service) waitForSnapshot(timeout time.Duration, match func(*state.Snapshot) bool) (*state.Snapshot, error) {
	if match == nil {
		return nil, errors.New("nil wait condition")
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	deadline := time.Now().Add(timeout)
	for {
		ch := s.renderSignal()

		snap, err := s.Inspect()
		if err != nil {
			return nil, err
		}
		if match(snap) {
			return snap, nil
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("timeout waiting for condition")
		}

		timer := time.NewTimer(remaining)
		select {
		case <-ch:
			if !timer.Stop() {
				<-timer.C
			}
			continue
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for condition")
		}
	}
}

func (s *Service) renderSignal() <-chan struct{} {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	ch := s.renderCh
	s.mu.RUnlock()
	if ch != nil {
		return ch
	}

	s.mu.Lock()
	if s.renderCh == nil {
		s.renderCh = make(chan struct{})
	}
	ch = s.renderCh
	s.mu.Unlock()
	return ch
}

func (s *Service) GetValue(locator string) (interface{}, error) {
	loc, err := s.resolveLocator(locator)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	frame := s.latestFrame
	s.mu.RUnlock()
	if frame == nil || frame.Snapshot == nil {
		return nil, fmt.Errorf("no snapshot available")
	}
	comp, ok := frame.Snapshot.GetComponent(loc.ComponentID)
	if !ok {
		return nil, fmt.Errorf("component not found: %s", loc.ComponentID)
	}
	if value, ok := comp.State["value"]; ok {
		return value, nil
	}
	if value, ok := comp.Props["value"]; ok {
		return value, nil
	}
	if value, ok := comp.State["values"]; ok {
		return value, nil
	}
	if value, ok := comp.State["selectedValue"]; ok {
		return value, nil
	}
	if value, ok := comp.State["selectedValues"]; ok {
		return value, nil
	}
	if value, ok := comp.State["selectedIndex"]; ok {
		return value, nil
	}
	if value, ok := comp.State["checkedIndices"]; ok {
		return value, nil
	}
	if value, ok := comp.State["selectedRow"]; ok {
		return value, nil
	}
	if value, ok := comp.State["selectedNode"]; ok {
		return value, nil
	}
	if value, ok := comp.State["checked"]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("component has no readable value: %s", loc.ComponentID)
}

func (s *Service) SetValue(locator string, value interface{}) error {
	loc, err := s.resolveLocator(locator)
	if err != nil {
		return err
	}
	s.setOverride(loc.ComponentID, "value", value)
	return s.invokeFiberMutationWithLocator(loc, func(fiber *rtui.Fiber) error {
		if ok := capabilitypkg.SetValue(fiber, value); !ok {
			return nil
		}
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) SetProp(locator, key string, value interface{}) error {
	if key == "" {
		return errors.New("empty property key")
	}
	loc, err := s.resolveLocator(locator)
	if err != nil {
		return err
	}
	s.setOverride(loc.ComponentID, key, value)
	return s.invokeFiberMutationWithLocator(loc, func(fiber *rtui.Fiber) error {
		if ok := capabilitypkg.SetProp(fiber, key, value); !ok {
			return nil
		}
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) Click(locator string) error {
	return s.invokeFiberMutation(locator, func(fiber *rtui.Fiber) error {
		if focusable := fiber.GetFocusableInstance(); focusable != nil {
			if fm := s.host.GetFocusManager(); fm != nil {
				_ = fm.SetFocusByID(fmt.Sprintf("%d", fiber.NodeID))
			}
		}
		if handler, ok := fiber.Instance.(rtui.ActionHandlerInstance); ok {
			act := runtimeaction.NewAction(runtimeaction.ActionClick).
				WithTarget(fiber.ActionTargetID).
				WithPayload(nil)
			if !handler.HandleAction(act) {
				return fmt.Errorf("click action not handled: %s", locator)
			}
			s.host.MarkDirty()
			return nil
		}
		return fmt.Errorf("component does not handle click: %s", locator)
	})
}

func (s *Service) Input(locator, text string) error {
	return s.invokeFiberMutation(locator, func(fiber *rtui.Fiber) error {
		if handler, ok := fiber.Instance.(rtui.ActionHandlerInstance); ok {
			act := runtimeaction.NewAction(runtimeaction.ActionInputText).
				WithTarget(fiber.ActionTargetID).
				WithPayload(text)
			if handler.HandleAction(act) {
				s.host.MarkDirty()
				return nil
			}
		}
		if ok := capabilitypkg.SetValue(fiber, text); ok {
			s.host.MarkDirty()
			return nil
		}
		return fmt.Errorf("component does not support input: %s", locator)
	})
}

func (s *Service) Dispatch(locator string, actionType string, payload interface{}) error {
	if strings.TrimSpace(actionType) == "" {
		return errors.New("empty action type")
	}
	return s.invokeFiberMutation(locator, func(fiber *rtui.Fiber) error {
		handler, ok := fiber.Instance.(rtui.ActionHandlerInstance)
		if !ok {
			return fmt.Errorf("component does not handle actions: %s", locator)
		}
		act := runtimeaction.NewAction(runtimeaction.ActionType(actionType)).
			WithTarget(fiber.ActionTargetID).
			WithPayload(payload)
		if !handler.HandleAction(act) {
			return fmt.Errorf("action not handled: %s", actionType)
		}
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) Navigate(direction Direction) error {
	if s == nil || s.host == nil {
		return errors.New("AI service is not available")
	}

	_, err := s.host.Invoke(context.Background(), func() (any, error) {
		if err := s.navigateWithoutInvoke(direction); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (s *Service) GetTree(kind string) (interface{}, error) {
	if s == nil || s.host == nil {
		return nil, errors.New("AI service is not available")
	}
	return s.host.Invoke(context.Background(), func() (any, error) {
		return snapshotpkg.BuildTreeWithOptions(s.host.GetRoot(), kind, snapshotpkg.TreeOptions{})
	})
}

func (s *Service) GetTreeCompact(kind string) (interface{}, error) {
	return s.GetTreeWithOptions(kind, snapshotpkg.TreeOptions{Compact: true})
}

func (s *Service) GetTreeWithOptions(kind string, opts snapshotpkg.TreeOptions) (interface{}, error) {
	if s == nil || s.host == nil {
		return nil, errors.New("AI service is not available")
	}
	return s.host.Invoke(context.Background(), func() (any, error) {
		return snapshotpkg.BuildTreeWithOptions(s.host.GetRoot(), kind, opts)
	})
}

func (s *Service) GetNode(locator string) (interface{}, error) {
	resolved, err := s.resolveLocator(locator)
	if err != nil {
		return nil, err
	}

	if resolved.NodeID == 0 {
		return nil, fmt.Errorf("locator has no node id: %s", locator)
	}

	return s.host.Invoke(context.Background(), func() (any, error) {
		return snapshotpkg.BuildNodeBundle(s.host.GetRoot(), resolved.NodeID)
	})
}

func (s *Service) GetFormData(locator string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.invokeFiberMutation(locator, func(fiber *rtui.Fiber) error {
		reader, ok := fiber.Instance.(interface{ GetValues() map[string]interface{} })
		if !ok {
			return fmt.Errorf("component is not a form: %s", locator)
		}
		result = cloneMap(reader.GetValues())
		return nil
	})
	return result, err
}

func (s *Service) SetFormField(locator, field string, value interface{}) error {
	if field == "" {
		return errors.New("empty form field")
	}
	loc, err := s.resolveLocator(locator)
	if err != nil {
		return err
	}
	s.setOverrideValueMap(loc.ComponentID, "values", field, value)
	return s.invokeFiberMutationWithLocator(loc, func(fiber *rtui.Fiber) error {
		writer, ok := fiber.Instance.(interface {
			SetValue(field string, value interface{})
		})
		if !ok {
			return nil
		}
		writer.SetValue(field, value)
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) Select(locator string, value interface{}) error {
	loc, err := s.resolveLocator(locator)
	if err != nil {
		return err
	}
	switch v := value.(type) {
	case string:
		s.setOverride(loc.ComponentID, "selected", v)
	case []string:
		s.setOverride(loc.ComponentID, "selecteds", append([]string(nil), v...))
	case int:
		s.setOverride(loc.ComponentID, "selectedIndex", v)
	case float64:
		s.setOverride(loc.ComponentID, "selectedIndex", int(v))
	}
	return s.invokeFiberMutationWithLocator(loc, func(fiber *rtui.Fiber) error {
		switch v := value.(type) {
		case string:
			if index, ok := findSelectionIndexForString(fiber, v); ok {
				if capabilitypkg.SelectIndex(fiber, index) || selectedIndexEquals(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
		case int:
			if capabilitypkg.SelectIndex(fiber, v) || selectedIndexEquals(fiber, v) {
				s.host.MarkDirty()
				return nil
			}
		case float64:
			if capabilitypkg.SelectIndex(fiber, int(v)) || selectedIndexEquals(fiber, int(v)) {
				s.host.MarkDirty()
				return nil
			}
		case map[string]interface{}:
			if index, ok := numericMapValue(v, "index"); ok {
				if capabilitypkg.SelectIndex(fiber, index) || selectedIndexEquals(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
			if index, ok := numericMapValue(v, "source_index"); ok {
				if capabilitypkg.ToggleSelectionIndex(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
			if index, ok := numericMapValue(v, "toggle_index"); ok {
				if capabilitypkg.ToggleSelectionIndex(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
			for _, key := range []string{"value", "text", "path"} {
				if text, ok := stringMapValue(v, key); ok {
					if index, ok := findSelectionIndexForString(fiber, text); ok {
						if capabilitypkg.SelectIndex(fiber, index) || selectedIndexEquals(fiber, index) {
							s.host.MarkDirty()
							return nil
						}
					}
				}
			}
		}
		if selector, ok := fiber.Instance.(interface{ SelectOption(string) }); ok {
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("select value must be string for option group: %T", value)
			}
			selector.SelectOption(text)
			s.host.MarkDirty()
			return nil
		}
		if setter, ok := fiber.Instance.(interface {
			SetProp(key string, value interface{})
		}); ok {
			switch v := value.(type) {
			case string:
				if idx, ok := findOptionIndexByValue(fiber, v); ok {
					setter.SetProp("selectedIndex", idx)
					s.host.MarkDirty()
					return nil
				}
			case int:
				setter.SetProp("selectedIndex", v)
				s.host.MarkDirty()
				return nil
			case float64:
				setter.SetProp("selectedIndex", int(v))
				s.host.MarkDirty()
				return nil
			}
		}
		return fmt.Errorf("component does not support select: %s", locator)
	})
}

func (s *Service) Batch(operations []mcppkg.BatchOperation, stopOnError bool, dryRun bool) (*mcppkg.BatchResult, error) {
	if s == nil || s.host == nil {
		return nil, errors.New("AI service is not available")
	}
	if s.cfg.ReadOnly {
		return nil, errors.New("AI service is read-only")
	}
	if len(operations) == 0 {
		return nil, errors.New("batch requires at least one operation")
	}

	result := &mcppkg.BatchResult{
		OK:          true,
		Total:       len(operations),
		StopOnError: stopOnError,
		Results:     make([]mcppkg.BatchOperationResult, 0, len(operations)),
	}
	var firstErr error
	previewCounts := make(map[string]*mcppkg.BatchPreviewTarget)
	s.resetBatchWarnings()

	_, err := s.host.Invoke(context.Background(), func() (any, error) {
		if dryRun {
			result.ValidatedOnly = true
			for idx, op := range operations {
				item, err := s.prepareBatchOperation(op)
				if err != nil {
					if firstErr == nil {
						firstErr = err
					}
					result.Results = append(result.Results, batchResultItem(idx, op, err, false, false, nil, nil, nil, nil))
					result.OK = false
					if stopOnError {
						result.StoppedOnError = true
						result.Preview = buildBatchPreview(previewCounts)
						break
					}
					continue
				}
				loc := item.Locator
				result.Results = append(result.Results, batchResultItem(idx, op, nil, true, false, &loc, item.Preview, item.Warnings, item.WarningCodes))
				accumulateBatchPreview(previewCounts, op, &loc)
			}
			if firstErr == nil {
				result.OK = true
			}
			result.Status = batchSummaryStatus(result, firstErr, dryRun)
			result.Preview = buildBatchPreview(previewCounts)
			result.Warnings = s.takeBatchWarnings()
			return nil, nil
		}

		if stopOnError {
			prepared := make([]preparedBatchOperation, 0, len(operations))
			precheckResults := make([]mcppkg.BatchOperationResult, 0, len(operations))
			result.ValidatedOnly = true
			for idx, op := range operations {
				item, err := s.prepareBatchOperation(op)
				if err != nil {
					firstErr = err
					result.OK = false
					result.StoppedOnError = true
					precheckResults = append(precheckResults, batchResultItem(idx, op, err, false, false, nil, nil, nil, nil))
					result.Results = precheckResults
					result.Status = batchSummaryStatus(result, firstErr, dryRun)
					result.Preview = buildBatchPreview(previewCounts)
					return nil, nil
				}
				item.Index = idx
				prepared = append(prepared, item)
				loc := item.Locator
				precheckResults = append(precheckResults, batchResultItem(idx, op, nil, true, false, &loc, item.Preview, item.Warnings, item.WarningCodes))
				accumulateBatchPreview(previewCounts, op, &loc)
			}
			result.ValidatedOnly = false
			result.Results = result.Results[:0]
			for _, item := range prepared {
				err := s.executePreparedBatchOperation(item)
				loc := item.Locator
				result.Results = append(result.Results, batchResultItem(item.Index, item.Operation, err, true, err == nil, &loc, item.Preview, item.Warnings, item.WarningCodes))
				accumulateBatchPreview(previewCounts, item.Operation, &loc)
				if err != nil {
					firstErr = err
					result.OK = false
					result.StoppedOnError = true
					break
				}
				result.Applied++
			}
			if firstErr == nil {
				result.OK = true
			}
			result.Status = batchSummaryStatus(result, firstErr, dryRun)
			result.Preview = buildBatchPreview(previewCounts)
			result.Warnings = s.takeBatchWarnings()
			return nil, nil
		}

		for idx, op := range operations {
			item, err := s.prepareBatchOperation(op)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				result.Results = append(result.Results, batchResultItem(idx, op, err, false, false, nil, nil, nil, nil))
				continue
			}
			item.Index = idx
			err = s.executePreparedBatchOperation(item)
			loc := item.Locator
			result.Results = append(result.Results, batchResultItem(idx, op, err, true, err == nil, &loc, item.Preview, item.Warnings, item.WarningCodes))
			accumulateBatchPreview(previewCounts, op, &loc)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			result.Applied++
		}
		result.OK = firstErr == nil
		result.Status = batchSummaryStatus(result, firstErr, dryRun)
		result.Preview = buildBatchPreview(previewCounts)
		result.Warnings = s.takeBatchWarnings()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type preparedBatchOperation struct {
	Index        int
	Operation    mcppkg.BatchOperation
	Locator      snapshotpkg.NodeLocator
	HasLocator   bool
	Preview      *mcppkg.BatchPreviewTarget
	Warnings     []string
	WarningCodes []mcppkg.WarningCode
}

func batchResultItem(index int, op mcppkg.BatchOperation, err error, validated bool, executed bool, locator *snapshotpkg.NodeLocator, preview *mcppkg.BatchPreviewTarget, warnings []string, warningCodes []mcppkg.WarningCode) mcppkg.BatchOperationResult {
	item := mcppkg.BatchOperationResult{
		Index:     index,
		Operation: op.Operation,
		Locator:   op.Locator,
		Direction: op.Direction,
		OK:        err == nil,
		Validated: validated,
		Executed:  executed,
	}
	item.Status = batchStatus(err, validated, executed)
	if len(warnings) > 0 {
		item.Warnings = append([]string(nil), warnings...)
	}
	if len(warningCodes) > 0 {
		item.WarningCodes = append([]mcppkg.WarningCode(nil), warningCodes...)
	}
	if locator != nil {
		item.ComponentID = locator.ComponentID
		item.NodeID = locator.NodeID
		item.Path = locator.Path
		item.ActionTargetID = locator.ActionTargetID
		item.Type = locator.Type
		item.Tag = locator.Tag
	}
	if preview != nil {
		item.Preview = preview
	} else if locator != nil {
		item.Preview = &mcppkg.BatchPreviewTarget{
			ComponentID:    locator.ComponentID,
			NodeID:         locator.NodeID,
			Path:           locator.Path,
			ActionTargetID: locator.ActionTargetID,
			Type:           locator.Type,
			Tag:            locator.Tag,
			Operation:      op.Operation,
			Direction:      op.Direction,
			Count:          1,
		}
	} else if op.Operation == "navigate" {
		item.Preview = &mcppkg.BatchPreviewTarget{
			Operation: op.Operation,
			Direction: op.Direction,
			Count:     1,
		}
	}
	if err != nil {
		item.Error = err.Error()
	}
	return item
}

func batchStatus(err error, validated bool, executed bool) string {
	if err != nil {
		return "failed"
	}
	if executed {
		return "executed"
	}
	if validated {
		return "validated_only"
	}
	return "skipped"
}

func batchSummaryStatus(result *mcppkg.BatchResult, firstErr error, dryRun bool) string {
	if firstErr != nil {
		return "failed"
	}
	if dryRun || result.ValidatedOnly {
		return "validated_only"
	}
	if result.Applied > 0 {
		return "executed"
	}
	return "skipped"
}

func (s *Service) prepareBatchOperation(op mcppkg.BatchOperation) (preparedBatchOperation, error) {
	item := preparedBatchOperation{Operation: op}
	switch op.Operation {
	case "set_value", "select", "click", "input":
		if strings.TrimSpace(op.Locator) == "" {
			return item, errors.New("empty locator")
		}
	case "set_prop":
		if strings.TrimSpace(op.Locator) == "" {
			return item, errors.New("empty locator")
		}
		if strings.TrimSpace(op.Key) == "" {
			return item, errors.New("empty property key")
		}
	case "set_form_field":
		if strings.TrimSpace(op.Locator) == "" {
			return item, errors.New("empty locator")
		}
		if strings.TrimSpace(op.Field) == "" {
			return item, errors.New("empty form field")
		}
	case "dispatch":
		if strings.TrimSpace(op.Locator) == "" {
			return item, errors.New("empty locator")
		}
		if strings.TrimSpace(op.ActionType) == "" {
			return item, errors.New("empty action type")
		}
	case "navigate":
		if _, err := normalizeBatchDirection(op.Direction); err != nil {
			return item, err
		}
		item.Preview = &mcppkg.BatchPreviewTarget{
			Operation: op.Operation,
			Direction: op.Direction,
			Count:     1,
		}
		item.Warnings = nil
		return item, nil
	default:
		return item, fmt.Errorf("unsupported batch operation: %s", op.Operation)
	}

	switch op.Operation {
	case "navigate":
		return item, nil
	default:
		loc, err := s.resolveLocator(op.Locator)
		if err != nil {
			return item, err
		}
		item.Locator = loc
		item.HasLocator = true
		item.Preview, item.Warnings, item.WarningCodes = s.buildOperationPreview(op, loc)
		return item, nil
	}
}

func (s *Service) executePreparedBatchOperation(item preparedBatchOperation) error {
	op := item.Operation
	switch op.Operation {
	case "set_value":
		return s.applySetValueWithLocator(item.Locator, op.Value)
	case "set_prop":
		return s.applySetPropWithLocator(item.Locator, op.Key, op.Value)
	case "set_form_field":
		return s.applySetFormFieldWithLocator(item.Locator, op.Field, op.Value)
	case "select":
		return s.applySelectWithLocator(item.Locator, op.Value, op.Locator)
	case "click":
		return s.applyClickWithLocator(item.Locator, op.Locator)
	case "input":
		return s.applyInputWithLocator(item.Locator, op.Locator, op.Text)
	case "dispatch":
		return s.applyDispatchWithLocator(item.Locator, op.Locator, op.ActionType, op.Payload)
	case "navigate":
		direction, err := normalizeBatchDirection(op.Direction)
		if err != nil {
			return err
		}
		return s.navigateWithoutInvoke(direction)
	default:
		return fmt.Errorf("unsupported batch operation: %s", op.Operation)
	}
}

func (s *Service) executeBatchOperation(op mcppkg.BatchOperation) error {
	item, err := s.prepareBatchOperation(op)
	if err != nil {
		return err
	}
	return s.executePreparedBatchOperation(item)
}

func (s *Service) applySetValueWithLocator(loc snapshotpkg.NodeLocator, value interface{}) error {
	s.setOverride(loc.ComponentID, "value", value)
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		if ok := capabilitypkg.SetValue(fiber, value); !ok {
			return nil
		}
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) applySetPropWithLocator(loc snapshotpkg.NodeLocator, key string, value interface{}) error {
	s.setOverride(loc.ComponentID, key, value)
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		if ok := capabilitypkg.SetProp(fiber, key, value); !ok {
			return nil
		}
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) applyClickWithLocator(loc snapshotpkg.NodeLocator, locator string) error {
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		if focusable := fiber.GetFocusableInstance(); focusable != nil {
			if fm := s.host.GetFocusManager(); fm != nil {
				_ = fm.SetFocusByID(fmt.Sprintf("%d", fiber.NodeID))
			}
		}
		if handler, ok := fiber.Instance.(rtui.ActionHandlerInstance); ok {
			act := runtimeaction.NewAction(runtimeaction.ActionClick).
				WithTarget(fiber.ActionTargetID).
				WithPayload(nil)
			if !handler.HandleAction(act) {
				return fmt.Errorf("click action not handled: %s", locator)
			}
			s.host.MarkDirty()
			return nil
		}
		return fmt.Errorf("component does not handle click: %s", locator)
	})
}

func (s *Service) applyInputWithLocator(loc snapshotpkg.NodeLocator, locator, text string) error {
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		if handler, ok := fiber.Instance.(rtui.ActionHandlerInstance); ok {
			act := runtimeaction.NewAction(runtimeaction.ActionInputText).
				WithTarget(fiber.ActionTargetID).
				WithPayload(text)
			if handler.HandleAction(act) {
				s.host.MarkDirty()
				return nil
			}
		}
		if ok := capabilitypkg.SetValue(fiber, text); ok {
			s.host.MarkDirty()
			return nil
		}
		return fmt.Errorf("component does not support input: %s", locator)
	})
}

func (s *Service) applyDispatchWithLocator(loc snapshotpkg.NodeLocator, locator, actionType string, payload interface{}) error {
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		handler, ok := fiber.Instance.(rtui.ActionHandlerInstance)
		if !ok {
			return fmt.Errorf("component does not handle actions: %s", locator)
		}
		act := runtimeaction.NewAction(runtimeaction.ActionType(actionType)).
			WithTarget(fiber.ActionTargetID).
			WithPayload(payload)
		if !handler.HandleAction(act) {
			return fmt.Errorf("action not handled: %s", actionType)
		}
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) applySetFormFieldWithLocator(loc snapshotpkg.NodeLocator, field string, value interface{}) error {
	s.setOverrideValueMap(loc.ComponentID, "values", field, value)
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		writer, ok := fiber.Instance.(interface {
			SetValue(field string, value interface{})
		})
		if !ok {
			return nil
		}
		writer.SetValue(field, value)
		s.host.MarkDirty()
		return nil
	})
}

func (s *Service) applySelectWithLocator(loc snapshotpkg.NodeLocator, value interface{}, locator string) error {
	switch v := value.(type) {
	case string:
		s.setOverride(loc.ComponentID, "selected", v)
	case []string:
		s.setOverride(loc.ComponentID, "selecteds", append([]string(nil), v...))
	case int:
		s.setOverride(loc.ComponentID, "selectedIndex", v)
	case float64:
		s.setOverride(loc.ComponentID, "selectedIndex", int(v))
	}
	return s.mutateFiberWithLocator(loc, func(fiber *rtui.Fiber) error {
		switch v := value.(type) {
		case string:
			if index, ok := findSelectionIndexForString(fiber, v); ok {
				if capabilitypkg.SelectIndex(fiber, index) || selectedIndexEquals(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
		case int:
			if capabilitypkg.SelectIndex(fiber, v) || selectedIndexEquals(fiber, v) {
				s.host.MarkDirty()
				return nil
			}
		case float64:
			if capabilitypkg.SelectIndex(fiber, int(v)) || selectedIndexEquals(fiber, int(v)) {
				s.host.MarkDirty()
				return nil
			}
		case map[string]interface{}:
			if index, ok := numericMapValue(v, "index"); ok {
				if capabilitypkg.SelectIndex(fiber, index) || selectedIndexEquals(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
			if index, ok := numericMapValue(v, "source_index"); ok {
				if capabilitypkg.ToggleSelectionIndex(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
			if index, ok := numericMapValue(v, "toggle_index"); ok {
				if capabilitypkg.ToggleSelectionIndex(fiber, index) {
					s.host.MarkDirty()
					return nil
				}
			}
			for _, key := range []string{"value", "text", "path"} {
				if text, ok := stringMapValue(v, key); ok {
					if index, ok := findSelectionIndexForString(fiber, text); ok {
						if capabilitypkg.SelectIndex(fiber, index) || selectedIndexEquals(fiber, index) {
							s.host.MarkDirty()
							return nil
						}
					}
				}
			}
		}
		if selector, ok := fiber.Instance.(interface{ SelectOption(string) }); ok {
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("select value must be string for option group: %T", value)
			}
			selector.SelectOption(text)
			s.host.MarkDirty()
			return nil
		}
		if setter, ok := fiber.Instance.(interface {
			SetProp(key string, value interface{})
		}); ok {
			switch v := value.(type) {
			case string:
				if idx, ok := findOptionIndexByValue(fiber, v); ok {
					setter.SetProp("selectedIndex", idx)
					s.host.MarkDirty()
					return nil
				}
			case int:
				setter.SetProp("selectedIndex", v)
				s.host.MarkDirty()
				return nil
			case float64:
				setter.SetProp("selectedIndex", int(v))
				s.host.MarkDirty()
				return nil
			}
		}
		return fmt.Errorf("component does not support select: %s", locator)
	})
}

func (s *Service) navigateWithoutInvoke(direction Direction) error {
	fm := s.host.GetFocusManager()
	if fm == nil {
		return errors.New("focus manager not available")
	}
	var ok bool
	switch direction {
	case DirectionNext, DirectionRight, DirectionDown:
		ok = fm.FocusNext()
	case DirectionPrev, DirectionLeft, DirectionUp:
		ok = fm.FocusPrev()
	case DirectionFirst:
		ok = fm.FocusFirst()
	case DirectionLast:
		ok = fm.FocusLast()
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}
	if !ok {
		return fmt.Errorf("navigation failed: %s", direction)
	}
	s.host.MarkDirty()
	return nil
}

func normalizeBatchDirection(direction string) (Direction, error) {
	dir := Direction(strings.TrimSpace(direction))
	switch dir {
	case DirectionNext, DirectionPrev, DirectionFirst, DirectionLast, DirectionUp, DirectionDown, DirectionLeft, DirectionRight:
		return dir, nil
	default:
		return "", fmt.Errorf("invalid direction: %s", direction)
	}
}

func (s *Service) buildOperationPreview(op mcppkg.BatchOperation, loc snapshotpkg.NodeLocator) (*mcppkg.BatchPreviewTarget, []string, []mcppkg.WarningCode) {
	var warnings []string
	var warningCodes []mcppkg.WarningCode
	preview := &mcppkg.BatchPreviewTarget{
		ComponentID:    loc.ComponentID,
		NodeID:         loc.NodeID,
		Path:           loc.Path,
		ActionTargetID: loc.ActionTargetID,
		Type:           loc.Type,
		Tag:            loc.Tag,
		Operation:      op.Operation,
		Direction:      op.Direction,
		Count:          1,
	}
	switch op.Operation {
	case "set_value":
		preview.Key = "value"
		if before, ok := s.componentValueSnapshot(loc.ComponentID, "value"); ok {
			preview.Before = before
		} else {
			warnings = append(warnings, fmt.Sprintf("preview.before missing for %s:%s", loc.ComponentID, "value"))
			warningCodes = append(warningCodes, mcppkg.WarningCodePreviewBeforeMissing)
		}
		preview.After = op.Value
	case "set_prop":
		preview.Key = op.Key
		if before, ok := s.componentValueSnapshot(loc.ComponentID, op.Key); ok {
			preview.Before = before
		} else {
			warnings = append(warnings, fmt.Sprintf("preview.before missing for %s:%s", loc.ComponentID, op.Key))
			warningCodes = append(warningCodes, mcppkg.WarningCodePreviewBeforeMissing)
		}
		preview.After = op.Value
	case "set_form_field":
		preview.Field = op.Field
		if before, ok := s.componentFormFieldSnapshot(loc.ComponentID, op.Field); ok {
			preview.Before = before
		} else {
			warnings = append(warnings, fmt.Sprintf("preview.before missing for %s:%s", loc.ComponentID, op.Field))
			warningCodes = append(warningCodes, mcppkg.WarningCodePreviewBeforeMissing)
		}
		preview.After = op.Value
	case "select":
		if before, ok := s.componentValueSnapshot(loc.ComponentID, "selectedIndex"); ok {
			preview.Before = before
		} else if before, ok := s.componentValueSnapshot(loc.ComponentID, "selected"); ok {
			preview.Before = before
		} else if before, ok := s.componentValueSnapshot(loc.ComponentID, "selectedRow"); ok {
			preview.Before = before
		} else if before, ok := s.componentValueSnapshot(loc.ComponentID, "selectedNode"); ok {
			preview.Before = before
		} else {
			warnings = append(warnings, fmt.Sprintf("preview.before missing for %s:selected", loc.ComponentID))
			warningCodes = append(warningCodes, mcppkg.WarningCodePreviewBeforeMissing)
		}
		preview.After = op.Value
	case "click", "dispatch", "input":
		preview.MayChange = true
	}
	return preview, warnings, warningCodes
}

func (s *Service) componentValueSnapshot(componentID, key string) (interface{}, bool) {
	s.mu.RLock()
	frame := s.latestFrame
	s.mu.RUnlock()
	if frame == nil || frame.Snapshot == nil {
		return nil, false
	}
	comp, ok := frame.Snapshot.GetComponent(componentID)
	if !ok {
		return nil, false
	}
	if value, ok := comp.State[key]; ok {
		return value, true
	}
	if value, ok := comp.Props[key]; ok {
		return value, true
	}
	return nil, false
}

func (s *Service) componentFormFieldSnapshot(componentID, field string) (interface{}, bool) {
	s.mu.RLock()
	frame := s.latestFrame
	s.mu.RUnlock()
	if frame == nil || frame.Snapshot == nil {
		return nil, false
	}
	comp, ok := frame.Snapshot.GetComponent(componentID)
	if !ok {
		return nil, false
	}
	if values, ok := comp.State["values"].(map[string]interface{}); ok {
		if value, ok := values[field]; ok {
			return value, true
		}
	}
	if values, ok := comp.Props["values"].(map[string]interface{}); ok {
		if value, ok := values[field]; ok {
			return value, true
		}
	}
	return nil, false
}

func accumulateBatchPreview(preview map[string]*mcppkg.BatchPreviewTarget, op mcppkg.BatchOperation, loc *snapshotpkg.NodeLocator) {
	if op.Operation == "navigate" {
		key := fmt.Sprintf("navigate|%s", op.Direction)
		target := preview[key]
		if target == nil {
			target = &mcppkg.BatchPreviewTarget{
				Operation: op.Operation,
				Direction: op.Direction,
			}
			preview[key] = target
		}
		target.Count++
		return
	}
	if loc == nil {
		return
	}
	key := fmt.Sprintf("%s|%d|%s|%s|%s", loc.ComponentID, loc.NodeID, loc.Path, loc.ActionTargetID, op.Operation)
	target := preview[key]
	if target == nil {
		target = &mcppkg.BatchPreviewTarget{
			ComponentID:    loc.ComponentID,
			NodeID:         loc.NodeID,
			Path:           loc.Path,
			ActionTargetID: loc.ActionTargetID,
			Type:           loc.Type,
			Tag:            loc.Tag,
			Operation:      op.Operation,
		}
		preview[key] = target
	}
	target.Count++
}

func buildBatchPreview(preview map[string]*mcppkg.BatchPreviewTarget) []mcppkg.BatchPreviewTarget {
	if len(preview) == 0 {
		return nil
	}
	out := make([]mcppkg.BatchPreviewTarget, 0, len(preview))
	for _, entry := range preview {
		out = append(out, *entry)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ComponentID != out[j].ComponentID {
			return out[i].ComponentID < out[j].ComponentID
		}
		if out[i].NodeID != out[j].NodeID {
			return out[i].NodeID < out[j].NodeID
		}
		if out[i].Operation != out[j].Operation {
			return out[i].Operation < out[j].Operation
		}
		return out[i].Direction < out[j].Direction
	})
	return out
}

func (s *Service) resetBatchWarnings() {
	s.mu.Lock()
	s.batchWarnings = nil
	s.mu.Unlock()
}

func (s *Service) appendBatchWarningf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	s.mu.Lock()
	s.batchWarnings = append(s.batchWarnings, message)
	s.mu.Unlock()
}

func (s *Service) takeBatchWarnings() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.batchWarnings) == 0 {
		return nil
	}
	out := append([]string(nil), s.batchWarnings...)
	s.batchWarnings = nil
	return out
}

func componentInfoFromState(comp state.ComponentState) ComponentInfo {
	return ComponentInfo{
		ID:       comp.ID,
		Type:     comp.Type,
		Props:    cloneMap(comp.Props),
		State:    cloneMap(comp.State),
		Rect:     comp.Rect,
		Visible:  comp.Visible,
		Disabled: comp.Disabled,
	}
}

func (s *Service) matchWaitCondition(snap *state.Snapshot, cond WaitCondition) bool {
	if snap == nil {
		return false
	}
	if cond.RenderSeqAtLeast > 0 {
		renderSeq, _ := snap.Metadata["render_seq"].(uint64)
		if renderSeq < cond.RenderSeqAtLeast {
			return false
		}
	}
	if cond.Not != nil && s.matchWaitCondition(snap, *cond.Not) {
		return false
	}
	if len(cond.All) > 0 {
		for _, child := range cond.All {
			if !s.matchWaitCondition(snap, child) {
				return false
			}
		}
	}
	if len(cond.Any) > 0 {
		matched := false
		for _, child := range cond.Any {
			if s.matchWaitCondition(snap, child) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	candidates := make([]state.ComponentState, 0)
	switch {
	case cond.ComponentID != "":
		comp, ok := snap.GetComponent(cond.ComponentID)
		if !ok {
			return false
		}
		candidates = append(candidates, comp)
	case cond.Locator != "" || cond.Selector != "":
		locators, err := s.findLocatorsForWait(cond)
		if err != nil || len(locators) == 0 {
			return false
		}
		for _, loc := range locators {
			if comp, ok := snap.GetComponent(loc.ComponentID); ok {
				candidates = append(candidates, comp)
			}
		}
	default:
		for _, comp := range snap.Components {
			if cond.ComponentType != "" && comp.Type != cond.ComponentType {
				continue
			}
			candidates = append(candidates, comp)
		}
	}
	if len(candidates) == 0 {
		return len(cond.All) > 0 || len(cond.Any) > 0
	}

	for _, comp := range candidates {
		if cond.ComponentType != "" && comp.Type != cond.ComponentType {
			continue
		}
		if cond.Visible != nil && comp.Visible != *cond.Visible {
			continue
		}
		if cond.Disabled != nil && comp.Disabled != *cond.Disabled {
			continue
		}
		if cond.Key == "" {
			return true
		}
		value, exists := comp.State[cond.Key]
		if !exists {
			value, exists = comp.Props[cond.Key]
		}
		if cond.Exists != nil && exists != *cond.Exists {
			continue
		}
		if cond.Equals != nil && fmt.Sprintf("%v", value) != fmt.Sprintf("%v", cond.Equals) {
			continue
		}
		if cond.Exists == nil && cond.Equals == nil && !exists {
			continue
		}
		return true
	}
	if len(cond.All) > 0 || len(cond.Any) > 0 || cond.Not != nil {
		return cond.ComponentID == "" && cond.ComponentType == "" &&
			cond.Selector == "" && cond.Locator == "" && cond.Key == "" &&
			cond.Equals == nil && cond.Exists == nil &&
			cond.Visible == nil && cond.Disabled == nil
	}
	return false
}

func (s *Service) findLocatorsForWait(cond WaitCondition) ([]snapshotpkg.NodeLocator, error) {
	s.mu.RLock()
	frame := s.latestFrame
	s.mu.RUnlock()
	if frame == nil {
		return nil, errors.New("no snapshot available")
	}
	if cond.Locator != "" {
		loc, err := s.resolveLocator(cond.Locator)
		if err != nil {
			return nil, err
		}
		return []snapshotpkg.NodeLocator{loc}, nil
	}
	return selectorpkg.Find(frame, cond.Selector)
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (s *Service) resolveLocator(locator string) (snapshotpkg.NodeLocator, error) {
	s.mu.RLock()
	frame := s.latestFrame
	s.mu.RUnlock()
	if frame == nil {
		return snapshotpkg.NodeLocator{}, errors.New("no snapshot available")
	}

	locators, err := selectorpkg.Find(frame, locator)
	if err == nil && len(locators) > 0 {
		return locators[0], nil
	}

	if loc, ok := frame.ByComponentID[locator]; ok {
		return loc, nil
	}
	if loc, ok := frame.ByPath[locator]; ok {
		return loc, nil
	}
	return snapshotpkg.NodeLocator{}, fmt.Errorf("locator not found: %s", locator)
}

func (s *Service) invokeFiberMutation(locator string, fn func(fiber *rtui.Fiber) error) error {
	loc, err := s.resolveLocator(locator)
	if err != nil {
		return err
	}
	return s.invokeFiberMutationWithLocator(loc, fn)
}

func (s *Service) invokeFiberMutationWithLocator(loc snapshotpkg.NodeLocator, fn func(fiber *rtui.Fiber) error) error {
	if s == nil || s.host == nil {
		return errors.New("AI service is not available")
	}
	if s.cfg.ReadOnly {
		return errors.New("AI service is read-only")
	}

	_, err := s.host.Invoke(context.Background(), func() (any, error) {
		if err := s.mutateFiberWithLocator(loc, fn); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (s *Service) mutateFiberWithLocator(loc snapshotpkg.NodeLocator, fn func(fiber *rtui.Fiber) error) error {
	root := snapshotpkg.GetFiberRootFromComponent(s.host.GetRoot())
	if root == nil {
		return errors.New("fiber root unavailable")
	}
	fiber := rtui.FindFiberByID(root, loc.NodeID)
	if fiber == nil {
		return fmt.Errorf("fiber not found: %d", loc.NodeID)
	}
	if err := fn(fiber); err != nil {
		return err
	}
	s.syncOverridesFromFiber(loc.ComponentID, fiber)
	return nil
}

func findOptionIndexByValue(fiber *rtui.Fiber, value string) (int, bool) {
	if fiber == nil || fiber.Props == nil {
		return 0, false
	}
	raw, ok := fiber.Props["options"]
	if !ok {
		return 0, false
	}
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return 0, false
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Pointer && !elem.IsNil() {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		field := elem.FieldByName("Value")
		if field.IsValid() && field.Kind() == reflect.String && field.String() == value {
			return i, true
		}
	}
	return 0, false
}

func selectedIndexEquals(fiber *rtui.Fiber, want int) bool {
	if fiber == nil || fiber.Instance == nil {
		return false
	}
	getter, ok := fiber.Instance.(interface{ GetSelectedIndex() int })
	return ok && getter.GetSelectedIndex() == want
}

func stringMapValue(in map[string]interface{}, key string) (string, bool) {
	value, ok := in[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	return text, ok
}

func findSelectionIndexForString(fiber *rtui.Fiber, target string) (int, bool) {
	if fiber == nil || fiber.Instance == nil {
		return 0, false
	}
	if getter, ok := fiber.Instance.(interface{ GetRows() []string }); ok {
		rows := getter.GetRows()
		for i, row := range rows {
			if row == target {
				return i, true
			}
		}
		for i, row := range rows {
			if strings.Contains(row, target) {
				return i, true
			}
		}
	}
	if index, ok := findTreeNodeIndexByString(fiber.Instance, "GetVisibleNodes", target); ok {
		return index, true
	}
	if index, ok := findTreeNodeIndexByString(fiber.Instance, "GetNodes", target); ok {
		return index, true
	}
	propsProvider, ok := fiber.Instance.(interface{ GetProps() rtui.Props })
	if !ok {
		return 0, false
	}
	props := propsProvider.GetProps()
	if rows, ok := props["rows"].([][]string); ok {
		for i, row := range rows {
			for _, cell := range row {
				if cell == target {
					return i, true
				}
			}
		}
		for i, row := range rows {
			for _, cell := range row {
				if strings.Contains(cell, target) {
					return i, true
				}
			}
		}
	}
	return 0, false
}

func findTreeNodeIndexByString(inst interface{}, methodName, target string) (int, bool) {
	if inst == nil {
		return 0, false
	}
	method := reflect.ValueOf(inst).MethodByName(methodName)
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return 0, false
	}
	results := method.Call(nil)
	if len(results) != 1 {
		return 0, false
	}
	slice := results[0]
	if slice.Kind() != reflect.Slice {
		return 0, false
	}
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		if elem.Kind() == reflect.Pointer && !elem.IsNil() {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		if field := elem.FieldByName("Path"); field.IsValid() && field.Kind() == reflect.String && field.String() == target {
			return i, true
		}
		if field := elem.FieldByName("Content"); field.IsValid() && field.Kind() == reflect.String && field.String() == target {
			return i, true
		}
	}
	return 0, false
}

func numericMapValue(in map[string]interface{}, key string) (int, bool) {
	value, ok := in[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func waitConditionFromMap(in map[string]any) WaitCondition {
	cond := WaitCondition{
		Locator:       asMapString(in["locator"]),
		Selector:      asMapString(in["selector"]),
		ComponentID:   asMapString(in["component_id"]),
		ComponentType: asMapString(in["component_type"]),
		Key:           asMapString(in["key"]),
		Equals:        in["equals"],
	}
	if v, ok := in["visible"].(bool); ok {
		cond.Visible = &v
	}
	if v, ok := in["disabled"].(bool); ok {
		cond.Disabled = &v
	}
	if v, ok := in["exists"].(bool); ok {
		cond.Exists = &v
	}
	if v, ok := numericMapValueAny(in["render_seq_at_least"]); ok {
		cond.RenderSeqAtLeast = uint64(v)
	}
	if anyItems, ok := in["any"].([]interface{}); ok {
		cond.Any = make([]WaitCondition, 0, len(anyItems))
		for _, item := range anyItems {
			if child, ok := item.(map[string]interface{}); ok {
				cond.Any = append(cond.Any, waitConditionFromMap(child))
			}
		}
	}
	if anyItems, ok := in["any"].([]map[string]any); ok {
		cond.Any = make([]WaitCondition, 0, len(anyItems))
		for _, child := range anyItems {
			cond.Any = append(cond.Any, waitConditionFromMap(child))
		}
	}
	if allItems, ok := in["all"].([]interface{}); ok {
		cond.All = make([]WaitCondition, 0, len(allItems))
		for _, item := range allItems {
			if child, ok := item.(map[string]interface{}); ok {
				cond.All = append(cond.All, waitConditionFromMap(child))
			}
		}
	}
	if allItems, ok := in["all"].([]map[string]any); ok {
		cond.All = make([]WaitCondition, 0, len(allItems))
		for _, child := range allItems {
			cond.All = append(cond.All, waitConditionFromMap(child))
		}
	}
	if notItem, ok := in["not"].(map[string]interface{}); ok {
		notCond := waitConditionFromMap(notItem)
		cond.Not = &notCond
	}
	return cond
}

func numericMapValueAny(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}

func asMapString(v any) string {
	if text, ok := v.(string); ok {
		return text
	}
	return ""
}

func callSelectedNodeValue(inst interface{}) (interface{}, bool) {
	if inst == nil {
		return nil, false
	}
	method := reflect.ValueOf(inst).MethodByName("GetSelectedNode")
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 2 {
		return nil, false
	}
	results := method.Call(nil)
	if len(results) != 2 || results[1].Kind() != reflect.Bool || !results[1].Bool() {
		return nil, false
	}
	return results[0].Interface(), true
}

func (s *Service) installOverrideHookLocked() {
	if s.host == nil || s.hookInstalled {
		return
	}
	root := s.host.GetRoot()
	if root == nil {
		return
	}
	provider, ok := root.(interface {
		GetHooks() *runtimerender.HookManager
	})
	if !ok {
		return
	}
	hooks := provider.GetHooks()
	if hooks == nil {
		return
	}
	hooks.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		s.applyOverrides(vnode)
		return vnode
	})
	s.hookInstalled = true
}

func (s *Service) setOverride(componentID, key string, value interface{}) {
	if componentID == "" || key == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	props := s.overrides[componentID]
	if props == nil {
		props = make(map[string]interface{})
		s.overrides[componentID] = props
	}
	props[key] = value
}

func (s *Service) setOverrideValueMap(componentID, propKey, field string, value interface{}) {
	if componentID == "" || propKey == "" || field == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	props := s.overrides[componentID]
	if props == nil {
		props = make(map[string]interface{})
		s.overrides[componentID] = props
	}
	current, _ := props[propKey].(map[string]interface{})
	if current == nil {
		current = make(map[string]interface{})
	}
	current[field] = value
	props[propKey] = current
}

func (s *Service) applyOverrides(vnode rtui.VNode) {
	if vnode == nil {
		return
	}
	if override := s.overrideForVNode(vnode); override != nil {
		props := cloneMap(map[string]interface{}(vnode.Props()))
		if props == nil {
			props = make(map[string]interface{})
		}
		for key, value := range override {
			props[key] = value
		}
		vnode.SetProps(props)
	}
	for _, child := range vnode.Children() {
		s.applyOverrides(child)
	}
}

func (s *Service) overrideForVNode(vnode rtui.VNode) map[string]interface{} {
	if s == nil || vnode == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, key := range []string{vnode.ID(), vnode.Key()} {
		if key == "" {
			continue
		}
		if override := s.overrides[key]; len(override) > 0 {
			return cloneMap(override)
		}
	}
	return nil
}

func (s *Service) syncOverridesFromFiber(componentID string, fiber *rtui.Fiber) {
	if componentID == "" || fiber == nil || fiber.Instance == nil {
		return
	}
	updates := make(map[string]interface{})

	if getter, ok := fiber.Instance.(interface{ GetValue() string }); ok {
		updates["value"] = getter.GetValue()
	}
	if getter, ok := fiber.Instance.(interface{ GetValues() map[string]interface{} }); ok {
		updates["values"] = getter.GetValues()
	}
	if getter, ok := fiber.Instance.(interface{ GetSelectedIndex() int }); ok {
		updates["selectedIndex"] = getter.GetSelectedIndex()
	}
	if getter, ok := fiber.Instance.(interface{ GetCheckedIndices() []int }); ok {
		updates["checkedIndices"] = getter.GetCheckedIndices()
	}
	if getter, ok := fiber.Instance.(interface{ SelectedValue() string }); ok {
		updates["selected"] = getter.SelectedValue()
	}
	if getter, ok := fiber.Instance.(interface{ SelectedValues() []string }); ok {
		updates["selecteds"] = getter.SelectedValues()
	}
	if getter, ok := fiber.Instance.(interface {
		GetSelectedRow() (string, bool)
	}); ok {
		if row, ok := getter.GetSelectedRow(); ok {
			updates["selectedRow"] = row
		}
	}
	if getter, ok := fiber.Instance.(interface {
		GetSelectedRow() ([]string, bool)
	}); ok {
		if row, ok := getter.GetSelectedRow(); ok {
			updates["selectedRow"] = row
		}
	}
	if node, ok := callSelectedNodeValue(fiber.Instance); ok {
		updates["selectedNode"] = node
	}
	if getter, ok := fiber.Instance.(interface {
		GetProp(key string) (interface{}, bool)
	}); ok {
		for _, key := range []string{"selectedIndex", "checked", "disabled"} {
			if value, ok := getter.GetProp(key); ok {
				updates[key] = value
			}
		}
	}
	if len(updates) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	props := s.overrides[componentID]
	if props == nil {
		props = make(map[string]interface{})
		s.overrides[componentID] = props
	}
	for key, value := range updates {
		props[key] = value
	}
}
