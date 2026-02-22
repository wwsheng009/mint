package visualizer

import (
	"fmt"
	"html"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
)

// PrintHTML prints the layout tree in HTML format.
func (v *Visualizer) PrintHTML() string {
	if v.rootID == "" {
		return "<html><body><h1>Empty layout tree</h1></body></html>"
	}

	var buf strings.Builder

	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n")
	buf.WriteString("<head>\n")
	buf.WriteString("  <meta charset=\"UTF-8\">\n")
	buf.WriteString("  <title>Layout Tree Visualization</title>\n")
	buf.WriteString("  <style>\n")
	buf.WriteString("    body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 20px; background: #f5f5f5; }\n")
	buf.WriteString("    .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }\n")
	buf.WriteString("    h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }\n")
	buf.WriteString("    .node { margin: 10px 0 padding: 15px; border: 1px solid #ddd; border-radius: 6px; background: #fff; }\n")
	buf.WriteString("    .node-title { font-weight: bold; color: #4CAF50; margin-bottom: 8px; }\n")
	buf.WriteString("    .node-title .tag { color: #2196F3; }\n")
	buf.WriteString("    .node-title .id { color: #999; font-size: 0.9em; margin-left: 8px; }\n")
	buf.WriteString("    .node-props { background: #f9f9f9; padding: 8px; border-radius: 4px; margin: 8px 0; font-family: 'Consolas', 'Monaco', monospace; font-size: 0.9em; }\n")
	buf.WriteString("    .prop { display: inline-block; margin: 2px 8px 2px 0; padding: 2px 6px; background: #e3f2fd; border-radius: 3px; }\n")
	buf.WriteString("    .constraints { display: flex; flex-wrap: wrap; gap: 8px; margin: 8px 0; }\n")
	buf.WriteString("    .constraint { flex: 1; min-width: 150px; padding: 8px; background: #fffbf0; border-left: 3px solid #ffc107; border-radius: 3px; }\n")
	buf.WriteString("    .constraint.warning { border-left-color: #ff9800; background: #fff3e0; }\n")
	buf.WriteString("    .constraint.error { border-left-color: #f44336; background: #ffebee; }\n")
	buf.WriteString("    .constraint-label { font-weight: bold; color: #666; font-size: 0.85em; display: block; margin-bottom: 4px; }\n")
	buf.WriteString("    .children { margin-left: 20px; padding-left: 20px; border-left: 2px solid #e0e0e0; margin-top: 10px; }\n")
	buf.WriteString("    .summary { background: #e8f5e9; padding: 15px; border-radius: 6px; margin-bottom: 20px; }\n")
	buf.WriteString("    .summary h2 { margin-top: 0; color: #2e7d32; }\n")
	buf.WriteString("    .summary-stat { display: inline-block; margin: 5px 15px 5px 0; }\n")
	buf.WriteString("    .summary-stat strong { color: #2e7d32; }\n")
	buf.WriteString("    .label { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 0.85em; margin-right: 5px; }\n")
	buf.WriteString("    .label.panel { background: #e3f2fd; color: #1976d2; }\n")
	buf.WriteString("    .label.border { background: #fff3e0; color: #f57c00; }\n")
	buf.WriteString("    .label.text { background: #f3e5f5; color: #7b1fa2; }\n")
	buf.WriteString("    .label.vstack { background: #e8f5e9; color: #388e3c; }\n")
	buf.WriteString("    .label.hstack { background: #fce4ec; color: #c2185b; }\n")
	buf.WriteString("    .label.grid { background: #e0f7fa; color: #0097a7; }\n")
	buf.WriteString("    .problem-list { background: #ffebee; padding: 12px; border-radius: 6px; margin: 10px 0; }\n")
	buf.WriteString("    .problem-list h3 { margin-top: 0; color: #c62828; }\n")
	buf.WriteString("    .problem { padding: 8px; margin: 4px 0; background: white; border-left: 3px solid #f44336; border-radius: 3px; }\n")
	buf.WriteString("    .tree-link { color: #4CAF50; text-decoration: none; }\n")
	buf.WriteString("    .tree-link:hover { text-decoration: underline; }\n")
	buf.WriteString("    .breadcrumb { margin: 10px 0; padding: 8px; background: #f5f5f5; border-radius: 4px; font-size: 0.9em; }\n")
	buf.WriteString("    .breadcrumb-item { color: #666; }\n")
	buf.WriteString("    .breadcrumb-separator { margin: 0 8px; color: #999; }\n")
	buf.WriteString("  </style>\n")
	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")
	buf.WriteString("  <div class=\"container\">\n")

	// Print summary
	buf.WriteString(v.printHTMLSummary())

	// Print problems if any
	problems := v.FindProblems()
	if len(problems) > 0 {
		buf.WriteString(v.printHTMLProblems(problems))
	}

	// Print layout tree
	buf.WriteString("    <h1>Layout Tree</h1>\n")
	v.printHTMLNodeRecursive(&buf, v.rootID, 0)

	buf.WriteString("  </div>\n")
	buf.WriteString("</body>\n")
	buf.WriteString("</html>\n")

	return buf.String()
}

