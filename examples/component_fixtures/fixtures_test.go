// Package component_fixtures_test contains tests for component fixtures
package component_fixtures_test

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/ui"
)

// TestFixtureDemo1FullApp tests the Demo1 full app fixture
func TestFixtureDemo1FullApp(t *testing.T) {
	fixture := component_fixtures.GetFixture("demo1_full_app")
	if fixture == nil {
		t.Fatal("demo1_full_app fixture not found")
	}

	vnode := fixture.Build()
	if vnode == nil {
		t.Fatal("Failed to build VNode from fixture")
	}

	// Convert to Fiber
	fiber := reconciler.CreateFiberFromVNode(vnode)
	if fiber == nil {
		t.Fatal("Failed to create Fiber from VNode")
	}

	// Verify node counts
	vnodeCount := component_fixtures.CountNodes(vnode)
	fiberCount := countFibers(fiber)

	t.Logf("VNode count: %d", vnodeCount)
	t.Logf("Fiber count: %d", fiberCount)

	if vnodeCount != fiberCount {
		t.Errorf("Node count mismatch: VNode=%d, Fiber=%d", vnodeCount, fiberCount)
	}
}

// TestFixtureDemo1Header tests the header component
func TestFixtureDemo1Header(t *testing.T) {
	fixture := component_fixtures.GetFixture("demo1_header")
	if fixture == nil {
		t.Fatal("demo1_header fixture not found")
	}

	vnode := fixture.Build()
	if vnode == nil {
		t.Fatal("Failed to build header VNode")
	}

	// Header should have children
	children := vnode.Children()
	if len(children) == 0 {
		t.Error("Header should have children")
	}

	t.Logf("Header has %d children", len(children))
}

// TestFixtureDemo1MainBody tests the main body component
func TestFixtureDemo1MainBody(t *testing.T) {
	fixture := component_fixtures.GetFixture("demo1_main_body")
	if fixture == nil {
		t.Fatal("demo1_main_body fixture not found")
	}

	vnode := fixture.Build()
	if vnode == nil {
		t.Fatal("Failed to build main body VNode")
	}

	// Main body should contain HStack with panels
	t.Logf("Main body built successfully, node count: %d", component_fixtures.CountNodes(vnode))
}

// TestFixtureDemo1Modal tests the modal component
func TestFixtureDemo1Modal(t *testing.T) {
	fixture := component_fixtures.GetFixture("demo1_modal")
	if fixture == nil {
		t.Fatal("demo1_modal fixture not found")
	}

	vnode := fixture.Build()
	if vnode == nil {
		t.Fatal("Failed to build modal VNode")
	}

	t.Logf("Modal built successfully, node count: %d", component_fixtures.CountNodes(vnode))
}

// TestAllStandardFixtures runs through all standard fixtures
func TestAllStandardFixtures(t *testing.T) {
	fixtures := component_fixtures.StandardFixtures()

	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			vnode := f.Build()
			if vnode == nil {
				t.Fatalf("Failed to build fixture: %s", f.Name)
			}

			// Convert to Fiber
			fiber := reconciler.CreateFiberFromVNode(vnode)
			if fiber == nil {
				t.Fatalf("Failed to create Fiber for fixture: %s", f.Name)
			}

			vnodeCount := component_fixtures.CountNodes(vnode)
			fiberCount := countFibers(fiber)

			if vnodeCount != fiberCount {
				t.Errorf("Node count mismatch: VNode=%d, Fiber=%d", vnodeCount, fiberCount)
			}

			t.Logf("%s: VNode=%d, Fiber=%d nodes", f.Name, vnodeCount, fiberCount)
		})
	}
}

// TestBuildDemo1AppWithConfig tests building Demo1 app with custom config
func TestBuildDemo1AppWithConfig(t *testing.T) {
	tests := []struct {
		name string
		opts []component_fixtures.Demo1ConfigOption
	}{
		{
			name: "default_config",
			opts: nil,
		},
		{
			name: "custom_count",
			opts: []component_fixtures.Demo1ConfigOption{component_fixtures.WithCount(42)},
		},
		{
			name: "custom_input",
			opts: []component_fixtures.Demo1ConfigOption{component_fixtures.WithInput("test input")},
		},
		{
			name: "custom_items",
			opts: []component_fixtures.Demo1ConfigOption{component_fixtures.WithItems([]string{"A", "B", "C"})},
		},
		{
			name: "custom_size",
			opts: []component_fixtures.Demo1ConfigOption{component_fixtures.WithSize(120, 40)},
		},
		{
			name: "combined_options",
			opts: []component_fixtures.Demo1ConfigOption{
				component_fixtures.WithCount(10),
				component_fixtures.WithInput("hello"),
				component_fixtures.WithItems([]string{"X", "Y", "Z"}),
				component_fixtures.WithSize(100, 30),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vnode := component_fixtures.BuildDemo1App(tc.opts...)
			if vnode == nil {
				t.Fatal("Failed to build Demo1 app")
			}

			fiber := reconciler.CreateFiberFromVNode(vnode)
			if fiber == nil {
				t.Fatal("Failed to create Fiber")
			}

			vnodeCount := component_fixtures.CountNodes(vnode)
			fiberCount := countFibers(fiber)

			if vnodeCount != fiberCount {
				t.Errorf("Node count mismatch: VNode=%d, Fiber=%d", vnodeCount, fiberCount)
			}
		})
	}
}

