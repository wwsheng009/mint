package selectcomp

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
)

type selectIntentSource interface {
	intent.TreeComponent
	EmitIntent(intent.Intent)
}

func getOverlayCallbacksProp(props rtui.Props) *overlayCallbacks {
	if props == nil {
		return nil
	}
	if callbacks, ok := props[overlayCallbacksProp].(*overlayCallbacks); ok {
		return callbacks
	}
	return nil
}

func emitSelectChangeFrom(
	source selectIntentSource,
	componentID string,
	mode SelectionMode,
	options []Option,
	selectedIndex int,
	selectedIndices []int,
) {
	if source == nil {
		return
	}
	source.EmitIntent(SelectChangeIntent{
		SelectedIndex:   selectedIndex,
		SelectedIndices: append([]int(nil), selectedIndices...),
		SelectedValue:   selectedValueFor(options, selectedIndex),
		SelectedValues:  selectedValuesFor(options, selectedIndices),
		SelectedLabel:   selectedLabelFor(options, selectedIndex),
		SelectedLabels:  selectedLabelsFor(options, selectedIndices),
		Mode:            mode,
		ComponentID:     componentID,
	})
}

func emitFieldValueChangedFrom(
	source selectIntentSource,
	changeIntent intent.Intent,
	changeIntentField intent.FieldIntent,
	formID string,
	mode SelectionMode,
	selectedIndex int,
	selectedIndices []int,
) {
	if source == nil {
		return
	}

	value := fieldValueFor(mode, selectedIndex, selectedIndices)
	if formID != "" {
		if changeIntentField != nil {
			intent.Emit(source, form.FieldChange(formID, changeIntentField.GetField(), value, true))
		}
		return
	}

	if changeIntentField != nil {
		source.EmitIntent(intent.FieldChangeIntent{
			Field: changeIntentField.GetField(),
			Value: value,
		})
		return
	}

	if changeIntent != nil {
		source.EmitIntent(changeIntent)
	}
}

func emitFieldBlurFrom(
	source selectIntentSource,
	changeIntentField intent.FieldIntent,
	formID string,
	mode SelectionMode,
	selectedIndex int,
	selectedIndices []int,
) {
	if source == nil || changeIntentField == nil || formID == "" {
		return
	}
	intent.Emit(source, form.FieldBlur(
		formID,
		changeIntentField.GetField(),
		fieldValueFor(mode, selectedIndex, selectedIndices),
	))
}

func selectedValueFor(options []Option, selectedIndex int) string {
	if selectedIndex >= 0 && selectedIndex < len(options) {
		return options[selectedIndex].Value
	}
	return ""
}

func selectedValuesFor(options []Option, selectedIndices []int) []string {
	values := make([]string, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		if idx >= 0 && idx < len(options) {
			values = append(values, options[idx].Value)
		}
	}
	return values
}

func selectedLabelFor(options []Option, selectedIndex int) string {
	if selectedIndex >= 0 && selectedIndex < len(options) {
		return options[selectedIndex].Label
	}
	return ""
}

func selectedLabelsFor(options []Option, selectedIndices []int) []string {
	labels := make([]string, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		if idx >= 0 && idx < len(options) {
			labels = append(labels, options[idx].Label)
		}
	}
	return labels
}

func fieldValueFor(mode SelectionMode, selectedIndex int, selectedIndices []int) string {
	if mode == SelectionMultiple {
		return joinIndices(selectedIndices)
	}
	return fmt.Sprintf("%d", selectedIndex)
}
