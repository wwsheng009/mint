package descriptions

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func TestNew(t *testing.T) {
	v := New([]Item{Field("Name", "Mint")})
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "descriptions" {
		t.Fatalf("Tag = %q, want descriptions", v.Tag())
	}
	if v.column != 3 {
		t.Fatalf("column = %d, want 3", v.column)
	}
	if !v.colon {
		t.Fatal("colon should default to true")
	}
	if v.layout != LayoutHorizontal {
		t.Fatalf("layout = %v, want horizontal", v.layout)
	}
}

func TestBuilderFluent(t *testing.T) {
	extra := textcomp.New("Copy")
	v := NewBuilder().
		Key("build-meta").
		Title("Build Info").
		Extra(extra).
		Column(2).
		Bordered(true).
		Colon(false).
		Vertical().
		Width(64).
		LabelWidth(14).
		ContentWidth(28).
		EmptyText("-").
		MaskText("masked").
		LabelStyle(style.NewStyle().Bold(true)).
		ContentStyle(style.NewStyle().Foreground(style.Color("cyan"))).
		Item(Field("Version", "v1.2.3").WithKey("version")).
		BuildVNode()

	if v.Key() != "build-meta" {
		t.Fatalf("Key = %q, want build-meta", v.Key())
	}
	if v.title != "Build Info" || v.extra != extra {
		t.Fatalf("header props not set: title=%q extra=%v", v.title, v.extra)
	}
	if v.column != 2 || !v.bordered || v.colon {
		t.Fatalf("layout props = (column=%d bordered=%v colon=%v)", v.column, v.bordered, v.colon)
	}
	if v.layout != LayoutVertical {
		t.Fatalf("layout = %v, want vertical", v.layout)
	}
	if v.width != 64 {
		t.Fatalf("width = %d, want 64", v.width)
	}
	if v.labelWidth != 14 || v.contentWidth != 28 {
		t.Fatalf("width props = (label=%d content=%d), want (14,28)", v.labelWidth, v.contentWidth)
	}
	if v.emptyText != "-" || v.maskText != "masked" {
		t.Fatalf("text props = (empty=%q mask=%q)", v.emptyText, v.maskText)
	}
}

