package runtime

import (
	"testing"
)

// =============================================================================
// Mock FocusableComponent
// =============================================================================

type mockFocusable struct {
	id     string
	focus  bool
	canFocus bool
}

func (m *mockFocusable) SetFocus(f bool) {
	m.focus = f
}

func (m *mockFocusable) IsFocusable() bool {
	return m.canFocus
}

func (m *mockFocusable) View() string {
	return m.id
}

// =============================================================================
// FocusManager Tests
// =============================================================================

func TestNewFocusManager(t *testing.T) {
	fm := NewFocusManager()
	if fm == nil {
		t.Fatal("NewFocusManager() should not return nil")
	}
	if !fm.IsEmpty() {
		t.Error("new FocusManager should be empty")
	}
	if fm.Count() != 0 {
		t.Error("new FocusManager should have count 0")
	}
	if fm.GetCurrent() != nil {
		t.Error("new FocusManager should have no current focus")
	}
	if fm.GetCurrentIndex() != -1 {
		t.Error("new FocusManager should have index -1")
	}
}

func TestFocusManager_SetFocusable(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1"},
		{ID: "item2"},
	}
	fm.SetFocusable(items)

	if fm.Count() != 2 {
		t.Errorf("Count = %d, want 2", fm.Count())
	}
}

func TestFocusManager_AddFocusable(t *testing.T) {
	fm := NewFocusManager()
	fm.AddFocusable(&FocusableItem{ID: "item1"})
	fm.AddFocusable(&FocusableItem{ID: "item2"})

	if fm.Count() != 2 {
		t.Errorf("Count = %d, want 2", fm.Count())
	}
}

func TestFocusManager_RemoveFocusable(t *testing.T) {
	fm := NewFocusManager()
	fm.AddFocusable(&FocusableItem{ID: "item1"})
	fm.AddFocusable(&FocusableItem{ID: "item2"})
	fm.AddFocusable(&FocusableItem{ID: "item3"})

	fm.RemoveFocusable("item2")

	if fm.Count() != 2 {
		t.Errorf("Count = %d, want 2", fm.Count())
	}
	if fm.FindByID("item2") != nil {
		t.Error("item2 should be removed")
	}
}

func TestFocusManager_RemoveFocusable_AdjustsIndex(t *testing.T) {
	fm := NewFocusManager()
	fm.SetFocusable([]*FocusableItem{
		{ID: "item1"},
		{ID: "item2"},
		{ID: "item3"},
	})
	fm.FocusAt(1) // Focus item2

	fm.RemoveFocusable("item1")
	if fm.GetCurrentIndex() != 0 {
		t.Errorf("after removing item before, index should be 0, got %d", fm.GetCurrentIndex())
	}

	fm.RemoveFocusable("item2")
	if fm.GetCurrentIndex() != -1 {
		t.Errorf("after removing focused item, index should be -1, got %d", fm.GetCurrentIndex())
	}
}

func TestFocusManager_RemoveFocusable_NonExistent(t *testing.T) {
	fm := NewFocusManager()
	fm.AddFocusable(&FocusableItem{ID: "item1"})

	// Should not panic
	fm.RemoveFocusable("nonexistent")
	if fm.Count() != 1 {
		t.Error("count should not change when removing non-existent item")
	}
}

func TestFocusManager_FocusNext(t *testing.T) {
	fm := NewFocusManager()
	item1 := &FocusableItem{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}}
	item2 := &FocusableItem{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}}
	fm.SetFocusable([]*FocusableItem{item1, item2})

	focused := fm.FocusNext()
	if focused == nil || focused.ID != "item1" {
		t.Error("FocusNext() should focus first item")
	}
	if !item1.Instance.(*mockFocusable).focus {
		t.Error("first item should have focus")
	}

	focused = fm.FocusNext()
	if focused == nil || focused.ID != "item2" {
		t.Error("FocusNext() should focus second item")
	}
	if item1.Instance.(*mockFocusable).focus {
		t.Error("first item should lose focus")
	}
}

func TestFocusManager_FocusNext_Wrap(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)
	fm.FocusAt(1) // Start at last item

	focused := fm.FocusNext()
	if focused == nil || focused.ID != "item1" {
		t.Error("FocusNext() should wrap to first item")
	}
}

func TestFocusManager_FocusNext_NoWrap(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)
	fm.SetWrap(false)
	fm.FocusAt(1) // Start at last item

	focused := fm.FocusNext()
	if focused != nil {
		t.Error("FocusNext() should return nil when no wrap and at end")
	}
}

func TestFocusManager_FocusNext_Empty(t *testing.T) {
	fm := NewFocusManager()
	if fm.FocusNext() != nil {
		t.Error("FocusNext() should return nil for empty manager")
	}
}

