// Package render provides parallel rendering pipeline with switchable layout engines
package render

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// LayoutEngineType - Defines which layout engine to use
// =============================================================================

// LayoutEngineType specifies which layout engine to use for rendering
type LayoutEngineType int

const (
	// LayoutEngineCompute uses runtime/compute (stable, production)
	LayoutEngineCompute LayoutEngineType = iota

	// LayoutEngineNew uses runtime/layout (experimental, under development)
	LayoutEngineNew

	// LayoutEngineBoth runs both engines in parallel for comparison
	LayoutEngineBoth
)

// String returns the string representation of the engine type
func (t LayoutEngineType) String() string {
	switch t {
	case LayoutEngineCompute:
		return "compute"
	case LayoutEngineNew:
		return "layout"
	case LayoutEngineBoth:
		return "both"
	default:
		return "unknown"
	}
}

// ParseLayoutEngineType parses a string to LayoutEngineType
func ParseLayoutEngineType(s string) LayoutEngineType {
	switch s {
	case "compute", "stable", "old":
		return LayoutEngineCompute
	case "layout", "new", "experimental":
		return LayoutEngineNew
	case "both", "parallel", "compare":
		return LayoutEngineBoth
	default:
		return LayoutEngineCompute
	}
}

// =============================================================================
// LayoutResult - Unified layout result interface
// =============================================================================

// LayoutResult represents the output of a layout computation
// Both compute.ComputedLayout and layout.LayoutResult can be adapted to this
type LayoutResult interface {
	// GetRootBox returns the root layout box for painting
	GetRootBox() PaintableBox

	// GetHitMap returns the hit map for event routing
	GetHitMap() *event.HitMap

	// GetRenderPlanes returns layered boxes for multi-layer rendering
	GetRenderPlanes() *layer.RenderPlanes
}

// PaintableBox represents a box that can be painted
// This is a minimal interface that both compute.ComputedBox and layout.LayoutBox can satisfy
type PaintableBox interface {
	// GetBounds returns the box bounds (x, y, width, height)
	GetBounds() (x, y, width, height int)

	// GetChildren returns child boxes
	GetChildren() []PaintableBox
}

// =============================================================================
// LayoutEngine - Unified layout engine interface
// =============================================================================

// LayoutEngine defines the interface for layout engines
type LayoutEngine interface {
	// Layout performs layout computation
	Layout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error)

	// LayoutFiber performs layout on a Fiber tree directly
	LayoutFiber(fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error)

	// GetType returns the engine type
	GetType() LayoutEngineType

	// GetStats returns cache statistics
	GetStats() CacheStats

	// ClearCache clears the layout cache
	ClearCache()

	// SetDebug enables/disables debug output
	SetDebug(debug bool)
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits   int
	Misses int
	Size   int
}

// =============================================================================
// ComputeEngineAdapter - Adapts compute.Engine to LayoutEngine interface
// =============================================================================

// ComputeEngineAdapter wraps compute.Engine to implement LayoutEngine
type ComputeEngineAdapter struct {
	engine *compute.Engine
}

// NewComputeEngineAdapter creates a new adapter for the compute engine
func NewComputeEngineAdapter() *ComputeEngineAdapter {
	return &ComputeEngineAdapter{
		engine: compute.NewEngine(),
	}
}

// Layout performs layout using the compute engine
func (a *ComputeEngineAdapter) Layout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	result, err := a.engine.Layout(vnode, fiber, constraints)
	if err != nil {
		return nil, err
	}
	return &computeLayoutResultAdapter{ComputedLayout: result}, nil
}

// LayoutFiber performs layout on a Fiber tree using the compute engine
func (a *ComputeEngineAdapter) LayoutFiber(fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	result, err := a.engine.LayoutFiber(fiber, constraints)
	if err != nil {
		return nil, err
	}
	return &computeLayoutResultAdapter{ComputedLayout: result}, nil
}

// GetType returns the engine type
func (a *ComputeEngineAdapter) GetType() LayoutEngineType {
	return LayoutEngineCompute
}

// GetStats returns cache statistics
func (a *ComputeEngineAdapter) GetStats() CacheStats {
	stats := a.engine.GetCacheStats()
	return CacheStats{
		Hits:   stats.Hits,
		Misses: stats.Misses,
		Size:   a.engine.CacheSize(),
	}
}

