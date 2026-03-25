package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	breadcrumbcomp "github.com/wwsheng009/mint/ui/components/breadcrumb"
)

func newBreadcrumbStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Breadcrumb E2E Fixture").Build(),
				breadcrumbcomp.NewBuilder().
					SetID("breadcrumb-wide").
					Labels("RootHub", "DevDocs").
					Item(breadcrumbcomp.Current("WideCurrent")).
					Separator(" > ").
					CurrentStyle(style.NewStyle().Bold(true)).
					Build(),
				breadcrumbcomp.NewBuilder().
					SetID("breadcrumb-narrow").
					Items([]breadcrumbcomp.Item{
						breadcrumbcomp.Crumb("NorthRoot"),
						breadcrumbcomp.Crumb("WorkspaceNode"),
						breadcrumbcomp.Crumb("RenderToken"),
						breadcrumbcomp.Current("CollapsedTail"),
					}).
					MaxWidth(20).
					Build(),
			})
	}
}

func TestE2EBreadcrumbWideRenderAndCurrentStyle(t *testing.T) {
	app, err := Run(newBreadcrumbStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("RootHub > DevDocs > WideCurrent")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("WideCurrent"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("breadcrumb-wide")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EBreadcrumbNarrowRenderCollapsesLeftSide(t *testing.T) {
	app, err := Run(newBreadcrumbStaticApp(), ui.WithSize(96, 20))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("… / ")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("CollapsedTail")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByID("breadcrumb-narrow")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("NorthRoot")); err == nil {
		t.Fatal("leftmost breadcrumb item should be collapsed in narrow breadcrumb")
	}
}
