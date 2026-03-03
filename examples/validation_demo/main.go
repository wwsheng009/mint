package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// ValidationDemo demonstrates the validation package
func ValidationDemo() ui.VNode {
	// Example 1: Individual validators
	requiredValidator := validation.Required()
	emailValidator := validation.Email()
	_ = validation.MinLength(3)
	_ = validation.MaxLength(20)

	// Example 2: Composite validators (AND)
	usernameValidator := validation.NewAllValidator(
		validation.Required(),
		validation.MinLength(3),
		validation.MaxLength(20),
	)

	// Example 3: Composite validators (OR)
	phoneOrEmailValidator := validation.NewAnyValidator(
		validation.Pattern(`^\d{10,11}$`), // phone
		validation.Email(),
	)
	_ = phoneOrEmailValidator

	// Example 4: Chain API (fluent)
	passwordValidator := validation.NewChain().
		Required().
		MinLength(8).
		MaxLength(32).
		Build()

	// Example 5: Custom validator
	customValidator := validation.Custom(
		func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return validation.ErrTypeMismatch
			}
			if len(str) < 6 || len(str) > 10 {
				return fmt.Errorf("密码长度必须在6-10之间")
			}
			return nil
		},
		"密码长度必须在6-10之间",
	)

	return ui.VStack(
		ui.NewTextBuilder("Validation Package Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),

		// Individual validators test
		ui.NewTextBuilder("1. Individual Validators:").Bold(true).Build(),
		ui.Text(fmt.Sprintf("   Required(''): %v", requiredValidator.Validate(""))),
		ui.Text(fmt.Sprintf("   Required('hello'): %v", requiredValidator.Validate("hello"))),
		ui.Text(fmt.Sprintf("   Email('test'): %v", emailValidator.Validate("test"))),
		ui.Text(fmt.Sprintf("   Email('test@test.com'): %v", emailValidator.Validate("test@test.com"))),
		ui.Text(""),

		// Composite validators test
		ui.NewTextBuilder("2. Composite Validators:").Bold(true).Build(),
		ui.Text(fmt.Sprintf("   Username 'ab': %v", usernameValidator.Validate("ab"))),
		ui.Text(fmt.Sprintf("   Username 'john': %v", usernameValidator.Validate("john"))),
		ui.Text(""),

		// Chain API test
		ui.NewTextBuilder("3. Chain API (Password):").Bold(true).Build(),
		ui.Text(fmt.Sprintf("   Password '123': %v", passwordValidator.Validate("123"))),
		ui.Text(fmt.Sprintf("   Password '12345678': %v", passwordValidator.Validate("12345678"))),
		ui.Text(""),

		// Custom validator test
		ui.NewTextBuilder("4. Custom Validator:").Bold(true).Build(),
		ui.Text(fmt.Sprintf("   Custom '12345': %v", customValidator.Validate("12345"))),
		ui.Text(fmt.Sprintf("   Custom '123456': %v", customValidator.Validate("123456"))),
		ui.Text(""),

		// OneOf validator
		ui.NewTextBuilder("5. OneOf (Enum):").Bold(true).Build(),
		ui.Text(fmt.Sprintf("   OneOf('red', 'red'): %v", validation.OneOf("red", "green", "blue").Validate("red"))),
		ui.Text(fmt.Sprintf("   OneOf('red', 'yellow'): %v", validation.OneOf("red", "green", "blue").Validate("yellow"))),
		ui.Text(""),

		ui.NewTextBuilder("Press q to quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(ValidationDemo,
		ui.WithWidth(60),
		ui.WithHeight(30),
		ui.WithTitle("Validation Demo"),
	)
	if err != nil {
		panic(err)
	}
}
