package events

import (
	"context"
	"reflect"
	"sync"
)

// Event is the interface that all events must satisfy.
type Event interface {
	Name() string
}

// Listener is the interface for event listeners.
type Listener interface {
	Handle(ctx context.Context, event Event) error
}

// Dispatcher manages event listeners and dispatches events.
type Dispatcher struct {
	listeners map[string][]Listener
	mu        sync.RWMutex
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string][]Listener),
	}
}

// Register registers a listener for a specific event type.
func (d *Dispatcher) Register(eventName string, listener Listener) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners[eventName] = append(d.listeners[eventName], listener)
}

// Dispatch dispatches an event to all registered listeners.
func (d *Dispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mu.RLock()
	listeners, ok := d.listeners[event.Name()]
	d.mu.RUnlock()

	if !ok {
		return nil
	}

	for _, listener := range listeners {
		if err := listener.Handle(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// GetEventName returns the name of the event type.
func GetEventName(event interface{}) string {
	if t := reflect.TypeOf(event); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
