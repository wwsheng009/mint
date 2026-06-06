package actionbridge

import (
	"testing"

	runtimepkg "github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/ui"
)

type bridgeActionRecorder struct {
	key     string
	handled int
}

func (r *bridgeActionRecorder) Key() string                      { return r.key }
func (r *bridgeActionRecorder) SetKey(key string)                { r.key = key }
func (r *bridgeActionRecorder) Init(ui.Props)                    {}
func (r *bridgeActionRecorder) Destroy()                         {}
func (r *bridgeActionRecorder) OnMount()                         {}
func (r *bridgeActionRecorder) OnUnmount()                       {}
func (r *bridgeActionRecorder) SetProps(ui.Props) bool           { return false }
func (r *bridgeActionRecorder) GetProps() ui.Props               { return nil }
func (r *bridgeActionRecorder) MarkDirty()                       {}
func (r *bridgeActionRecorder) IsDirty() bool                    { return false }
func (r *bridgeActionRecorder) GetContext() *ui.ComponentContext { return nil }
func (r *bridgeActionRecorder) HandleAction(*action.Action) bool { r.handled++; return true }

func TestDispatchFromFiberBubblesToAncestorActionHandler(t *testing.T) {
	parent := &bridgeActionRecorder{key: "parent"}
	root := &ui.Fiber{NodeID: 1, ActionTargetID: "1", Instance: parent}
	child := &ui.Fiber{NodeID: 2, ActionTargetID: "2", Return: root}
	bridge := New(action.NewRouter(nil))

	if !bridge.DispatchFromFiber(child, action.ActionScroll, 1) {
		t.Fatal("scroll should bubble to ancestor ActionHandlerInstance")
	}
	if parent.handled != 1 {
		t.Fatalf("ancestor handled count = %d, want 1", parent.handled)
	}
}

func TestDispatchFromFiberSkipsUnregisteredRouterTargets(t *testing.T) {
	parent := &bridgeActionRecorder{key: "parent"}
	root := &ui.Fiber{NodeID: 1, ActionTargetID: "1", Instance: parent}
	child := &ui.Fiber{NodeID: 2, ActionTargetID: "2", Return: root}
	router := action.NewRouter(nil)
	router.SetRoot(&runtimepkg.LayoutNode{ID: "root"})
	bridge := New(router)

	if !bridge.DispatchFromFiber(child, action.ActionScroll, 1) {
		t.Fatal("unregistered child router target should not stop fiber bubbling")
	}
	if parent.handled != 1 {
		t.Fatalf("ancestor handled count = %d, want 1", parent.handled)
	}
}