// ClearCache clears the layout cache
func (a *ComputeEngineAdapter) ClearCache() {
	a.engine.ClearCache()
}

// SetDebug enables/disables debug output
func (a *ComputeEngineAdapter) SetDebug(debug bool) {
	a.engine.SetDebug(debug)
}

// GetEngine returns the underlying compute engine
func (a *ComputeEngineAdapter) GetEngine() *compute.Engine {
	return a.engine
}

// computeLayoutResultAdapter adapts compute.ComputedLayout to LayoutResult
type computeLayoutResultAdapter struct {
	*compute.ComputedLayout
}

// GetRootBox returns the root box for painting
func (a *computeLayoutResultAdapter) GetRootBox() PaintableBox {
	if a.ComputedLayout == nil || a.Root == nil {
		return nil
	}
	return &computedBoxAdapter{ComputedBox: a.Root}
}

// GetHitMap returns the hit map
func (a *computeLayoutResultAdapter) GetHitMap() *event.HitMap {
	return a.ComputedLayout.HitMap
}

// GetRenderPlanes returns render planes
func (a *computeLayoutResultAdapter) GetRenderPlanes() *layer.RenderPlanes {
	if a.ComputedLayout == nil || a.ComputedLayout.RenderPlanes == nil {
		return nil
	}
	if planes, ok := a.ComputedLayout.RenderPlanes.(*layer.RenderPlanes); ok {
		return planes
	}
	return nil
}

// computedBoxAdapter adapts compute.ComputedBox to PaintableBox
type computedBoxAdapter struct {
	*compute.ComputedBox
}

// GetBounds returns the box bounds
func (a *computedBoxAdapter) GetBounds() (x, y, width, height int) {
	return a.Box.X, a.Box.Y, a.Box.Width, a.Box.Height
}

// GetChildren returns child boxes
func (a *computedBoxAdapter) GetChildren() []PaintableBox {
	if a.ComputedBox == nil {
		return nil
	}
	children := make([]PaintableBox, len(a.Children))
	for i, child := range a.Children {
		children[i] = &computedBoxAdapter{ComputedBox: child}
	}
	return children
}

// =============================================================================
// NewLayoutEngineAdapter - Adapts layout.Engine to LayoutEngine interface
// =============================================================================

// NewLayoutEngineAdapter wraps layout.Engine to implement LayoutEngine
type NewLayoutEngineAdapter struct {
	engine *layout.Engine
}

// NewNewLayoutEngineAdapter creates a new adapter for the layout engine
func NewNewLayoutEngineAdapter() *NewLayoutEngineAdapter {
	return &NewLayoutEngineAdapter{
		engine: layout.NewEngine(),
	}
}

// Layout performs layout using the new layout engine
func (a *NewLayoutEngineAdapter) Layout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	// Convert VNode/Fiber to layout.Node via adapter
	node := NewFiberToNodeAdapter(fiber, vnode)

	// Convert constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// Perform layout
	result := a.engine.Layout(node, layoutConstraints)

	return &newLayoutResultAdapter{
		result:     result,
		layoutType: LayoutEngineNew,
	}, nil
}

// LayoutFiber performs layout on a Fiber tree using the new layout engine
func (a *NewLayoutEngineAdapter) LayoutFiber(fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	// Convert Fiber to layout.Node
	node := NewFiberToNodeAdapter(fiber, nil)

	// Convert constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// Perform layout
	result := a.engine.Layout(node, layoutConstraints)

	return &newLayoutResultAdapter{
		result:     result,
		layoutType: LayoutEngineNew,
	}, nil
}

// GetType returns the engine type
func (a *NewLayoutEngineAdapter) GetType() LayoutEngineType {
	return LayoutEngineNew
}

// GetStats returns cache statistics
func (a *NewLayoutEngineAdapter) GetStats() CacheStats {
	stats := a.engine.GetStats()
	return CacheStats{
		Hits:   int(stats.CacheHits),
		Misses: int(stats.CacheMisses),
		Size:   0, // Not tracked in new engine
	}
}

