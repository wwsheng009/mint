package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
	statisticcomp "github.com/wwsheng009/mint/ui/components/statistic"
)

func newStatisticStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Statistic E2E Fixture").Build(),
				statisticcomp.NewBuilder().
					SetID("stat-revenue").
					Title("Revenue").
					Value(12345.67).
					Prefix("$").
					Suffix(" USD").
					Precision(2).
					Up().
					Extra(ui.NewTextBuilder("Compared to yesterday").Build()).
					Build(),
				statisticcomp.NewBuilder().
					SetID("stat-requests").
					Title("Requests").
					Value(999).
					Loading(true).
					Bordered(true).
					Width(28).
					Build(),
			})
	}
}

func TestE2EStatisticFormattedValueTrendAndExtraRender(t *testing.T) {
	app, err := Run(newStatisticStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Revenue")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("↑")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("$")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("12,345.67")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("USD")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Compared to yesterday")); err != nil {
		t.Fatal(err)
	}
}

func TestE2EStatisticLoadingAndStylesRender(t *testing.T) {
	app, err := Run(newStatisticStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Requests")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("...")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┌")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("┘")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Revenue"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("12,345.67"), StyleExpect{
		HasBold: true,
		Bold:    true,
		HasFG:   true,
		FG:      fwtheme.Text(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("↑"), StyleExpect{
		HasBold: true,
		Bold:    true,
		HasFG:   true,
		FG:      fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("..."), StyleExpect{
		HasBold: true,
		Bold:    true,
		HasFG:   true,
		FG:      fwtheme.Text(),
	}); err != nil {
		t.Fatal(err)
	}
}
