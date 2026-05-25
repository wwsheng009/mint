package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	filterbarcomp "github.com/wwsheng009/mint/ui/components/filterbar"
)

type filterBarTestIntent struct {
	name string
}

func (i filterBarTestIntent) IntentType() string { return i.name }

func newFilterBarStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		bar := filterbarcomp.NewBuilder().
			Key("provider-filterbar").
			Title("Filter Toolbar").
			Summary("scope: production | page: logs").
			Width(60).
			LabelWidth(4).
			Field(filterbarcomp.Search("query", "Query", "openai").
				WithPlaceholder("provider").
				WithWidth(12).
				ForField("query").
				OnSubmit(filterBarTestIntent{"submit-search"})).
			Field(filterbarcomp.Select("status", "Status", []filterbarcomp.Option{
				{Value: "all", Label: "All"},
				{Value: "healthy", Label: "Healthy"},
				{Value: "degraded", Label: "Degraded"},
			}).WithSelectedIndex(2).WithWidth(12).ForField("status")).
			Action(filterbarcomp.Button("refresh", "Refresh", filterBarTestIntent{"refresh"}).Primary()).
			Action(filterbarcomp.Button("reset", "Reset", intent.Focus("provider-filterbar"))).
			Action(filterbarcomp.Button("export", "Export", filterBarTestIntent{"export"}).WithDisabledReason("Select at least one provider.")).
			Build()

		table := ui.DataTable(
			[]ui.TableColumn{
				{Title: "Provider", Width: 18},
				{Title: "Status", Width: 12},
			},
			[][]string{
				{"openai", "degraded"},
				{"azure", "healthy"},
			},
			ui.DataTablePageSize(5),
			ui.DataTableEmptyText("No providers"),
			ui.DataTableOperationalStyle(),
		)

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("FilterBar E2E Fixture").Build(),
				bar,
				table,
			})
	}
}

func TestE2EFilterBarFieldsActionsAndTableOrderRender(t *testing.T) {
	app, err := Run(newFilterBarStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Filter Toolbar", "scope: production", "Query", "openai", "Status", "Degraded", "Refresh", "Reset", "Export", "Disabled: Export: Select at least one provider.", "Provider"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}

	filterPoint, err := app.ResolvePoint(ByText("Filter Toolbar"))
	if err != nil {
		t.Fatal(err)
	}
	tablePoint, err := app.ResolvePoint(ByText("Provider"))
	if err != nil {
		t.Fatal(err)
	}
	if filterPoint.Y >= tablePoint.Y {
		t.Fatalf("expected filter toolbar above table, got filterY=%d tableY=%d", filterPoint.Y, tablePoint.Y)
	}
}
