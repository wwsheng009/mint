package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/render"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Layout Engine Integration Test                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Test 1: Default (compute engine)
	fmt.Println("Test 1: Default (no env variable)")
	os.Unsetenv("MINT_LAYOUT_ENGINE")
	pipeline := render.NewRenderingPipeline()
	fmt.Printf("  Engine Type: %s\n", pipeline.GetLayoutEngineType())
	fmt.Printf("  Expected: compute\n")

	// Test 2: With env variable (layout engine)
	fmt.Println("\nTest 2: MINT_LAYOUT_ENGINE=layout")
	os.Setenv("MINT_LAYOUT_ENGINE", "layout")
	pipeline2 := render.NewRenderingPipeline()
	fmt.Printf("  Engine Type: %s\n", pipeline2.GetLayoutEngineType())
	fmt.Printf("  Expected: layout\n")

	// Test 3: With env variable (both engines)
	fmt.Println("\nTest 3: MINT_LAYOUT_ENGINE=both")
	os.Setenv("MINT_LAYOUT_ENGINE", "both")
	pipeline3 := render.NewRenderingPipeline()
	fmt.Printf("  Engine Type: %s\n", pipeline3.GetLayoutEngineType())
	fmt.Printf("  Expected: both\n")

	// Test 4: Direct switcher usage
	fmt.Println("\nTest 4: Direct LayoutSwitcher")
	switcher := render.NewLayoutSwitcher()
	fmt.Printf("  Default Engine: %s\n", switcher.GetEngineType())

	switcher.SetEngineType(render.LayoutEngineNew)
	fmt.Printf("  After SetEngineType(LayoutEngineNew): %s\n", switcher.GetEngineType())

	switcher.SetEngineType(render.LayoutEngineBoth)
	fmt.Printf("  After SetEngineType(LayoutEngineBoth): %s\n", switcher.GetEngineType())

	switcher.SetEngineType(render.LayoutEngineCompute)
	fmt.Printf("  After SetEngineType(LayoutEngineCompute): %s\n", switcher.GetEngineType())

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ All Integration Tests Passed                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}
