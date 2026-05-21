package events

import (
	"context"
	"errors"
	"testing"
)

type testEvent struct{}

func (testEvent) Name() string { return "test_event" }

type testListener struct {
	handled int
	err     error
	ctx     context.Context
	event   Event
}

func (l *testListener) Handle(ctx context.Context, event Event) error {
	l.handled++
	l.ctx = ctx
	l.event = event
	return l.err
}

func TestNewDispatcher(t *testing.T) {
	d := NewDispatcher()

	if d == nil {
		t.Fatal("NewDispatcher devolvio nil")
	}

	if d.listeners == nil {
		t.Fatal("listeners no fue inicializado")
	}
}

func TestDispatchWithoutListeners(t *testing.T) {
	d := NewDispatcher()

	err := d.Dispatch(context.Background(), testEvent{})
	if err != nil {
		t.Fatalf("Dispatch devolvio error inesperado: %v", err)
	}
}

func TestRegisterAndDispatchCallsAllListeners(t *testing.T) {
	d := NewDispatcher()
	event := testEvent{}
	ctx := context.WithValue(context.Background(), "trace", "123")

	l1 := &testListener{}
	l2 := &testListener{}

	d.Register(event.Name(), l1)
	d.Register(event.Name(), l2)

	err := d.Dispatch(ctx, event)
	if err != nil {
		t.Fatalf("Dispatch devolvio error inesperado: %v", err)
	}

	if l1.handled != 1 || l2.handled != 1 {
		t.Fatalf("se esperaban 2 listeners ejecutados una vez, got l1=%d l2=%d", l1.handled, l2.handled)
	}

	if l1.ctx != ctx || l2.ctx != ctx {
		t.Fatal("Dispatch no propago el contexto correcto")
	}

	if l1.event != event || l2.event != event {
		t.Fatal("Dispatch no propago el evento correcto")
	}
}

func TestDispatchStopsOnFirstError(t *testing.T) {
	d := NewDispatcher()
	event := testEvent{}

	wantErr := errors.New("listener fallo")
	failing := &testListener{err: wantErr}
	next := &testListener{}

	d.Register(event.Name(), failing)
	d.Register(event.Name(), next)

	err := d.Dispatch(context.Background(), event)
	if !errors.Is(err, wantErr) {
		t.Fatalf("se esperaba error %v, got %v", wantErr, err)
	}

	if failing.handled != 1 {
		t.Fatalf("listener con error deberia ejecutarse 1 vez, got %d", failing.handled)
	}

	if next.handled != 0 {
		t.Fatalf("listener posterior no deberia ejecutarse, got %d", next.handled)
	}
}

func TestGetEventName(t *testing.T) {
	tests := []struct {
		name  string
		event interface{}
		want  string
	}{
		{name: "valor", event: testEvent{}, want: "testEvent"},
		{name: "puntero", event: &testEvent{}, want: "testEvent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEventName(tt.event)
			if got != tt.want {
				t.Fatalf("GetEventName() = %q, want %q", got, tt.want)
			}
		})
	}
}
