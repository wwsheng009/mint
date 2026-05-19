package e2e

import (
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/framework"
	runtimeaction "github.com/wwsheng009/mint/runtime/action"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

const defaultTimeout = 2 * time.Second

// App is the Phase 1 interactive E2E harness.
type App struct {
	testApp *ui.TestableApp
	runtime *runtimeintent.Runtime
	driver  *Driver
	timeout time.Duration

	traceMu          sync.RWMutex
	rawInputs        []RawInputEvent
	messageEvents    []MessageEvent
	actionEvents     []ActionEvent
	traceEvent       []TraceEvent
	focusTransitions []FocusTransition
}

// Run starts an interactive E2E app using the standard test runtime.
func Run(app ui.ComponentFunc, opts ...ui.Option) (*App, error) {
	testApp, err := ui.RunTest(app, opts...)
	if err != nil {
		return nil, err
	}
	return newApp(testApp)
}

// RunWithSandbox starts an interactive E2E app backed by MockSandbox.
func RunWithSandbox(app ui.ComponentFunc, opts ...ui.Option) (*App, error) {
	testApp, err := ui.RunTestWithSandbox(app, opts...)
	if err != nil {
		return nil, err
	}
	return newApp(testApp)
}

func newApp(testApp *ui.TestableApp) (*App, error) {
	runtime := rtui.GetGlobalIntentRuntime()
	if runtime == nil || runtime.Dispatcher == nil {
		_ = testApp.Close()
		return nil, fmt.Errorf("global intent runtime not initialized")
	}
	runtime.Dispatcher.EnableLog(true)
	runtime.Dispatcher.ClearLogs()
	app := &App{
		testApp:          testApp,
		runtime:          runtime,
		timeout:          defaultTimeout,
		rawInputs:        make([]RawInputEvent, 0, 16),
		messageEvents:    make([]MessageEvent, 0, 16),
		actionEvents:     make([]ActionEvent, 0, 16),
		traceEvent:       make([]TraceEvent, 0, 32),
		focusTransitions: make([]FocusTransition, 0, 16),
	}
	if fwApp := testApp.GetFrameworkApp(); fwApp != nil {
		fwApp.SetTestMessageProbe(app.recordMessage)
		fwApp.SetTestActionProbe(app.recordAction)
	}
	app.driver = &Driver{app: app}
	if err := app.AwaitIdle(); err != nil {
		_ = app.Close()
		return nil, err
	}
	return app, nil
}

// Close closes the E2E app.
func (a *App) Close() error {
	if a == nil || a.testApp == nil {
		return nil
	}
	return a.testApp.Close()
}

// Driver returns the interactive driver.
func (a *App) Driver() *Driver {
	return a.driver
}

// FrameworkApp exposes the underlying framework app.
func (a *App) FrameworkApp() *framework.App {
	return a.testApp.GetFrameworkApp()
}

// IntentRuntime exposes the underlying intent runtime.
func (a *App) IntentRuntime() *runtimeintent.Runtime {
	return a.runtime
}

// RenderString returns the current rendered text snapshot.
func (a *App) RenderString() string {
	return a.testApp.GetRenderString()
}

// ForceRender forces an immediate render.
func (a *App) ForceRender() {
	a.testApp.ForceRender()
}

// HitMap returns the latest app HitMap.
func (a *App) HitMap() *runtimeevent.HitMap {
	if a.FrameworkApp() == nil {
		return nil
	}
	return a.FrameworkApp().GetHitMap()
}

// IntentLogs returns recent intent dispatch logs.
func (a *App) IntentLogs() []runtimeintent.DispatchLog {
	if a.runtime == nil || a.runtime.Dispatcher == nil {
		return nil
	}
	return a.runtime.Dispatcher.GetLogs()
}

// ClearIntentLogs clears recent intent dispatch logs.
func (a *App) ClearIntentLogs() {
	if a.runtime == nil || a.runtime.Dispatcher == nil {
		return
	}
	a.runtime.Dispatcher.ClearLogs()
}

// RawInputs returns the injected raw input trace.
func (a *App) RawInputs() []RawInputEvent {
	a.traceMu.RLock()
	defer a.traceMu.RUnlock()
	out := make([]RawInputEvent, len(a.rawInputs))
	copy(out, a.rawInputs)
	return out
}

// ClearRawInputs clears the raw input trace.
func (a *App) ClearRawInputs() {
	a.traceMu.Lock()
	defer a.traceMu.Unlock()
	a.rawInputs = a.rawInputs[:0]
	a.messageEvents = a.messageEvents[:0]
	a.actionEvents = a.actionEvents[:0]
	a.traceEvent = a.traceEvent[:0]
	a.focusTransitions = a.focusTransitions[:0]
}

// MessageEvents returns captured Msg trace events.
func (a *App) MessageEvents() []MessageEvent {
	a.traceMu.RLock()
	defer a.traceMu.RUnlock()
	out := make([]MessageEvent, len(a.messageEvents))
	copy(out, a.messageEvents)
	return out
}

// ActionEvents returns captured Action trace events.
func (a *App) ActionEvents() []ActionEvent {
	a.traceMu.RLock()
	defer a.traceMu.RUnlock()
	out := make([]ActionEvent, len(a.actionEvents))
	copy(out, a.actionEvents)
	return out
}

// TraceEvents returns merged raw-input and intent trace events ordered by timestamp.
func (a *App) TraceEvents() []TraceEvent {
	a.traceMu.RLock()
	localEvents := make([]TraceEvent, len(a.traceEvent))
	copy(localEvents, a.traceEvent)
	a.traceMu.RUnlock()

	intentEvents := make([]TraceEvent, 0, len(a.IntentLogs()))
	for _, logEntry := range a.IntentLogs() {
		intentEvents = append(intentEvents, TraceEvent{
			Kind:      TraceIntentDispatch,
			Name:      logEntry.Type,
			Timestamp: logEntry.Timestamp,
			Payload:   logEntry,
		})
	}

	merged := append(localEvents, intentEvents...)
	sortTraceEvents(merged)
	return merged
}

// AwaitIdle waits until render/focus/intent log state stabilizes.
func (a *App) AwaitIdle() error {
	return a.AwaitIdleFor(a.timeout)
}

// AwaitIdleFor waits for a bounded time until the app state stabilizes.
func (a *App) AwaitIdleFor(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	stableRounds := 0
	var prev idleSnapshot

	for time.Now().Before(deadline) {
		if a.runtime != nil && a.runtime.Dispatcher != nil {
			a.runtime.Dispatcher.ProcessQueue(5 * time.Millisecond)
		}
		snapshot := a.captureIdleSnapshot()
		if snapshot == prev {
			stableRounds++
			if stableRounds >= 3 {
				return nil
			}
		} else {
			stableRounds = 0
			prev = snapshot
		}
		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("await idle timed out after %s", timeout)
}

// AwaitIntent waits until an intent with the given type appears in dispatch logs.
func (a *App) AwaitIntent(intentType string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, logEntry := range a.IntentLogs() {
			if logEntry.Type == intentType {
				return nil
			}
		}
		if err := a.AwaitIdleFor(100 * time.Millisecond); err != nil {
			// Keep polling until overall timeout.
		}
	}
	return fmt.Errorf("intent %q not observed within %s", intentType, timeout)
}

// AwaitMessage waits until a message with the given name appears.
func (a *App) AwaitMessage(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, event := range a.MessageEvents() {
			if event.Name == name {
				return nil
			}
		}
		if err := a.AwaitIdleFor(100 * time.Millisecond); err != nil {
			// Keep polling until timeout.
		}
	}
	return fmt.Errorf("message %q not observed within %s", name, timeout)
}

