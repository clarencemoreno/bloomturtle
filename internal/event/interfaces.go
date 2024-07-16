package event

import (
	"context"
)

// Event represents a generic event.
type Event interface{}

// EventListener is an interface for listening to events.
type EventListener interface {
	HandleEvent(ctx context.Context, event Event) error
}

// EventPublisher is an interface for publishing events.
type EventPublisher interface {
	Start()
	PublishEvent(event Event) error
	AddListener(listener EventListener)
	Shutdown(ctx context.Context) error
}
