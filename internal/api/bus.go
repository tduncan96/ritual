package api

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
	Cron                    // 3
)

type Event struct {
	SubList SubList
	Payload map[string]any
}

type EventBus struct {
	mu          sync.Mutex
	subscribers map[SubList][]chan Event
}

var Bus *EventBus

func NewBus() *EventBus {
	return &EventBus{subscribers: make(map[SubList][]chan Event)}
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
	ch := Bus.Subscribe(subLists...)
	defer Bus.Unsubscribe(ch, subLists...)
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
		case Cron:
		}
	}
}

func (bus *EventBus) Publish(event Event) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for _, ch := range bus.subscribers[event.SubList] {
		ch <- event
	}
}