// TestBuildVNodeTree tests the tree builder helper
func TestBuildVNodeTree(t *testing.T) {
	tests := []struct {
		depth   int
		breadth int
	}{
		{1, 1},
		{2, 2},
		{3, 3},
		{2, 5},
		{4, 2},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("depth_%d_breadth_%d", tc.depth, tc.breadth), func(t *testing.T) {
			vnode := component_fixtures.BuildVNodeTree(tc.depth, tc.breadth)
			if vnode == nil {
				t.Fatal("Failed to build tree")
			}

			count := component_fixtures.CountNodes(vnode)
			t.Logf("Tree (depth=%d, breadth=%d): %d nodes", tc.depth, tc.breadth, count)
		})
	}
}

// TestBuildKeyedVNodeTree tests the keyed tree builder helper
func TestBuildKeyedVNodeTree(t *testing.T) {
	vnode := component_fixtures.BuildKeyedVNodeTree(2, 3, "root")
	if vnode == nil {
		t.Fatal("Failed to build keyed tree")
	}

	// Convert to Fiber
	fiber := reconciler.CreateFiberFromVNode(vnode)
	if fiber == nil {
		t.Fatal("Failed to create Fiber")
	}

	// Verify keys are preserved
	keys := collectKeys(fiber)
	t.Logf("Keys found: %v", keys)

	if len(keys) == 0 {
		t.Error("No keys found in Fiber tree")
	}

	// Verify unique NodeIDs
	nodeIDs := collectNodeIDs(fiber)
	uniqueNodeIDs := make(map[uint64]bool)
	for _, id := range nodeIDs {
		if uniqueNodeIDs[id] {
			t.Errorf("Duplicate NodeID found: %d", id)
		}
		uniqueNodeIDs[id] = true
	}

	t.Logf("Found %d unique NodeIDs", len(uniqueNodeIDs))
}

// TestBuildMixedKeyedTree tests mixed keyed/non-keyed tree
func TestBuildMixedKeyedTree(t *testing.T) {
	vnode := component_fixtures.BuildMixedKeyedTree(3, 2)
	if vnode == nil {
		t.Fatal("Failed to build mixed tree")
	}

	fiber := reconciler.CreateFiberFromVNode(vnode)
	if fiber == nil {
		t.Fatal("Failed to create Fiber")
	}

	count := component_fixtures.CountNodes(vnode)
	t.Logf("Mixed tree: %d nodes", count)

	// Should have 1 root + 3 keyed + 2 non-keyed = 6 nodes
	if count != 6 {
		t.Errorf("Expected 6 nodes, got %d", count)
	}
}

// =============================================================================
// Helper functions
// =============================================================================

func countFibers(fiber *ui.Fiber) int {
	if fiber == nil {
		return 0
	}
	count := 1
	for _, child := range fiber.GetChildFibers() {
		count += countFibers(child)
	}
	return count
}

func collectKeys(fiber *ui.Fiber) []string {
	keys := make([]string, 0)
	var traverse func(f *ui.Fiber)
	traverse = func(f *ui.Fiber) {
		if f == nil {
			return
		}
		if f.Key != "" {
			keys = append(keys, f.Key)
		}
		traverse(f.Child)
		traverse(f.Sibling)
	}
	traverse(fiber)
	return keys
}

func collectNodeIDs(fiber *ui.Fiber) []uint64 {
	ids := make([]uint64, 0)
	var traverse func(f *ui.Fiber)
	traverse = func(f *ui.Fiber) {
		if f == nil {
			return
		}
		ids = append(ids, f.NodeID)
		traverse(f.Child)
		traverse(f.Sibling)
	}
	traverse(fiber)
	return ids
}