// printHTMLSummary prints the summary section in HTML.
func (v *Visualizer) printHTMLSummary() string {
	if v.rootID == "" {
		return ""
	}

	var buf strings.Builder

	buf.WriteString("    <div class=\"summary\">\n")
	buf.WriteString("      <h2>Layout Summary</h2>\n")

	totalNodes := len(v.nodes)
	maxDepth := v.calculateDepth(v.rootID, 0)
	rootNode := v.nodes[v.rootID]

	buf.WriteString(fmt.Sprintf("      <div class=\"summary-stat\"><strong>Total Nodes:</strong> %d</div>\n", totalNodes))
	buf.WriteString(fmt.Sprintf("      <div class=\"summary-stat\"><strong>Max Depth:</strong> %d</div>\n", maxDepth))
	buf.WriteString(fmt.Sprintf("      <div class=\"summary-stat\"><strong>Root Size:</strong> %dw × %dh</div>\n", rootNode.Bounds.Width, rootNode.Bounds.Height))
	buf.WriteString(fmt.Sprintf("      <div class=\"summary-stat\"><strong>Root Position:</strong> (%d, %d)</div>\n", rootNode.Bounds.X, rootNode.Bounds.Y))
	buf.WriteString("<br>\n")

	// Count nodes by type
	typeCounts := make(map[string]int)
	for _, node := range v.nodes {
		typeCounts[node.Tag]++
	}

	buf.WriteString("      <div class=\"summary-stat\"><strong>Node Types:</strong></div>\n")
	for tag, count := range typeCounts {
		labelClass := tag
		if _, ok := map[string]bool{"panel": true, "border": true, "text": true, "vstack": true, "hstack": true, "grid": true}[labelClass]; !ok {
			labelClass = "panel"
		}
		buf.WriteString(fmt.Sprintf("      <span class=\"label %s\">%s</span> %d\n", labelClass, html.EscapeString(tag), count))
	}

	buf.WriteString("    </div>\n")

	return buf.String()
}

// printHTMLProblems prints the problems section in HTML.
func (v *Visualizer) printHTMLProblems(problems []string) string {
	var buf strings.Builder

	buf.WriteString("    <div class=\"problem-list\">\n")
	buf.WriteString(fmt.Sprintf("      <h3>⚠️  Found %d Layout Problems</h3>\n", len(problems)))

	for _, problem := range problems {
		buf.WriteString(fmt.Sprintf("      <div class=\"problem\">%s</div>\n", html.EscapeString(problem)))
	}

	buf.WriteString("    </div>\n")

	return buf.String()
}

