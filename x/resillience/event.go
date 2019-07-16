package resilience

//go:generate stringer -type Event -trimprefix=Event
type Event int

const (
	EventNew     Event = iota // new and start
	EventClose                // close
	EventExpired              // restart
)