// AwaitAction waits until an action of the given type appears.
func (a *App) AwaitAction(actionType runtimeaction.ActionType, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, event := range a.ActionEvents() {
			if event.Type == actionType {
				return nil
			}
		}
		if err := a.AwaitIdleFor(100 * time.Millisecond); err != nil {
			// Keep polling until timeout.
		}
	}
	return fmt.Errorf("action %q not observed within %s", actionType, timeout)
}

// AwaitTrace waits until a trace event matching the given matcher appears.
func (a *App) AwaitTrace(match TraceMatch, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, event := range a.TraceEvents() {
			if match.Match(event) {
				return nil
			}
		}
		if err := a.AwaitIdleFor(100 * time.Millisecond); err != nil {
			// Keep polling until timeout.
		}
	}
	return fmt.Errorf("trace match %+v not observed within %s", match, timeout)
}

// AwaitFocus waits until the current focus matches the locator.
func (a *App) AwaitFocus(locator Locator, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := a.AwaitIdleFor(100 * time.Millisecond); err != nil {
			// Keep polling until timeout.
		}
		if err := a.AssertFocus(locator); err == nil {
			return nil
		}
	}
	return fmt.Errorf("focus %v not observed within %s", locator, timeout)
}

// Eventually retries the assertion function until it passes or times out.
func (a *App) Eventually(timeout time.Duration, interval time.Duration, fn func(*App) error) error {
	if interval <= 0 {
		interval = 20 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := fn(a); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(interval)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("condition not met")
	}
	return fmt.Errorf("eventually timed out after %s: %w", timeout, lastErr)
}