// printHTMLNodeRecursive recursively prints nodes in HTML format.
func (v *Visualizer) printHTMLNodeRecursive(buf *strings.Builder, nodeID string, depth int) {
	node := v.nodes[nodeID]
	if node == nil {
		return
	}

	// Node container
	buf.WriteString("    <div class=\"node\" id=\"" + html.EscapeString("node-"+nodeID) + "\">\n")

	// Node title
	buf.WriteString("      <div class=\"node-title\">\n")
	buf.WriteString("        <span class=\"tag\">" + html.EscapeString(node.Tag) + "</span>\n")
	buf.WriteString(fmt.Sprintf("        <span class=\"id\">%s</span>\n", html.EscapeString(shortID(node.ID))))
	buf.WriteString("      </div>\n")

	// Breadcrumb (path to root)
	if depth > 0 {
		buf.WriteString("      <div class=\"breadcrumb\">\n")
		path := v.buildBreadcrumb(nodeID)
		buf.WriteString(path)
		buf.WriteString("      </div>\n")
	}

	// Constraints section
	buf.WriteString("      <div class=\"constraints\">\n")

	// Input constraints
	inputClass := "constraint"
	if v.hasConstraintWarning(node) {
		inputClass += " warning"
	}
	if v.hasConstraintError(node) {
		inputClass += " error"
	}

	buf.WriteString(fmt.Sprintf("        <div class=\"%s\">\n", inputClass))
	buf.WriteString("          <span class=\"constraint-label\">Input Constraints</span>\n")
	buf.WriteString("          <code>" + html.EscapeString(formatConstraints(node.InputConstraints)) + "</code>\n")
	buf.WriteString("        </div>\n")

	// Size
	buf.WriteString("        <div class=\"constraint\">\n")
	buf.WriteString("          <span class=\"constraint-label\">Measured Size</span>\n")
	buf.WriteString(fmt.Sprintf("          <code>%dw × %dh</code>\n", node.Bounds.Width, node.Bounds.Height))
	buf.WriteString("        </div>\n")

	// Position
	buf.WriteString("        <div class=\"constraint\">\n")
	buf.WriteString("          <span class=\"constraint-label\">Position</span>\n")
	buf.WriteString(fmt.Sprintf("          <code>(%d, %d)</code>\n", node.Bounds.X, node.Bounds.Y))
	buf.WriteString("        </div>\n")

	buf.WriteString("      </div>\n")

	// Output constraints if present
	if node.OutputConstraints != (layout.Constraints{}) {
		buf.WriteString("      <div class=\"node-props\">\n")
		buf.WriteString("        <strong>Constraints to Children:</strong><br>\n")
		buf.WriteString("        <code>" + html.EscapeString(formatConstraints(node.OutputConstraints)) + "</code>\n")
		buf.WriteString("      </div>\n")
	}

	// Additional properties
	if len(node.Props) > 0 {
		buf.WriteString("      <div class=\"node-props\">\n")
		for k, val := range node.Props {
			buf.WriteString(fmt.Sprintf("        <span class=\"prop\">%s = %s</span>\n",
				html.EscapeString(k), html.EscapeString(fmt.Sprint(val))))
		}
		buf.WriteString("      </div>\n")
	}

	// Children
	if len(node.Children) > 0 {
		buf.WriteString("      <div class=\"children\">\n")
		for _, childID := range node.Children {
			v.printHTMLNodeRecursive(buf, childID, depth+1)
		}
		buf.WriteString("      </div>\n")
	}

	buf.WriteString("    </div>\n")
}

// buildBreadcrumb builds a breadcrumb trail from root to node.
func (v *Visualizer) buildBreadcrumb(nodeID string) string {
	var path []string

	currentID := nodeID
	for currentID != "" {
		node := v.nodes[currentID]
		if node == nil {
			break
		}
		link := fmt.Sprintf("<a href=\"#node-%s\" class=\"tree-link\">%s</a>", node.ID, html.EscapeString(node.Tag))
		path = append([]string{link}, path...)
		currentID = node.ParentID
	}

	separator := "<span class=\"breadcrumb-separator»</span>"
	return strings.Join(path, separator)
}

// PrintHTMLOneline prints a single-line HTML representation.
func (v *Visualizer) PrintHTMLOneline() string {
	if v.rootID == "" {
		return "<div>Empty layout tree</div>"
	}

	var buf strings.Builder
	buf.WriteString("<div class=\"layout-tree\">")

	v.printHTMLInlineRecursive(&buf, v.rootID)

	buf.WriteString("</div>")
	return buf.String()
}

