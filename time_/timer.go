package time_

import (
	"github.com/searKing/golib/sync_/atomic_"
	"time"
)

type Timer struct {
	*time.Timer
	chanConsumed atomic_.Bool
}

//saw channel read, must be called after receiving value from timer chan
func (t *Timer) ChanConsumed() {
	t.chanConsumed.Store(true)
}

func (t *Timer) Reset(d time.Duration) bool {
	ret := t.Stop()
	if !ret && !t.chanConsumed.Load() {
		<-t.C
	}
	t.Timer.Reset(d)
	t.chanConsumed.Store(false)
	return ret
}

func NewTimer(d time.Duration) *Timer {
	return &Timer{
		Timer: time.NewTimer(d),
	}
}

func WrapTimer(t *time.Timer) *Timer {
	return &Timer{
		Timer: t,
	}
}

func After(d time.Duration) <-chan time.Time {
	return NewTimer(d).C
}

func AfterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{}
	t.Timer = time.AfterFunc(d, func() {
		t.ChanConsumed()
		f()
	})
	return t
}