// ClearCache clears the layout cache
func (a *NewLayoutEngineAdapter) ClearCache() {
	a.engine.Invalidate()
}

// SetDebug enables/disables debug output
func (a *NewLayoutEngineAdapter) SetDebug(debug bool) {
	// Layout engine doesn't have debug mode yet
}

// GetEngine returns the underlying layout engine
func (a *NewLayoutEngineAdapter) GetEngine() *layout.Engine {
	return a.engine
}

// newLayoutResultAdapter adapts layout.LayoutResult to LayoutResult
type newLayoutResultAdapter struct {
	result     *layout.LayoutResult
	layoutType LayoutEngineType
}

// GetRootBox returns the root box for painting
func (a *newLayoutResultAdapter) GetRootBox() PaintableBox {
	if a.result == nil || a.result.Root == nil {
		return nil
	}
	return &layoutBoxAdapter{LayoutBox: a.result.Root}
}

// GetHitMap returns the hit map (converted from layout.HitMap)
func (a *newLayoutResultAdapter) GetHitMap() *event.HitMap {
	if a.result == nil || a.result.HitMap == nil {
		return nil
	}
	// Convert layout.HitMap to event.HitMap
	return convertLayoutHitMap(a.result.HitMap)
}

// GetRenderPlanes returns render planes (not supported yet)
func (a *newLayoutResultAdapter) GetRenderPlanes() *layer.RenderPlanes {
	// Not yet supported in new layout engine
	return nil
}

// layoutBoxAdapter adapts layout.LayoutBox to PaintableBox
type layoutBoxAdapter struct {
	*layout.LayoutBox
}

// GetBounds returns the box bounds
func (a *layoutBoxAdapter) GetBounds() (x, y, width, height int) {
	if a.LayoutBox == nil {
		return 0, 0, 0, 0
	}
	return a.LayoutBox.X, a.LayoutBox.Y, a.LayoutBox.Width, a.LayoutBox.Height
}

// GetChildren returns child boxes
func (a *layoutBoxAdapter) GetChildren() []PaintableBox {
	if a.LayoutBox == nil {
		return nil
	}
	children := make([]PaintableBox, len(a.LayoutBox.Children))
	for i, child := range a.LayoutBox.Children {
		children[i] = &layoutBoxAdapter{LayoutBox: child}
	}
	return children
}

// convertLayoutHitMap converts layout.HitMap to event.HitMap
func convertLayoutHitMap(hm *layout.HitMap) *event.HitMap {
	if hm == nil {
		return nil
	}
	result := event.NewHitMap()
	// Iterate through layout.HitMap entries and convert
	// Note: This is a simplified conversion; full implementation would need
	// to handle all entries properly
	return result
}

// =============================================================================
// LayoutSwitcher - Manages switching between layout engines
// =============================================================================

// LayoutSwitcher manages switching between layout engines
type LayoutSwitcher struct {
	mu sync.RWMutex

	// Active engine type
	activeType LayoutEngineType

	// Engine instances
	computeEngine *ComputeEngineAdapter
	newEngine     *NewLayoutEngineAdapter

	// Comparison mode settings
	compareResults   bool
	logDifferences   bool
	tolerancePercent float64 // Tolerance for size differences (0-100)

	// Statistics
	stats SwitcherStats
}

// SwitcherStats tracks switching statistics
type SwitcherStats struct {
	mu              sync.RWMutex
	TotalRenders    int64
	ComputeRenders  int64
	NewEngineRenders int64
	BothRenders     int64
	Differences     int64
	Errors          int64
}

// NewLayoutSwitcher creates a new layout switcher
func NewLayoutSwitcher() *LayoutSwitcher {
	switcher := &LayoutSwitcher{
		activeType:       LayoutEngineCompute, // Default to stable
		compareResults:   false,
		logDifferences:   true,
		tolerancePercent: 1.0, // 1% tolerance
	}

	// Initialize both engines
	switcher.computeEngine = NewComputeEngineAdapter()
	switcher.newEngine = NewNewLayoutEngineAdapter()

	// Check environment variable for initial engine type
	if envEngine := os.Getenv("MINT_LAYOUT_ENGINE"); envEngine != "" {
		switcher.activeType = ParseLayoutEngineType(envEngine)
		log.PipelineLogger.Debug("[LayoutSwitcher] Engine type from env: %s", switcher.activeType)
	}

	return switcher
}

