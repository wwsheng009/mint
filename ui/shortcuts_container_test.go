package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestContainerAndDisplayShortcuts(t *testing.T) {
	if vnode := Text("Hello"); vnode.Tag() != "text" {
		t.Fatalf("Text().Tag() = %q, want text", vnode.Tag())
	}
	if vnode := Divider(); vnode.Tag() != "divider" {
		t.Fatalf("Divider().Tag() = %q, want divider", vnode.Tag())
	}
	if vnode := HDivider(); vnode.Tag() != "divider" {
		t.Fatalf("HDivider().Tag() = %q, want divider", vnode.Tag())
	}
	if vnode := VDivider(); vnode.Tag() != "divider" {
		t.Fatalf("VDivider().Tag() = %q, want divider", vnode.Tag())
	}
	if vnode := Panel(Text("Body")); vnode.Tag() != "panel" {
		t.Fatalf("Panel().Tag() = %q, want panel", vnode.Tag())
	}
	tablePanel := TablePanel("Requests", Text("rows"), 126)
	if tablePanel.Tag() != "panel" {
		t.Fatalf("TablePanel().Tag() = %q, want panel", tablePanel.Tag())
	}
	if got := tablePanel.Props()["title"]; got != "Requests" {
		t.Fatalf("TablePanel title = %v, want Requests", got)
	}
	if got := tablePanel.Props()["width"]; got != 126 {
		t.Fatalf("TablePanel width = %v, want 126", got)
	}
	contentPanel := ContentPanel("Actions", Text("buttons"), "actions unavailable", 126)
	if contentPanel.Tag() != "panel" {
		t.Fatalf("ContentPanel().Tag() = %q, want panel", contentPanel.Tag())
	}
	if got := contentPanel.Props()["title"]; got != "Actions" {
		t.Fatalf("ContentPanel title = %v, want Actions", got)
	}
	if got := contentPanel.Props()["width"]; got != 126 {
		t.Fatalf("ContentPanel width = %v, want 126", got)
	}
	if got, ok := contentPanel.Props()["content"].(VNode); !ok || got.Tag() != "text" {
		t.Fatalf("ContentPanel content = %T, want text VNode", contentPanel.Props()["content"])
	}
	stackPanel := StackPanel("Alert History", []VNode{Text("Table"), Text("Toolbar")}, "history unavailable", 62)
	if stackPanel.Tag() != "panel" {
		t.Fatalf("StackPanel().Tag() = %q, want panel", stackPanel.Tag())
	}
	if got := stackPanel.Props()["title"]; got != "Alert History" {
		t.Fatalf("StackPanel title = %v, want Alert History", got)
	}
	stackContent, ok := stackPanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("StackPanel content = %T, want VNode", stackPanel.Props()["content"])
	}
	if stackContent.Tag() != "vstack" || len(stackContent.Children()) != 2 {
		t.Fatalf("StackPanel content = %s children %d, want vstack with two children", stackContent.Tag(), len(stackContent.Children()))
	}
	paddedStackPanel := PaddedStackPanel("Login", []VNode{Text("Base URL")}, "login form unavailable", 92, 1)
	if paddedStackPanel.Tag() != "panel" {
		t.Fatalf("PaddedStackPanel().Tag() = %q, want panel", paddedStackPanel.Tag())
	}
	if got := paddedStackPanel.Props()["title"]; got != "Login" {
		t.Fatalf("PaddedStackPanel title = %v, want Login", got)
	}
	if got := paddedStackPanel.Props()["width"]; got != 92 {
		t.Fatalf("PaddedStackPanel width = %v, want 92", got)
	}
	if got := paddedStackPanel.Props()["padding"]; got != 1 {
		t.Fatalf("PaddedStackPanel padding = %v, want 1", got)
	}
	paddedContent, ok := paddedStackPanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("PaddedStackPanel content = %T, want VNode", paddedStackPanel.Props()["content"])
	}
	if paddedContent.Tag() != "vstack" || len(paddedContent.Children()) != 1 {
		t.Fatalf("PaddedStackPanel content = %s children %d, want vstack with one child", paddedContent.Tag(), len(paddedContent.Children()))
	}
	panelRow := PanelRow(TablePanel("Left", Text("left"), 40), TablePanel("Right", Text("right"), 60))
	if panelRow.Tag() != "hstack" {
		t.Fatalf("PanelRow().Tag() = %q, want hstack", panelRow.Tag())
	}
	if got := panelRow.Props()["gap"]; got != 1 {
		t.Fatalf("PanelRow gap = %v, want 1", got)
	}
	if len(panelRow.Children()) != 2 {
		t.Fatalf("PanelRow children = %d, want 2", len(panelRow.Children()))
	}
	for index, child := range panelRow.Children() {
		if child.Tag() != "panel" {
			t.Fatalf("PanelRow child %d tag = %q, want panel", index, child.Tag())
		}
	}
	operationsPanel := OperationsPanel("Runtime Operations", []VNode{Text("Effective Sections")}, "runtime diagnostics unavailable", 126)
	if operationsPanel.Tag() != "panel" {
		t.Fatalf("OperationsPanel().Tag() = %q, want panel", operationsPanel.Tag())
	}
	if got := operationsPanel.Props()["title"]; got != "Runtime Operations" {
		t.Fatalf("OperationsPanel title = %v, want Runtime Operations", got)
	}
	if got := operationsPanel.Props()["width"]; got != 126 {
		t.Fatalf("OperationsPanel width = %v, want 126", got)
	}
	content, ok := operationsPanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("OperationsPanel content = %T, want VNode", operationsPanel.Props()["content"])
	}
	if content.Tag() != "vstack" || len(content.Children()) != 1 {
		t.Fatalf("OperationsPanel content = %s children %d, want vstack with one child", content.Tag(), len(content.Children()))
	}
	scopedOperationsPanel := OperationsPanelWithScope("Runtime Operations", []VNode{Text("Effective Sections")}, "runtime=available", "runtime diagnostics unavailable", 126)
	scopedContent, ok := scopedOperationsPanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("OperationsPanelWithScope content = %T, want VNode", scopedOperationsPanel.Props()["content"])
	}
	if scopedContent.Tag() != "vstack" || len(scopedContent.Children()) != 2 {
		t.Fatalf("OperationsPanelWithScope content = %s children %d, want vstack with scope", scopedContent.Tag(), len(scopedContent.Children()))
	}
	if got := scopedContent.Children()[1].Props()["content"]; got != "Scope: runtime=available" {
		t.Fatalf("OperationsPanelWithScope scope = %v, want Scope line", got)
	}
	scopedStackPanel := StackPanelWithScope("Distribution", []VNode{Text("chart")}, "source=analytics", "distribution unavailable", 126)
	scopedStackContent, ok := scopedStackPanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("StackPanelWithScope content = %T, want VNode", scopedStackPanel.Props()["content"])
	}
	if scopedStackContent.Tag() != "vstack" || len(scopedStackContent.Children()) != 2 {
		t.Fatalf("StackPanelWithScope content = %s children %d, want vstack with scope", scopedStackContent.Tag(), len(scopedStackContent.Children()))
	}
	if got := scopedStackContent.Children()[1].Props()["content"]; got != "Scope: source=analytics" {
		t.Fatalf("StackPanelWithScope scope = %v, want Scope line", got)
	}
	scopedTablePanel := TablePanelWithScope("Provider Summary", Text("table"), "provider=all", "summary unavailable", 72)
	scopedTableContent, ok := scopedTablePanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("TablePanelWithScope content = %T, want VNode", scopedTablePanel.Props()["content"])
	}
	if scopedTableContent.Tag() != "vstack" || len(scopedTableContent.Children()) != 2 {
		t.Fatalf("TablePanelWithScope content = %s children %d, want vstack with scope", scopedTableContent.Tag(), len(scopedTableContent.Children()))
	}
	if got := scopedTableContent.Children()[1].Props()["content"]; got != "Scope: provider=all" {
		t.Fatalf("TablePanelWithScope scope = %v, want Scope line", got)
	}
	linesPanel := LinesPanel("Capabilities", []PanelLine{
		PanelSuccessLine("ops-monitor/overview: available"),
		PanelMutedLine("applied filters: group=default"),
		PanelWarningLine("unsupported filters: provider"),
		PanelTextLine("custom", "cyan"),
	}, "No capabilities.", 126)
	if linesPanel.Tag() != "panel" {
		t.Fatalf("LinesPanel().Tag() = %q, want panel", linesPanel.Tag())
	}
	if got := linesPanel.Props()["title"]; got != "Capabilities" {
		t.Fatalf("LinesPanel title = %v, want Capabilities", got)
	}
	linesContent, ok := linesPanel.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("LinesPanel content = %T, want VNode", linesPanel.Props()["content"])
	}
	if linesContent.Tag() != "vstack" || len(linesContent.Children()) != 4 {
		t.Fatalf("LinesPanel content = %s children %d, want vstack with four children", linesContent.Tag(), len(linesContent.Children()))
	}
	lineStyle, ok := linesContent.Children()[2].Props()["style"].(style.Style)
	if !ok {
		t.Fatalf("LinesPanel third line style = %T, want style.Style", linesContent.Children()[2].Props()["style"])
	}
	if lineStyle.FG != "yellow" {
		t.Fatalf("LinesPanel third line fg = %q, want yellow", lineStyle.FG)
	}
	noticePanel := NoticePanel("Boundary", []string{"This page is read-only in this phase."}, 62)
	if noticePanel.Tag() != "panel" {
		t.Fatalf("NoticePanel().Tag() = %q, want panel", noticePanel.Tag())
	}
	if got := noticePanel.Props()["title"]; got != "Boundary" {
		t.Fatalf("NoticePanel title = %v, want Boundary", got)
	}
	if got := noticePanel.Props()["width"]; got != 62 {
		t.Fatalf("NoticePanel width = %v, want 62", got)
	}
	if vnode := ScrollView(Text("Body")); vnode.Tag() != "scrollview" {
		t.Fatalf("ScrollView().Tag() = %q, want scrollview", vnode.Tag())
	}
	if vnode := PageViewport(Text("Body"), 40, 8); vnode.Tag() != "pageviewport" {
		t.Fatalf("PageViewport().Tag() = %q, want pageviewport", vnode.Tag())
	} else if vnode.Props()["width"] != 40 || vnode.Props()["height"] != 8 {
		t.Fatalf("PageViewport size = %v/%v, want 40/8", vnode.Props()["width"], vnode.Props()["height"])
	}
	if vnode := PageViewportOffset(Text("Body"), 40, 8, 3); vnode.Props()["scrollOffset"] != 3 {
		t.Fatalf("PageViewportOffset scrollOffset = %v, want 3", vnode.Props()["scrollOffset"])
	}
	if vnode := Empty("No records"); vnode.Tag() != "empty" {
		t.Fatalf("Empty().Tag() = %q, want empty", vnode.Tag())
	}
}

