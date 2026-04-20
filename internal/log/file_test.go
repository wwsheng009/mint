package log

import (
	"sync"
	"testing"
)

func TestGetLogOutputDefaultIsFile(t *testing.T) {
	t.Setenv("TUI_LOG_OUTPUT", "")
	resetLogOutputForTest()

	if got := getLogOutput(); got != OutputFile {
		t.Fatalf("getLogOutput() = %v, want %v", got, OutputFile)
	}
}

func TestGetLogOutputExplicitModes(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want LogOutput
	}{
		{name: "console", env: "console", want: OutputConsole},
		{name: "both", env: "both", want: OutputBoth},
		{name: "file", env: "file", want: OutputFile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TUI_LOG_OUTPUT", tt.env)
			resetLogOutputForTest()

			if got := getLogOutput(); got != tt.want {
				t.Fatalf("getLogOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func resetLogOutputForTest() {
	outputOnce = sync.Once{}
	logOutput = OutputFile
}
