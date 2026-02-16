package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockIdentifiableNode 模拟可标识节点
type MockIdentifiableNode struct {
	*MockNode
	stableID uint64
}

func NewMockIdentifiableNode(id string, width, height, stableID uint64) *MockIdentifiableNode {
	return &MockIdentifiableNode{
		MockNode:  NewMockNode(id, int(width), int(height)),
		stableID: stableID,
	}
}

func (m *MockIdentifiableNode) GetStableID() uint64 {
	return m.stableID
}

// MockVersionedNode 模拟可版本节点
type MockVersionedNode struct {
	*MockNode
	version uint64
}

func NewMockVersionedNode(id string, width, height int, version uint64) *MockVersionedNode {
	return &MockVersionedNode{
		MockNode: NewMockNode(id, width, height),
		version:   version,
	}
}

func (m *MockVersionedNode) GetVersion() uint64 {
	return m.version
}

// MockLayoutInfoNode 模拟布局信息提供者节点
type MockLayoutInfoNode struct {
	*MockNode
	layoutInfo *LayoutInfo
}

func NewMockLayoutInfoNode(id string, width, height int) *MockLayoutInfoNode {
	return &MockLayoutInfoNode{
		MockNode: NewMockNode(id, width, height),
		layoutInfo: &LayoutInfo{
			ID:     0,
			Version: 0,
		},
	}
}

func (m *MockLayoutInfoNode) GetLayoutInfo() *LayoutInfo {
	m.layoutInfo.LayoutBox = &LayoutBox{
		ID:      m.MockNode.id,
		X:       m.MockNode.position.X,
		Y:       m.MockNode.position.Y,
		Width:   m.MockNode.size.Width,
		Height:  m.MockNode.size.Height,
	}
	return m.layoutInfo
}

// MockIdentifiableVersionedNode 同时实现Identifiable和Versioned
type MockIdentifiableVersionedNode struct {
	*MockNode
	stableID uint64
	version  uint64
}

func NewMockIdentifiableVersionedNode(id string, width, height int, stableID, version uint64) *MockIdentifiableVersionedNode {
	return &MockIdentifiableVersionedNode{
		MockNode: NewMockNode(id, width, height),
		stableID: stableID,
		version:   version,
	}
}

func (m *MockIdentifiableVersionedNode) GetStableID() uint64 {
	return m.stableID
}

func (m *MockIdentifiableVersionedNode) GetVersion() uint64 {
	return m.version
}

func (m *MockIdentifiableVersionedNode) GetLayoutInfo() *LayoutInfo {
	return &LayoutInfo{
		ID:     m.stableID,
		Version: m.version,
		LayoutBox: &LayoutBox{
			ID:      m.MockNode.id,
			X:       m.MockNode.position.X,
			Y:       m.MockNode.position.Y,
			Width:   m.MockNode.size.Width,
			Height:  m.MockNode.size.Height,
		},
	}
}

func TestIdentifiable_Interface(t *testing.T) {
	t.Run("implement Identifiable interface", func(t *testing.T) {
		node := NewMockIdentifiableNode("node1", 100, 100, 12345)
		
		// Verify it implements the interface
		var identifiable Identifiable = node
		assert.Equal(t, uint64(12345), identifiable.GetStableID())
	})

	t.Run("unique stable IDs", func(t *testing.T) {
		nodes := []Identifiable{
			NewMockIdentifiableNode("node1", 10, 10, 1),
			NewMockIdentifiableNode("node2", 10, 10, 2),
			NewMockIdentifiableNode("node3", 10, 10, 3),
		}

		ids := make(map[uint64]bool)
		for _, node := range nodes {
			id := node.GetStableID()
			assert.False(t, ids[id], "ID should be unique: %d", id)
			ids[id] = true
		}

		assert.Equal(t, 3, len(ids))
	})

	t.Run("stable ID consistency", func(t *testing.T) {
		node := NewMockIdentifiableNode("node1", 100, 100, 12345)

		// ID should remain consistent across calls
		id1 := node.GetStableID()
		id2 := node.GetStableID()
		id3 := node.GetStableID()

		assert.Equal(t, uint64(12345), id1)
		assert.Equal(t, id1, id2)
		assert.Equal(t, id2, id3)
	})
}

