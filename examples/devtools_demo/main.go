// DevTools Demo - A comprehensive example demonstrating Mint DevTools usage
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/observation"
	"github.com/wwsheng009/mint/devtools/observation/v1"
	"github.com/wwsheng009/mint/devtools/observation/v2"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	demo := os.Args[1]

	switch demo {
	case "basic":
		runBasicDemo()
	case "observation":
		runObservationDemo()
	case "patterns":
		runPatternDemo()
	case "causal":
		runCausalDemo()
	case "insights":
		runInsightsDemo()
	case "all":
		runBasicDemo()
		fmt.Println("\n" + strings.Repeat("=", 60))
		runObservationDemo()
		fmt.Println("\n" + strings.Repeat("=", 60))
		runPatternDemo()
		fmt.Println("\n" + strings.Repeat("=", 60))
		runCausalDemo()
		fmt.Println("\n" + strings.Repeat("=", 60))
		runInsightsDemo()
	default:
		fmt.Printf("Unknown demo: %s\n", demo)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("DevTools Demo - Usage:")
	fmt.Println()
	fmt.Println("  go run main.go <demo_name>")
	fmt.Println()
	fmt.Println("Available demos:")
	fmt.Println("  basic       - Basic DevTools usage (enable, collect, disable)")
	fmt.Println("  observation - V1 Statistics layer (counts, top N, percentiles)")
	fmt.Println("  patterns    - V2 Pattern detection (oscillation, same-field, high-freq)")
	fmt.Println("  causal      - Causal chain tracking (event -> mutation -> layout)")
	fmt.Println("  insights    - Confidence-based insights generation")
	fmt.Println("  all         - Run all demos")
}

// runBasicDemo demonstrates basic DevTools usage
func runBasicDemo() {
	fmt.Println("=== Basic DevTools Demo ===")
	fmt.Println()

	// Create a new DevTools instance
	dt := devtools.New()

	// Enable DevTools
	fmt.Println("1. Enabling DevTools...")
	dt.Enable()
	fmt.Printf("   DevTools enabled: %v\n", dt.IsEnabled())

	// Simulate frame processing
	fmt.Println("\n2. Simulating frame processing...")
	for i := 0; i < 5; i++ {
		dt.BeginFrame()
		nodeID := fmt.Sprintf("node-%d", i)
		dt.RecordEvent("keypress", nodeID, "bubble", map[string]interface{}{
			"key":  fmt.Sprintf("key-%d", i),
			"code": i,
		})
		time.Sleep(50 * time.Millisecond)
		dt.EndFrame()
		fmt.Printf("   Frame %d processed\n", i+1)
	}

	// Access the event bus
	fmt.Println("\n3. Accessing event bus...")
	bus := dt.GetEventBus()
	fmt.Printf("   Event bus: %p\n", bus)

	// Highlight a component
	fmt.Println("\n4. Highlighting component...")
	dt.Highlight("test-component", 10, 10, 100, 50)
	if overlay := dt.GetOverlay(); overlay != nil {
		fmt.Printf("   Component 'test-component' highlighted: %v\n", overlay.IsShown("test-component"))
	}

	// Disable DevTools
	fmt.Println("\n5. Disabling DevTools...")
	dt.Disable()
	fmt.Printf("   DevTools enabled: %v\n", dt.IsEnabled())

	// Clean shutdown
	fmt.Println("\n6. Shutting down...")
	_ = dt.Shutdown()
	fmt.Println("   Done!")
}

