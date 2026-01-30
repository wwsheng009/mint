// Package v1 provides pure statistics collection for DevTools.
//
// This file implements the analysis level system with zero overhead when disabled.
package v1

import (
	"sync"
	"sync/atomic"
)

// Level represents the analysis level.
type Level int

const (
	LevelNone Level = iota // 0: Completely disabled (zero overhead)
	LevelBasic             // 1: Basic statistics (atomic counters only)
	LevelEnhanced          // 2: Enhanced statistics (with history window)
	LevelAdvanced          // 3: Advanced (with ranking and percentiles)
)

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelBasic:
		return "Basic"
	case LevelEnhanced:
		return "Enhanced"
	case LevelAdvanced:
		return "Advanced"
	default:
		return "None"
	}
}

// LevelController manages the analysis level.
type LevelController struct {
	mu    sync.RWMutex
	level atomic.Value // Level
}

// NewLevelController creates a new level controller.
func NewLevelController(initialLevel Level) *LevelController {
	lc := &LevelController{}
	lc.level.Store(initialLevel)
	return lc
}

// SetLevel sets the analysis level.
func (lc *LevelController) SetLevel(level Level) {
	lc.level.Store(level)
}

// GetLevel returns the current analysis level.
func (lc *LevelController) GetLevel() Level {
	return lc.level.Load().(Level)
}

// IsEnabled returns true if any analysis is enabled.
func (lc *LevelController) IsEnabled() bool {
	return lc.GetLevel() > LevelNone
}

// ShouldCollectBasicStats returns true if basic stats should be collected.
func (lc *LevelController) ShouldCollectBasicStats() bool {
	return lc.GetLevel() >= LevelBasic
}

// ShouldCollectEnhancedStats returns true if enhanced stats should be collected.
func (lc *LevelController) ShouldCollectEnhancedStats() bool {
	return lc.GetLevel() >= LevelEnhanced
}

// ShouldCollectAdvancedStats returns true if advanced stats should be collected.
func (lc *LevelController) ShouldCollectAdvancedStats() bool {
	return lc.GetLevel() >= LevelAdvanced
}

// ExpectedOverhead returns the expected overhead percentage for the current level.
func (lc *LevelController) ExpectedOverhead() float64 {
	switch lc.GetLevel() {
	case LevelNone:
		return 0.0
	case LevelBasic:
		return 1.0    // < 1%
	case LevelEnhanced:
		return 3.0    // < 3%
	case LevelAdvanced:
		return 5.0    // < 5%
	default:
		return 0.0
	}
}
