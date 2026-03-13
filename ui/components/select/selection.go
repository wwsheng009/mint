package selectcomp

type SelectionMode int

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
	SelectionTags
)

func isMultiSelectionMode(mode SelectionMode) bool {
	return mode == SelectionMultiple || mode == SelectionTags
}

func isTagsSelectionMode(mode SelectionMode) bool {
	return mode == SelectionTags
}
