package ui

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/breadcrumb"
)

func TestBreadcrumbShortcuts(t *testing.T) {
	items := []breadcrumb.Item{
		NewBreadcrumbItem("Home"),
		NewBreadcrumbItem("Workspace").WithIcon("W"),
		NewBreadcrumbItem("Breadcrumb").AsCurrent(),
	}

	vnode := Breadcrumb(items)
	bc, ok := vnode.(*breadcrumb.VNode)
	if !ok {
		t.Fatal("Breadcrumb should return *breadcrumb.VNode")
	}
	if got := len(bc.Items()); got != 3 {
		t.Fatalf("Breadcrumb items len = %d, want 3", got)
	}

	built := NewBreadcrumbBuilder().
		Items(items[:2]).
		Item(breadcrumb.Current("Docs")).
		Separator(" > ").
		Build()
	if got := built.Separator(); got != " > " {
		t.Fatalf("Separator = %q, want %q", got, " > ")
	}
	if got := len(built.Items()); got != 3 {
		t.Fatalf("Builder items len = %d, want 3", got)
	}
}