// FocusSnapshot returns the current focused fiber snapshot.
func (a *App) FocusSnapshot() (FocusSnapshot, bool) {
	root := a.testApp.GetDeclarativeRoot()
	if root == nil {
		return FocusSnapshot{}, false
	}
	focusMgr := root.GetFocusManager()
	if focusMgr == nil {
		return FocusSnapshot{}, false
	}
	fiber := focusMgr.GetCurrent()
	if fiber == nil {
		return FocusSnapshot{}, false
	}
	snapshot := FocusSnapshot{
		Index:       focusMgr.CurrentIndex(),
		Type:        int(fiber.Type),
		NodeID:      fiber.NodeID,
		ComponentID: fiberComponentID(fiber),
		TargetID:    fiber.ActionTargetID,
		Key:         fiber.Key,
		ID:          fiber.ID,
		Tag:         fiber.Tag,
		Focused:     true,
		FocusCount:  focusMgr.Count(),
	}
	if fiber.Instance != nil {
		if ctx := fiber.Instance.GetContext(); ctx != nil {
			snapshot.ComponentContextID = ctx.ComponentID
		}
	}
	if hitMap := a.HitMap(); hitMap != nil {
		if entry := hitMap.FindByID(fiber.NodeID); entry != nil {
			snapshot.Bounds = entry.Bounds
		}
	}
	return snapshot, true
}

// RootFiber returns the current Fiber root if available.
func (a *App) RootFiber() *rtui.Fiber {
	root := a.testApp.GetDeclarativeRoot()
	if root == nil {
		return nil
	}
	return root.GetFiberRoot()
}

// BoundsOf resolves a locator to the rendered bounds.
func (a *App) BoundsOf(locator Locator) (layout.Rect, error) {
	switch locator.kind {
	case locatorFocused:
		focus, ok := a.FocusSnapshot()
		if !ok {
			return layout.Rect{}, fmt.Errorf("no focused element")
		}
		return focus.Bounds, nil
	case locatorAt:
		return layout.Rect{X: locator.x, Y: locator.y, Width: 1, Height: 1}, nil
	case locatorText:
		point, err := a.findTextPoint(locator.value)
		if err != nil {
			return layout.Rect{}, err
		}
		return layout.Rect{X: point.X, Y: point.Y, Width: 1, Height: 1}, nil
	default:
		fiber, err := a.ResolveFiber(locator)
		if err != nil {
			return layout.Rect{}, err
		}
		if hitMap := a.HitMap(); hitMap != nil {
			if entry := hitMap.FindByID(fiber.NodeID); entry != nil {
				return entry.Bounds, nil
			}
		}
		return layout.Rect{}, fmt.Errorf("bounds unavailable for locator %q", locator.kind)
	}
}

// CellAt returns a structured snapshot of one rendered cell.
func (a *App) CellAt(x, y int) (CellSnapshot, error) {
	buffer := a.testApp.GetBuffer()
	if buffer == nil {
		return CellSnapshot{}, fmt.Errorf("render buffer unavailable")
	}
	if y < 0 || y >= buffer.Height || x < 0 || x >= buffer.Width {
		return CellSnapshot{}, fmt.Errorf("cell (%d,%d) out of bounds", x, y)
	}
	cell := buffer.Cells[y][x]
	return CellSnapshot{
		X:              x,
		Y:              y,
		Cluster:        cell.Cluster,
		Style:          cell.Style,
		Width:          cell.Width,
		IsContinuation: cell.IsContinuation,
	}, nil
}

// AssertFocus asserts the current focus matches a locator.
func (a *App) AssertFocus(locator Locator) error {
	focus, ok := a.FocusSnapshot()
	if !ok {
		return fmt.Errorf("no focused element")
	}
	if err := locator.matchFocus(focus); err != nil {
		return err
	}
	return nil
}