// runObservationDemo demonstrates V1 statistics collection
func runObservationDemo() {
	fmt.Println("=== V1 Statistics Layer Demo ===")
	fmt.Println()

	// Create observation layer with default config
	cfg := observation.DefaultConfig()
	cfg.InitialLevel = v1.LevelEnhanced
	layer := observation.NewLayer(cfg)
	layer.LinkComponents()

	fmt.Println("1. Setting observation level to Enhanced...")
	layer.SetLevel(v1.LevelEnhanced)
	fmt.Printf("   Current level: %s\n", layer.GetLevel())

	// Simulate component mutations
	fmt.Println("\n2. Simulating component mutations...")
	components := []devtools.NodeID{"button_submit", "text_input", "label_status", "panel_main", "icon_save"}

	// Simulate different mutation patterns
	for i := 0; i < 50; i++ {
		nodeID := components[i%len(components)]
		fieldType := "state_update"
		fieldValue := i

		layer.RecordMutation(nodeID, fieldType, fieldValue)

		// Some components mutate more frequently
		if i%3 == 0 {
			layer.RecordMutation("button_submit", "click_count", i)
		}
		if i%5 == 0 {
			layer.RecordMutation("text_input", "text_change", i)
		}
	}

	fmt.Println("   Recorded 50 mutations across 5 components")

	// Get metrics
	fmt.Println("\n3. Getting metrics snapshot...")
	metrics := layer.GetMetrics()
	fmt.Printf("   Total frames: %d\n", metrics.TotalFrames)
	fmt.Printf("   Total mutations: %d\n", metrics.TotalMutations)
	fmt.Printf("   Total layouts: %d\n", metrics.TotalLayouts)

	// Get Top N components
	fmt.Println("\n4. Top 3 components by mutation count...")
	topN := layer.GetTopN(v1.MetricMutations, 3)
	for i, rank := range topN {
		fmt.Printf("   %d. %s: %d mutations\n", i+1, rank.NodeID, rank.Value)
	}

	// Get distribution
	fmt.Println("\n5. Mutation count distribution...")
	dist := layer.GetDistribution(v1.MetricMutations)
	if dist != nil {
		fmt.Printf("   Min: %d, Max: %d\n", dist.Min, dist.Max)
		fmt.Printf("   Mean: %.2f, StdDev: %.2f\n", dist.Mean, dist.StdDev)
		fmt.Printf("   Median: %d, P90: %d, P95: %d, P99: %d\n", dist.Median, dist.P90, dist.P95, dist.P99)
	}

	// Get component-specific metrics
	fmt.Println("\n6. Component-specific metrics (button_submit)...")
	cm := layer.GetComponentMetrics("button_submit")
	if cm != nil {
		fmt.Printf("   Mutations: %d\n", cm.MutationCount)
		fmt.Printf("   Layouts: %d\n", cm.LayoutCount)
		fmt.Printf("   Repaints: %d\n", cm.RepaintCount)
	}

	// Time series data
	fmt.Println("\n7. Time series data for button_submit...")
	layer.RecordMutation("button_submit", "test", 1)
	layer.RecordMutation("button_submit", "test", 2)
	layer.RecordMutation("button_submit", "test", 3)
	ts := layer.GetTimeSeries("button_submit")
	fmt.Printf("   Time series points: %d\n", len(ts))

	fmt.Println("\n8. Statistics summary...")
	stats := layer.GetStats()
	fmt.Printf("   Level: %s\n", stats.Level)
	fmt.Printf("   Expected overhead: %.2f%%\n", stats.Overhead*100)
	fmt.Printf("   Active patterns: %d\n", stats.PatternStats.ActivePatterns)
	fmt.Printf("   Active insights: %d\n", stats.InsightStats.ActiveInsights)
}

