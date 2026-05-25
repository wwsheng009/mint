package transfer

import (
	"reflect"
	"testing"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/list"
)

func TestNewDefaults(t *testing.T) {
	node := New()
	if node.Tag() != "transfer" {
		t.Fatalf("Tag = %q, want transfer", node.Tag())
	}
	if node.listWidth != defaultListWidth || node.listHeight != defaultListHeight {
		t.Fatalf("unexpected list defaults: width=%d height=%d", node.listWidth, node.listHeight)
	}
	if node.titles != defaultTitles {
		t.Fatalf("unexpected titles: %#v", node.titles)
	}
}

func TestBuilderFluentAPI(t *testing.T) {
	node := NewBuilder().
		Key("members").
		ComponentID("members-transfer").
		Titles("Available", "Chosen").
		Operations(">>", "<<").
		BulkOperations(true).
		BulkOperationLabels("All >>", "<< All").
		Searchable(true).
		SearchPlaceholders("Find available", "Find chosen").
		InitialSearchValues("alp", "").
		InitialTargetKeys([]string{"b"}).
		ListWidth(30).
		ListHeight(10).
		PageSize(5).
		Width(72).
		Items([]Item{
			NewItem("a", "Alpha"),
			NewItem("b", "Beta"),
		}).
		BuildVNode()

	if node.Key() != "members" {
		t.Fatalf("Key = %q, want members", node.Key())
	}
	if node.componentID != "members-transfer" {
		t.Fatalf("componentID = %q, want members-transfer", node.componentID)
	}
	if node.titles != [2]string{"Available", "Chosen"} {
		t.Fatalf("titles = %#v", node.titles)
	}
	if node.operations != [2]string{">>", "<<"} {
		t.Fatalf("operations = %#v", node.operations)
	}
	if !node.bulkOperations || node.bulkOperationLabels != [2]string{"All >>", "<< All"} {
		t.Fatalf("bulk config = enabled:%v labels:%#v, want enabled with custom labels", node.bulkOperations, node.bulkOperationLabels)
	}
	if !node.searchable || node.searchControlled {
		t.Fatalf("search config = searchable:%v controlled:%v, want searchable uncontrolled", node.searchable, node.searchControlled)
	}
	if node.searchPlaceholders != [2]string{"Find available", "Find chosen"} || node.sourceSearch != "alp" {
		t.Fatalf("search values = placeholders:%#v source:%q target:%q", node.searchPlaceholders, node.sourceSearch, node.targetSearch)
	}
	if !reflect.DeepEqual(node.targetKeys, []string{"b"}) {
		t.Fatalf("targetKeys = %#v, want []string{\"b\"}", node.targetKeys)
	}
	if node.targetKeysControlled {
		t.Fatal("InitialTargetKeys should keep uncontrolled mode")
	}
	if node.listWidth != 30 || node.listHeight != 10 || node.pageSize != 5 || node.width != 72 {
		t.Fatalf("unexpected sizing: listWidth=%d listHeight=%d pageSize=%d width=%d", node.listWidth, node.listHeight, node.pageSize, node.width)
	}
}

func TestRuntimeChildrenBuildsDualLists(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma"},
		},
		propTargetKeys: []string{"c"},
	})
	inst.selectedSourceKeys = []string{"a"}

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}

	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack", root.Tag())
	}
	if len(root.Children()) != 3 {
		t.Fatalf("root children len = %d, want 3", len(root.Children()))
	}

	leftWrapper := root.Children()[0]
	leftList := leftWrapper.Children()[0]
	if leftList.Tag() != "list" {
		t.Fatalf("left tag = %q, want list", leftList.Tag())
	}
	if header, _ := leftList.Props()["header"].(string); header != "Source (2)" {
		t.Fatalf("left header = %q, want Source (2)", header)
	}
	if checked, _ := leftList.Props()["checkedIndices"].([]int); !reflect.DeepEqual(checked, []int{0}) {
		t.Fatalf("left checkedIndices = %#v, want []int{0}", checked)
	}

	rightWrapper := root.Children()[2]
	rightList := rightWrapper.Children()[0]
	if header, _ := rightList.Props()["header"].(string); header != "Target (1)" {
		t.Fatalf("right header = %q, want Target (1)", header)
	}
}

