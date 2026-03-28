package animation

import (
	"time"
)

// Common animation builders for quick animation creation

// FadeIn creates a fade-in animation (opacity 0 to 1).
func FadeIn(id string, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationFade, duration).
		WithFromTo(0.0, 1.0).
		WithEasingName("ease-out-quad")
}

// FadeOut creates a fade-out animation (opacity 1 to 0).
func FadeOut(id string, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationFade, duration).
		WithFromTo(1.0, 0.0).
		WithEasingName("ease-in-quad")
}

// CreateSlideUp creates a slide-up animation.
func CreateSlideUp(id string, distance, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(SlideUp).
		WithFromTo(distance, 0).
		WithEasingName("ease-out-cubic")
}

// CreateSlideDown creates a slide-down animation.
func CreateSlideDown(id string, distance, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(SlideDown).
		WithFromTo(0, distance).
		WithEasingName("ease-out-cubic")
}

// CreateSlideLeft creates a slide-left animation.
func CreateSlideLeft(id string, distance, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(SlideLeft).
		WithFromTo(distance, 0).
		WithEasingName("ease-out-cubic")
}

// CreateSlideRight creates a slide-right animation.
func CreateSlideRight(id string, distance, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(SlideRight).
		WithFromTo(0, distance).
		WithEasingName("ease-out-cubic")
}

// ScaleUp creates a scale-up animation (scale 1 to larger).
func ScaleUp(id string, fromScale, toScale float64, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationScale, duration).
		WithFromTo(fromScale, toScale).
		WithEasingName("ease-out-back")
}

// ScaleDown creates a scale-down animation (scale larger to 1).
func ScaleDown(id string, fromScale, toScale float64, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationScale, duration).
		WithFromTo(fromScale, toScale).
		WithEasingName("ease-in-back")
}

// Progress creates a progress animation (0 to 100).
func Progress(id string, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationProgress, duration).
		WithFromTo(0, 100).
		WithEasingName("linear")
}

// Typewriter creates a typewriter text animation.
func Typewriter(id string, text string, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationCustom, duration).
		WithFromTo("", text).
		WithEasingName("linear")
}

// Pulse creates a pulsing animation (scales up and down repeatedly).
func Pulse(id string, minScale, maxScale float64, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationScale, duration).
		WithFromTo(minScale, maxScale).
		WithEasingName("ease-in-out-sine").
		WithRepeat(-1).     // Infinite
		WithAlternate(true) // Ping-pong
}

// Shake creates a shake animation (oscillates left and right).
func Shake(id string, distance int, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(SlideLeft).
		WithFromTo(-distance, distance).
		WithEasingName("ease-in-out-sine").
		WithRepeat(3). // Shake 3 times
		WithAlternate(true)
}

// Bounce creates a bounce animation.
func Bounce(id string, height int, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(SlideUp).
		WithFromTo(0, height).
		WithEasingName("ease-out-bounce")
}

// Rotate creates a rotation animation (for components that support rotation).
func Rotate(id string, fromAngle, toAngle float64, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationRotate, duration).
		WithFromTo(fromAngle, toAngle).
		WithEasingName("ease-in-out-cubic")
}

// ElasticPop creates an elastic pop-in animation.
func ElasticPop(id string, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationScale, duration).
		WithFromTo(0.0, 1.0).
		WithEasingName("ease-out-elastic")
}

// SlideInFrom creates a slide-in animation from a specific direction.
func SlideInFrom(id string, dir SlideDirection, distance int, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(dir).
		WithFromTo(distance, 0).
		WithEasingName("ease-out-cubic")
}

// SlideOutTo creates a slide-out animation to a specific direction.
func SlideOutTo(id string, dir SlideDirection, distance int, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(dir).
		WithFromTo(0, distance).
		WithEasingName("ease-in-cubic")
}

// Wait creates a delay animation (does nothing for the specified duration).
func Wait(id string, duration time.Duration) *Animation {
	return NewAnimation(id, AnimationCustom, duration).
		WithFromTo(nil, nil)
}

