package defined

type Event int16 // 事件定义

const (
	E_RELOAD_ALL Event = iota
	E_RELOAD_AD
	E_RELOAD_LONGLAT
)

var (
	EventChan chan Event
)

func init() {
	if EventChan == nil {
		EventChan = make(chan Event, 1024)
	}
}

func AddEvent(event Event) {
	EventChan <- event
}
