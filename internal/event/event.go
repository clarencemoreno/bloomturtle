package event

import (
	"context"
	"errors"
	"sync"
)

// Event represents a generic event.
type Event interface{}

// EventListener is an interface for listening to events.
type EventListener interface {
	HandleEvent(ctx context.Context, event Event) error
}

// BaseEventPublisher is a basic event publisher implementation.
type BaseEventPublisher struct {
	mu        sync.Mutex
	eventChan chan Event
	listeners []EventListener
	ctx       context.Context
	cancel    context.CancelFunc
	closed    bool // Add a flag to track if the channel is closed
}

// NewBaseEventPublisher creates a new BaseEventPublisher.
func NewBaseEventPublisher() *BaseEventPublisher {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseEventPublisher{
		eventChan: make(chan Event, 10), // Create a buffered channel with capacity 10
		listeners: make([]EventListener, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the event publishing loop.
func (p *BaseEventPublisher) Start() {
	go p.run()
}

// PublishEvent publishes an event to all listeners.
func (p *BaseEventPublisher) PublishEvent(event Event) error {
	if p.closed {
		return ErrPublisherClosed
	}
	p.eventChan <- event
	return nil
}

// AddListener adds an event listener to the publisher.
func (p *BaseEventPublisher) AddListener(listener EventListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listeners = append(p.listeners, listener)
}

// Shutdown stops the event publisher and waits for it to finish.
func (p *BaseEventPublisher) Shutdown(ctx context.Context) error {
	p.cancel()
	close(p.eventChan)
	p.closed = true // Set the closed flag
	return nil
}

func (p *BaseEventPublisher) run() {
	for {
		select {
		case event := <-p.eventChan:
			p.mu.Lock()
			for _, listener := range p.listeners {
				listener.HandleEvent(p.ctx, event)
			}
			p.mu.Unlock()
		case <-p.ctx.Done():
			return
		}
	}
}

// ErrPublisherClosed is returned when trying to publish an event on a closed publisher.
var ErrPublisherClosed = errors.New("event publisher is closed")
