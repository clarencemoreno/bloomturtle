package event

import (
	"context"
	"errors"
)

// BaseEventPublisher is a basic event publisher implementation.
type BaseEventPublisher struct {
	eventChan chan Event
	listeners []EventListener
	ctx       context.Context
	cancel    context.CancelFunc
	closed    bool
}

// NewBaseEventPublisher creates a new BaseEventPublisher.
func NewBaseEventPublisher() *BaseEventPublisher {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseEventPublisher{
		eventChan: make(chan Event, 10), // Buffered channel with capacity 10
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

	select {
	case p.eventChan <- event:
		return nil
	case <-p.ctx.Done():
		return ErrPublisherClosed
	}
}

// AddListener adds an event listener to the publisher.
func (p *BaseEventPublisher) AddListener(listener EventListener) {
	p.listeners = append(p.listeners, listener)
}

// Shutdown stops the event publisher and waits for it to finish.
func (p *BaseEventPublisher) Shutdown(ctx context.Context) error {
	if p.closed {
		return ErrPublisherClosed
	}
	p.cancel()
	close(p.eventChan)
	p.closed = true

	// Wait for the run goroutine to finish
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return nil
	}
}

func (p *BaseEventPublisher) run() {
	for {
		select {
		case event, ok := <-p.eventChan:
			if !ok {
				return // Channel closed, exit goroutine
			}
			for _, listener := range p.listeners {
				listener.HandleEvent(p.ctx, event)
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// ErrPublisherClosed is returned when trying to publish an event on a closed publisher.
var ErrPublisherClosed = errors.New("event publisher is closed")
