package layout

import (
	"sync"
)

// ==============================================================================
// Dirty Tracking (V3)
// ==============================================================================
// 脏标记跟踪器，支持增量布局优化

// DirtyTracker 脏标记跟踪器
type DirtyTracker struct {
	mu    sync.RWMutex
	dirty map[string]bool
}

// NewDirtyTracker 创建新的脏标记跟踪器
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		dirty: make(map[string]bool),
	}
}

// MarkLayoutDirty 标记节点为脏
func (dt *DirtyTracker) MarkLayoutDirty(id string) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.dirty[id] = true
}

// IsLayoutDirty 检查节点是否需要布局
func (dt *DirtyTracker) IsLayoutDirty(id string) bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.dirty[id]
}

// MarkSubtreeDirty 标记整个子树为脏
func (dt *DirtyTracker) MarkSubtreeDirty(node Node) {
	dt.markRecursive(node)
}

func (dt *DirtyTracker) markRecursive(node Node) {
	if node == nil {
		return
	}
	dt.MarkLayoutDirty(node.ID())
	for _, child := range node.Children() {
		dt.markRecursive(child)
	}
}

// Clear 清除所有脏标记
func (dt *DirtyTracker) Clear() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.dirty = make(map[string]bool)
}

// ClearKey 清除特定节点的脏标记
func (dt *DirtyTracker) ClearKey(id string) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	delete(dt.dirty, id)
}

// Size 返回脏标记数量
func (dt *DirtyTracker) Size() int {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return len(dt.dirty)
}

// HasAny 检查是否有任何脏标记
func (dt *DirtyTracker) HasAny() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return len(dt.dirty) > 0
}