// FocusTransitions returns recorded focus transitions.
func (a *App) FocusTransitions() []FocusTransition {
	a.traceMu.RLock()
	defer a.traceMu.RUnlock()
	out := make([]FocusTransition, len(a.focusTransitions))
	copy(out, a.focusTransitions)
	return out
}

// ClearFocusTransitions clears recorded focus transitions.
func (a *App) ClearFocusTransitions() {
	a.traceMu.Lock()
	defer a.traceMu.Unlock()
	a.focusTransitions = a.focusTransitions[:0]
}

// AssertFocusTransition asserts a recorded transition from one locator to another exists.
func (a *App) AssertFocusTransition(from Locator, to Locator) error {
	for _, transition := range a.FocusTransitions() {
		if err := from.matchFocus(transition.From); err != nil {
			continue
		}
		if err := to.matchFocus(transition.To); err != nil {
			continue
		}
		return nil
	}
	return fmt.Errorf("focus transition %v -> %v not found", from, to)
}

// AssertVisible asserts a locator resolves to visible content/bounds.
func (a *App) AssertVisible(locator Locator) error {
	switch locator.kind {
	case locatorText:
		_, err := a.findTextPoint(locator.value)
		return err
	default:
		bounds, err := a.BoundsOf(locator)
		if err != nil {
			return err
		}
		if bounds.Width <= 0 || bounds.Height <= 0 {
			return fmt.Errorf("locator %q resolved to non-visible bounds %v", locator.kind, bounds)
		}
		return nil
	}
}

// AssertHit asserts that a point hits the expected locator.
func (a *App) AssertHit(point Point, locator Locator) error {
	hitMap := a.HitMap()
	if hitMap == nil {
		return fmt.Errorf("hitmap unavailable")
	}
	entry := hitMap.HitTest(point.X, point.Y)
	if entry == nil {
		return fmt.Errorf("no hit at (%d,%d)", point.X, point.Y)
	}

	switch locator.kind {
	case locatorTargetID:
		if entry.TargetFiber == nil || entry.TargetFiber.GetActionTargetID() != locator.value {
			return fmt.Errorf("hit targetID = %q, want %q", safeTargetID(entry), locator.value)
		}
		return nil
	case locatorAt:
		if !entry.Bounds.Contains(locator.x, locator.y) {
			return fmt.Errorf("hit bounds %v do not contain point (%d,%d)", entry.Bounds, locator.x, locator.y)
		}
		return nil
	}

	fiber, err := a.ResolveFiber(locator)
	if err != nil {
		return err
	}
	if entry.NodeID != fiber.NodeID {
		return fmt.Errorf("hit nodeID = %d, want %d", entry.NodeID, fiber.NodeID)
	}
	return nil
}

// AssertTargetID resolves a locator to a point and asserts the hit target ID there.
func (a *App) AssertTargetID(locator Locator, targetID string) error {
	point, err := a.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return a.AssertHit(point, ByTargetID(targetID))
}

// AssertBounds asserts the locator resolves to the expected bounds.
func (a *App) AssertBounds(locator Locator, expect layout.Rect) error {
	actual, err := a.BoundsOf(locator)
	if err != nil {
		return err
	}
	if actual != expect {
		return fmt.Errorf("bounds = %v, want %v", actual, expect)
	}
	return nil
}

// AssertCellStyleAt asserts style at a specific cell.
func (a *App) AssertCellStyleAt(x, y int, expect StyleExpect) error {
	cell, err := a.CellAt(x, y)
	if err != nil {
		return err
	}
	return expect.Match(cell.Style)
}

// AssertIntentSequence asserts the observed intent log contains the given subsequence in order.
func (a *App) AssertIntentSequence(intentTypes ...string) error {
	if len(intentTypes) == 0 {
		return nil
	}
	logs := a.IntentLogs()
	index := 0
	for _, logEntry := range logs {
		if logEntry.Type == intentTypes[index] {
			index++
			if index == len(intentTypes) {
				return nil
			}
		}
	}
	return fmt.Errorf("intent sequence %v not found in logs", intentTypes)
}

// LastIntentLog returns the most recent intent dispatch log.
func (a *App) LastIntentLog() (runtimeintent.DispatchLog, bool) {
	logs := a.IntentLogs()
	if len(logs) == 0 {
		return runtimeintent.DispatchLog{}, false
	}
	return logs[len(logs)-1], true
}