func TestVersioned_Interface(t *testing.T) {
	t.Run("implement Versioned interface", func(t *testing.T) {
		node := NewMockVersionedNode("node1", 100, 100, 5)

		// Verify it implements the interface
		var versioned Versioned = node
		assert.Equal(t, uint64(5), versioned.GetVersion())
	})

	t.Run("version increments", func(t *testing.T) {
		node := NewMockVersionedNode("node1", 100, 100, 1)

		assert.Equal(t, uint64(1), node.GetVersion())

		// Simulate version increment
		node.version = 2
		assert.Equal(t, uint64(2), node.GetVersion())
	})

	t.Run("version tracking", func(t *testing.T) {
		nodes := []Versioned{
			NewMockVersionedNode("node1", 10, 10, 1),
			NewMockVersionedNode("node2", 10, 10, 1),
			NewMockVersionedNode("node3", 10, 10, 2),
		}

		versions := make(map[uint64]int)
		for _, node := range nodes {
			ver := node.GetVersion()
			versions[ver]++
		}

		assert.Equal(t, 2, versions[1])
		assert.Equal(t, 1, versions[2])
	})
}

func TestLayoutInfoProvider_Interface(t *testing.T) {
	t.Run("implement LayoutInfoProvider interface", func(t *testing.T) {
		node := NewMockLayoutInfoNode("node1", 100, 50)

		// Verify it implements the interface
		var provider LayoutInfoProvider = node
		info := provider.GetLayoutInfo()

		assert.NotNil(t, info)
		assert.NotNil(t, info.LayoutBox)
	})

	t.Run("layout info contains box", func(t *testing.T) {
		node := NewMockLayoutInfoNode("node1", 100, 50)
		node.MockNode.position.X = 10
		node.MockNode.position.Y = 20

		info := node.GetLayoutInfo()

		assert.NotNil(t, info.LayoutBox)
		assert.Equal(t, "node1", info.LayoutBox.ID)
		assert.Equal(t, 10, info.LayoutBox.X)
		assert.Equal(t, 20, info.LayoutBox.Y)
		assert.Equal(t, 100, info.LayoutBox.Width)
		assert.Equal(t, 50, info.LayoutBox.Height)
	})

	t.Run("layout info is updated", func(t *testing.T) {
		node := NewMockLayoutInfoNode("node1", 100, 50)
		node.MockNode.position.X = 10
		node.MockNode.position.Y = 20

		info1 := node.GetLayoutInfo()
		assert.Equal(t, 10, info1.LayoutBox.X)

		// Update position
		node.MockNode.position.X = 30
		node.MockNode.position.Y = 40
		info2 := node.GetLayoutInfo()

		assert.Equal(t, 30, info2.LayoutBox.X)
		assert.Equal(t, 40, info2.LayoutBox.Y)
	})
}

func TestLayoutInfo_Struct(t *testing.T) {
	t.Run("create LayoutInfo", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "test",
			X:       10,
			Y:       20,
			Width:   100,
			Height:  50,
		}

		info := &LayoutInfo{
			ID:        12345,
			Version:   5,
			LayoutBox: box,
		}

		assert.Equal(t, uint64(12345), info.ID)
		assert.Equal(t, uint64(5), info.Version)
		assert.Equal(t, box, info.LayoutBox)
	})

	t.Run("nil LayoutBox", func(t *testing.T) {
		info := &LayoutInfo{
			ID:        12345,
			Version:   5,
			LayoutBox: nil,
		}

		assert.Equal(t, uint64(12345), info.ID)
		assert.Equal(t, uint64(5), info.Version)
		assert.Nil(t, info.LayoutBox)
	})
}