func TestOperationalMetricRowShortcut(t *testing.T) {
	if item := MetricRowValue("Status", ""); item.Value != "-" {
		t.Fatalf("MetricRowValue blank = %v, want -", item.Value)
	}
	if item := MetricRowCompactValue("Trace", "trace-1234567890-extra", 12); item.Value != "trace-123..." {
		t.Fatalf("MetricRowCompactValue = %v, want trace-123...", item.Value)
	}
	if item := MetricRowCount("Errors", -1); item.Value != 0 {
		t.Fatalf("MetricRowCount = %v, want 0", item.Value)
	}
	if item := MetricRowItemWithWidth(MetricRowValue("Filters", "status=active"), 36); item.Width != 36 {
		t.Fatalf("MetricRowItemWithWidth width = %v, want 36", item.Width)
	}

	row := OperationalMetricRow([]MetricItem{
		{Title: "Runtime", Value: "healthy"},
		MetricRowItemWithWidth(MetricRowValue("Alerts", 2), 28),
	})
	if row.Tag() != "hstack" {
		t.Fatalf("OperationalMetricRow().Tag() = %q, want hstack", row.Tag())
	}
	if got := row.Props()["gap"]; got != 1 {
		t.Fatalf("gap = %v, want 1", got)
	}
	children := row.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	if got := children[0].Props()["width"]; got != 20 {
		t.Fatalf("metric width = %v, want 20", got)
	}
	if got := children[1].Props()["width"]; got != 28 {
		t.Fatalf("custom metric width = %v, want 28", got)
	}
}