func TestRuntimeChildrenBulkOperations(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:         "members",
		propBulkOperations:      true,
		propBulkOperationLabels: [2]string{"All Send", "All Return"},
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma", Disabled: true},
			{Key: "d", Title: "Delta"},
		},
		propTargetKeys: []string{"d"},
	})

	children := inst.RuntimeChildren()
	operations := children[0].Children()[1]
	buttons := operations.Children()
	if len(buttons) != 4 {
		t.Fatalf("operation buttons len = %d, want 4", len(buttons))
	}

	expected := []struct {
		key      string
		label    string
		disabled bool
	}{
		{"members-to-target", ">", true},
		{"members-to-source", "<", true},
		{"members-to-target-all", "All Send", false},
		{"members-to-source-all", "All Return", false},
	}
	for index, item := range expected {
		button := buttons[index]
		if button.Key() != item.key {
			t.Fatalf("button[%d] key = %q, want %q", index, button.Key(), item.key)
		}
		if label, _ := button.Props()["label"].(string); label != item.label {
			t.Fatalf("button[%d] label = %q, want %q", index, label, item.label)
		}
		if disabled, _ := button.Props()["disabled"].(bool); disabled != item.disabled {
			t.Fatalf("button[%d] disabled = %v, want %v", index, disabled, item.disabled)
		}
	}
}

func TestRuntimeChildrenSearchableFiltersLists(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:        "members",
		propSearchable:         true,
		propSearchPlaceholders: [2]string{"Find source", "Find target"},
		propSourceSearch:       "alp",
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma"},
			{Key: "d", Title: "Delta"},
		},
		propTargetKeys: []string{"d"},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	leftWrapper := root.Children()[0]
	if len(leftWrapper.Children()) != 2 {
		t.Fatalf("searchable wrapper children len = %d, want 2", len(leftWrapper.Children()))
	}
	searchInput := leftWrapper.Children()[0]
	if searchInput.Tag() != "input" || searchInput.ID() != inst.sourceSearchField() {
		t.Fatalf("search input tag/id = %q/%q, want input/%q", searchInput.Tag(), searchInput.ID(), inst.sourceSearchField())
	}
	leftList := leftWrapper.Children()[1]
	if header, _ := leftList.Props()["header"].(string); header != "Source (1/3)" {
		t.Fatalf("left header = %q, want Source (1/3)", header)
	}
	if rows, _ := leftList.Props()["rows"].([]string); !reflect.DeepEqual(rows, []string{"Alpha"}) {
		t.Fatalf("left rows = %#v, want Alpha only", rows)
	}

	if !inst.HandleIntent(list.SelectionChangeWithID(inst.sourceListID(), list.SelectionMultiple, []int{0}, []string{"Alpha"})) {
		t.Fatal("selection intent should be handled")
	}
	if !reflect.DeepEqual(inst.selectedSourceKeys, []string{"a"}) {
		t.Fatalf("selectedSourceKeys = %#v, want []string{\"a\"}", inst.selectedSourceKeys)
	}
	if !inst.HandleIntent(MoveToTargetWithID("members")) {
		t.Fatal("move-to-target intent should be handled")
	}
	if !reflect.DeepEqual(inst.targetKeys, []string{"a", "d"}) {
		t.Fatalf("targetKeys = %#v, want []string{\"a\", \"d\"}", inst.targetKeys)
	}
}

func TestRuntimeChildrenPagedLists(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "members",
		propPageSize:    2,
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma"},
			{Key: "d", Title: "Delta"},
			{Key: "e", Title: "Epsilon"},
		},
	})

	children := inst.RuntimeChildren()
	leftWrapper := children[0].Children()[0]
	if len(leftWrapper.Children()) != 2 {
		t.Fatalf("paged wrapper children len = %d, want list plus pager", len(leftWrapper.Children()))
	}
	leftList := leftWrapper.Children()[0]
	if header, _ := leftList.Props()["header"].(string); header != "Source (1-2/5)" {
		t.Fatalf("left header = %q, want Source (1-2/5)", header)
	}
	if rows, _ := leftList.Props()["rows"].([]string); !reflect.DeepEqual(rows, []string{"Alpha", "Beta"}) {
		t.Fatalf("left rows = %#v, want first page", rows)
	}
	pager := leftWrapper.Children()[1]
	if pager.Tag() != "hstack" || len(pager.Children()) != 3 {
		t.Fatalf("pager = %q children:%d, want hstack with three children", pager.Tag(), len(pager.Children()))
	}
}

