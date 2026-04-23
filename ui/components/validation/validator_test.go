package validation

import (
	"strings"
	"testing"
)

func TestFuncValidator_WithMessageOverridesReturnedError(t *testing.T) {
	validator := Pattern(`^\d+$`).WithMessage("digits only")

	err := validator.Validate("abc")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "digits only" {
		t.Fatalf("expected custom message %q, got %q", "digits only", err.Error())
	}
}

func TestCompositeValidator_WithMessageOverridesReturnedError(t *testing.T) {
	validator := NewAllValidator(MinLength(3), MaxLength(5)).WithMessage("length out of range")

	err := validator.Validate("ab")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "length out of range" {
		t.Fatalf("expected composite custom message %q, got %q", "length out of range", err.Error())
	}
}

func TestCompositeValidator_ModeAnyReturnsCustomMessage(t *testing.T) {
	validator := NewAnyValidator(Pattern(`^\d+$`), Pattern(`^[a-z]+$`)).WithMessage("must be digits or lowercase letters")

	err := validator.Validate("ABC")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "must be digits or lowercase letters" {
		t.Fatalf("expected custom any-mode message %q, got %q", "must be digits or lowercase letters", err.Error())
	}
}

func TestEmailValidator_UsesConfiguredMessageWithoutLeakingInnerDetail(t *testing.T) {
	err := Email().Validate("not-an-email")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "请输入有效的邮箱地址" {
		t.Fatalf("expected email validator message only, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "格式不正确") {
		t.Fatalf("expected wrapped inner detail to stay hidden, got %q", err.Error())
	}
}
