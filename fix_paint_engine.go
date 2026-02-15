package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
 inputFile := "internal/render/paint_engine.go"
	outputFile := "internal/render/paint_engine.go"

	// Read input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	fixed := make([]string, 0, len(lines))

	// Patterns to fix - add nil check before GetVNode() calls
	patterns := []struct{
		find    string
		wrapLine string
	}{
		{
			find: `		vnode := box.GetVNode\(\)`,
			wrapLine: `		vnode := box.GetVNode()
		if vnode == nil {
			// No VNode
			return
		}`,
		},
		{
			find: `	if _, ok := box\.GetVNode\(\)\.Type\(\)`,
			wrapLine: `	if _, ok := vnode.(interface { Type() rtui.VNodeType }); ok`,
		},
		{
			find: `	switch vnode\.Type\(\)`,
			wrapLine: `	switch vnode.Type()`,
		},
		{
			find: `	paintable, ok := vnode\.\(interface`,
			wrapLine: `	paintable, ok := vnode.(interface`,
		},
		{
			find: `	text = rtui\.GetTextContent\(box\.GetVNode\(\)\)`,
			wrapLine: `	text = rtui.GetTextContent(vnode)`,
		},
		{
			find: `	content = rtui\.GetTextContent\(box\.GetVNode\(\)\)`,
			wrapLine: `	content = rtui.GetTextContent(vnode)`,
		},
		{
			find: `	buffer\.SetStringAligned\(box\.Box\.X, box\.Box\.Y, text, box\.GetVNode\(\)\.Style\(\), maxX\)`,
			wrapLine: `	buffer.SetStringAligned(box.Box.X, box.Box.Y, text, vnode.Style(), maxX)`,
		},
		{
			find: `	nodeStyle := box\.GetVNode\(\)\.Style\(\)`,
			wrapLine: `	nodeStyle := vnode.Style()`,
		},
		{
			find: `	box\.GetVNode\(\)\.SetStyle\(inheritedStyle\)`,
			wrapLine: `	vnode.SetStyle(inheritedStyle)`,
		},
		{
			find: `	children := box\.GetVNode\(\)\.Children\(\)`,
			wrapLine: `	children := vnode.Children()`,
		},
		{
			find: `	if tagger, ok := box\.GetVNode\(\)\.\(interface{ Tag\(\) string }\); ok`,
			wrapLine: `	if tagger, ok := vnode.(interface { Tag() string }); ok`,
		},
		{
			find: `	if labeled, ok := box\.GetVNode\(\)\.\(interface{ GetBorderStyle\(\) rtui\.BorderStyle }\); ok`,
			wrapLine: `	if labeled, ok := vnode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok`,
		},
		{
			find: `	if colored, ok := box\.GetVNode\(\)\.\(interface{ GetBorderColor\(\) string }\); ok`,
			wrapLine: `	if colored, ok := vnode.(interface{ GetBorderColor() string }); ok`,
		},
		{
			find: `	if labeled, ok := box\.GetVNode\(\)\.\(interface{ GetBorderLabel\(\) string }\); ok`,
			wrapLine: `	if labeled, ok := vnode.(interface{ GetBorderLabel() string }); ok`,
		},
	}

	// Track vnode variable to ensure proper scoping
	vnodeVar := "vnode"
	setVnodeVar := false

	for i, line := range lines {
		fixedLine := line
		setVnodeVarCurrent := setVnodeVar

		// Apply fixes
		for _, pattern := range patterns {
			if strings.Contains(line, pattern.find) {
				fixedLine = strings.Replace(line, pattern.find, pattern.wrapLine)
				// Track vnode variable usage
				if strings.Contains(pattern.wrapLine, "vnode := ") {
					setVnodeVarCurrent = true
				}
				break
			}
		}

		// Ensure vnode variable is set before each use
		if setVnodeVarCurrent {
			// Find indentation of current line
			indent := ""
			for j := 0; j < len(line); j++ {
				if line[j] == ' ' || line[j] == '\t' {
					indent += string(line[j])
				} else {
					break
				}
			}

			// Insert vnode declaration after line
			fixedLine = line + "\n" + indent + "vnode := box.GetVNode()"
			setVnodeVarCurrent = false
		}

		fixed[i] = fixedLine
	}

	// Write output
	output := strings.Join(fixed, "\n")
	err = os.WriteFile(outputFile, []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fixed %d lines in %s\n", len(fixed), outputFile)
}
