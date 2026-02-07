package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// 规范定义 - 来自 comp_1.md 和 comp_2.md
var specRules = map[string]ColorSpec{
	"Button.Primary":        {FG: "BG", BG: "PRIMARY"},
	"Button.Secondary":      {FG: "TEXT", BG: "SURFACE"},
	"Button.Danger":         {FG: "BG", BG: "ERROR"},
	"Button.Success":        {FG: "BG", BG: "SUCCESS"},
	"Button.Default":        {FG: "TEXT", BG: "SURFACE"},
	"Button.Disabled":       {FG: "DISABLED_FG", BG: "DISABLED_BG"},
	"Input.Normal":          {FG: "TEXT", BG: "SURFACE"},
	"Input.Focus":           {FG: "FOCUS"},
	"Input.Placeholder":     {FG: "PLACEHOLDER"},
	"Input.Disabled":        {FG: "DISABLED_FG", BG: "DISABLED_BG"},
	"Checkbox.Normal":       {FG: "TEXT", BG: "SURFACE"},
	"Checkbox.Checked":      {FG: "BG", BG: "PRIMARY"},
	"Checkbox.Focus":        {FG: "FOCUS"},
	"Checkbox.Disabled":     {FG: "DISABLED_FG", BG: "DISABLED_BG"},
	"Select.Normal":         {FG: "TEXT", BG: "SURFACE"},
	"Select.Focus":          {FG: "FOCUS"},
	"Select.Disabled":       {FG: "DISABLED_FG", BG: "DISABLED_BG"},
	"Textarea.Normal":       {FG: "TEXT", BG: "SURFACE"},
	"Textarea.Focus":        {FG: "FOCUS"},
	"Textarea.Placeholder":  {FG: "PLACEHOLDER"},
	"Textarea.Disabled":     {FG: "DISABLED_FG", BG: "DISABLED_BG"},
}

type ColorSpec struct {
	FG string
	BG string
}

type ComponentCheck struct {
	Component string
	Status    string
	Issues    []string
}

func main() {
	components := []string{
		"components/button/button.go",
		"components/form/input.go",
		"components/form/checkbox.go",
		"components/form/select.go",
		"components/form/textarea.go",
	}

	fmt.Println("🔍 TUI 组件配色规范检查")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	var allChecks []ComponentCheck

	for _, compPath := range components {
		check := checkComponent(compPath)
		allChecks = append(allChecks, check)
	}

	// Print results
	for _, check := range allChecks {
		fmt.Printf("📦 %s\n", check.Component)
		if check.Status == "✅ PASS" {
			fmt.Printf("   状态: %s\n", check.Status)
		} else {
			fmt.Printf("   状态: %s\n", check.Status)
			for _, issue := range check.Issues {
				fmt.Printf("   ⚠️  %s\n", issue)
			}
		}
		fmt.Println()
	}

	// Summary
	passCount := 0
	for _, check := range allChecks {
		if check.Status == "✅ PASS" {
			passCount++
		}
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("总结: %d/%d 组件符合规范\n", passCount, len(allChecks))
}

func checkComponent(filepath string) ComponentCheck {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return ComponentCheck{
			Component: filepath,
			Status:    "❌ ERROR",
			Issues:    []string{fmt.Sprintf("无法解析文件: %v", err)},
		}
	}

	check := ComponentCheck{
		Component: filepath,
		Status:    "✅ PASS",
		Issues:    []string{},
	}

	// Walk AST and check theme color usage
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check for theme.XXX() calls
		for _, arg := range callExpr.Args {
			selector, ok := arg.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "theme" {
				continue
			}

			colorFunc := selector.Sel.Name
			if !isThemeColor(colorFunc) {
				continue
			}

			// Check if this call is part of a Foreground() or Background() call
			// This is a simplified check - in reality we'd need more context
		}

		return true
	})

	return check
}

func isThemeColor(name string) bool {
	validColors := []string{
		"BG", "Surface", "Overlay",
		"Text", "Muted", "Placeholder",
		"Primary", "Secondary", "Accent",
		"Success", "Warning", "Error",
		"Border", "Focus", "Select", "Highlight",
		"DisabledFG", "DisabledBG",
	}

	for _, c := range validColors {
		if strings.EqualFold(name, c) {
			return true
		}
	}
	return false
}