// runPatternDemo demonstrates V2 pattern detection
func runPatternDemo() {
	fmt.Println("=== V2 Pattern Detection Demo ===")
	fmt.Println()

	// Create observation layer with pattern detection enabled
	cfg := observation.DefaultConfig()
	cfg.InitialLevel = v1.LevelAdvanced // Enables pattern detection
	layer := observation.NewLayer(cfg)
	layer.LinkComponents()

	fmt.Println("1. Setting observation level to Advanced (enables pattern detection)...")
	layer.Enable(v1.LevelAdvanced)

	// Demo 1: Oscillation pattern (A -> B -> A -> B)
	fmt.Println("\n2. Detecting oscillation pattern...")
	oscillatingNode := devtools.NodeID("oscillating_toggle")
	for i := 0; i < 10; i++ {
		value := i % 2 // Alternates between 0 and 1
		layer.RecordMutation(oscillatingNode, "toggle_state", value)
		time.Sleep(10 * time.Millisecond)
	}
	patterns := layer.GetPatterns(oscillatingNode)
	fmt.Printf("   Recorded %d value oscillations\n", 10)
	if len(patterns) > 0 {
		for _, p := range patterns {
			fmt.Printf("   Pattern detected: %s (confidence: %.2f)\n", p.Type, p.Confidence)
		}
	} else {
		fmt.Println("   Oscillation pattern may need more cycles to detect")
	}

	// Demo 2: Same-field rapid changes
	fmt.Println("\n3. Detecting same-field rapid changes...")
	sameFieldNode := devtools.NodeID("rapid_field_node")
	for i := 0; i < 6; i++ {
		layer.RecordMutation(sameFieldNode, "user_input", i)
		time.Sleep(10 * time.Millisecond)
	}
	patterns = layer.GetPatterns(sameFieldNode)
	fmt.Printf("   Recorded 6 rapid changes to 'user_input' field\n")
	if len(patterns) > 0 {
		for _, p := range patterns {
			fmt.Printf("   Pattern detected: %s (confidence: %.2f)\n", p.Type, p.Confidence)
		}
	}

	// Demo 3: High frequency updates
	fmt.Println("\n4. Detecting high-frequency updates...")
	highFreqNode := devtools.NodeID("high_freq_component")
	for i := 0; i < 100; i++ {
		layer.RecordMutation(highFreqNode, "counter", i)
	}
	patterns = layer.GetPatterns(highFreqNode)
	fmt.Printf("   Recorded 100 rapid updates\n")
	if len(patterns) > 0 {
		for _, p := range patterns {
			if p.Type == v2.PatternHighFrequency {
				fmt.Printf("   Pattern detected: %s (confidence: %.2f)\n", p.Type, p.Confidence)
				for _, e := range p.Evidence {
					fmt.Printf("   Evidence: %s\n", e.Description)
				}
			}
		}
	}

	// Get all patterns
	fmt.Println("\n5. All detected patterns...")
	allPatterns := layer.GetAllPatterns()
	fmt.Printf("   Components with patterns: %d\n", len(allPatterns))
	for nodeID, pats := range allPatterns {
		fmt.Printf("   %s: %d patterns\n", nodeID, len(pats))
	}

	// Pattern statistics
	fmt.Println("\n6. Pattern detection statistics...")
	stats := layer.GetStats()
	ps := stats.PatternStats
	fmt.Printf("   Total patterns: %d\n", ps.TotalPatterns)
	fmt.Printf("   Oscillations: %d\n", ps.Oscillations)
	fmt.Printf("   Same field: %d\n", ps.SameFields)
	fmt.Printf("   High frequency: %d\n", ps.HighFrequency)
	fmt.Printf("   Average confidence: %.2f\n", ps.AvgConfidence)
}

// runCausalDemo demonstrates causal chain tracking
func runCausalDemo() {
	fmt.Println("=== Causal Chain Tracking Demo ===")
	fmt.Println()

	fmt.Println("1. Building causal chain...")
	fmt.Println("   Event: User presses 'Enter' key")
	fmt.Println("      -> Mutation: form_submit = true")
	fmt.Println("         -> Layout: Button repositioned")
	fmt.Println("            -> Repaint: Screen region updated")

	fmt.Println("\n2. Causal chain types:")
	fmt.Println("   EdgeEventToMutation  - Event causes state change")
	fmt.Println("   EdgeMutationToLayout - State causes layout recalculation")
	fmt.Println("   EdgeLayoutToRepaint  - Layout causes screen update")

	fmt.Println("\n3. In a real application, the causal chain tracks:")
	fmt.Println("   - Which events triggered which mutations")
	fmt.Println("   - Which mutations caused layout changes")
	fmt.Println("   - Which layout operations required repaints")
	fmt.Println("   - Enables: 'Why did this component re-render?'")

	fmt.Println("\n4. Use cases:")
	fmt.Println("   - Debug unexpected renders")
	fmt.Println("   - Identify performance bottlenecks")
	fmt.Println("   - Time-travel debugging (replay exact sequence)")
	fmt.Println("   - Deterministic replay for bug reproduction")

	fmt.Println("\n5. Example edge types in the causal graph:")
	fmt.Printf("   EdgeEventToMutation = %d\n", devtools.EdgeEventToMutation)
	fmt.Printf("   EdgeMutationToLayout = %d\n", devtools.EdgeMutationToLayout)
	fmt.Printf("   EdgeLayoutToRepaint = %d\n", devtools.EdgeLayoutToRepaint)
}

