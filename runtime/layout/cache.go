package layout

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"
)

// ==============================================================================
// Layout Cache (V3)
// ==============================================================================
// 布局结果缓存，避免重复计算



// Cache 布局缓存
type Cache struct {
	entries map[string]*CachedLayout
	maxSize int
}

// CachedLayout 缓存的布局结果
type CachedLayout struct {
	Result     *LayoutResult
	Timestamp  time.Time
	HitCount   int
}

// Get 获取缓存
func (c *Cache) Get(node Node, constraints Constraints) *LayoutResult {
	// 叶子节点优先使用缓存
	if !c.isLeafNode(node) {
		return nil
	}

	key := c.makeKey(node, constraints)
	if entry, ok := c.entries[key]; ok {
		entry.HitCount++
		// 返回克隆避免修改缓存
		return c.cloneResult(entry.Result)
	}
	return nil
}

// isLeafNode 检查是否为叶子节点
func (c *Cache) isLeafNode(node Node) bool {
	if node == nil {
		return false
	}
	return len(node.Children()) == 0
}

// Put 存入缓存
func (c *Cache) Put(node Node, constraints Constraints, result *LayoutResult) {
	key := c.makeKey(node, constraints)

	// 如果缓存已满，删除最旧的条目
	if len(c.entries) >= c.maxSize {
		c.evict()
	}

	c.entries[key] = &CachedLayout{
		Result:     result,
		Timestamp:  time.Now(),
		HitCount:   0,
	}
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.entries = make(map[string]*CachedLayout)
}

// RemoveByNode 删除特定节点的缓存
// 注意：由于缓存键使用节点树哈希（SHA256），无法精确匹配特定节点
// 此方法会清空所有缓存作为暂时的解决方案
// TODO: 改进缓存策略，使用节点 ID 作为键的一部分
func (c *Cache) RemoveByNode(id string) {
	c.Clear()
}

// evict 驱逐最旧的条目
func (c *Cache) evict() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// makeKey 生成缓存键
func (c *Cache) makeKey(node Node, constraints Constraints) string {
	// 实际应该基于节点树结构
	constraintKey := c.constraintsKey(constraints)
	nodesHash := c.nodesHash(node)

	return constraintKey + ":" + nodesHash
}

// constraintsKey 约束键
func (c *Cache) constraintsKey(constraints Constraints) string {
	return fmt.Sprintf("%d,%d,%d,%d",
		constraints.MinWidth, constraints.MaxWidth,
		constraints.MinHeight, constraints.MaxHeight)
}

// nodesHash 节点哈希
func (c *Cache) nodesHash(node Node) string {
	h := sha256.New()
	c.nodesHashRecursive(node, h)
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// nodesHashRecursive 递归计算节点哈希
func (c *Cache) nodesHashRecursive(node Node, h hash.Hash) {
	if node == nil {
		return
	}
	h.Write([]byte(node.ID()))
	h.Write([]byte(node.Type()))

	// 如果实现了 Measurable，可能需要包含其属性（这里简化）

	for _, child := range node.Children() {
		c.nodesHashRecursive(child, h)
	}
}

// cloneResult 克隆布局结果
func (c *Cache) cloneResult(result *LayoutResult) *LayoutResult {
	if result == nil {
		return nil
	}
	clone := &LayoutResult{
		ContentSize: result.ContentSize,
		Dirty:       result.Dirty,
	}

	if result.Root != nil {
		clone.Root = c.cloneBox(result.Root)
	}

	clone.Boxes = make([]LayoutBox, len(result.Boxes))
	copy(clone.Boxes, result.Boxes)

	return clone
}

// cloneBox 克隆布局盒子
func (c *Cache) cloneBox(box *LayoutBox) *LayoutBox {
	if box == nil {
		return nil
	}
	clone := &LayoutBox{
		ID:       box.ID,
		X:        box.X,
		Y:        box.Y,
		Width:    box.Width,
		Height:   box.Height,
		Baseline: box.Baseline,
	}

	if len(box.Children) > 0 {
		clone.Children = make([]*LayoutBox, len(box.Children))
		for i, child := range box.Children {
			clone.Children[i] = c.cloneBox(child)
		}
	}

	return clone
}
