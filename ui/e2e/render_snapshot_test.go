package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const updateRenderSnapshotsEnv = "MINT_UPDATE_E2E_SNAPSHOTS"

func assertRenderSnapshot(t *testing.T, app *App, relativePath string) {
	t.Helper()

	snapshotPath := filepath.Join("testdata", relativePath)
	actual := normalizeRenderSnapshot(app.RenderString())

	if os.Getenv(updateRenderSnapshotsEnv) != "" {
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755); err != nil {
			t.Fatalf("create snapshot dir for %q: %v", snapshotPath, err)
		}
		if err := os.WriteFile(snapshotPath, []byte(actual+"\n"), 0o644); err != nil {
			t.Fatalf("write snapshot %q: %v", snapshotPath, err)
		}
		return
	}

	expectedBytes, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("read snapshot %q: %v\n如需生成基线，可设置 %s=1 后重新运行测试。", snapshotPath, err, updateRenderSnapshotsEnv)
	}

	expected := normalizeRenderSnapshot(string(expectedBytes))
	if actual == expected {
		return
	}

	if dir, saveErr := app.SaveDiagnosticsTemp("mint-e2e-snapshot-"); saveErr == nil {
		t.Logf("render snapshot mismatch diagnostics saved to %s", dir)
	}

	t.Fatalf(
		"render snapshot mismatch for %s\nexpected:\n%s\n\nactual:\n%s",
		relativePath,
		expected,
		actual,
	)
}

func normalizeRenderSnapshot(render string) string {
	render = strings.ReplaceAll(render, "\r\n", "\n")
	lines := strings.Split(render, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	lines = trimTrailingEmptyLines(lines)
	return strings.Join(lines, "\n")
}

func trimTrailingEmptyLines(lines []string) []string {
	end := len(lines)
	for end > 0 && lines[end-1] == "" {
		end--
	}
	return lines[:end]
}
