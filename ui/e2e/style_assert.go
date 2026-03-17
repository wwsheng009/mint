package e2e

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/style"
)

// StyleExpect matches a subset of style properties.
type StyleExpect struct {
	HasFG bool
	FG    style.Color

	HasBG bool
	BG    style.Color

	HasBold bool
	Bold    bool

	HasItalic bool
	Italic    bool

	HasUnderline bool
	Underline    bool

	HasReverse bool
	Reverse    bool
}

func (s StyleExpect) Match(actual style.Style) error {
	if s.HasFG && actual.FG != s.FG {
		return fmt.Errorf("style fg = %q, want %q", actual.FG, s.FG)
	}
	if s.HasBG && actual.BG != s.BG {
		return fmt.Errorf("style bg = %q, want %q", actual.BG, s.BG)
	}
	if s.HasBold && actual.IsBold() != s.Bold {
		return fmt.Errorf("style bold = %v, want %v", actual.IsBold(), s.Bold)
	}
	if s.HasItalic && actual.IsItalic() != s.Italic {
		return fmt.Errorf("style italic = %v, want %v", actual.IsItalic(), s.Italic)
	}
	if s.HasUnderline && actual.IsUnderline() != s.Underline {
		return fmt.Errorf("style underline = %v, want %v", actual.IsUnderline(), s.Underline)
	}
	if s.HasReverse && actual.IsReverse() != s.Reverse {
		return fmt.Errorf("style reverse = %v, want %v", actual.IsReverse(), s.Reverse)
	}
	return nil
}
