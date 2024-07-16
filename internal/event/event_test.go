package event

import (
	"context"
	"testing"
	"time"
)

type mockListener struct {
	receivedEvents []Event
}

func (ml *mockListener) HandleEvent(ctx context.Context, event Event) error {
	ml.receivedEvents = append(ml.receivedEvents, event)
	return nil
}

func TestEventPublisher(t *testing.T) {
	publisher := NewBaseEventPublisher()
	go publisher.run()
	defer publisher.Shutdown(context.Background())

	listener := &mockListener{}
	publisher.AddListener(listener)

	event := "test event"
	err := publisher.PublishEvent(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	found := false
	for _, e := range listener.receivedEvents {
		if e == event {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected event %v to be received", event)
	}
}

// func TestEventPublisherFullChannel(t *testing.T) {
// 	// Create a new BaseEventPublisher with a buffered channel (capacity 1)
// 	publisher := &BaseEventPublisher{
// 		eventChan: make(chan Event, 1), // Buffered channel with capacity 1
// 		listeners: make([]EventListener, 0),
// 		ctx:       context.Background(),
// 		cancel:    func() {}, // No-op cancel function
// 	}
// 	go publisher.run()
// 	defer publisher.Shutdown(context.Background())

// 	listener := &mockListener{}
// 	publisher.AddListener(listener)

// 	// Fill the channel completely
// 	err := publisher.PublishEvent("test event")
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	// Try to publish another event, which should block
// 	done := make(chan bool)
// 	go func() {
// 		err := publisher.PublishEvent("another event")
// 		if err != nil {
// 			t.Errorf("unexpected error when publishing to a full channel: %v", err)
// 		}
// 		done <- true
// 	}()

// 	// Wait for a short time to see if the publish blocks
// 	select {
// 	case <-done:
// 		t.Errorf("expected publish to block, but it completed")
// 	case <-time.After(100 * time.Millisecond):
// 		// Publish is blocking, as expected
// 	}
// }

func TestEventPublisherShutdown(t *testing.T) {
	publisher := NewBaseEventPublisher()
	go publisher.run()

	// Publish an event before shutdown
	event := "test event before shutdown"
	err := publisher.PublishEvent(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	publisher.Shutdown(context.Background())

	// Try to publish an event after shutdown, which should fail
	err = publisher.PublishEvent("test event after shutdown")
	if err == nil {
		t.Errorf("expected error when publishing after shutdown")
	}
}