func TestPageIntentMovesVisiblePage(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "members",
		propPageSize:    2,
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma"},
			{Key: "d", Title: "Delta"},
		},
	})

	if !inst.HandleIntent(PageWithID("members", SearchSideSource, 1)) {
		t.Fatal("page intent should be handled")
	}
	if inst.sourcePage != 1 {
		t.Fatalf("sourcePage = %d, want 1", inst.sourcePage)
	}
	children := inst.RuntimeChildren()
	leftList := children[0].Children()[0].Children()[0]
	if header, _ := leftList.Props()["header"].(string); header != "Source (3-4/4)" {
		t.Fatalf("left header = %q, want Source (3-4/4)", header)
	}
	if rows, _ := leftList.Props()["rows"].([]string); !reflect.DeepEqual(rows, []string{"Gamma", "Delta"}) {
		t.Fatalf("left rows = %#v, want second page", rows)
	}
}

func TestSearchFieldChangeUpdatesVisibleRows(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "members",
		propSearchable:  true,
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma"},
		},
	})

	if !inst.HandleIntent(runtimeintent.FieldChangeIntent{Field: inst.sourceSearchField(), Value: "bet"}) {
		t.Fatal("source search field change should be handled")
	}
	if inst.sourceSearch != "bet" {
		t.Fatalf("sourceSearch = %q, want bet", inst.sourceSearch)
	}
	children := inst.RuntimeChildren()
	leftList := children[0].Children()[0].Children()[1]
	if header, _ := leftList.Props()["header"].(string); header != "Source (1/3)" {
		t.Fatalf("left header = %q, want Source (1/3)", header)
	}
	if rows, _ := leftList.Props()["rows"].([]string); !reflect.DeepEqual(rows, []string{"Beta"}) {
		t.Fatalf("left rows = %#v, want Beta only", rows)
	}
}

func TestHandleSelectionAndMoveRight(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "members",
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
			{Key: "c", Title: "Gamma"},
		},
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleIntent(list.SelectionChangeWithID(inst.sourceListID(), list.SelectionMultiple, []int{0, 1}, []string{"Alpha", "Beta"})) {
		t.Fatal("source selection intent should be handled")
	}
	if !reflect.DeepEqual(inst.selectedSourceKeys, []string{"a", "b"}) {
		t.Fatalf("selectedSourceKeys = %#v, want []string{\"a\", \"b\"}", inst.selectedSourceKeys)
	}

	if !inst.HandleIntent(MoveToTargetWithID("members")) {
		t.Fatal("move-to-target intent should be handled")
	}
	if !reflect.DeepEqual(inst.targetKeys, []string{"a", "b"}) {
		t.Fatalf("targetKeys = %#v, want []string{\"a\", \"b\"}", inst.targetKeys)
	}
	if len(inst.selectedSourceKeys) != 0 {
		t.Fatalf("selectedSourceKeys should be cleared, got %#v", inst.selectedSourceKeys)
	}

	if len(emitted) == 0 {
		t.Fatal("expected change intent emission")
	}
	change, ok := emitted[0].(ChangeIntent)
	if !ok {
		t.Fatalf("first emitted intent type = %T, want transfer.ChangeIntent", emitted[0])
	}
	if change.Direction != MoveDirectionToTarget {
		t.Fatalf("change direction = %q, want %q", change.Direction, MoveDirectionToTarget)
	}
	if !reflect.DeepEqual(change.MovedKeys, []string{"a", "b"}) {
		t.Fatalf("movedKeys = %#v, want []string{\"a\", \"b\"}", change.MovedKeys)
	}
}

func TestMoveAllToTargetMovesVisibleEnabledItems(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:  "members",
		propSearchable:   true,
		propSourceSearch: "a",
		propItems: []Item{
			{Key: "alpha", Title: "Alpha"},
			{Key: "beta", Title: "Beta"},
			{Key: "gamma", Title: "Gamma", Disabled: true},
			{Key: "delta", Title: "Delta"},
		},
		propTargetKeys: []string{"delta"},
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleIntent(MoveAllToTargetWithID("members")) {
		t.Fatal("move-all-to-target intent should be handled")
	}
	if !reflect.DeepEqual(inst.targetKeys, []string{"alpha", "beta", "delta"}) {
		t.Fatalf("targetKeys = %#v, want []string{\"alpha\", \"beta\", \"delta\"}", inst.targetKeys)
	}
	if !reflect.DeepEqual(inst.selectedSourceKeys, []string(nil)) {
		t.Fatalf("selectedSourceKeys = %#v, want nil", inst.selectedSourceKeys)
	}
	if len(emitted) == 0 {
		t.Fatal("expected change intent emission")
	}
	change, ok := emitted[0].(ChangeIntent)
	if !ok {
		t.Fatalf("first emitted intent type = %T, want transfer.ChangeIntent", emitted[0])
	}
	if change.Direction != MoveDirectionToTarget {
		t.Fatalf("change direction = %q, want %q", change.Direction, MoveDirectionToTarget)
	}
	if !reflect.DeepEqual(change.MovedKeys, []string{"alpha", "beta"}) {
		t.Fatalf("movedKeys = %#v, want []string{\"alpha\", \"beta\"}", change.MovedKeys)
	}
}

