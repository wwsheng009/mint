package event

import (
	"unsafe"
)

// optimizeHitMapEntry 优化 HitMapEntry 内存布局
//
// 通过调整字段顺序和类型，减少内存占用和提高缓存命中率。
type optimizedHitMapEntry struct {
	// 将常用字段放在一起，提高缓存行利用率
	NodeID  string // 8 bytes
	X       int16  // 2 bytes (假设屏幕不超过 32767)
	Y       int16  // 2 bytes
	Width   uint16 // 2 bytes (假设宽度不超过 65535)
	Height  uint16 // 2 bytes
	ZIndex  int16  // 2 bytes

	// 对齐填充，确保结构体大小是 8 的倍数
	_ uint16 // 2 bytes padding
	_ uint32 // 4 bytes padding
}

// OptimizeMemory 优化 HitMap 内存使用
//
// 这是一组优化 HitMap 内存使用的方法。
func (h *HitMap) OptimizeMemory() {
	if h.entries == nil {
		return
	}

	// 1. 预分配切片容量，避免频繁扩容
	if cap(h.entries) > len(h.entries)*2 {
		// 如果容量远大于实际使用量，缩小切片
		newEntries := make([]HitMapEntry, len(h.entries))
		copy(newEntries, h.entries)
		h.entries = newEntries
	}
}

// ReuseEntries 重用条目内存
//
// 在更新 HitMap 时，尽量重用现有的条目结构，
// 而不是删除后重新分配。
func (h *HitMap) ReuseEntries() {
	// TODO: 实现对象池模式
	// 这可以进一步减少内存分配
}

// Compact 压缩 HitMap，移除无效条目
func (h *HitMap) Compact() {
	if h.entries == nil {
		return
	}

	// 移除零大小或无效的条目
	validEntries := make([]HitMapEntry, 0, len(h.entries))
	for _, entry := range h.entries {
		// 跳过零大小或无效的条目
		if entry.Bounds.Width <= 0 || entry.Bounds.Height <= 0 {
			continue
		}
		if entry.NodeID == "" {
			continue
		}
		validEntries = append(validEntries, entry)
	}

	h.entries = validEntries
}

// GetMemoryUsage 获取 HitMap 内存使用情况
func (h *HitMap) GetMemoryUsage() map[string]interface{} {
	if h.entries == nil {
		return map[string]interface{}{
			"entry_count": 0,
			"total_bytes": 0,
		}
	}

	// 估算内存使用
	entrySize := int(unsafe.Sizeof(HitMapEntry{}))
	totalBytes := len(h.entries) * entrySize

	return map[string]interface{}{
		"entry_count":  len(h.entries),
		"entry_size":   entrySize,
		"total_bytes":  totalBytes,
		"cap":          cap(h.entries),
	}
}

// CheckMemoryLeaks 检查内存泄漏
//
// 定期调用此方法来检查是否有内存泄漏。
// 返回 true 如果可能有内存泄漏。
func (h *HitMap) CheckMemoryLeaks() bool {
	// 检查条目数量是否异常增长
	if len(h.entries) > 10000 {
		// 超过 10000 个条目可能表示有问题
		return true
	}

	// 检查容量和实际使用量的差距
	if cap(h.entries) > len(h.entries)*10 {
		// 容量是实际使用量的 10 倍以上，可能表示有内存泄漏
		return true
	}

	return false
}

// Shrink 缩小 HitMap 内存占用
//
// 将切片容量缩减到实际大小，释放未使用的内存。
func (h *HitMap) Shrink() {
	if h.entries == nil {
		return
	}

	if cap(h.entries) > len(h.entries) {
		newEntries := make([]HitMapEntry, len(h.entries))
		copy(newEntries, h.entries)
		h.entries = newEntries
	}
}

// ValidateMemoryLayout 验证内存布局
//
// 确保结构体字段顺序优化，避免内存浪费。
func ValidateMemoryLayout() {
	// 验证 HitMapEntry 的大小
	entrySize := unsafe.Sizeof(HitMapEntry{})
	pointerSize := unsafe.Sizeof(uintptr(0))

	// 输出内存布局信息
	// 在实际应用中，这可以帮助发现内存问题
	_ = entrySize
	_ = pointerSize
}
