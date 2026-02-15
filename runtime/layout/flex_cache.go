package layout

import (
	"fmt"
	"strings"
	"sync"
)

// ==============================================================================
// Flex Distribution Cache (V3)
// ==============================================================================
// 缓存 Flex 布局的分布信息，避免重复计算

// FlexDistributionInfo 记录 Flex 布局的分布信息
type FlexDistributionInfo struct {
	TotalFlexFactor int  // flex-grow 总和
	FixedSize       int  // 固定尺寸子节点总尺寸
	ChildCount      int  // 子节点数量
	MaxCrossSize    int  // 最大交叉轴尺寸
	Valid           bool // 是否有效
	Version         uint64 // 版本号（用于失效）
}

// FlexCache Flex 分布缓存
type FlexCache struct {
	entries map[string]*FlexDistributionInfo
	mu      sync.RWMutex
}

// NewFlexCache 创建新的 Flex 缓存
func NewFlexCache() *FlexCache {
	return &FlexCache{
		entries: make(map[string]*FlexDistributionInfo),
	}
}

// Get 获取或计算 Flex 分布信息
func (fc *FlexCache) Get(
	nodeID string,
	children []Node,
	flexibleIndices []int,
	isRow bool,
	computeFunc func() *FlexDistributionInfo,
) *FlexDistributionInfo {
	key := fmt.Sprintf("%s:%d:%v", nodeID, len(children), isRow)

	fc.mu.RLock()
	if info, ok := fc.entries[key]; ok && info.Valid {
		fc.mu.RUnlock()
		return info
	}
	fc.mu.RUnlock()

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Double-check after acquiring write lock
	if info, ok := fc.entries[key]; ok && info.Valid {
		return info
	}

	info := computeFunc()
	fc.entries[key] = info
	return info
}

// Invalidate 失效指定节点的缓存
func (fc *FlexCache) Invalidate(nodeID string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	for key, info := range fc.entries {
		// 匹配以 nodeID: 开头的键
		if strings.HasPrefix(key, nodeID+":") {
			info.Valid = false
			delete(fc.entries, key)
		}
	}
}

// InvalidateAll 失效所有缓存
func (fc *FlexCache) InvalidateAll() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.entries = make(map[string]*FlexDistributionInfo)
}

// GetStats 获取缓存统计
func (fc *FlexCache) GetStats() (total, valid int) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	total = len(fc.entries)
	for _, info := range fc.entries {
		if info.Valid {
			valid++
		}
	}
	return
}

// Clear 清空缓存（同 InvalidateAll）
func (fc *FlexCache) Clear() {
	fc.InvalidateAll()
}