// AssertLastIntent asserts the latest intent matches the expected type.
func (a *App) AssertLastIntent(intentType string) error {
	logEntry, ok := a.LastIntentLog()
	if !ok {
		return fmt.Errorf("no intent logs recorded")
	}
	if logEntry.Type != intentType {
		return fmt.Errorf("last intent = %q, want %q", logEntry.Type, intentType)
	}
	return nil
}

// AssertIntentHandled asserts that at least one intent of the given type was handled successfully.
func (a *App) AssertIntentHandled(intentType string) error {
	for _, logEntry := range a.IntentLogs() {
		if logEntry.Type != intentType {
			continue
		}
		if logEntry.Handled && logEntry.Error == nil {
			return nil
		}
	}
	return fmt.Errorf("handled intent %q not found", intentType)
}

// LastMessage returns the most recent captured Msg event.
func (a *App) LastMessage() (MessageEvent, bool) {
	events := a.MessageEvents()
	if len(events) == 0 {
		return MessageEvent{}, false
	}
	return events[len(events)-1], true
}

// LastAction returns the most recent captured Action event.
func (a *App) LastAction() (ActionEvent, bool) {
	events := a.ActionEvents()
	if len(events) == 0 {
		return ActionEvent{}, false
	}
	return events[len(events)-1], true
}

// AssertLastMessage asserts the latest message matches the expected name.
func (a *App) AssertLastMessage(name string) error {
	event, ok := a.LastMessage()
	if !ok {
		return fmt.Errorf("no message events recorded")
	}
	if event.Name != name {
		return fmt.Errorf("last message = %q, want %q", event.Name, name)
	}
	return nil
}

// AssertLastAction asserts the latest action matches the expected type.
func (a *App) AssertLastAction(actionType runtimeaction.ActionType) error {
	event, ok := a.LastAction()
	if !ok {
		return fmt.Errorf("no action events recorded")
	}
	if event.Type != actionType {
		return fmt.Errorf("last action = %q, want %q", event.Type, actionType)
	}
	return nil
}

// AssertActionHandled asserts that at least one action of the given type was handled in the expected stage.
func (a *App) AssertActionHandled(actionType runtimeaction.ActionType, stage string) error {
	for _, event := range a.ActionEvents() {
		if event.Type != actionType {
			continue
		}
		if !event.Handled {
			continue
		}
		if stage != "" && event.Stage != stage {
			continue
		}
		return nil
	}
	if stage != "" {
		return fmt.Errorf("handled action %q at stage %q not found", actionType, stage)
	}
	return fmt.Errorf("handled action %q not found", actionType)
}

// AssertTraceContains asserts at least one trace event matches the given matcher.
func (a *App) AssertTraceContains(match TraceMatch) error {
	for _, event := range a.TraceEvents() {
		if match.Match(event) {
			return nil
		}
	}
	return fmt.Errorf("trace match %+v not found", match)
}

// AssertTraceSequence asserts the merged trace contains the given subsequence in order.
func (a *App) AssertTraceSequence(matches ...TraceMatch) error {
	if len(matches) == 0 {
		return nil
	}
	trace := a.TraceEvents()
	index := 0
	for _, event := range trace {
		if matches[index].Match(event) {
			index++
			if index == len(matches) {
				return nil
			}
		}
	}
	return fmt.Errorf("trace sequence %+v not found", matches)
}

// AssertMessageSequence asserts the observed Msg sequence contains the given subsequence in order.
func (a *App) AssertMessageSequence(names ...string) error {
	if len(names) == 0 {
		return nil
	}
	events := a.MessageEvents()
	index := 0
	for _, event := range events {
		if event.Name == names[index] {
			index++
			if index == len(names) {
				return nil
			}
		}
	}
	return fmt.Errorf("message sequence %v not found", names)
}

// AssertActionSequence asserts the observed Action sequence contains the given subsequence in order.
func (a *App) AssertActionSequence(types ...runtimeaction.ActionType) error {
	if len(types) == 0 {
		return nil
	}
	events := a.ActionEvents()
	index := 0
	for _, event := range events {
		if event.Type == types[index] {
			index++
			if index == len(types) {
				return nil
			}
		}
	}
	return fmt.Errorf("action sequence %v not found", types)
}

// AssertStyle resolves a locator to a point and checks the cell style there.
func (a *App) AssertStyle(locator Locator, expect StyleExpect) error {
	point, err := a.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return a.AssertCellStyleAt(point.X, point.Y, expect)
}

