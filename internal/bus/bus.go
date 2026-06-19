package bus

import (
	"encoding/json"
	"log/slog"
	"sync"

	"ritual/internal/cron"
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

func CronSubscription(cr *cron.CronRunner, subLists ...SubList) {
	ch := GlobalBus.Subscribe(subLists...)
	defer GlobalBus.Unsubscribe(ch, subLists...)
	for event := range ch {
		switch event.SubList {
		case LifeCycle:
			switch event.Method {
			case PUT:
				cr.Cron.Start()
			case DELETE:
				cr.Cron.Stop()
			}
		case Database:
			var ids []int64
			if err := json.Unmarshal(event.Payload, &ids); err != nil {
				slog.Error("error unmarshaling event payload", "error", err)
				return
			}
			switch event.Method {
			case POST:
				cr.Cron.Stop()
				if err := cr.UpdateRunner(ids); err != nil {
					slog.Error("error updating cron runner from event payload", "error", err, "ids", ids)
				} else {
					cr.Cron.Start()
					slog.Info("cron runner jobs updated", "ids", ids)
				}
			case DELETE:
				cr.Cron.Stop()
				cr.RemoveRunnerJob(ids)
				cr.Cron.Start()
			}
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
