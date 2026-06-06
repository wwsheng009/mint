package msg

import (
	"fmt"
	"time"
	"unicode/utf8"
)

// PasteMsg represents a terminal paste event carrying the full pasted text.
//
// String() deliberately avoids printing the pasted content because it may
// contain credentials or other sensitive values.
type PasteMsg struct {
	BaseMsg
	Text string
}

// NewPasteMsg creates a paste message with the full pasted text.
func NewPasteMsg(text string) *PasteMsg {
	return &PasteMsg{
		BaseMsg: BaseMsg{
			TypeValue:      MsgTypePaste,
			TimestampValue: time.Now(),
		},
		Text: text,
	}
}

// String returns a non-sensitive description of the paste message.
func (p *PasteMsg) String() string {
	return fmt.Sprintf("PasteMsg{runes=%d bytes=%d}", utf8.RuneCountInString(p.Text), len(p.Text))
}