func TestFocusManager_FocusPrev(t *testing.T) {
	fm := NewFocusManager()
	item1 := &FocusableItem{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}}
	item2 := &FocusableItem{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}}
	fm.SetFocusable([]*FocusableItem{item1, item2})
	fm.FocusAt(1)

	focused := fm.FocusPrev()
	if focused == nil || focused.ID != "item1" {
		t.Error("FocusPrev() should focus first item")
	}
}

func TestFocusManager_FocusPrev_Wrap(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)
	fm.FocusAt(0) // Start at first item

	focused := fm.FocusPrev()
	if focused == nil || focused.ID != "item2" {
		t.Error("FocusPrev() should wrap to last item")
	}
}

func TestFocusManager_FocusPrev_NoWrap(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)
	fm.SetWrap(false)
	fm.FocusAt(0) // Start at first item

	focused := fm.FocusPrev()
	if focused != nil {
		t.Error("FocusPrev() should return nil when no wrap and at start")
	}
}

func TestFocusManager_Focus(t *testing.T) {
	fm := NewFocusManager()
	item1 := &FocusableItem{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}}
	item2 := &FocusableItem{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}}
	fm.SetFocusable([]*FocusableItem{item1, item2})

	if !fm.Focus("item2") {
		t.Error("Focus() should return true for existing item")
	}
	if fm.GetCurrent().ID != "item2" {
		t.Error("Focus() should set focus to item2")
	}
	if fm.Focus("nonexistent") {
		t.Error("Focus() should return false for non-existent item")
	}
}

func TestFocusManager_FocusAt(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)

	if !fm.FocusAt(1) {
		t.Error("FocusAt(1) should return true")
	}
	if fm.GetCurrentIndex() != 1 {
		t.Error("FocusAt(1) should set index to 1")
	}

	if fm.FocusAt(5) {
		t.Error("FocusAt(5) should return false for out of range")
	}
	if fm.FocusAt(-1) {
		t.Error("FocusAt(-1) should return false for negative index")
	}
}

func TestFocusManager_ClearFocus(t *testing.T) {
	fm := NewFocusManager()
	item1 := &FocusableItem{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}}
	fm.SetFocusable([]*FocusableItem{item1})
	fm.Focus("item1")

	fm.ClearFocus()
	if fm.GetCurrent() != nil {
		t.Error("ClearFocus() should clear current focus")
	}
	if item1.Instance.(*mockFocusable).focus {
		t.Error("item should lose focus after ClearFocus")
	}
}

func TestFocusManager_HasFocus(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)
	fm.Focus("item1")

	if !fm.HasFocus("item1") {
		t.Error("HasFocus(item1) should return true")
	}
	if fm.HasFocus("item2") {
		t.Error("HasFocus(item2) should return false")
	}
	if fm.HasFocus("nonexistent") {
		t.Error("HasFocus(nonexistent) should return false")
	}
}

func TestFocusManager_Count(t *testing.T) {
	fm := NewFocusManager()
	if fm.Count() != 0 {
		t.Errorf("Count() = %d, want 0", fm.Count())
	}

	fm.AddFocusable(&FocusableItem{ID: "item1"})
	fm.AddFocusable(&FocusableItem{ID: "item2"})
	if fm.Count() != 2 {
		t.Errorf("Count() = %d, want 2", fm.Count())
	}
}

func TestFocusManager_IsEmpty(t *testing.T) {
	fm := NewFocusManager()
	if !fm.IsEmpty() {
		t.Error("IsEmpty() should return true for new manager")
	}

	fm.AddFocusable(&FocusableItem{ID: "item1"})
	if fm.IsEmpty() {
		t.Error("IsEmpty() should return false after adding item")
	}
}

func TestFocusManager_SetWrap(t *testing.T) {
	fm := NewFocusManager()
	fm.SetWrap(false)
	// Just verify it doesn't panic - actual behavior tested in FocusNext/Prev
}

func TestFocusManager_SetCyclic(t *testing.T) {
	fm := NewFocusManager()
	fm.SetCyclic(true)
	// Just verify it doesn't panic
}

func TestFocusManager_FocusFirst(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)

	if !fm.FocusFirst() {
		t.Error("FocusFirst() should return true")
	}
	if fm.GetCurrent().ID != "item1" {
		t.Error("FocusFirst() should focus first item")
	}
}

func TestFocusManager_FocusFirst_Empty(t *testing.T) {
	fm := NewFocusManager()
	if fm.FocusFirst() {
		t.Error("FocusFirst() should return false for empty manager")
	}
}