// ResolvePoint resolves a locator into a clickable point.
func (a *App) ResolvePoint(locator Locator) (Point, error) {
	switch locator.kind {
	case locatorAt:
		return Point{X: locator.x, Y: locator.y}, nil
	case locatorText:
		return a.findTextPoint(locator.value)
	case locatorComponentID, locatorID, locatorKey, locatorTag:
		fiber, err := a.ResolveFiber(locator)
		if err != nil {
			return Point{}, err
		}
		return a.hitMapCenter(func(entry *runtimeevent.HitMapEntry) bool {
			return entry.NodeID == fiber.NodeID
		})
	case locatorTargetID:
		return a.hitMapCenter(func(entry *runtimeevent.HitMapEntry) bool {
			return entry.TargetFiber != nil && entry.TargetFiber.GetActionTargetID() == locator.value
		})
	case locatorFocused:
		focus, ok := a.FocusSnapshot()
		if !ok {
			return Point{}, fmt.Errorf("no focused element")
		}
		return Point{
			X: focus.Bounds.X + maxInt(0, focus.Bounds.Width/2),
			Y: focus.Bounds.Y + maxInt(0, focus.Bounds.Height/2),
		}, nil
	default:
		return Point{}, fmt.Errorf("unsupported locator kind %q", locator.kind)
	}
}

// ResolveFiber resolves a locator to a Fiber when possible.
func (a *App) ResolveFiber(locator Locator) (*rtui.Fiber, error) {
	switch locator.kind {
	case locatorFocused:
		root := a.testApp.GetDeclarativeRoot()
		if root == nil || root.GetFocusManager() == nil {
			return nil, fmt.Errorf("focus manager unavailable")
		}
		fiber := root.GetFocusManager().GetCurrent()
		if fiber == nil {
			return nil, fmt.Errorf("no focused fiber")
		}
		return fiber, nil
	case locatorTargetID:
		if hitMap := a.HitMap(); hitMap != nil {
			for _, entry := range hitMap.AllEntries() {
				entry := entry
				if entry.TargetFiber != nil && entry.TargetFiber.GetActionTargetID() == locator.value {
					if root := a.RootFiber(); root != nil {
						return rtui.FindFiberByID(root, entry.NodeID), nil
					}
				}
			}
		}
		return nil, fmt.Errorf("targetID %q not found", locator.value)
	}

	root := a.RootFiber()
	if root == nil {
		return nil, fmt.Errorf("fiber root unavailable")
	}

	var found *rtui.Fiber
	rtui.WalkFiberDepthFirst(root, func(fiber *rtui.Fiber) bool {
		if matchFiber(locator, fiber) {
			found = fiber
			return false
		}
		return true
	})
	if found == nil {
		return nil, fmt.Errorf("fiber not found for locator %q(%s)", locator.kind, locator.value)
	}
	return found, nil
}

func (a *App) hitMapCenter(match func(entry *runtimeevent.HitMapEntry) bool) (Point, error) {
	hitMap := a.HitMap()
	if hitMap == nil {
		return Point{}, fmt.Errorf("hitmap unavailable")
	}
	for _, entry := range hitMap.AllEntries() {
		entry := entry
		if !match(&entry) {
			continue
		}
		return Point{
			X: entry.Bounds.X + maxInt(0, entry.Bounds.Width/2),
			Y: entry.Bounds.Y + maxInt(0, entry.Bounds.Height/2),
		}, nil
	}
	return Point{}, fmt.Errorf("locator not found in hitmap")
}

func (a *App) findTextPoint(text string) (Point, error) {
	buffer := a.testApp.GetBuffer()
	if buffer == nil {
		return Point{}, fmt.Errorf("render buffer unavailable")
	}
	for y := 0; y < buffer.Height; y++ {
		rowText, positions := rowTextAndPositions(buffer, y)
		index := findSubstring(rowText, text)
		if index < 0 {
			continue
		}
		if index >= len(positions) {
			return Point{}, fmt.Errorf("text %q found at rune index %d on row %d, but row position map has length %d", text, index, y, len(positions))
		}
		x := positions[index]
		return Point{X: x, Y: y}, nil
	}
	return Point{}, fmt.Errorf("text %q not found", text)
}

type idleSnapshot struct {
	renderHash  uint64
	focusIndex  int
	focusType   int
	intentCount int
	hitMapSize  int
}

