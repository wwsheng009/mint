package ui

import "testing"

func TestControlBuilderFactories(t *testing.T) {
	if vnode := NewSliderBuilder().Label("Volume").Value(40).Build(); vnode == nil {
		t.Fatal("NewSliderBuilder().Build() returned nil")
	}
	if vnode := NewRateBuilder().Label("Score").Value(4).Build(); vnode == nil {
		t.Fatal("NewRateBuilder().Build() returned nil")
	}
	if vnode := NewOptionGroupBuilder([]OptionGroupOption{
		NewOptionGroupOption("a", "Alpha"),
		NewOptionGroupOption("b", "Beta"),
	}).Multiple().Build(); vnode == nil {
		t.Fatal("NewOptionGroupBuilder().Build() returned nil")
	}
	if vnode := NewVirtualListBuilder().Items([]string{"A", "B"}).Height(5).Build(); vnode == nil {
		t.Fatal("NewVirtualListBuilder().Build() returned nil")
	}
	if vnode := NewToastBuilder("Saved").Success().Build(); vnode == nil {
		t.Fatal("NewToastBuilder().Build() returned nil")
	}
}

func TestControlSeedConstructors(t *testing.T) {
	if vnode := NewSlider(); vnode == nil {
		t.Fatal("NewSlider() returned nil")
	} else if vnode.Tag() != "slider" {
		t.Fatalf("NewSlider().Tag() = %q, want slider", vnode.Tag())
	}
	if vnode := NewRate(); vnode == nil {
		t.Fatal("NewRate() returned nil")
	} else if vnode.Tag() != "rate" {
		t.Fatalf("NewRate().Tag() = %q, want rate", vnode.Tag())
	}
	if vnode := NewOptionGroup([]OptionGroupOption{NewOptionGroupOption("x", "Choice")}); vnode == nil {
		t.Fatal("NewOptionGroup() returned nil")
	} else if vnode.Tag() != "optiongroup" {
		t.Fatalf("NewOptionGroup().Tag() = %q, want optiongroup", vnode.Tag())
	}
	if vnode := NewToast("Saved"); vnode == nil {
		t.Fatal("NewToast() returned nil")
	} else if vnode.Tag() != "toast" {
		t.Fatalf("NewToast().Tag() = %q, want toast", vnode.Tag())
	}
}

func TestControlShortcuts(t *testing.T) {
	if vnode := Slider().Value(30).Build(); vnode == nil {
		t.Fatal("Slider().Build() returned nil")
	} else if vnode.Tag() != "slider" {
		t.Fatalf("Slider().Build().Tag() = %q, want slider", vnode.Tag())
	}
	if vnode := Rate().Value(3).Build(); vnode == nil {
		t.Fatal("Rate().Build() returned nil")
	} else if vnode.Tag() != "rate" {
		t.Fatalf("Rate().Build().Tag() = %q, want rate", vnode.Tag())
	}
	if vnode := OptionGroup([]OptionGroupOption{
		NewOptionGroupOption("1", "One"),
		NewOptionGroupOption("2", "Two"),
	}).Horizontal().Build(); vnode == nil {
		t.Fatal("OptionGroup().Build() returned nil")
	} else if vnode.Tag() != "optiongroup" {
		t.Fatalf("OptionGroup().Build().Tag() = %q, want optiongroup", vnode.Tag())
	}
	if vnode := VirtualListOfSize([]string{"A", "B", "C"}, 24, 6); vnode == nil {
		t.Fatal("VirtualListOfSize() returned nil")
	} else if vnode.Tag() != "virtuallist" {
		t.Fatalf("VirtualListOfSize().Tag() = %q, want virtuallist", vnode.Tag())
	} else {
		props := vnode.Props()
		if got := props["width"]; got != 24 {
			t.Fatalf("VirtualListOfSize().Props()[\"width\"] = %v, want 24", got)
		}
		if got := props["height"]; got != 6 {
			t.Fatalf("VirtualListOfSize().Props()[\"height\"] = %v, want 6", got)
		}
	}
	if vnode := Toast("Saved"); vnode == nil {
		t.Fatal("Toast() returned nil")
	} else if vnode.Tag() != "toast" {
		t.Fatalf("Toast().Tag() = %q, want toast", vnode.Tag())
	}
	if vnode := ToastInfo("Saved"); vnode == nil {
		t.Fatal("ToastInfo() returned nil")
	} else if vnode.Tag() != "toast" {
		t.Fatalf("ToastInfo().Tag() = %q, want toast", vnode.Tag())
	}
}
