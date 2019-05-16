package resilience

type Event int

const (
	EventNew     Event = iota // new and start
	EventClose                // close
	EventExpired              // restart
)

func (e Event) String() string {
	switch e {
	case EventNew:
		return "new"
	case EventClose:
		return "close"
	case EventExpired:
		return "expired"
	default:
		return "unknown event"
	}
}
