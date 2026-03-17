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
	return d.mouseSequence(
		false,
		10*time.Millisecond,
		d.mouseRawInput(x, y, platform.MouseLeft, platform.MousePress),
		d.mouseRawInput(x, y, platform.MouseLeft, platform.MouseRelease),
	)
}

func (d *Driver) Press(locator Locator) error {
	point, err := d.app.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return d.PressAt(point.X, point.Y)
}

func (d *Driver) PressAt(x, y int) error {
	return d.mouseSequence(false, 0, d.mouseRawInput(x, y, platform.MouseLeft, platform.MousePress))
}

func (d *Driver) Move(locator Locator) error {
	point, err := d.app.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return d.MoveAt(point.X, point.Y)
}

func (d *Driver) MoveAt(x, y int) error {
	return d.mouseSequence(false, 0, d.mouseRawInput(x, y, platform.MouseNone, platform.MouseMotion))
}

func (d *Driver) Release(locator Locator) error {
	point, err := d.app.ResolvePoint(locator)
	if err != nil {
		return err
	}
	return d.ReleaseAt(point.X, point.Y)
}

func (d *Driver) ReleaseAt(x, y int) error {
	return d.mouseSequence(false, 0, d.mouseRawInput(x, y, platform.MouseLeft, platform.MouseRelease))
}

func (d *Driver) Drag(from Locator, to Locator) error {
	start, err := d.app.ResolvePoint(from)
	if err != nil {
		return err
	}
	end, err := d.app.ResolvePoint(to)
	if err != nil {
		return err
	}
	return d.DragAt(start.X, start.Y, end.X, end.Y)
}

func (d *Driver) DragAt(fromX, fromY, toX, toY int) error {
	return d.mouseSequence(
		true,
		10*time.Millisecond,
		d.mouseRawInput(fromX, fromY, platform.MouseLeft, platform.MousePress),
		d.mouseRawInput(toX, toY, platform.MouseLeft, platform.MouseMotion),
		d.mouseRawInput(toX, toY, platform.MouseLeft, platform.MouseRelease),
	)
}

func (d *Driver) mouseRawInput(x, y int, button platform.MouseButton, action platform.MouseAction) platform.RawInput {
	return platform.RawInput{
		Type:        platform.InputMouse,
		MouseX:      x,
		MouseY:      y,
		MouseButton: button,
		MouseAction: action,
		Timestamp:   time.Now(),
	}
}

func (d *Driver) mouseSequence(awaitEach bool, stepDelay time.Duration, raws ...platform.RawInput) error {
	before, beforeOK := d.app.FocusSnapshot()
	for i, raw := range raws {
		label := "mouse"
		switch raw.MouseAction {
		case platform.MousePress:
			label = "mouse_press"
		case platform.MouseRelease:
			label = "mouse_release"
		case platform.MouseMotion:
			label = "mouse_move"
		}
		d.app.recordRawInput(raw, label)
		if err := d.app.FrameworkApp().InjectEvent(raw); err != nil {
			return err
		}
		if awaitEach {
			if err := d.app.AwaitIdle(); err != nil {
				return err
			}
		}
		if stepDelay > 0 && i < len(raws)-1 {
			time.Sleep(stepDelay)
		}
	}
	if !awaitEach {
		if err := d.app.AwaitIdle(); err != nil {
			return err
		}
	}
	after, afterOK := d.app.FocusSnapshot()
	d.app.recordFocusTransition(before, beforeOK, after, afterOK)
	return nil
}
