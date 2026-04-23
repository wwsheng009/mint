package selectcomp

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestPopupRows_IncludeGroupHeaders(t *testing.T) {
	rows := buildPopupRows(
		[]Option{
			{Group: "Asia", Value: "cn", Label: "China"},
			{Group: "Asia", Value: "jp", Label: "Japan"},
			{Group: "Europe", Value: "de", Label: "Germany"},
		},
		SelectionSingle,
		true,
		"search",
		"",
	)

	if !rows.showFilter {
		t.Fatal("expected filter row to be shown")
	}
	if len(rows.scrollable) < 5 {
		t.Fatalf("expected grouped rows, got %d", len(rows.scrollable))
	}
	if rows.scrollable[0].kind != popupRowGroup || rows.scrollable[0].text != "[Asia]" {
		t.Fatalf("first row = %#v, want Asia group header", rows.scrollable[0])
	}
	if rows.scrollable[1].kind != popupRowOption || rows.scrollable[1].optionIndex != 0 {
		t.Fatalf("second row = %#v, want first option row", rows.scrollable[1])
	}
}

func TestPopupRows_FilterMatchesValueAndLabel(t *testing.T) {
	rowsByLabel := buildPopupRows(
		[]Option{
			{Value: "cn", Label: "China"},
			{Value: "us", Label: "United States"},
		},
		SelectionSingle,
		true,
		"search",
		"United",
	)
	if len(rowsByLabel.scrollable) != 1 || rowsByLabel.scrollable[0].optionIndex != 1 {
		t.Fatalf("filter by label result = %#v, want only index 1", rowsByLabel.scrollable)
	}

	rowsByValue := buildPopupRows(
		[]Option{
			{Value: "cn", Label: "China"},
			{Value: "us", Label: "United States"},
		},
		SelectionSingle,
		true,
		"search",
		"cn",
	)
	if len(rowsByValue.scrollable) != 1 || rowsByValue.scrollable[0].optionIndex != 0 {
		t.Fatalf("filter by value result = %#v, want only index 0", rowsByValue.scrollable)
	}
}

func TestPopupRows_TagsModeShowsCreateRow(t *testing.T) {
	rows := buildPopupRows(
		[]Option{{Value: "go", Label: "go"}},
		SelectionTags,
		true,
		"search",
		"mint",
	)

	if len(rows.scrollable) == 0 {
		t.Fatal("expected at least one popup row")
	}
	if rows.scrollable[0].kind != popupRowCreateTag {
		t.Fatalf("first row kind = %v, want create tag", rows.scrollable[0].kind)
	}
}

func TestInstance_FilterOptionCommitsFilteredSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propOptions: []Option{
			{Value: "cn", Label: "China"},
			{Value: "us", Label: "United States"},
			{Value: "jp", Label: "Japan"},
		},
		propSelectionMode: SelectionSingle,
		propFilterOption:  true,
	})

	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should open dropdown")
	}
	if !inst.HandleAction(action.NewAction(action.ActionInputText).WithPayload("Japan")) {
		t.Fatal("input text should update filter")
	}
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should commit filtered option")
	}

	if inst.SelectedValue() != "jp" {
		t.Fatalf("selected value = %q, want %q", inst.SelectedValue(), "jp")
	}
}

func TestInstance_TagsModeCreateAndSelectTag(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propOptions: []Option{
			{Value: "go", Label: "Go"},
		},
		propSelectionMode: SelectionTags,
		propFilterOption:  true,
	})

	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should open tags dropdown")
	}
	if !inst.HandleAction(action.NewAction(action.ActionInputText).WithPayload("Mint")) {
		t.Fatal("input text should update tags query")
	}
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should create and select tag")
	}

	if len(inst.options) != 2 {
		t.Fatalf("options len = %d, want 2", len(inst.options))
	}
	values := inst.SelectedValues()
	if len(values) != 1 || values[0] != "Mint" {
		t.Fatalf("selected values = %v, want [Mint]", values)
	}
	if got := inst.fieldValue(); got != "Mint" {
		t.Fatalf("fieldValue = %q, want %q", got, "Mint")
	}
}
