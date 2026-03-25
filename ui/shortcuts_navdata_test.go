package ui

import "testing"

func TestNavDataShortcuts(t *testing.T) {
	listNode := List().
		Rows([]string{"A", "B"}).
		Build()
	if listNode == nil || listNode.Tag() != "list" {
		t.Fatalf("List().Build().Tag() = %q, want list", listNode.Tag())
	}

	treeNode := TreeView().
		Nodes([]TreeNode{{Content: "root", NodeID: 1}}).
		Build()
	if treeNode == nil || treeNode.Tag() != "treeview" {
		t.Fatalf("TreeView().Build().Tag() = %q, want treeview", treeNode.Tag())
	}

	drawerNode := Drawer(Text("Body"))
	if drawerNode == nil || drawerNode.Tag() != "drawer" {
		t.Fatalf("Drawer().Tag() = %q, want drawer", drawerNode.Tag())
	}

	titled := DrawerTitled("Settings", Text("Body"))
	if titled == nil || titled.Tag() != "drawer" {
		t.Fatalf("DrawerTitled().Tag() = %q, want drawer", titled.Tag())
	}
}

func TestNewDrawerBuilder(t *testing.T) {
	vnode := NewDrawerBuilder().
		Title("Settings").
		Content(Text("Body")).
		Opened().
		Build()
	if vnode == nil {
		t.Fatal("NewDrawerBuilder().Build() returned nil")
	}
	if vnode.Tag() != "drawer" {
		t.Fatalf("Tag() = %q, want drawer", vnode.Tag())
	}
}
