package grpc_service

import (
	pb "coursera/hw7_microservice/gen"
	"sync"
)

type SimpleEventLogger struct {
	mu          sync.Mutex
	subscribers map[chan *pb.Event]struct{}
}

func NewSimpleEventLogger() *SimpleEventLogger {
	return &SimpleEventLogger{
		subscribers: make(map[chan *pb.Event]struct{}),
	}
}

func (l *SimpleEventLogger) LogEvent(event *pb.Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for ch := range l.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (l *SimpleEventLogger) Subscribe() chan *pb.Event {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch := make(chan *pb.Event, 100)
	l.subscribers[ch] = struct{}{}
	return ch
}

func (l *SimpleEventLogger) Unsubscribe(ch chan *pb.Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.subscribers[ch]; ok {
		delete(l.subscribers, ch)
		close(ch)
	}
}
