package e2e

import (
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
)

// Driver injects high-level user interactions.
type Driver struct {
	app *App
}

func (d *Driver) Key(key rune) error {
	before, beforeOK := d.app.FocusSnapshot()
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       key,
		Timestamp: time.Now(),
	}
	d.app.recordRawInput(raw, "key")
	if err := d.app.FrameworkApp().InjectEvent(raw); err != nil {
		return err
	}
	if err := d.app.AwaitIdle(); err != nil {
		return err
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}

func (d *Driver) KeyWithMod(key rune, mod platform.KeyModifier) error {
	before, beforeOK := d.app.FocusSnapshot()
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       key,
		Modifiers: mod,
		Timestamp: time.Now(),
	}
	d.app.recordRawInput(raw, "key_mod")
	if err := d.app.FrameworkApp().InjectEvent(raw); err != nil {
		return err
	}
	if err := d.app.AwaitIdle(); err != nil {
		return err
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}

func (d *Driver) Special(key platform.SpecialKey) error {
	before, beforeOK := d.app.FocusSnapshot()
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Special:   key,
		Timestamp: time.Now(),
	}
	d.app.recordRawInput(raw, "special")
	if err := d.app.FrameworkApp().InjectEvent(raw); err != nil {
		return err
	}
	if err := d.app.AwaitIdle(); err != nil {
		return err
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}

func (d *Driver) SpecialWithMod(key platform.SpecialKey, mod platform.KeyModifier) error {
	before, beforeOK := d.app.FocusSnapshot()
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Special:   key,
		Modifiers: mod,
		Timestamp: time.Now(),
	}
	d.app.recordRawInput(raw, "special_mod")
	if err := d.app.FrameworkApp().InjectEvent(raw); err != nil {
		return err
	}
	if err := d.app.AwaitIdle(); err != nil {
		return err
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}

func (d *Driver) Type(text string) error {
	before, beforeOK := d.app.FocusSnapshot()
	for _, r := range text {
		raw := platform.RawInput{
			Type:      platform.InputKeyPress,
			Key:       r,
			Timestamp: time.Now(),
		}
		d.app.recordRawInput(raw, "type")
		if err := d.app.FrameworkApp().InjectEvent(raw); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := d.app.AwaitIdle(); err != nil {
		return err
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}

func (d *Driver) Click(locator Locator) error {
	point, err := d.app.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return d.ClickAt(point.X, point.Y)
}

func (d *Driver) ClickAt(x, y int) error {
	before, beforeOK := d.app.FocusSnapshot()
	press := platform.RawInput{
		Type:        platform.InputMouse,
		MouseX:      x,
		MouseY:      y,
		MouseButton: platform.MouseLeft,
		MouseAction: platform.MousePress,
		Timestamp:   time.Now(),
	}
	d.app.recordRawInput(press, "mouse_press")
	if err := d.app.FrameworkApp().InjectEvent(press); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	release := platform.RawInput{
		Type:        platform.InputMouse,
		MouseX:      x,
		MouseY:      y,
		MouseButton: platform.MouseLeft,
		MouseAction: platform.MouseRelease,
		Timestamp:   time.Now(),
	}
	d.app.recordRawInput(release, "mouse_release")
	if err := d.app.FrameworkApp().InjectEvent(release); err != nil {
		return err
	}
	if err := d.app.AwaitIdle(); err != nil {
		return err
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}