func TestPanelPresetWrapsDescriptionsAndActions(t *testing.T) {
	action := textcomp.New("Open")
	node := Panel(
		"logs.selection",
		"Log Detail",
		56,
		12,
		36,
		[]Item{Value("Request", "req-1")},
		action,
	)
	if node.Tag() != "panel" {
		t.Fatalf("panel tag = %q, want panel", node.Tag())
	}
	props := node.Props()
	if got := props["key"]; got != "logs.selection" {
		t.Fatalf("panel key = %v, want logs.selection", got)
	}
	if got := props["title"]; got != "Log Detail" {
		t.Fatalf("panel title = %v, want Log Detail", got)
	}
	if got := props["width"]; got != 56 {
		t.Fatalf("panel width = %v, want 56", got)
	}
	content, ok := props["content"].(rtui.VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", props["content"])
	}
	if content.Tag() != "vstack" {
		t.Fatalf("content tag = %q, want vstack", content.Tag())
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	details := children[0]
	if details.Tag() != "descriptions" || details.Key() != "logs.selection.details" {
		t.Fatalf("details = tag:%q key:%q", details.Tag(), details.Key())
	}
	if got := details.Props()[propLabelWidth]; got != 12 {
		t.Fatalf("details label width = %v, want 12", got)
	}
	if got := details.Props()[propContentWidth]; got != 36 {
		t.Fatalf("details content width = %v, want 36", got)
	}
}

func TestContextStripBuildsCompactDescriptionNode(t *testing.T) {
	node := ContextStrip(ContextStripConfig{
		Key:          "jobs.selection.context",
		Width:        62,
		Column:       3,
		LabelWidth:   7,
		ContentWidth: 12,
		Items: []Item{
			CompactValue("ID", "job-background-refresh", 12),
			StateValue("Status", "running", "running"),
			EnabledValue("Enabled", true),
		},
	})

	if node.Tag() != "descriptions" {
		t.Fatalf("context tag = %q, want descriptions", node.Tag())
	}
	if node.Key() != "jobs.selection.context" {
		t.Fatalf("context key = %q, want jobs.selection.context", node.Key())
	}
	props := node.Props()
	if got := props[propWidth]; got != 62 {
		t.Fatalf("context width = %v, want 62", got)
	}
	if got := props[propColumn]; got != 3 {
		t.Fatalf("context column = %v, want 3", got)
	}
	if got := props[propLabelWidth]; got != 7 {
		t.Fatalf("context label width = %v, want 7", got)
	}
	if got := props[propContentWidth]; got != 12 {
		t.Fatalf("context content width = %v, want 12", got)
	}
	items, ok := props[propItems].([]Item)
	if !ok {
		t.Fatalf("context items = %T, want []Item", props[propItems])
	}
	if len(items) != 3 {
		t.Fatalf("context item count = %d, want 3", len(items))
	}
	if items[0].Value != "job-backg..." {
		t.Fatalf("compact context id = %v, want job-backg...", items[0].Value)
	}
}

func TestPanelWithContextOrdersContextBeforeDetailsAndActions(t *testing.T) {
	action := textcomp.New("Next")
	node := PanelWithContext(
		"jobs.selection",
		"Job Detail",
		62,
		14,
		40,
		ContextStripConfig{
			LabelWidth:   7,
			ContentWidth: 12,
			Items: []Item{
				CompactValue("ID", "job-1", 12),
				StateValue("Status", "running", "running"),
			},
		},
		[]Item{Value("Current Run", "run-1")},
		action,
	)

	if node.Tag() != "panel" {
		t.Fatalf("panel tag = %q, want panel", node.Tag())
	}
	content, ok := node.Props()["content"].(rtui.VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", node.Props()["content"])
	}
	if content.Tag() != "vstack" {
		t.Fatalf("content tag = %q, want vstack", content.Tag())
	}
	children := content.Children()
	if len(children) != 3 {
		t.Fatalf("children len = %d, want context/details/action", len(children))
	}
	if children[0].Tag() != "descriptions" || children[0].Key() != "jobs.selection.context" {
		t.Fatalf("context child = tag:%q key:%q", children[0].Tag(), children[0].Key())
	}
	if children[1].Tag() != "descriptions" || children[1].Key() != "jobs.selection.details" {
		t.Fatalf("details child = tag:%q key:%q", children[1].Tag(), children[1].Key())
	}
	if children[2] != action {
		t.Fatal("action child not preserved")
	}
}

func TestDetailPanelConfigWrapsContextDetailsAndActions(t *testing.T) {
	action := textcomp.New("Open Trace")
	node := DetailPanel(DetailPanelConfig{
		Key:          "logs.selection",
		Title:        "Log Detail",
		Width:        56,
		LabelWidth:   12,
		ContentWidth: 36,
		Context: ContextStripConfig{
			Column:       2,
			LabelWidth:   8,
			ContentWidth: 18,
			Items: []Item{
				CompactValue("Request", "request-1", 18),
				StateValue("Status", "failed", "failed"),
			},
		},
		Items: []Item{
			CompactValue("Path", "/v1/chat/completions", 36),
			MaskedValue("Key", "provider-key-demo", 6, 4),
		},
		Actions: []rtui.VNode{action},
	})

	if node.Tag() != "panel" {
		t.Fatalf("panel tag = %q, want panel", node.Tag())
	}
	props := node.Props()
	if got := props["key"]; got != "logs.selection" {
		t.Fatalf("panel key = %v, want logs.selection", got)
	}
	if got := props["title"]; got != "Log Detail" {
		t.Fatalf("panel title = %v, want Log Detail", got)
	}
	if got := props["width"]; got != 56 {
		t.Fatalf("panel width = %v, want 56", got)
	}
	content, ok := props["content"].(rtui.VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", props["content"])
	}
	if content.Tag() != "vstack" {
		t.Fatalf("content tag = %q, want vstack", content.Tag())
	}
	children := content.Children()
	if len(children) != 3 {
		t.Fatalf("children len = %d, want context/details/action", len(children))
	}

	context := children[0]
	if context.Tag() != "descriptions" || context.Key() != "logs.selection.context" {
		t.Fatalf("context child = tag:%q key:%q", context.Tag(), context.Key())
	}
	if got := context.Props()[propWidth]; got != 56 {
		t.Fatalf("context width = %v, want inherited panel width 56", got)
	}
	if got := context.Props()[propColumn]; got != 2 {
		t.Fatalf("context column = %v, want 2", got)
	}

	details := children[1]
	if details.Tag() != "descriptions" || details.Key() != "logs.selection.details" {
		t.Fatalf("details child = tag:%q key:%q", details.Tag(), details.Key())
	}
	if got := details.Props()[propLabelWidth]; got != 12 {
		t.Fatalf("details label width = %v, want 12", got)
	}
	if got := details.Props()[propContentWidth]; got != 36 {
		t.Fatalf("details content width = %v, want 36", got)
	}
	if children[2] != action {
		t.Fatal("action child not preserved")
	}
}

func TestDetailPanelEmptyWhenShowsEmptyStateAndKeepsActions(t *testing.T) {
	action := textcomp.New("Next")
	node := DetailPanel(DetailPanelConfig{
		Key:       "jobs.selection",
		Title:     "Job Detail",
		Width:     62,
		EmptyWhen: true,
		EmptyText: "No job selected.",
		EmptyHint: "Clear filters or refresh jobs.",
		Context: ContextStripConfig{
			Items: []Item{CompactValue("ID", "job-1", 12)},
		},
		Items:   []Item{Value("Name", "Sync")},
		Actions: []rtui.VNode{action},
	})

	if node.Tag() != "panel" {
		t.Fatalf("panel tag = %q, want panel", node.Tag())
	}
	props := node.Props()
	if got := props["key"]; got != "jobs.selection" {
		t.Fatalf("panel key = %v, want jobs.selection", got)
	}
	if got := props["title"]; got != "Job Detail" {
		t.Fatalf("panel title = %v, want Job Detail", got)
	}
	if got := props["width"]; got != 62 {
		t.Fatalf("panel width = %v, want 62", got)
	}
	content, ok := props["content"].(rtui.VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", props["content"])
	}
	if content.Tag() != "vstack" {
		t.Fatalf("content tag = %q, want vstack", content.Tag())
	}
	children := content.Children()
	if len(children) != 3 {
		t.Fatalf("children len = %d, want empty/hint/action", len(children))
	}
	if children[0].Tag() != "empty" {
		t.Fatalf("first child tag = %q, want empty", children[0].Tag())
	}
	if got := children[0].Props()["description"]; got != "No job selected." {
		t.Fatalf("empty description = %v, want No job selected.", got)
	}
	if children[1].Tag() != "text" || children[1].Props()["content"] != "Clear filters or refresh jobs." {
		t.Fatalf("empty hint = %s %+v, want subtle hint text", children[1].Tag(), children[1].Props())
	}
	if children[2] != action {
		t.Fatal("action child not preserved")
	}
}

func TestDetailPanelEmptyHintBuildsRecoveryScopeText(t *testing.T) {
	got := DetailPanelEmptyHint("Refresh jobs or reset filters.",
		textcomp.KeyValuePart{Label: "status", Value: " active "},
		textcomp.KeyValuePart{Label: "search", Value: ""},
		textcomp.KeyValuePart{Label: "last", Value: "failed\n"},
	)
	want := "Refresh jobs or reset filters. Scope: status=active / last=failed"
	if got != want {
		t.Fatalf("empty hint = %q, want %q", got, want)
	}

	if got := DetailPanelEmptyHint("Refresh jobs."); got != "Refresh jobs." {
		t.Fatalf("hint without scope = %q, want action only", got)
	}
	if got := DetailPanelEmptyHint("", textcomp.KeyValuePart{Label: "source", Value: "unavailable"}); got != "Scope: source=unavailable" {
		t.Fatalf("hint without action = %q, want scope only", got)
	}
	if got := DetailPanelEmptyHintWithScopeWidth("Clear search.", 18, textcomp.KeyValuePart{Label: "search", Value: "abcdefghijklmnopqrstuvwxyz"}); got != "Clear search. Scope: search=abcdefgh..." {
		t.Fatalf("compact hint = %q, want compact scope", got)
	}
}

func TestDetailPanelEmptyWhenDefaultsEmptyText(t *testing.T) {
	node := DetailPanel(DetailPanelConfig{
		Title:     "Selection",
		EmptyWhen: true,
	})
	content, ok := node.Props()["content"].(rtui.VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", node.Props()["content"])
	}
	if content.Tag() != "empty" {
		t.Fatalf("content tag = %q, want empty", content.Tag())
	}
	if got := content.Props()["description"]; got != "No selection available." {
		t.Fatalf("empty description = %v, want No selection available.", got)
	}
}

func TestEmptyPanelPresetWrapsEmptyState(t *testing.T) {
	node := EmptyPanel("runtime.http", "HTTP Client", 56, "HTTP client diagnostics unavailable")
	if node.Tag() != "panel" {
		t.Fatalf("panel tag = %q, want panel", node.Tag())
	}
	props := node.Props()
	if got := props["key"]; got != "runtime.http" {
		t.Fatalf("panel key = %v, want runtime.http", got)
	}
	if got := props["title"]; got != "HTTP Client" {
		t.Fatalf("panel title = %v, want HTTP Client", got)
	}
	if got := props["width"]; got != 56 {
		t.Fatalf("panel width = %v, want 56", got)
	}
	content, ok := props["content"].(rtui.VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", props["content"])
	}
	if content.Tag() != "empty" {
		t.Fatalf("content tag = %q, want empty", content.Tag())
	}
	if got := content.Props()["description"]; got != "HTTP client diagnostics unavailable" {
		t.Fatalf("empty description = %v, want HTTP client diagnostics unavailable", got)
	}
}

func TestNormalizeItemsAssignsKeysAndSpan(t *testing.T) {
	items := normalizeItems([]Item{
		Field("Name", "Mint").WithKey("dup").WithSpan(0),
		Field("Env", "Prod").WithKey("dup").WithSpan(2),
	})
	if items[0].Key != "dup" {
		t.Fatalf("items[0].Key = %q, want dup", items[0].Key)
	}
	if items[1].Key != "dup-1" {
		t.Fatalf("items[1].Key = %q, want dup-1", items[1].Key)
	}
	if items[0].Span != 1 {
		t.Fatalf("items[0].Span = %d, want 1", items[0].Span)
	}
}

func TestPackRowsRespectsSpan(t *testing.T) {
	rows := packRows([]Item{
		Field("Name", "Mint"),
		Field("Version", "v1.2.3").WithSpan(2),
		Field("Commit", "abc"),
	}, 3)

	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(rows))
	}
	if len(rows[0]) != 2 || rows[0][0].span != 1 || rows[0][1].span != 2 {
		t.Fatalf("row 0 = %#v", rows[0])
	}
	if len(rows[1]) != 1 || rows[1][0].item.Label != "Commit" {
		t.Fatalf("row 1 = %#v", rows[1])
	}
}

