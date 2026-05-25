package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
)

func TestE2EDataTableOperationalStates(t *testing.T) {
	appFn := func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("DataTable E2E Fixture").Build(),
				ui.DataTable(
					[]ui.TableColumn{
						{Title: "Provider", Width: 16},
						{Title: "Status", Width: 12},
					},
					[][]string{
						{"openai", "healthy"},
						{"azure", "degraded"},
					},
					ui.DataTableServerPagination(2, 25, 76),
					ui.DataTableOperationalStyle(),
				),
				ui.DataTable(
					[]ui.TableColumn{{Title: "Job", Width: 20}},
					[][]string{{"sync"}},
					ui.DataTableLoading(true),
					ui.DataTableLoadingText("Loading jobs..."),
					ui.DataTableOperationalStyle(),
				),
				ui.DataTable(
					[]ui.TableColumn{{Title: "Alert", Width: 34}},
					[][]string{{"latency"}},
					ui.DataTableErrorText("alerts API unavailable"),
					ui.DataTableOperationalStyle(),
				),
			})
	}

	app, err := Run(appFn, ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"DataTable E2E Fixture",
		"Page 2/4 · Total 76 · Size 25",
		"Loading jobs...",
		"Loading",
		"alerts API unavailable",
		"Error · alerts API unavailable",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}
}
