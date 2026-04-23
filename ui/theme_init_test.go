package ui

import "testing"

func TestRunTest_InitializesValidTheme(t *testing.T) {
	testApp, err := RunTest(func() VNode {
		return Text("theme check")
	}, WithWidth(20), WithHeight(4))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	if !testApp.fwApp.IsThemeEnabled() {
		t.Fatal("RunTest should initialize a valid theme")
	}
	if testApp.fwApp.GetTheme() == "" {
		t.Fatal("RunTest should expose the active theme name")
	}
}