func (a *App) captureIdleSnapshot() idleSnapshot {
	snapshot := idleSnapshot{
		focusIndex:  a.testApp.GetFocusedIndex(),
		focusType:   a.testApp.GetFocusedType(),
		intentCount: len(a.IntentLogs()),
	}
	if hitMap := a.HitMap(); hitMap != nil {
		snapshot.hitMapSize = hitMap.Size()
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(a.RenderString()))
	snapshot.renderHash = hasher.Sum64()
	return snapshot
}

func rowTextAndPositions(buffer *paint.Buffer, y int) (string, []int) {
	if buffer == nil || y < 0 || y >= buffer.Height {
		return "", nil
	}
	var builder []rune
	positions := make([]int, 0, buffer.Width)
	for x := 0; x < buffer.Width; x++ {
		cell := buffer.Cells[y][x]
		if cell.IsContinuation {
			continue
		}
		cluster := cell.Cluster
		if cluster == "" {
			cluster = " "
		}
		for _, r := range cluster {
			builder = append(builder, r)
			positions = append(positions, x)
		}
	}
	return string(builder), positions
}

func findSubstring(source, target string) int {
	if target == "" {
		return -1
	}
	byteIndex := strings.Index(source, target)
	if byteIndex < 0 {
		return -1
	}
	return len([]rune(source[:byteIndex]))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func safeTargetID(entry *runtimeevent.HitMapEntry) string {
	if entry == nil || entry.TargetFiber == nil {
		return ""
	}
	return entry.TargetFiber.GetActionTargetID()
}

func (a *App) recordRawInput(raw platform.RawInput, label string) {
	if raw.Timestamp.IsZero() {
		raw.Timestamp = time.Now()
	}
	rawEvent := RawInputEvent{
		Label:     label,
		Raw:       raw,
		Timestamp: raw.Timestamp,
	}
	trace := TraceEvent{
		Kind:      TraceRawInput,
		Name:      rawEvent.Name(),
		Timestamp: rawEvent.Timestamp,
		Payload:   rawEvent,
	}
	a.traceMu.Lock()
	defer a.traceMu.Unlock()
	a.rawInputs = append(a.rawInputs, rawEvent)
	a.traceEvent = append(a.traceEvent, trace)
}

func (a *App) recordMessage(msg runtimemsg.Msg) {
	if msg == nil {
		return
	}
	event := MessageEvent{
		MsgType:    msg.Type(),
		Name:       string(msg.Type()),
		Timestamp:  time.Now(),
		SourceType: fmt.Sprintf("%T", msg),
	}
	a.traceMu.Lock()
	defer a.traceMu.Unlock()
	a.messageEvents = append(a.messageEvents, event)
	a.traceEvent = append(a.traceEvent, TraceEvent{
		Kind:      TraceMsg,
		Name:      event.Name,
		Timestamp: event.Timestamp,
		Payload:   event,
	})
}

func (a *App) recordAction(act *runtimeaction.Action, handled bool, stage string) {
	if act == nil {
		return
	}
	event := ActionEvent{
		Type:      act.Type,
		Name:      string(act.Type),
		Source:    act.Source,
		Target:    act.Target,
		TargetID:  act.TargetID,
		Handled:   handled,
		Stage:     stage,
		Timestamp: time.Now(),
	}
	a.traceMu.Lock()
	defer a.traceMu.Unlock()
	a.actionEvents = append(a.actionEvents, event)
	a.traceEvent = append(a.traceEvent, TraceEvent{
		Kind:      TraceAction,
		Name:      event.Name,
		Timestamp: event.Timestamp,
		Payload:   event,
	})
}

func (a *App) recordFocusTransition(from FocusSnapshot, fromOK bool, to FocusSnapshot, toOK bool) {
	if !fromOK && !toOK {
		return
	}
	if fromOK && toOK && from.Equal(to) {
		return
	}
	transition := FocusTransition{
		From:      from,
		FromOK:    fromOK,
		To:        to,
		ToOK:      toOK,
		Timestamp: time.Now(),
	}
	a.traceMu.Lock()
	defer a.traceMu.Unlock()
	a.focusTransitions = append(a.focusTransitions, transition)
	a.traceEvent = append(a.traceEvent, TraceEvent{
		Kind:      TraceFocusTransition,
		Name:      transition.Name(),
		Timestamp: transition.Timestamp,
		Payload:   transition,
	})
}

// RawInputEvent captures one injected raw input.
type RawInputEvent struct {
	Label     string
	Raw       platform.RawInput
	Timestamp time.Time
}

func (e RawInputEvent) Name() string {
	switch e.Raw.Type {
	case platform.InputKeyPress:
		if e.Raw.Special != platform.KeyUnknown {
			return "key:" + e.Raw.Special.String()
		}
		if e.Raw.Key != 0 {
			return "key:" + string(e.Raw.Key)
		}
		return "key"
	case platform.InputMouse:
		return fmt.Sprintf("mouse:%s:%s", mouseButtonString(e.Raw.MouseButton), mouseActionString(e.Raw.MouseAction))
	case platform.InputResize:
		return fmt.Sprintf("resize:%dx%d", e.Raw.Width, e.Raw.Height)
	default:
		return e.Label
	}
}

type TraceKind string

const (
	TraceRawInput        TraceKind = "raw_input"
	TraceMsg             TraceKind = "msg"
	TraceAction          TraceKind = "action"
	TraceIntentDispatch  TraceKind = "intent_dispatch"
	TraceFocusTransition TraceKind = "focus_transition"
)

// TraceEvent is the flattened Phase 1/2 trace event format.
type TraceEvent struct {
	Kind      TraceKind
	Name      string
	Timestamp time.Time
	Payload   any
}

// TraceMatch matches a subset of trace event fields.
type TraceMatch struct {
	Kind TraceKind
	Name string
}

func (m TraceMatch) Match(event TraceEvent) bool {
	if m.Kind != "" && event.Kind != m.Kind {
		return false
	}
	if m.Name != "" && event.Name != m.Name {
		return false
	}
	return true
}

// MessageEvent captures one observed Msg passed through framework.App.processMsg.
type MessageEvent struct {
	MsgType    runtimemsg.MsgType
	Name       string
	Timestamp  time.Time
	SourceType string
}

// ActionEvent captures one mapped Action observed during processMsg.
type ActionEvent struct {
	Type      runtimeaction.ActionType
	Name      string
	Source    string
	Target    string
	TargetID  uint64
	Handled   bool
	Stage     string
	Timestamp time.Time
}

// FocusTransition captures one focus move observed around a driver action.
type FocusTransition struct {
	From      FocusSnapshot
	FromOK    bool
	To        FocusSnapshot
	ToOK      bool
	Timestamp time.Time
}

func (t FocusTransition) Name() string {
	return fmt.Sprintf("%s->%s", focusLabel(t.From, t.FromOK), focusLabel(t.To, t.ToOK))
}

func focusLabel(snapshot FocusSnapshot, ok bool) string {
	if !ok {
		return "none"
	}
	if snapshot.ComponentID != "" {
		return snapshot.ComponentID
	}
	if snapshot.ID != "" {
		return snapshot.ID
	}
	if snapshot.Key != "" {
		return snapshot.Key
	}
	if snapshot.Tag != "" {
		return snapshot.Tag
	}
	return fmt.Sprintf("node-%d", snapshot.NodeID)
}

func sortTraceEvents(events []TraceEvent) {
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].Timestamp.Before(events[i].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}

func mouseButtonString(button platform.MouseButton) string {
	switch button {
	case platform.MouseLeft:
		return "left"
	case platform.MouseMiddle:
		return "middle"
	case platform.MouseRight:
		return "right"
	default:
		return "none"
	}
}

func mouseActionString(action platform.MouseAction) string {
	switch action {
	case platform.MousePress:
		return "press"
	case platform.MouseRelease:
		return "release"
	case platform.MouseMotion:
		return "move"
	case platform.MouseWheelUp:
		return "wheel_up"
	case platform.MouseWheelDown:
		return "wheel_down"
	default:
		return "unknown"
	}
}

func matchFiber(locator Locator, fiber *rtui.Fiber) bool {
	if fiber == nil {
		return false
	}
	switch locator.kind {
	case locatorComponentID:
		return fiberComponentID(fiber) == locator.value
	case locatorID:
		return fiber.ID == locator.value
	case locatorKey:
		return fiber.Key == locator.value
	case locatorTag:
		return fiber.Tag == locator.value
	default:
		return false
	}
}

func fiberComponentID(fiber *rtui.Fiber) string {
	if fiber == nil || fiber.Props == nil {
		return ""
	}
	if value, ok := fiber.Props["componentID"].(string); ok {
		return value
	}
	return ""
}

// Point is a resolved screen coordinate.
type Point struct {
	X int
	Y int
}

// CellSnapshot captures one rendered cell.
type CellSnapshot struct {
	X              int
	Y              int
	Cluster        string
	Style          style.Style
	Width          int
	IsContinuation bool
}