// printHTMLInlineRecursive prints nodes inline (compact format).
func (v *Visualizer) printHTMLInlineRecursive(buf *strings.Builder, nodeID string) {
	node := v.nodes[nodeID]
	if node == nil {
		return
	}

	// Node with tag and size
	buf.WriteString(fmt.Sprintf("<div class=\"node-inline\" title=\"%s\">%s <span class=\"size\">%dw×%dh</span>",
		html.EscapeString(formatConstraints(node.InputConstraints)),
		html.EscapeString(node.Tag),
		node.Bounds.Width,
		node.Bounds.Height))

	// Children
	if len(node.Children) > 0 {
		for _, childID := range node.Children {
			v.printHTMLInlineRecursive(buf, childID)
		}
	}

	buf.WriteString("</div>")
}

// PrintHTMLIndex prints an index of all nodes.
func (v *Visualizer) PrintHTMLIndex() string {
	if v.rootID == "" {
		return "<div>Empty layout tree</div>"
	}

	var buf strings.Builder

	buf.WriteString("<div class=\"node-index\">\n")
	buf.WriteString("  <h2>Node Index</h2>\n")
	buf.WriteString("  <table>\n")
	buf.WriteString("    <thead>\n")
	buf.WriteString("      <tr><th>ID</th><th>Tag</th><th>Size</th><th>Position</th><th>Parent</th></tr>\n")
	buf.WriteString("    </thead>\n")
	buf.WriteString("    <tbody>\n")

	// Sort nodes by ID
	sortedIDs := v.getSortedNodeIDs()

	for _, id := range sortedIDs {
		node := v.nodes[id]
		buf.WriteString("      <tr>\n")
		buf.WriteString(fmt.Sprintf("        <td><a href=\"#node-%s\">%s</a></td>\n", id, html.EscapeString(shortID(id))))
		buf.WriteString(fmt.Sprintf("        <td><span class=\"label %s\">%s</span></td>\n",
			getLabelClass(node.Tag), html.EscapeString(node.Tag)))
		buf.WriteString(fmt.Sprintf("        <td>%dw × %dh</td>\n", node.Bounds.Width, node.Bounds.Height))
		buf.WriteString(fmt.Sprintf("        <td>(%d, %d)</td>\n", node.Bounds.X, node.Bounds.Y))
		if node.ParentID != "" {
			buf.WriteString(fmt.Sprintf("        <td><a href=\"#node-%s\">%s</a></td>\n", node.ParentID, html.EscapeString(shortID(node.ParentID))))
		} else {
			buf.WriteString("        <td>-</td>\n")
		}
		buf.WriteString("      </tr>\n")
	}

	buf.WriteString("    </tbody>\n")
	buf.WriteString("  </table>\n")
	buf.WriteString("</div>\n")

	return buf.String()
}

// Helper functions for HTML output

func (v *Visualizer) hasConstraintWarning(node *NodeState) bool {
	return (node.Dimension.Height > node.InputConstraints.MaxHeight ||
		node.Dimension.Width > node.InputConstraints.MaxWidth) &&
		!v.hasConstraintError(node)
}

func (v *Visualizer) hasConstraintError(node *NodeState) bool {
	return (node.Dimension.Height > node.InputConstraints.MaxHeight && node.InputConstraints.MaxHeight < layout.MaxInt) ||
		(node.Dimension.Width > node.InputConstraints.MaxWidth && node.InputConstraints.MaxWidth < layout.MaxInt)
}

func getLabelClass(tag string) string {
	labelClass := tag
	if _, ok := map[string]bool{"panel": true, "border": true, "text": true, "vstack": true, "hstack": true, "grid": true}[labelClass]; !ok {
		labelClass = "panel"
	}
	return labelClass
}

func (v *Visualizer) getSortedNodeIDs() []string {
	ids := make([]string, 0, len(v.nodes))
	for id := range v.nodes {
		ids = append(ids, id)
	}
	// Simple sort (optional - could use sort package)
	return ids
}
