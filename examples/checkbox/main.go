package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// CheckboxDemo demonstrates checkbox component
func CheckboxDemo() ui.VNode {
	acceptTerms, setAcceptTerms := ui.UseStateBool(false)
	acceptUpdates, setAcceptUpdates := ui.UseStateBool(false)
	acceptPrivacy, setAcceptPrivacy := ui.UseStateBool(false)

	// Count checked checkboxes
	checkedCount := 0
	if acceptTerms {
		checkedCount++
	}
	if acceptUpdates {
		checkedCount++
	}
	if acceptPrivacy {
		checkedCount++
	}

	return ui.VStack(
		ui.NewTextBuilder("Checkbox Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Select your preferences:").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.CheckboxBuilder().
			Label("I accept the terms and conditions").
			Checked(acceptTerms).
			OnChange(setAcceptTerms).
			Build(),
		ui.CheckboxBuilder().
			Label("Subscribe to updates").
			Checked(acceptUpdates).
			OnChange(setAcceptUpdates).
			Build(),
		ui.CheckboxBuilder().
			Label("I have read the privacy policy").
			Checked(acceptPrivacy).
			OnChange(setAcceptPrivacy).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Checked: %d/3", checkedCount)).
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Space/Enter: toggle | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(CheckboxDemo,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Checkbox Demo"),
	)
	if err != nil {
		panic(err)
	}
}