// SetEngineType sets the active engine type
func (s *LayoutSwitcher) SetEngineType(engineType LayoutEngineType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeType = engineType
	log.PipelineLogger.Debug("[LayoutSwitcher] Switched to engine: %s", engineType)
}

// GetEngineType returns the active engine type
func (s *LayoutSwitcher) GetEngineType() LayoutEngineType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeType
}

// SetCompareResults enables/disables result comparison in Both mode
func (s *LayoutSwitcher) SetCompareResults(compare bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.compareResults = compare
}

// SetTolerance sets the tolerance percentage for comparison
func (s *LayoutSwitcher) SetTolerance(percent float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tolerancePercent = percent
}

// Layout performs layout using the active engine
func (s *LayoutSwitcher) Layout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	s.mu.RLock()
	engineType := s.activeType
	s.mu.RUnlock()

	s.stats.mu.Lock()
	s.stats.TotalRenders++
	s.stats.mu.Unlock()

	switch engineType {
	case LayoutEngineCompute:
		s.stats.mu.Lock()
		s.stats.ComputeRenders++
		s.stats.mu.Unlock()
		return s.computeEngine.Layout(vnode, fiber, constraints)

	case LayoutEngineNew:
		s.stats.mu.Lock()
		s.stats.NewEngineRenders++
		s.stats.mu.Unlock()
		return s.newEngine.Layout(vnode, fiber, constraints)

	case LayoutEngineBoth:
		s.stats.mu.Lock()
		s.stats.BothRenders++
		s.stats.mu.Unlock()
		return s.layoutBoth(vnode, fiber, constraints)

	default:
		return s.computeEngine.Layout(vnode, fiber, constraints)
	}
}

// layoutBoth runs both engines and optionally compares results
func (s *LayoutSwitcher) layoutBoth(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (LayoutResult, error) {
	var computeResult, newResult LayoutResult
	var computeErr, newErr error
	var wg sync.WaitGroup

	// Run both engines in parallel
	wg.Add(2)
	go func() {
		defer wg.Done()
		computeResult, computeErr = s.computeEngine.Layout(vnode, fiber, constraints)
	}()
	go func() {
		defer wg.Done()
		newResult, newErr = s.newEngine.Layout(vnode, fiber, constraints)
	}()
	wg.Wait()

	// Track errors
	if computeErr != nil || newErr != nil {
		s.stats.mu.Lock()
		s.stats.Errors++
		s.stats.mu.Unlock()
	}

	// Compare results if enabled
	if s.compareResults && computeResult != nil && newResult != nil {
		if diff := s.compareLayoutResults(computeResult, newResult); len(diff) > 0 {
			s.stats.mu.Lock()
			s.stats.Differences++
			s.stats.mu.Unlock()
			if s.logDifferences {
				log.PipelineLogger.Debug("[LayoutSwitcher] Differences found: %v", diff)
			}
		}
	}

	// Return compute result by default (stable)
	if computeResult != nil {
		return computeResult, computeErr
	}
	return newResult, newErr
}

// compareLayoutResults compares two layout results and returns differences
func (s *LayoutSwitcher) compareLayoutResults(computeResult, newResult LayoutResult) []string {
	var differences []string

	computeBox := computeResult.GetRootBox()
	newBox := newResult.GetRootBox()

	if computeBox == nil && newBox == nil {
		return nil
	}
	if computeBox == nil || newBox == nil {
		return []string{"one result is nil"}
	}

	// Compare bounds
	cx, cy, cw, ch := computeBox.GetBounds()
	nx, ny, nw, nh := newBox.GetBounds()

	if cx != nx || cy != ny {
		differences = append(differences, 
			fmt.Sprintf("position differs: compute=(%d,%d) new=(%d,%d)", cx, cy, nx, ny))
	}

	// Check size with tolerance
	if cw > 0 && nw > 0 {
		widthDiff := float64(abs(cw-nw)) / float64(cw) * 100
		if widthDiff > s.tolerancePercent {
			differences = append(differences,
				fmt.Sprintf("width differs: compute=%d new=%d (%.1f%%)", cw, nw, widthDiff))
		}
	}

	if ch > 0 && nh > 0 {
		heightDiff := float64(abs(ch-nh)) / float64(ch) * 100
		if heightDiff > s.tolerancePercent {
			differences = append(differences,
				fmt.Sprintf("height differs: compute=%d new=%d (%.1f%%)", ch, nh, heightDiff))
		}
	}

	return differences
}

// GetStats returns statistics for both engines
func (s *LayoutSwitcher) GetStats() (computeStats, newStats CacheStats, switcherStats SwitcherStats) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	computeStats = s.computeEngine.GetStats()
	newStats = s.newEngine.GetStats()
	switcherStats = s.stats
	return
}

