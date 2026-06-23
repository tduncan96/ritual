package bus

import (
	"slices"
	"sync"
)

type SubList int

const (
	LifeCycle SubList = iota // 0
	Database                 // 1
)

type Method int

const (
	GET    Method = iota // 0
	POST                 // 1
	PUT                  // 2
	DELETE               // 3
)

type Event struct {
	SubList SubList `json:"sub_list"`
	Method  Method  `json:"method"`
	Payload []byte  `json:"payload"`
}

type EventBus struct {
	mu          sync.Mutex
	subscribers map[SubList][]chan Event
}

var GlobalBus *EventBus

func MakeBus() {
	GlobalBus = &EventBus{subscribers: make(map[SubList][]chan Event)}
}

func (bus *EventBus) Subscribe(subLists ...SubList) <-chan Event {
	ch := make(chan Event, 16)
	bus.mu.Lock()
	for _, list := range subLists {
		bus.subscribers[list] = append(bus.subscribers[list], ch)
	}
	bus.mu.Unlock()
	return ch
}

func (bus *EventBus) Unsubscribe(ch <-chan Event, subLists ...SubList) {
	bus.mu.Lock()
	for _, list := range subLists {
		subs := bus.subscribers[list]

		i := slices.IndexFunc(subs, func(c chan Event) bool { return c == ch })
		if i >= 0 {
			bus.subscribers[list] = slices.Delete(subs, i, i+1)
		}
	}
	bus.mu.Unlock()
}

func (bus *EventBus) Publish(events ...Event) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for _, event := range events {
		for _, ch := range bus.subscribers[event.SubList] {
			ch <- event
		}
	}
}
