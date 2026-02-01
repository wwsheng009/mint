// replay/player_test.go - 回放器测试
package replay

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
)

func TestNewPlayer(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
	}

	p := NewPlayer(events)

	if p.Length() != 2 {
		t.Errorf("NewPlayer() Length() = %v, want 2", p.Length())
	}

	if p.Index() != 0 {
		t.Errorf("NewPlayer() Index() = %v, want 0", p.Index())
	}
}

func TestPlayerPlay(t *testing.T) {
	p := NewPlayer(nil)

	err := p.Play()
	if err != nil {
		t.Fatalf("Play() error = %v", err)
	}

	if !p.IsPlaying() {
		t.Error("Play() IsPlaying() = false, want true")
	}
}

func TestPlayerPause(t *testing.T) {
	p := NewPlayer(nil)
	p.Play()

	err := p.Pause()
	if err != nil {
		t.Fatalf("Pause() error = %v", err)
	}

	if p.IsPlaying() {
		t.Error("Pause() IsPlaying() = true, want false")
	}
}

func TestPlayerStop(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
	}
	p := NewPlayer(events)
	p.Play()
	p.Next() // 移动索引

	err := p.Stop()
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if p.Index() != 0 {
		t.Errorf("Stop() Index() = %v, want 0", p.Index())
	}

	if p.IsPlaying() {
		t.Error("Stop() IsPlaying() = true, want false")
	}
}

func TestPlayerNext(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
	}
	p := NewPlayer(events)

	event, err := p.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	if event.Key != 'a' {
		t.Errorf("Next() Key = %v, want 'a'", event.Key)
	}

	if p.Index() != 1 {
		t.Errorf("Next() Index() = %v, want 1", p.Index())
	}
}

func TestPlayerNextEmpty(t *testing.T) {
	p := NewPlayer(nil)

	_, err := p.Next()
	if err == nil {
		t.Error("Next() should return error on empty events")
	}
}

func TestPlayerPrevious(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
	}
	p := NewPlayer(events)
	p.Next()
	p.Next() // index = 2

	event, err := p.Previous()
	if err != nil {
		t.Fatalf("Previous() error = %v", err)
	}

	if event.Key != 'b' {
		t.Errorf("Previous() Key = %v, want 'b'", event.Key)
	}

	if p.Index() != 1 {
		t.Errorf("Previous() Index() = %v, want 1", p.Index())
	}
}

func TestPlayerPreviousEmpty(t *testing.T) {
	p := NewPlayer(nil)

	_, err := p.Previous()
	if err == nil {
		t.Error("Previous() should return error on empty events")
	}
}

func TestPlayerSeek(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
		{Type: platform.InputKeyPress, Key: 'c'},
	}
	p := NewPlayer(events)

	err := p.Seek(2)
	if err != nil {
		t.Fatalf("Seek() error = %v", err)
	}

	if p.Index() != 2 {
		t.Errorf("Seek() Index() = %v, want 2", p.Index())
	}
}

func TestPlayerSeekOutOfBounds(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
	}
	p := NewPlayer(events)

	err := p.Seek(10)
	if err == nil {
		t.Error("Seek() out of bounds should return error")
	}
}

func TestPlayerSetSpeed(t *testing.T) {
	p := NewPlayer(nil)

	p.SetSpeed(2.5)

	if p.Speed() != 2.5 {
		t.Errorf("SetSpeed() Speed() = %v, want 2.5", p.Speed())
	}
}

func TestPlayerHasNext(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
	}
	p := NewPlayer(events)

	if !p.HasNext() {
		t.Error("HasNext() = false, want true")
	}

	p.Next()
	p.Next()

	if p.HasNext() {
		t.Error("HasNext() = true, want false at end")
	}
}

func TestPlayerHasPrevious(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
	}
	p := NewPlayer(events)

	if p.HasPrevious() {
		t.Error("HasPrevious() = true, want false at start")
	}

	p.Next()

	if !p.HasPrevious() {
		t.Error("HasPrevious() = false, want true after Next()")
	}
}

func TestPlayerReset(t *testing.T) {
	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
	}
	p := NewPlayer(events)
	p.Play()
	p.Next()
	p.SetSpeed(2.0)

	p.Reset()

	if p.Index() != 0 {
		t.Errorf("Reset() Index() = %v, want 0", p.Index())
	}

	if p.IsPlaying() {
		t.Error("Reset() IsPlaying() = true, want false")
	}

	if p.Speed() != 1.0 {
		t.Errorf("Reset() Speed() = %v, want 1.0", p.Speed())
	}
}

func TestRecorder(t *testing.T) {
	r := NewRecorder("test recording")
	r.Start()

	r.Record(platform.RawInput{Type: platform.InputKeyPress, Key: 'a'})
	r.Record(platform.RawInput{Type: platform.InputKeyPress, Key: 'b'})

	recording := r.Stop()

	if recording.Metadata.Title != "test recording" {
		t.Errorf("Recording Title = %v, want 'test recording'", recording.Metadata.Title)
	}

	if len(recording.Events) != 2 {
		t.Errorf("Recording Events length = %v, want 2", len(recording.Events))
	}

	if recording.Metadata.EventCount != 2 {
		t.Errorf("Recording EventCount = %v, want 2", recording.Metadata.EventCount)
	}
}

func TestRecorderDuration(t *testing.T) {
	r := NewRecorder("test")
	r.Start()

	time.Sleep(10 * time.Millisecond)

	r.Stop()

	if r.recording.Metadata.Duration < 10*time.Millisecond {
		t.Error("Recording Duration should be at least 10ms")
	}
}
