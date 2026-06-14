package bus

import (
	"encoding/json"
	"fmt"
	"sync"
)

type SubList int

const (
	Shutdown SubList = iota // 0
	Logging                 // 1
	DBWrites                // 2
)

type Event struct {
	SubList SubList
	Action  string
	Payload []byte
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
		for i, v := range subs {
			if v == ch {
				subs[i] = subs[len(subs)-1]
				subs[len(subs)-1] = nil
				subs = subs[:len(subs)-1]
				break
			}
		}
	}
	bus.mu.Unlock()
}

func Subscription(subLists ...SubList) {
	ch := GlobalBus.Subscribe(subLists...)
	defer GlobalBus.Unsubscribe(ch, subLists...)
	for event := range ch {
		switch event.SubList {
		case Shutdown:
		case Logging:
			entry, err := json.Marshal(event.Payload)
			if err != nil {
				fmt.Printf("error marshaling log event payload: %v", err)
			} else {
				fmt.Println(string(entry))
			}
		case DBWrites:
		}
	}
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
