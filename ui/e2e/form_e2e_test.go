package e2e

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
)

func newFormHelpRequiredFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Form E2E Help Required Fixture").Build(),
				ui.NewForm("loginForm").
					Layout(ui.FormVertical).
					AddChildren(
						ui.FormInputItem(
							"baseURL",
							"Gateway URL",
							"http://127.0.0.1:8080",
							ui.FormInputForForm("loginForm"),
							ui.FormInputWidth(48),
							ui.FormInputRequired(),
							ui.FormInputHelp("Use the Admin API base URL."),
						),
					),
			})
	}
}

func TestE2EFormItemHelpAndRequiredRender(t *testing.T) {
	app, err := Run(newFormHelpRequiredFixture(), ui.WithSize(96, 14))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Form E2E Help Required Fixture",
		"Gateway URL *",
		"Use the Admin API base URL.",
		"http://127.0.0.1:8080",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}
