# Complex Component Refactor Plan

## 1. Background

Mint currently supports three different patterns for complex interactive components:

1. Declarative popup subtree with external state
   - Example: `menu`
2. In-tree overlay sibling with local layout math
   - Example: `statusbar` overlay help / toolbar tooltip
3. Leaf instance with runtime registry + middleware + popup side channel
   - Example: previous `select` overlay popup implementation

Pattern 3 is the fragile one. It mixes:

- persistent `Instance` state
- runtime registry lookups
- middleware-based outside dismiss
- portal layout / anchor positioning
- hit testing on a popup that is not modeled as a first-class child subtree

This produces repeated regressions:

- popup rendering and hit-testing get out of sync
- direct target dispatch and middleware dismissal fight each other
- anchor positioning is harder to reason about
- component behavior depends on non-local runtime state

## 2. Refactor Goals

The refactor must satisfy the following:

1. Complex popup components must render their popup as a real child subtree.
2. Popup child trees may depend on persistent runtime state.
3. Popup interaction should use local instance-tree intent bubble or normal action routing.
4. Registries, if needed, should only be used for cross-cutting concerns such as outside-dismiss.
5. Popup rendering, layout, hit-testing, and dismissal must share the same declarative tree.

## 3. Extracted Design Rules

From the existing stable components:

### 3.1 Menu rule

`menu` is stable because the popup is declared as a normal subtree:

- trigger/menu bar is a normal node
- popup portal wrapper is a normal node
- popup surface is a normal node
- outside-click middleware only closes state

The middleware does not create popup rendering behavior.

### 3.2 Statusbar tooltip rule

`statusbar` overlay help is stable because the overlay is still a tree node:

- no external registry for rendering
- no extra hit-map synthesis for rendering
- position is computed from already-owned local bounds

### 3.3 New rule for popup controls

All popup controls should follow:

- trigger `Instance` owns persistent runtime state
- popup subtree is emitted declaratively from that runtime state
- popup child instance communicates back to trigger via local intent bubble
- middleware only handles outside-dismiss and other global concerns

## 4. Proposed Runtime Contract

Add a runtime interface:

```go
type RuntimeChildrenProvider interface {
    ComponentInstance
    RuntimeChildren() []VNode
}
```

Purpose:

- allow a persistent `Instance` to synthesize child VNodes from runtime state
- keep VNode declarative while still enabling popup/overlay subtrees
- avoid forcing complex components into ad hoc registries for rendering

Reconciler behavior:

- for element fibers, reconcile `Props["children"]`
- then append `Instance.RuntimeChildren()` if implemented

This keeps popup children inside the normal Fiber/Layout/Paint/HitMap pipeline.

## 5. Select Refactor Plan

### Phase 1 - Runtime child subtree support

Changes:

- add `RuntimeChildrenProvider` to `runtime/ui`
- update reconciler begin-work for element fibers to append runtime children

Expected result:

- `select` can generate popup subtree from persistent `Instance` state

### Phase 2 - Refactor `select` popup rendering

Changes:

- remove overlay popup rendering responsibility from `VNode.Children()`
- keep `VNode` responsible only for declarative configuration
- move popup subtree generation into `select.Instance.RuntimeChildren()`
- render popup as:
  - portal wrapper node
  - popup surface child node

Expected result:

- popup becomes first-class in Fiber tree
- layout, paint, hit-testing, and portal positioning all see the same popup subtree

### Phase 3 - Local popup interaction via intent bubble

Changes:

- add popup-local control intents:
  - open / close
  - highlight index
  - commit index
- popup child instance emits local intents upward
- trigger `select.Instance` handles those intents

Expected result:

- popup no longer needs registry access to mutate trigger state
- popup behavior is routed through normal instance-tree semantics

### Phase 4 - Limit registries to cross-cutting concerns

Changes:

- retain open-select registry only for outside-click detection
- retain popup instance registry only for geometric containment fallback
- remove registry dependency for popup rendering and state ownership

Expected result:

- registries are no longer part of core rendering architecture

### Phase 5 - Shared popup foundation

Follow-up work, not required for the first implementation:

- extract reusable popup surface model and popup dismiss middleware
- reuse across:
  - `select`
  - future `combobox`
  - date-picker / color-picker / command palette style controls

## 6. Action Runtime Follow-up

Complex popup components are especially sensitive to event ordering.

The long-term rule is:

- middleware must run exactly once
- middleware must run before direct target dispatch
- target dispatch and fallback routing must share one ordered pipeline

This refactor keeps the recent App/Router ordering fix as a prerequisite.

## 7. Implemented in This Iteration

This iteration implements Phase 1–4 for `select`:

1. Added `RuntimeChildrenProvider`
2. Reconciler now appends runtime-generated child VNodes for element instances
3. Refactored `select` overlay popup to be emitted from `select.Instance.RuntimeChildren()`
4. Refactored popup child to use local intent bubble back to the owner `select`
5. Reduced `select` registry usage to dismissal / containment concerns only
6. Kept `overlayPopup` as explicit opt-in
7. Kept demo integration explicit through:
   - stable owner ID
   - `selectcomp.Install(app)`

## 8. Validation Matrix

Required validation after each phase:

### Component tests

- `go test ./ui/components/select`

Covers:

- single select
- multi select
- popup rendering
- overlay popup anchoring
- outside click dismissal

### App routing tests

- `go test ./framework -run "TestApp_ProcessMsg_(UntargetedClickDispatchesActionRouterMiddleware|TargetedClickRunsMiddlewareBeforeFiberDispatch|UnhandledNavigationFallsThroughToFocusedFiber|TypedNilMouseTargetFiber_NoPanic)"`

### Example builds

- `go build ./examples/mvp_components_demo`
- `go build ./examples/fiber_firsts/select_demo`

## 9. Refactor Boundaries

This plan intentionally does not:

- rewrite the entire Fiber-first runtime
- replace portal layout engine
- redesign statusbar tooltip
- introduce `combobox` yet

Those should happen only after the popup subtree model is proven stable in `select`.

## 10. Next Steps

After this iteration, the recommended order is:

1. extract shared popup surface helpers from `select`
2. refactor `menu` and `select` to converge on one popup foundation
3. introduce a separate `combobox` for editable search/filter inputs
4. remove remaining direct-dispatch / middleware ambiguity from App routing