// Sequence creates a sequence of animations that play one after another.
func Sequence(id string, animations ...*Animation) *Animation {
	steps := make([]*Animation, 0, len(animations))
	totalDuration := time.Duration(0)
	for _, anim := range animations {
		if anim == nil {
			continue
		}
		clone := anim.Clone()
		steps = append(steps, clone)
		totalDuration += animationTimelineDuration(clone)
	}

	seq := NewAnimation(id, AnimationCustom, totalDuration).
		WithEasingName("linear")
	if len(steps) == 0 {
		return seq.WithFromTo(nil, nil)
	}

	seq.From = steps[0].From
	seq.To = sequenceFinalValue(steps)
	seq.Current = seq.From
	seq.OnProgress = func(_ float64) {
		elapsed := seq.Elapsed
		current := seq.From

		for _, step := range steps {
			stepDuration := animationTimelineDuration(step)
			if elapsed < stepDuration {
				current, _ = sampleAnimationAt(step, elapsed)
				seq.Current = current
				return
			}

			current, _ = sampleAnimationAt(step, stepDuration)
			elapsed -= stepDuration
		}

		seq.Current = current
	}
	return seq
}

// Parallel creates multiple animations that play simultaneously.
// Returns them as a slice for easy addition to a manager.
func Parallel(animations ...*Animation) []*Animation {
	return animations
}

// Delay creates a delayed version of an animation.
func Delay(anim *Animation, delay time.Duration) *Animation {
	anim.Delay = delay
	return anim
}

// SpeedUp creates a faster version of an animation.
func SpeedUp(anim *Animation, factor float64) *Animation {
	anim.Duration = time.Duration(float64(anim.Duration) / factor)
	return anim
}

// SlowDown creates a slower version of an animation.
func SlowDown(anim *Animation, factor float64) *Animation {
	anim.Duration = time.Duration(float64(anim.Duration) * factor)
	return anim
}

func animationTimelineDuration(anim *Animation) time.Duration {
	if anim == nil {
		return 0
	}

	playCount := 1
	if anim.Repeat > 0 {
		playCount = anim.Repeat
	}

	total := anim.Delay + anim.Duration*time.Duration(playCount)
	if playCount > 1 {
		total += anim.RepeatDelay * time.Duration(playCount-1)
	}
	return total
}

func sampleAnimationAt(anim *Animation, elapsed time.Duration) (interface{}, bool) {
	if anim == nil {
		return nil, true
	}
	if elapsed <= 0 {
		return anim.From, false
	}
	if elapsed < anim.Delay {
		return anim.From, false
	}

	playCount := 1
	if anim.Repeat > 0 {
		playCount = anim.Repeat
	}
	if anim.Duration <= 0 {
		return playFinalValue(anim, playCount-1), true
	}

	elapsed -= anim.Delay
	for play := 0; play < playCount; play++ {
		if elapsed < anim.Duration {
			from, to := playRange(anim, play)
			progress := float64(elapsed) / float64(anim.Duration)
			if anim.Easing != nil {
				progress = anim.Easing(progress)
			}
			return interpolate(from, to, progress), false
		}

		elapsed -= anim.Duration
		finalValue := playFinalValue(anim, play)
		if play == playCount-1 {
			return finalValue, true
		}
		if elapsed < anim.RepeatDelay {
			return finalValue, false
		}
		elapsed -= anim.RepeatDelay
	}

	return playFinalValue(anim, playCount-1), true
}

func playRange(anim *Animation, play int) (interface{}, interface{}) {
	if anim.Alternate && play%2 == 1 {
		return anim.To, anim.From
	}
	return anim.From, anim.To
}

func playFinalValue(anim *Animation, play int) interface{} {
	_, to := playRange(anim, play)
	return to
}

func sequenceFinalValue(steps []*Animation) interface{} {
	if len(steps) == 0 {
		return nil
	}
	last := steps[len(steps)-1]
	value, _ := sampleAnimationAt(last, animationTimelineDuration(last))
	return value
}
