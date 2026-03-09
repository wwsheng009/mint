package table

type SelectionMode int

const (
	SelectionNone SelectionMode = iota
	SelectionSingle
	SelectionMultiple
)