// ClearCache clears caches in both engines
func (s *LayoutSwitcher) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.computeEngine.ClearCache()
	s.newEngine.ClearCache()
}

// SetDebug sets debug mode for both engines
func (s *LayoutSwitcher) SetDebug(debug bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.computeEngine.SetDebug(debug)
	s.newEngine.SetDebug(debug)
}

// GetComputeEngine returns the compute engine adapter
func (s *LayoutSwitcher) GetComputeEngine() *ComputeEngineAdapter {
	return s.computeEngine
}

// GetNewEngine returns the new layout engine adapter
func (s *LayoutSwitcher) GetNewEngine() *NewLayoutEngineAdapter {
	return s.newEngine
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// =============================================================================
// ParallelRenderingPipeline - Main pipeline with switchable engines
// =============================================================================

// ParallelRenderingPipeline is a rendering pipeline with switchable layout engines
type ParallelRenderingPipeline struct {
	switcher    *LayoutSwitcher
	paintEngine *PaintEngine
	lastHitMap  *event.HitMap
	layerMgr    *layer.Manager
}

// NewParallelRenderingPipeline creates a new parallel rendering pipeline
func NewParallelRenderingPipeline() *ParallelRenderingPipeline {
	return &ParallelRenderingPipeline{
		switcher:    NewLayoutSwitcher(),
		paintEngine: NewPaintEngine(),
	}
}

// SetLayoutEngineType sets the active layout engine type
func (p *ParallelRenderingPipeline) SetLayoutEngineType(engineType LayoutEngineType) {
	p.switcher.SetEngineType(engineType)
}

// GetLayoutEngineType returns the active layout engine type
func (p *ParallelRenderingPipeline) GetLayoutEngineType() LayoutEngineType {
	return p.switcher.GetEngineType()
}

// SetCompareResults enables/disables result comparison
func (p *ParallelRenderingPipeline) SetCompareResults(compare bool) {
	p.switcher.SetCompareResults(compare)
}

// SetLayoutDebug enables/disables layout debug output
func (p *ParallelRenderingPipeline) SetLayoutDebug(debug bool) {
	p.switcher.SetDebug(debug)
}

// SetPaintDebug enables/disables paint debug output
func (p *ParallelRenderingPipeline) SetPaintDebug(debug bool) {
	p.paintEngine.SetDebug(debug)
}

// Render performs the complete rendering pipeline with the active layout engine
func (p *ParallelRenderingPipeline) Render(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) error {
	if vnode == nil {
		return nil
	}

	log.PipelineLogger.Debug("ParallelRender started with engine: %s", p.switcher.GetEngineType())

	// Phase 1: Layout using the active engine
	result, err := p.switcher.Layout(vnode, fiber, constraints)
	if err != nil {
		log.PipelineLogger.Debug("Layout FAILED: %v, falling back to legacy", err)
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	log.PipelineLogger.Debug("Layout complete, starting Paint phase")

	// Phase 2: Paint using computed layout
	// Note: For now, we need to convert our LayoutResult back to compute.ComputedLayout
	// for the existing PaintEngine. This will be improved later.
	if computeResult, ok := result.(*computeLayoutResultAdapter); ok {
		err = p.paintEngine.Paint(computeResult.ComputedLayout, buffer)
		p.lastHitMap = computeResult.GetHitMap()
	} else {
		// For new layout engine, we need a different painting approach
		// For now, fall back to legacy rendering
		log.PipelineLogger.Debug("New layout engine result - using legacy paint fallback")
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	log.PipelineLogger.Debug("Paint complete, err=%v", err)

	if p.lastHitMap != nil {
		log.PipelineLogger.Debug("Saved HitMap: %d entries", p.lastHitMap.Size())
	}

	return err
}

// RenderLayers renders with multi-layer support
func (p *ParallelRenderingPipeline) RenderLayers(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) error {
	// For now, delegate to regular Render
	// Full layer support will be added when new layout engine supports it
	return p.Render(vnode, fiber, constraints, buffer)
}

// renderLegacy fallback rendering
func (p *ParallelRenderingPipeline) renderLegacy(vnode rtui.VNode, x, y int, buffer *paint.Buffer) error {
	tempNode := NewDeclarativeNode(vnode)
	tempNode.PaintVNode(vnode, x, y, buffer)
	return nil
}

// GetSwitcher returns the layout switcher for direct access
func (p *ParallelRenderingPipeline) GetSwitcher() *LayoutSwitcher {
	return p.switcher
}

// GetPaintEngine returns the paint engine
func (p *ParallelRenderingPipeline) GetPaintEngine() *PaintEngine {
	return p.paintEngine
}

// GetLastHitMap returns the last HitMap from rendering
func (p *ParallelRenderingPipeline) GetLastHitMap() *event.HitMap {
	return p.lastHitMap
}

// GetStats returns statistics
func (p *ParallelRenderingPipeline) GetStats() (computeStats, newStats CacheStats, switcherStats SwitcherStats) {
	return p.switcher.GetStats()
}

// ClearCache clears all caches
func (p *ParallelRenderingPipeline) ClearCache() {
	p.switcher.ClearCache()
}

// =============================================================================
// Benchmark Support
// =============================================================================

// BenchmarkResult contains benchmark results for a single render pass
type BenchmarkResult struct {
	EngineType     LayoutEngineType
	LayoutDuration time.Duration
	PaintDuration  time.Duration
	TotalDuration  time.Duration
	CacheHits      int
	CacheMisses    int
	Error          error
}

// BenchmarkRender performs a benchmarked render pass
func (p *ParallelRenderingPipeline) BenchmarkRender(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) *BenchmarkResult {
	result := &BenchmarkResult{
		EngineType: p.switcher.GetEngineType(),
	}

	startTotal := time.Now()

	// Layout phase
	startLayout := time.Now()
	layoutResult, err := p.switcher.Layout(vnode, fiber, constraints)
	result.LayoutDuration = time.Since(startLayout)

	if err != nil {
		result.Error = err
		result.TotalDuration = time.Since(startTotal)
		return result
	}

	// Get cache stats
	computeStats, newStats, _ := p.switcher.GetStats()
	if result.EngineType == LayoutEngineCompute {
		result.CacheHits = computeStats.Hits
		result.CacheMisses = computeStats.Misses
	} else {
		result.CacheHits = newStats.Hits
		result.CacheMisses = newStats.Misses
	}

	// Paint phase
	startPaint := time.Now()
	if computeResult, ok := layoutResult.(*computeLayoutResultAdapter); ok {
		p.paintEngine.Paint(computeResult.ComputedLayout, buffer)
	}
	result.PaintDuration = time.Since(startPaint)

	result.TotalDuration = time.Since(startTotal)
	return result
}

// BenchmarkBoth runs both engines and returns comparison results
func (p *ParallelRenderingPipeline) BenchmarkBoth(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) (computeResult, newResult *BenchmarkResult) {
	// Benchmark compute engine
	p.SetLayoutEngineType(LayoutEngineCompute)
	computeResult = p.BenchmarkRender(vnode, fiber, constraints, buffer)

	// Benchmark new engine
	p.SetLayoutEngineType(LayoutEngineNew)
	// Note: For new engine, painting might not work yet
	newResult = &BenchmarkResult{
		EngineType: LayoutEngineNew,
	}
	startLayout := time.Now()
	_, newResult.Error = p.switcher.Layout(vnode, fiber, constraints)
	newResult.LayoutDuration = time.Since(startLayout)

	// Restore to compute
	p.SetLayoutEngineType(LayoutEngineCompute)

	return
}
