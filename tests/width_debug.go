package main

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/mattn/go-runewidth"
)

func main() {
	fmt.Println("Testing character width calculations")
	fmt.Println("====================================")

	testChars := []rune{
		'→',  // U+2192 RIGHTWARDS ARROW
		'←',  // U+2190 LEFTWARDS ARROW
		'↑',  // U+2191 UPWARDS ARROW
		'↓',  // U+2193 DOWNWARDS ARROW
		'-',  // U+002D HYPHEN-MINUS (ASCII)
		'─',  // U+2500 BOX DRAWINGS LIGHT HORIZONTAL
		'你', // Chinese character (wide)
	}

	for _, r := range testChars {
		fmt.Printf("\nCharacter: %c (U+%04X)\n", r, r)
		fmt.Printf("  paint.RuneWidth:        %d\n", paint.RuneWidth(r))
		fmt.Printf("  paint.StringWidth:      %d\n", paint.StringWidth(string(r)))
		fmt.Printf("  runewidth.RuneWidth:    %d\n", runewidth.RuneWidth(r))
		fmt.Printf("  runewidth.StringWidth:  %d\n", runewidth.StringWidth(string(r)))
		fmt.Printf("  paint.CellWidthOfRune:  %d\n", paint.CellWidthOfRune(r))

		// Check if it's box-drawing character
		isBoxDrawing := r >= 0x2500 && r <= 0x257F
		fmt.Printf("  Is Box Drawing (U+2500-U+257F): %v\n", isBoxDrawing)

		// Expected width
		expected := 1
		if r < 128 {
			fmt.Printf("  Expected: %d (ASCII)\n", expected)
		} else if isBoxDrawing {
			fmt.Printf("  Expected: %d (Box Drawing, forced to 1)\n", expected)
		} else if r >= 0x4E00 && r <= 0x9FFF {
			expected = 2
			fmt.Printf("  Expected: %d (Chinese wide char)\n", expected)
		} else {
			fmt.Printf("  Expected: 1 (Unicode symbol)\n")
		}

		// Check if mismatch
		actual := paint.RuneWidth(r)
		if actual != expected {
			fmt.Printf("  >>> MISMATCH! Got %d, expected %d <<<\n", actual, expected)
		}
	}

	// Test StringWidth on strings
	fmt.Println("\n\nString width tests")
	fmt.Println("================")
	testStrings := []string{
		"→",
		"--->",
		"你好→",
		"Button →",
	}
	for _, s := range testStrings {
		fmt.Printf("\nString: %q\n", s)
		fmt.Printf("  paint.StringWidth:     %d\n", paint.StringWidth(s))
		fmt.Printf("  runewidth.StringWidth: %d\n", runewidth.StringWidth(s))
	}
}