func TestNewEmptyBuilder(t *testing.T) {
	vnode := NewEmptyBuilder().
		Description("Nothing here").
		Image("[ ]").
		Build()
	if vnode == nil {
		t.Fatal("NewEmptyBuilder().Build() returned nil")
	}
	if vnode.Tag() != "empty" {
		t.Fatalf("Tag() = %q, want empty", vnode.Tag())
	}
}

func TestDescriptionsPanelShortcut(t *testing.T) {
	fallbackItem := NewDescriptionsFallbackValue("Source", "", "n/a")
	if fallbackItem.Value != "n/a" {
		t.Fatalf("fallback item value = %v, want n/a", fallbackItem.Value)
	}
	compactItem := NewDescriptionsCompactValue("Path", "abcdefghijklmnopqrstuvwxyz", 12)
	if compactItem.Value != "abcdefghi..." {
		t.Fatalf("compact item value = %v, want abcdefghi...", compactItem.Value)
	}
	compactFallbackItem := NewDescriptionsCompactFallbackValue("Reason", "", "missing", 4)
	if compactFallbackItem.Value != "m..." {
		t.Fatalf("compact fallback item value = %v, want m...", compactFallbackItem.Value)
	}
	countItem := NewDescriptionsCountValue("Retries", -2)
	if countItem.Value != "0" {
		t.Fatalf("count item value = %v, want 0", countItem.Value)
	}
	ratioItem := NewDescriptionsRatioValue("Queue", 3, 5)
	if ratioItem.Value != "3/5" {
		t.Fatalf("ratio item value = %v, want 3/5", ratioItem.Value)
	}
	maskedItem := NewDescriptionsMaskedValue("Account", "account-billing-prod", 4, 4)
	if maskedItem.Value != "acco...prod" {
		t.Fatalf("masked item value = %v, want acco...prod", maskedItem.Value)
	}
	maskedFallbackItem := NewDescriptionsMaskedFallbackValue("Token", "", "-", 4, 4)
	if maskedFallbackItem.Value != "-" {
		t.Fatalf("masked fallback item value = %v, want -", maskedFallbackItem.Value)
	}
	stateItem := NewDescriptionsStateValue("Status", "failed", "failed")
	if stateItem.Content == nil || stateItem.Content.Tag() != "text" {
		t.Fatalf("state item content = %+v, want text node", stateItem.Content)
	}
	boolStateItem := NewDescriptionsBoolStateValue("Running", true, "yes", "no", "running", "idle")
	if boolStateItem.Content == nil || boolStateItem.Content.Props()["content"] != "yes" {
		t.Fatalf("bool state item content = %+v, want yes text", boolStateItem.Content)
	}
	enabledItem := NewDescriptionsEnabledValue("Enabled", false)
	if enabledItem.Content == nil || enabledItem.Content.Props()["content"] != "disabled" {
		t.Fatalf("enabled item content = %+v, want disabled text", enabledItem.Content)
	}

	node := DescriptionsPanel(
		"runtime.detail",
		"Runtime",
		56,
		12,
		36,
		[]DescriptionsItem{NewDescriptionsValue("Status", "healthy")},
		Text("Refresh"),
	)
	if node.Tag() != "panel" {
		t.Fatalf("DescriptionsPanel().Tag() = %q, want panel", node.Tag())
	}
	props := node.Props()
	if got := props["key"]; got != "runtime.detail" {
		t.Fatalf("key = %v, want runtime.detail", got)
	}
	if got := props["title"]; got != "Runtime" {
		t.Fatalf("title = %v, want Runtime", got)
	}
	content, ok := props["content"].(VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", props["content"])
	}
	if content.Tag() != "vstack" {
		t.Fatalf("content tag = %q, want vstack", content.Tag())
	}

	emptyNode := DescriptionsEmptyPanel("runtime.http", "HTTP Client", 56, "HTTP client diagnostics unavailable")
	if emptyNode.Tag() != "panel" {
		t.Fatalf("DescriptionsEmptyPanel().Tag() = %q, want panel", emptyNode.Tag())
	}
	emptyProps := emptyNode.Props()
	if got := emptyProps["key"]; got != "runtime.http" {
		t.Fatalf("empty panel key = %v, want runtime.http", got)
	}
	emptyContent, ok := emptyProps["content"].(VNode)
	if !ok {
		t.Fatalf("empty panel content = %T, want VNode", emptyProps["content"])
	}
	if emptyContent.Tag() != "empty" {
		t.Fatalf("empty panel content tag = %q, want empty", emptyContent.Tag())
	}
}