func TestMoveAllToTargetHonorsSourceSearch(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:  "members",
		propSourceSearch: "alp",
		propItems: []Item{
			{Key: "alpha", Title: "Alpha"},
			{Key: "beta", Title: "Beta"},
			{Key: "gamma", Title: "Gamma"},
			{Key: "delta", Title: "Delta"},
		},
		propTargetKeys: []string{"delta"},
	})

	if !inst.HandleIntent(MoveAllToTargetWithID("members")) {
		t.Fatal("move-all-to-target intent should be handled")
	}
	if !reflect.DeepEqual(inst.targetKeys, []string{"alpha", "delta"}) {
		t.Fatalf("targetKeys = %#v, want []string{\"alpha\", \"delta\"}", inst.targetKeys)
	}
}

func TestMoveAllToTargetHonorsCurrentPage(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "members",
		propPageSize:    2,
		propItems: []Item{
			{Key: "alpha", Title: "Alpha"},
			{Key: "beta", Title: "Beta"},
			{Key: "gamma", Title: "Gamma"},
			{Key: "delta", Title: "Delta"},
		},
	})

	if !inst.HandleIntent(PageWithID("members", SearchSideSource, 1)) {
		t.Fatal("page intent should be handled")
	}
	if !inst.HandleIntent(MoveAllToTargetWithID("members")) {
		t.Fatal("move-all-to-target intent should be handled")
	}
	if !reflect.DeepEqual(inst.targetKeys, []string{"gamma", "delta"}) {
		t.Fatalf("targetKeys = %#v, want current page keys", inst.targetKeys)
	}
}

func TestMoveAllToSourceHonorsTargetSearch(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:  "members",
		propTargetSearch: "bet",
		propItems: []Item{
			{Key: "alpha", Title: "Alpha"},
			{Key: "beta", Title: "Beta"},
			{Key: "gamma", Title: "Gamma"},
			{Key: "delta", Title: "Delta"},
		},
		propTargetKeys: []string{"alpha", "beta", "delta"},
	})

	if !inst.HandleIntent(MoveAllToSourceWithID("members")) {
		t.Fatal("move-all-to-source intent should be handled")
	}
	if !reflect.DeepEqual(inst.targetKeys, []string{"alpha", "delta"}) {
		t.Fatalf("targetKeys = %#v, want []string{\"alpha\", \"delta\"}", inst.targetKeys)
	}
}

func TestDisabledItemsAreRejectedFromSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			{Key: "a", Title: "Alpha", Disabled: true},
			{Key: "b", Title: "Beta"},
		},
	})

	if !inst.HandleIntent(list.SelectionChangeWithID(inst.sourceListID(), list.SelectionMultiple, []int{0}, []string{"Alpha"})) {
		t.Fatal("selection intent should be handled")
	}
	if len(inst.selectedSourceKeys) != 0 {
		t.Fatalf("disabled item should not stay selected, got %#v", inst.selectedSourceKeys)
	}
	if inst.sourceListVersion != 1 {
		t.Fatalf("sourceListVersion = %d, want 1", inst.sourceListVersion)
	}
}

func TestFieldBindingEmitsJoinedTargetKeys(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "members",
		propItems: []Item{
			{Key: "a", Title: "Alpha"},
			{Key: "b", Title: "Beta"},
		},
		propChangeIntent: runtimeintent.BindField("members"),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	inst.HandleIntent(list.SelectionChangeWithID(inst.sourceListID(), list.SelectionMultiple, []int{0, 1}, []string{"Alpha", "Beta"}))
	inst.HandleIntent(MoveToTargetWithID("members"))

	if len(emitted) < 2 {
		t.Fatalf("emitted intents len = %d, want at least 2", len(emitted))
	}
	fieldChange, ok := emitted[1].(runtimeintent.FieldChangeIntent)
	if !ok {
		t.Fatalf("second emitted intent type = %T, want intent.FieldChangeIntent", emitted[1])
	}
	if fieldChange.Field != "members" || fieldChange.Value != "a,b" {
		t.Fatalf("field change = %#v, want field=members value=a,b", fieldChange)
	}
}