func TestCombinedInterfaces(t *testing.T) {
	t.Run("node implements all interfaces", func(t *testing.T) {
		node := NewMockIdentifiableVersionedNode("node1", 100, 50, 12345, 5)

		// Node implements all three interfaces through embedding
		assert.Equal(t, uint64(12345), node.GetStableID())
		assert.Equal(t, uint64(5), node.GetVersion())

		info := node.GetLayoutInfo()
		assert.NotNil(t, info)
		assert.Equal(t, uint64(12345), info.ID)
		assert.Equal(t, uint64(5), info.Version)
		assert.NotNil(t, info.LayoutBox)
	})
}

func TestInterfaceIntegration(t *testing.T) {
	t.Run("Node does not implement interfaces by default", func(t *testing.T) {
		node := NewMockNode("node1", 100, 50)

		// MockNode does not implement the new interfaces
		// We can't use type assertions on concrete types, so we just verify behavior

		// Try to access interface methods (will fail at runtime)
		// Instead, we just verify the node itself works
		assert.Equal(t, "node1", node.ID())
	})

	t.Run("selective interface implementation", func(t *testing.T) {
		// Only Identifiable
		identifiableNode := NewMockIdentifiableNode("node1", 100, 50, 12345)
		
		// Verify it has the method
		assert.Equal(t, uint64(12345), identifiableNode.GetStableID())

		// Only Versioned
		versionedNode := NewMockVersionedNode("node2", 100, 50, 5)
		
		// Verify it has the method
		assert.Equal(t, uint64(5), versionedNode.GetVersion())
	})
}

func TestLayoutInfo_Consistency(t *testing.T) {
	t.Run("layout info matches node state", func(t *testing.T) {
		node := NewMockIdentifiableVersionedNode("node1", 100, 50, 12345, 5)
		node.MockNode.position.X = 10
		node.MockNode.position.Y = 20

		info := node.GetLayoutInfo()

		// Verify info matches node state
		assert.Equal(t, "node1", info.LayoutBox.ID)
		assert.Equal(t, 10, info.LayoutBox.X)
		assert.Equal(t, 20, info.LayoutBox.Y)

		width, height := node.GetSize()
		assert.Equal(t, width, info.LayoutBox.Width)
		assert.Equal(t, height, info.LayoutBox.Height)
	})

	t.Run("version tracking with layout info", func(t *testing.T) {
		node := NewMockIdentifiableVersionedNode("node1", 100, 50, 12345, 1)

		info1 := node.GetLayoutInfo()
		assert.Equal(t, uint64(1), info1.Version)

		// Update version
		node.version = 2
		info2 := node.GetLayoutInfo()
		assert.Equal(t, uint64(2), info2.Version)
		assert.Equal(t, uint64(12345), info2.ID)
	})
}

func TestInterface_Semantics(t *testing.T) {
	t.Run("Identifiable provides stable identity", func(t *testing.T) {
		node := NewMockIdentifiableNode("node1", 100, 50, 12345)

		// Stable ID should not change even if node ID changes
		node.MockNode.id = "node2"
		assert.Equal(t, uint64(12345), node.GetStableID())
	})

	t.Run("Versioned tracks changes", func(t *testing.T) {
		node := NewMockVersionedNode("node1", 100, 50, 1)

		// Version should be manually incremented by the node implementation
		// when content changes
		node.version++
		assert.Equal(t, uint64(2), node.GetVersion())
	})
}

// Benchmark tests
func BenchmarkIdentifiable_GetStableID(b *testing.B) {
	node := NewMockIdentifiableNode("node1", 100, 50, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = node.GetStableID()
	}
}

func BenchmarkVersioned_GetVersion(b *testing.B) {
	node := NewMockVersionedNode("node1", 100, 50, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = node.GetVersion()
	}
}

func BenchmarkLayoutInfoProvider_GetLayoutInfo(b *testing.B) {
	node := NewMockLayoutInfoNode("node1", 100, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = node.GetLayoutInfo()
	}
}
