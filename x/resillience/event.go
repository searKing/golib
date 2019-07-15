//go:generate stringer -type=Event
package resilience

type Event int

const (
	EventNew     Event = iota // new and start
	EventClose                // close
	EventExpired              // restart
)