// runInsightsDemo demonstrates confidence-based insights
func runInsightsDemo() {
	fmt.Println("=== Confidence-Based Insights Demo ===")
	fmt.Println()

	// Create observation layer
	cfg := observation.DefaultConfig()
	cfg.InitialLevel = v1.LevelAdvanced
	layer := observation.NewLayer(cfg)
	layer.LinkComponents()
	layer.Enable(v1.LevelAdvanced)

	fmt.Println("1. Confidence Model (5-signal scoring):")
	cm := v2.NewConfidenceModel()
	weights := cm.GetWeights()
	fmt.Printf("   Statistical: %.2f (data distribution fit)\n", weights.Statistical)
	fmt.Printf("   Pattern:     %.2f (pattern match strength)\n", weights.Pattern)
	fmt.Printf("   Causal:      %.2f (causal link confidence - HIGHEST)\n", weights.Causal)
	fmt.Printf("   Context:     %.2f (context penalty factor)\n", weights.Context)
	fmt.Printf("   Historical:  %.2f (historical baseline)\n", weights.Historical)

	fmt.Println("\n2. Calculating confidence for a signal...")
	scores := &v2.SignalScores{
		Statistical: 0.8,
		Pattern:     0.6,
		Causal:      0.9,
		Context:     0.1,
		Historical:  0.7,
	}
	confidence := cm.Calculate(scores)
	fmt.Printf("   Signal scores: Statistical=%.2f, Pattern=%.2f, Causal=%.2f, Context=%.2f, Historical=%.2f\n",
		scores.Statistical, scores.Pattern, scores.Causal, scores.Context, scores.Historical)
	fmt.Printf("   Overall confidence: %.2f\n", confidence)
	fmt.Printf("   Confidence level: %s\n", v2.FromFloat64(confidence))

	// Generate some patterns to get insights
	fmt.Println("\n3. Generating patterns for insights...")
	for i := 0; i < 20; i++ {
		layer.RecordMutation(devtools.NodeID("insight_test_node"), "value", i%3)
	}

	fmt.Println("\n4. Generating insights from patterns...")
	insights := layer.GeneratePatternInsights()
	fmt.Printf("   Generated %d insights\n", len(insights))

	if len(insights) > 0 {
		fmt.Println("\n5. Example insight:")
		insight := insights[0]
		fmt.Printf("   ID: %s\n", insight.ID)
		fmt.Printf("   Type: %s\n", insight.Type)
		fmt.Printf("   Title: %s\n", insight.Title)
		fmt.Printf("   Confidence: %.2f (%s)\n", insight.Confidence, insight.Level)
		fmt.Printf("   Severity: %s\n", insight.Severity)

		if len(insight.Suggestions) > 0 {
			fmt.Println("   Suggestions:")
			for _, s := range insight.Suggestions {
				fmt.Printf("      - %s\n", s.Action)
				fmt.Printf("        Reason: %s\n", s.Reason)
				fmt.Printf("        Impact: %s\n", s.ExpectedImpact)
			}
		}
	}

	fmt.Println("\n6. High-confidence insights (threshold 0.7)...")
	highConf := layer.GetHighConfidenceInsights(0.7)
	fmt.Printf("   Found %d high-confidence insights\n", len(highConf))

	fmt.Println("\n7. Insight statistics...")
	stats := layer.GetStats()
	is := stats.InsightStats
	fmt.Printf("   Total insights: %d\n", is.TotalInsights)
	fmt.Printf("   Active insights: %d\n", is.ActiveInsights)
	fmt.Printf("   High confidence count: %d\n", is.HighConfidenceCount)
	fmt.Printf("   Average confidence: %.2f\n", is.AvgConfidence)

	fmt.Println("\n8. Insight levels:")
	fmt.Println("   ConfidenceNone:     < 0.35")
	fmt.Println("   ConfidenceLow:      0.35 - 0.50")
	fmt.Println("   ConfidenceMedium:   0.50 - 0.70")
	fmt.Println("   ConfidenceHigh:     0.70 - 0.85")
	fmt.Println("   ConfidenceVeryHigh: > 0.85")
}