func TestFocusManager_FocusLast(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)

	if !fm.FocusLast() {
		t.Error("FocusLast() should return true")
	}
	if fm.GetCurrent().ID != "item2" {
		t.Error("FocusLast() should focus last item")
	}
}

func TestFocusManager_FocusLast_Empty(t *testing.T) {
	fm := NewFocusManager()
	if fm.FocusLast() {
		t.Error("FocusLast() should return false for empty manager")
	}
}

func TestFocusManager_FocusNone(t *testing.T) {
	fm := NewFocusManager()
	item1 := &FocusableItem{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}}
	fm.SetFocusable([]*FocusableItem{item1})
	fm.Focus("item1")

	fm.FocusNone()
	if fm.GetCurrent() != nil {
		t.Error("FocusNone() should clear focus")
	}
}

func TestFocusManager_FindByID(t *testing.T) {
	fm := NewFocusManager()
	item1 := &FocusableItem{ID: "item1"}
	item2 := &FocusableItem{ID: "item2"}
	fm.SetFocusable([]*FocusableItem{item1, item2})

	found := fm.FindByID("item2")
	if found != item2 {
		t.Error("FindByID() should find item2")
	}

	if fm.FindByID("nonexistent") != nil {
		t.Error("FindByID() should return nil for non-existent item")
	}
}

func TestFocusManager_NextDirection(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)

	focused := fm.NextDirection(1)
	if focused == nil || focused.ID != "item1" {
		t.Error("NextDirection(1) should focus first item")
	}

	focused = fm.NextDirection(-1)
	if focused == nil || focused.ID != "item2" {
		t.Error("NextDirection(-1) should focus last item (wrap)")
	}

	focused = fm.NextDirection(0)
	if focused == nil || focused.ID != "item2" {
		t.Error("NextDirection(0) should return current item")
	}
}

func TestFocusManager_SetFocusChangeCallback(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1", Instance: &mockFocusable{id: "1", canFocus: true}},
		{ID: "item2", Instance: &mockFocusable{id: "2", canFocus: true}},
	}
	fm.SetFocusable(items)

	var calledWith []*FocusableItem
	fm.SetFocusChangeCallback(func(focused, previous *FocusableItem) {
		calledWith = []*FocusableItem{focused, previous}
	})

	fm.Focus("item2")
	if len(calledWith) != 2 {
		t.Fatal("callback should be called with 2 args")
	}
	if calledWith[0] == nil || calledWith[0].ID != "item2" {
		t.Error("callback focused should be item2")
	}
	if calledWith[1] != nil {
		t.Error("callback previous should be nil for first focus")
	}
}

func TestFocusManager_GetFocusable(t *testing.T) {
	fm := NewFocusManager()
	items := []*FocusableItem{
		{ID: "item1"},
		{ID: "item2"},
	}
	fm.SetFocusable(items)

	got := fm.GetFocusable()
	if len(got) != 2 {
		t.Errorf("GetFocusable() length = %d, want 2", len(got))
	}
}

func TestFocusManager_UpdateFromLayout(t *testing.T) {
	fm := NewFocusManager()
	fm.Focus("item1") // Set initial focus

	// Create a layout result with focusable items
	result := LayoutResult{
		Boxes: []LayoutBox{
			{
				NodeID: "box1",
				Node: &LayoutNode{
					ID: "node1",
					Component: &ComponentRef{
						Instance: &mockFocusable{id: "comp1", canFocus: true},
					},
				},
			},
		},
	}

	fm.UpdateFromLayout(result)
	// Just verify it doesn't panic and updates focusable list
	if fm.Count() != 1 {
		t.Errorf("UpdateFromLayout should result in 1 focusable, got %d", fm.Count())
	}
}

func TestCollectFocusableFromNode(t *testing.T) {
	root := &LayoutNode{
		ID: "root",
		Component: &ComponentRef{
			Instance: &mockFocusable{id: "root", canFocus: true},
		},
		Children: []*LayoutNode{
			{
				ID: "child1",
				Component: &ComponentRef{
					Instance: &mockFocusable{id: "c1", canFocus: true},
				},
			},
			{
				ID: "child2",
				Component: &ComponentRef{
					Instance: &mockFocusable{id: "c2", canFocus: false}, // Not focusable
				},
			},
		},
	}

	items := CollectFocusableFromNode(root)
	if len(items) != 2 { // root and child1
		t.Errorf("CollectFocusableFromNode() = %d items, want 2", len(items))
	}
}

func TestCollectFocusableFromNode_NilRoot(t *testing.T) {
	items := CollectFocusableFromNode(nil)
	if len(items) != 0 {
		t.Errorf("CollectFocusableFromNode(nil) = %d items, want 0", len(items))
	}
}
