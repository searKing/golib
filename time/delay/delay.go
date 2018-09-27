package delay

import "time"

const DefaultInitDuration = 5 * time.Millisecond
const DefaultStepTimes = 2
const DefaultMaxDuration = 1 * time.Second

const ZeroDuration = 0

func NewDefaultDelay() *Delay {
	return &Delay{
		InitDuration: DefaultInitDuration,
		MaxDuration:  DefaultMaxDuration,
		DelayAgainHandler: func(delay time.Duration) time.Duration {
			return delay * DefaultStepTimes
		},
	}
}

type Delay struct {
	delay             time.Duration
	InitDuration      time.Duration
	MaxDuration       time.Duration
	DelayAgainHandler func(delay time.Duration) time.Duration
}

func (d *Delay) Update() {
	if d.delay == ZeroDuration {
		d.delay = d.InitDuration
	} else {
		if d.DelayAgainHandler != nil {
			d.delay = d.DelayAgainHandler(d.delay)
		}
	}
	if max := d.MaxDuration; d.delay > max {
		d.delay = max
	}
}
func (d *Delay) Reset() {
	d.delay = ZeroDuration
}
func (d *Delay) Duration() time.Duration {
	return d.delay
}
