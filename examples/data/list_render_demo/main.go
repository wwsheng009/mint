package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/list"
)

// hitTestEntry is a sample hittest record used to populate the list.
type hitTestEntry struct {
	NodeID    string
	Bounds    string
	ZOrder    int
	HitTest   string
	Clickable bool
}

var sampleEntries = []hitTestEntry{
	{"button-ok", "(10,20,80,25)", 5, "YES", true},
	{"input-name", "(10,50,120,25)", 4, "NO", false},
	{"modal", "(0,0,120,40)", 3, "YES", true},
	{"container", "(5,5,110,30)", 2, "NO", false},
	{"root", "(0,0,120,40)", 1, "NO", false},
}

func buildRows() []string {
	rows := make([]string, 0, len(sampleEntries)+1)
	rows = append(rows, fmt.Sprintf("%-3s %-15s %-12s %-2s %-2s", "Z", "Node", "Bounds", "H", "C"))
	for i := range sampleEntries {
		e := sampleEntries[len(sampleEntries)-1-i]
		hitMark := "·"
		if e.HitTest == "YES" {
			hitMark = "✓"
		}
		clickMark := "·"
		if e.Clickable {
			clickMark = "Y"
		}
		rows = append(rows, fmt.Sprintf("%-3d %-15s %-12s %-2s %-2s",
			e.ZOrder, e.NodeID, e.Bounds, hitMark, clickMark))
	}
	return rows
}

func buildListVNode(rows []string) *list.VNode {
	headerStyle := style.Style{}.Bold(true)
	headerStyle.FG = style.Color("green")

	return list.NewBuilder().
		Header("🎯 Hit Test Data").
		Rows(rows).
		HeaderStyle(headerStyle).
		RowStyleFn(func(index int, text string) style.Style {
			if index == 0 {
				return style.Style{}.Bold(true).Foreground(style.Color("cyan"))
			}
			if strings.Contains(text, "✓") {
				return style.Style{}.Bold(true).Foreground(style.Color("yellow"))
			}
			return style.Style{}
		}).
		BuildVNode()
}

// ── Section 1: Size measurement ─────────────────────────────────────────────

func runSizeMeasurement(listVNode *list.VNode) {
	fmt.Println("\n=== 1. Size Measurement ===")

	constraints := runtime.BoxConstraints{
		MinWidth: 0, MaxWidth: 80,
		MinHeight: 0, MaxHeight: 30,
	}

	size := listVNode.Measure(constraints)
	fmt.Printf("Measured size: %dx%d\n", size.Width, size.Height)
	fmt.Printf("Configured:    (auto-sized)\n")

	if size.Width > 0 && size.Height > 0 {
		fmt.Println("✅ Measurement successful")
	} else {
		fmt.Println("❌ Measurement returned zero size")
	}
}

// ── Section 2: DrawCmd render analysis ──────────────────────────────────────

func describeStyle(s style.Style) string {
	desc := "Style["
	if s.FG != "" {
		desc += "FG:" + string(s.FG) + " "
	}
	if s.BG != "" {
		desc += "BG:" + string(s.BG) + " "
	}
	if s.IsBold() {
		desc += "Bold "
	}
	if s.IsItalic() {
		desc += "Italic "
	}
	if s.IsUnderline() {
		desc += "Underline "
	}
	desc += "]"
	return desc
}

func runDrawCmdAnalysis(listVNode *list.VNode) {
	fmt.Println("\n=== 2. DrawCmd Render Analysis ===")

	cmds := listVNode.Paint(0, 0)

	fmt.Printf("Total DrawCmds: %d\n", len(cmds))
	for i, cmd := range cmds {
		fmt.Printf("\nCmd %d: pos=(%d,%d) text=%q style=%s\n",
			i+1, cmd.X, cmd.Y, cmd.Text, describeStyle(cmd.Style))
	}
}

// ── Section 3: Render verification ──────────────────────────────────────────

func runRenderVerification(listVNode *list.VNode) {
	fmt.Println("\n=== 3. Render Verification ===")

	cmds := listVNode.Paint(0, 0)

	var (
		headerFound     bool
		colHeaderFound  bool
		hitHighlight    bool
		totalLines      int
	)

	for _, cmd := range cmds {
		if cmd.Text == "" {
			continue
		}
		totalLines++
		trimmed := strings.TrimSpace(cmd.Text)

		if strings.Contains(cmd.Text, "Hit Test Data") {
			headerFound = true
			fmt.Println("✅ Header found")
		}
		if strings.Contains(trimmed, "Z") && strings.Contains(trimmed, "Node") &&
			strings.Contains(trimmed, "H") && strings.Contains(trimmed, "C") {
			colHeaderFound = true
			fmt.Println("✅ Column header found")
		}
		if strings.Contains(cmd.Text, "✓") && cmd.Style.ToANSI() != "" {
			hitHighlight = true
			fmt.Println("✅ Hit-row highlight applied")
		}
	}

	fmt.Printf("\nStats: total_lines=%d header=%v col_header=%v hit_highlight=%v\n",
		totalLines, headerFound, colHeaderFound, hitHighlight)

	if headerFound && colHeaderFound && hitHighlight {
		fmt.Println("\n🎉 All expected content rendered correctly")
	} else {
		fmt.Println("\n❌ Some expected content is missing")
	}
}

// ── Section 4: Style summary ─────────────────────────────────────────────────

func runStyleSummary(listVNode *list.VNode) {
	fmt.Println("\n=== 4. Style Summary ===")

	cmds := listVNode.Paint(0, 0)

	unique := make(map[string]int)
	for _, cmd := range cmds {
		if cmd.Text == "" {
			continue
		}
		ansi := cmd.Style.ToANSI()
		if ansi == "" {
			ansi = "(default)"
		}
		unique[ansi]++
	}
	for s, n := range unique {
		fmt.Printf("  %s: %d lines\n", s, n)
	}
}

func main() {
	fmt.Println("=== List VNode Render Demo ===")
	fmt.Println("Builds a list with hittest data, measures it, and analyses DrawCmds.")

	rows := buildRows()
	listVNode := buildListVNode(rows)

	runSizeMeasurement(listVNode)
	runDrawCmdAnalysis(listVNode)
	runRenderVerification(listVNode)
	runStyleSummary(listVNode)

	fmt.Println("\n=== Done ===")
}
