package paint

import (
	"github.com/wwsheng009/mint/internal/log"
)

// =============================================================================
// Line Diff Engine (Renderer 2.0)
// =============================================================================
// 行级差异检测引擎，用于替代 cell/region diff
// 核心原则：只要一行任意 cell 改变 → 整行重新渲染
// 这解决了终端作为 line-oriented device 的特性

// LineDiffResult 表示行级 diff 结果
type LineDiffResult struct {
	// ChangedLines 是变化的行号列表
	ChangedLines []int

	// HasChanges 表示是否有任何变化
	HasChanges bool

	// ScrollAmount 表示滚动行数（正数=向上滚动，负数=向下滚动）
	// 0 表示无滚动
	ScrollAmount int

	// HasScroll 表示是否检测到滚动
	HasScroll bool
}

// LineDiff 执行行级 diff 比较
// 使用行哈希进行 O(1) 行比较
func LineDiff(front, back *Buffer) LineDiffResult {
	return LineDiffInto(front, back, []int{})
}

// LineDiffInto is LineDiff with caller-owned storage for ChangedLines.
// The returned ChangedLines slice aliases scratch.
func LineDiffInto(front, back *Buffer, scratch []int) LineDiffResult {
	result := LineDiffResult{
		ChangedLines: scratch[:0],
		HasChanges:   false,
		ScrollAmount: 0,
		HasScroll:    false,
	}

	// 边界检查
	if front == nil || back == nil {
		return result
	}

	// 尺寸不同，全屏重绘
	if front.Width != back.Width || front.Height != back.Height {
		result.HasChanges = true
		for y := 0; y < back.Height; y++ {
			result.ChangedLines = append(result.ChangedLines, y)
		}
		return result
	}

	// 确保 back buffer 有哈希
	if back.LineHash == nil || len(back.LineHash) != back.Height {
		back.Rehash()
	}

	// 确保 front buffer 有哈希
	if front.LineHash == nil || len(front.LineHash) != front.Height {
		front.Rehash()
	}

	// 尝试检测滚动
	scrollAmount, hasScroll := detectScroll(front, back)
	if hasScroll {
		result.HasScroll = true
		result.ScrollAmount = scrollAmount
		log.RenderLogger.Debug("[LineDiff] Detected scroll: amount=%d", scrollAmount)
	}

	// 比较每一行
	height := minInt(front.Height, back.Height)
	for y := 0; y < height; y++ {
		if front.LineHash[y] != back.LineHash[y] {
			result.ChangedLines = append(result.ChangedLines, y)
		}
	}

	result.HasChanges = len(result.ChangedLines) > 0 || hasScroll

	log.RenderLogger.Debug("[LineDiff] ChangedLines=%d, HasScroll=%v",
		len(result.ChangedLines), hasScroll)

	return result
}

// detectScroll 检测是否发生了滚动
// 通过比较行哈希序列来识别滚动模式
func detectScroll(front, back *Buffer) (int, bool) {
	if front == nil || back == nil {
		return 0, false
	}

	// 如果两个 buffer 完全相同，不需要滚动
	allSame := true
	height := minInt(front.Height, back.Height)
	for y := 0; y < height; y++ {
		if y < len(front.LineHash) && y < len(back.LineHash) {
			if front.LineHash[y] != back.LineHash[y] {
				allSame = false
				break
			}
		}
	}
	if allSame {
		return 0, false
	}

	maxShift := minInt(front.Height, back.Height) / 2
	if maxShift < 1 {
		return 0, false
	}

	// 检测向上滚动（内容向上移动）
	for shift := 1; shift <= maxShift; shift++ {
		match := true
		matchedLines := 0

		for y := 0; y < front.Height-shift; y++ {
			if y+shift >= len(back.LineHash) || y >= len(front.LineHash) {
				match = false
				break
			}
			if front.LineHash[y+shift] != back.LineHash[y] {
				match = false
				break
			}
			matchedLines++
		}

		// 需要至少匹配一半的行才算有效滚动
		if match && matchedLines > front.Height/2 {
			return shift, true
		}
	}

	// 检测向下滚动（内容向下移动）
	for shift := 1; shift <= maxShift; shift++ {
		match := true
		matchedLines := 0

		for y := 0; y < front.Height-shift; y++ {
			if y >= len(back.LineHash) || y+shift >= len(front.LineHash) {
				match = false
				break
			}
			if front.LineHash[y] != back.LineHash[y+shift] {
				match = false
				break
			}
			matchedLines++
		}

		if match && matchedLines > front.Height/2 {
			return -shift, true
		}
	}

	return 0, false
}

// DiffLinesSimple 简单行级 diff（不检测滚动）
// 用于需要精确行比较的场景
func DiffLinesSimple(front, back *Buffer) []int {
	if front == nil || back == nil {
		return []int{}
	}

	// 尺寸不同，返回所有行
	if front.Width != back.Width || front.Height != back.Height {
		lines := make([]int, back.Height)
		for y := 0; y < back.Height; y++ {
			lines[y] = y
		}
		return lines
	}

	// 确保哈希存在
	if back.LineHash == nil || len(back.LineHash) != back.Height {
		back.Rehash()
	}
	if front.LineHash == nil || len(front.LineHash) != front.Height {
		front.Rehash()
	}

	var changed []int
	height := minInt(front.Height, back.Height)

	for y := 0; y < height; y++ {
		if front.LineHash[y] != back.LineHash[y] {
			changed = append(changed, y)
		}
	}

	return changed
}

// EqualLine 比较两行是否相等（使用哈希快速比较）
func EqualLine(front, back *Buffer, y int) bool {
	if front == nil || back == nil {
		return false
	}
	if y >= front.Height || y >= back.Height {
		return false
	}
	if y >= len(front.LineHash) || y >= len(back.LineHash) {
		return false
	}
	return front.LineHash[y] == back.LineHash[y]
}
