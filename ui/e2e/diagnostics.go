package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
)

// DiagnosticsSnapshot is a serializable failure-diagnostics bundle.
type DiagnosticsSnapshot struct {
	CapturedAt       time.Time                `json:"captured_at"`
	Render           string                   `json:"render"`
	Focus            *FocusSnapshot           `json:"focus,omitempty"`
	RawInputs        []RawInputEvent          `json:"raw_inputs,omitempty"`
	Messages         []MessageEvent           `json:"messages,omitempty"`
	Actions          []ActionEvent            `json:"actions,omitempty"`
	Intents          []IntentDispatchSnapshot `json:"intents,omitempty"`
	FocusTransitions []FocusTransition        `json:"focus_transitions,omitempty"`
	Trace            []TraceSummaryEvent      `json:"trace,omitempty"`
	HitMap           []HitEntrySnapshot       `json:"hitmap,omitempty"`
}

// IntentDispatchSnapshot is the JSON-safe form of runtimeintent.DispatchLog.
type IntentDispatchSnapshot struct {
	Type      string    `json:"type"`
	Priority  string    `json:"priority"`
	Lane      string    `json:"lane"`
	Timestamp time.Time `json:"timestamp"`
	Handled   bool      `json:"handled"`
	Error     string    `json:"error,omitempty"`
}

// TraceSummaryEvent is a payload-free trace event summary.
type TraceSummaryEvent struct {
	Kind      TraceKind `json:"kind"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

// HitEntrySnapshot is the JSON-safe form of one hit map entry.
type HitEntrySnapshot struct {
	NodeID   uint64 `json:"node_id"`
	TargetID string `json:"target_id,omitempty"`
	ZOrder   int    `json:"z_order"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// DiagnosticsSnapshot captures the current diagnostic state for failure analysis.
func (a *App) DiagnosticsSnapshot() DiagnosticsSnapshot {
	snapshot := DiagnosticsSnapshot{
		CapturedAt:       time.Now(),
		Render:           a.RenderString(),
		RawInputs:        a.RawInputs(),
		Messages:         a.MessageEvents(),
		Actions:          a.ActionEvents(),
		FocusTransitions: a.FocusTransitions(),
	}

	if focus, ok := a.FocusSnapshot(); ok {
		copyFocus := focus
		snapshot.Focus = &copyFocus
	}

	intentLogs := a.IntentLogs()
	snapshot.Intents = make([]IntentDispatchSnapshot, 0, len(intentLogs))
	for _, logEntry := range intentLogs {
		snapshot.Intents = append(snapshot.Intents, intentLogSnapshot(logEntry))
	}

	traceEvents := a.TraceEvents()
	snapshot.Trace = make([]TraceSummaryEvent, 0, len(traceEvents))
	for _, event := range traceEvents {
		snapshot.Trace = append(snapshot.Trace, TraceSummaryEvent{
			Kind:      event.Kind,
			Name:      event.Name,
			Timestamp: event.Timestamp,
		})
	}

	if hitMap := a.HitMap(); hitMap != nil {
		entries := hitMap.AllEntries()
		snapshot.HitMap = make([]HitEntrySnapshot, 0, len(entries))
		for _, entry := range entries {
			snapshot.HitMap = append(snapshot.HitMap, HitEntrySnapshot{
				NodeID:   entry.NodeID,
				TargetID: safeTargetID(&entry),
				ZOrder:   entry.ZOrder,
				X:        entry.Bounds.X,
				Y:        entry.Bounds.Y,
				Width:    entry.Bounds.Width,
				Height:   entry.Bounds.Height,
			})
		}
	}

	return snapshot
}

// SaveDiagnostics writes a diagnostics bundle to the target directory.
func (a *App) SaveDiagnostics(dir string) error {
	if dir == "" {
		return fmt.Errorf("diagnostics directory cannot be empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	snapshot := a.DiagnosticsSnapshot()

	renderPath := filepath.Join(dir, "render.txt")
	if err := os.WriteFile(renderPath, []byte(snapshot.Render), 0o644); err != nil {
		return err
	}

	jsonPath := filepath.Join(dir, "diagnostics.json")
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, payload, 0o644); err != nil {
		return err
	}

	tracePath := filepath.Join(dir, "trace.json")
	tracePayload, err := json.MarshalIndent(snapshot.Trace, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tracePath, tracePayload, 0o644); err != nil {
		return err
	}

	return nil
}

// SaveDiagnosticsTemp creates a temporary directory, writes a diagnostics bundle
// into it, and returns the directory path.
func (a *App) SaveDiagnosticsTemp(prefix string) (string, error) {
	if prefix == "" {
		prefix = "mint-e2e-"
	}
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", err
	}
	if err := a.SaveDiagnostics(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// SaveDiagnosticsOnFailure writes a diagnostics bundle only when the test has failed.
// It logs the output directory through t.Logf and returns the directory path when written.
func (a *App) SaveDiagnosticsOnFailure(t *testing.T, prefix string) (string, error) {
	if t == nil {
		return "", fmt.Errorf("testing.T cannot be nil")
	}
	if !t.Failed() {
		return "", nil
	}
	dir, err := a.SaveDiagnosticsTemp(prefix)
	if err != nil {
		return "", err
	}
	t.Logf("e2e diagnostics saved to %s", dir)
	return dir, nil
}

func intentLogSnapshot(logEntry runtimeintent.DispatchLog) IntentDispatchSnapshot {
	snapshot := IntentDispatchSnapshot{
		Type:      logEntry.Type,
		Priority:  logEntry.Priority.String(),
		Lane:      fmt.Sprint(logEntry.Lane),
		Timestamp: logEntry.Timestamp,
		Handled:   logEntry.Handled,
	}
	if logEntry.Error != nil {
		snapshot.Error = logEntry.Error.Error()
	}
	return snapshot
}