func TestValueItemsRenderEmptyAndSensitiveValues(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn:    1,
		propEmptyText: "-",
		propMaskText:  "hidden",
		propItems: []Item{
			Value("Missing", nil).WithKey("missing"),
			Value("Blank", "").WithKey("blank"),
			SensitiveField("Token", "agw_example_token").WithKey("token"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "-") {
		t.Fatal("expected empty placeholder")
	}
	if !containsVNodeText(root, "hidden") {
		t.Fatal("expected sensitive value to be masked")
	}
	if containsVNodeText(root, "agw_example_token") {
		t.Fatal("sensitive value should not be rendered")
	}
}

func TestFallbackAndCompactValueItems(t *testing.T) {
	items := normalizeItems([]Item{
		FallbackValue("Blank", " ", "n/a").WithKey("blank"),
		CompactValue("ASCII", "abcdefghijklmnopqrstuvwxyz", 10).WithKey("ascii"),
		CompactFallbackValue("Wide", "服务端运行状态很长", "-", 8).WithKey("wide"),
	})

	if items[0].Value != "n/a" {
		t.Fatalf("fallback value = %v, want n/a", items[0].Value)
	}
	if items[1].Value != "abcdefg..." {
		t.Fatalf("compact ascii value = %v, want abcdefg...", items[1].Value)
	}
	if got := items[2].Value.(string); paint.StringWidth(got) > 8 {
		t.Fatalf("compact wide value width = %d, want <= 8 (%q)", paint.StringWidth(got), got)
	}
}

func TestCountAndRatioValueItemsUseOperationalText(t *testing.T) {
	items := normalizeItems([]Item{
		CountValue("Retries", -3).WithKey("retries"),
		RatioValue("Queue", -1, 12).WithKey("queue"),
		RatioValue("Active", 4, 7).WithKey("active"),
	})

	if items[0].Value != "0" {
		t.Fatalf("count value = %v, want 0", items[0].Value)
	}
	if items[1].Value != "0/12" {
		t.Fatalf("negative ratio value = %v, want 0/12", items[1].Value)
	}
	if items[2].Value != "4/7" {
		t.Fatalf("ratio value = %v, want 4/7", items[2].Value)
	}
}

func TestMaskedValueItemsDoNotStoreRawSensitiveText(t *testing.T) {
	items := normalizeItems([]Item{
		MaskedValue("Account", "account-billing-prod", 4, 4).WithKey("account"),
		MaskedFallbackValue("Missing", " ", "-", 2, 4).WithKey("missing"),
	})

	if items[0].Value != "acco...prod" {
		t.Fatalf("masked account value = %v, want acco...prod", items[0].Value)
	}
	if items[0].Value == "account-billing-prod" {
		t.Fatal("masked item should not keep the raw account value")
	}
	if items[1].Value != "-" {
		t.Fatalf("masked fallback value = %v, want -", items[1].Value)
	}
}

func TestStateValueItemUsesSemanticText(t *testing.T) {
	item := StateValue("Status", "failed", "failed").WithKey("status")
	if item.HasValue {
		t.Fatal("state value should use content node instead of raw value")
	}
	if item.Content == nil {
		t.Fatal("state value content is nil")
	}
	if item.Content.Tag() != "text" {
		t.Fatalf("state value content tag = %q, want text", item.Content.Tag())
	}
	if got := item.Content.Props()["content"]; got != "failed" {
		t.Fatalf("state value content = %v, want failed", got)
	}
	lineStyle, ok := item.Content.Props()["style"].(style.Style)
	if !ok {
		t.Fatalf("state value style = %T, want style.Style", item.Content.Props()["style"])
	}
	if lineStyle.FG != "red" {
		t.Fatalf("state value fg = %q, want red", lineStyle.FG)
	}
}

func TestBoolStateAndEnabledValueItemsUseSemanticText(t *testing.T) {
	running := BoolStateValue("Running", true, "yes", "no", "running", "idle")
	if running.Content == nil || running.Content.Props()["content"] != "yes" {
		t.Fatalf("running content = %+v, want yes text", running.Content)
	}
	runningStyle, ok := running.Content.Props()["style"].(style.Style)
	if !ok {
		t.Fatalf("running style = %T, want style.Style", running.Content.Props()["style"])
	}
	if runningStyle.FG != "yellow" {
		t.Fatalf("running fg = %q, want yellow", runningStyle.FG)
	}

	enabled := EnabledValue("Enabled", false)
	if enabled.Content == nil || enabled.Content.Props()["content"] != "disabled" {
		t.Fatalf("enabled content = %+v, want disabled text", enabled.Content)
	}
	enabledStyle, ok := enabled.Content.Props()["style"].(style.Style)
	if !ok {
		t.Fatalf("enabled style = %T, want style.Style", enabled.Content.Props()["style"])
	}
	if enabledStyle.FG != "red" {
		t.Fatalf("enabled fg = %q, want red", enabledStyle.FG)
	}
}

func TestRuntimeChildrenApplyStableLabelAndContentWidths(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn:       1,
		propLabelWidth:   12,
		propContentWidth: 24,
		propItems: []Item{
			Value("Runtime", "healthy").WithKey("runtime"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]

	label := findVNodeByKey(root, "descriptions-label-runtime")
	if label == nil {
		t.Fatal("expected label width wrapper")
	}
	if got := label.Props()["width"]; got != 12 {
		t.Fatalf("label width = %v, want 12", got)
	}

	content := findVNodeByKey(root, "descriptions-content-width-runtime")
	if content == nil {
		t.Fatal("expected content width wrapper")
	}
	if got := content.Props()["width"]; got != 24 {
		t.Fatalf("content width = %v, want 24", got)
	}
}

func TestInstanceRuntimeChildrenContainsHeaderAndValues(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:  "Build Info",
		propExtra:  textcomp.New("Copy"),
		propColumn: 2,
		propItems: []Item{
			Field("Name", "Mint").WithKey("name"),
			Field("Version", "v1.2.3").WithKey("version"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "Build Info") {
		t.Fatal("expected title text in runtime children")
	}
	if !containsVNodeText(root, "Copy") {
		t.Fatal("expected extra text in runtime children")
	}
	if !containsVNodeText(root, "Name:") {
		t.Fatal("expected label text with colon")
	}
	if !containsVNodeText(root, "Mint") || !containsVNodeText(root, "v1.2.3") {
		t.Fatal("expected values in runtime children")
	}
}

func TestInstanceVerticalLayoutOmitsColon(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn: 1,
		propColon:  false,
		propLayout: LayoutVertical,
		propItems: []Item{
			Field("Status", "Ready").WithKey("status"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "Status") {
		t.Fatal("expected label text without colon")
	}
	if containsVNodeText(root, "Status:") {
		t.Fatal("did not expect label text with colon")
	}
}

func TestInstanceEffectiveColumnsResponsiveToWidth(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn: 3,
		propWidth:  60,
		propItems:  []Item{Field("A", "1")},
	})
	if got := inst.effectiveColumns(); got != 2 {
		t.Fatalf("effectiveColumns = %d, want 2", got)
	}

	inst.SetProps(rtui.Props{
		propColumn: 3,
		propWidth:  40,
		propItems:  []Item{Field("A", "1")},
	})
	if got := inst.effectiveColumns(); got != 1 {
		t.Fatalf("effectiveColumns = %d, want 1", got)
	}
}

func TestNormalizeItemsClampsNegativeWidths(t *testing.T) {
	items := normalizeItems([]Item{
		Value("Runtime", "healthy").WithLabelWidth(-1).WithContentWidth(-2),
	})
	if items[0].LabelWidth != 0 {
		t.Fatalf("label width = %d, want 0", items[0].LabelWidth)
	}
	if items[0].ContentWidth != 0 {
		t.Fatalf("content width = %d, want 0", items[0].ContentWidth)
	}
}

func containsVNodeText(node rtui.VNode, want string) bool {
	if node == nil {
		return false
	}
	if contentProvider, ok := node.(interface{ Content() string }); ok && contentProvider.Content() == want {
		return true
	}
	if props := node.Props(); props != nil {
		if props.GetString("content") == want {
			return true
		}
	}
	for _, child := range node.Children() {
		if containsVNodeText(child, want) {
			return true
		}
	}
	return false
}

func findVNodeByKey(node rtui.VNode, key string) rtui.VNode {
	if node == nil {
		return nil
	}
	if node.Key() == key {
		return node
	}
	for _, child := range node.Children() {
		if found := findVNodeByKey(child, key); found != nil {
			return found
		}
	}
	return nil
}
