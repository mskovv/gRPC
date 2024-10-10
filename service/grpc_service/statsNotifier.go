package grpc_service

import (
	pb "coursera/hw7_microservice/gen"
	"sync"
)

type SimpleStatsNotify struct {
	subscribers map[chan *pb.Stat]struct{}
	mu          sync.Mutex
}

func NewStatsNotifier() *SimpleStatsNotify {
	return &SimpleStatsNotify{
		subscribers: make(map[chan *pb.Stat]struct{}),
	}
}

func (s *SimpleStatsNotify) Subscribe() chan *pb.Stat {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *pb.Stat, 100)
	s.subscribers[ch] = struct{}{}
	return ch
}

func (s *SimpleStatsNotify) Unsubscribe(ch chan *pb.Stat) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[ch]; ok {
		delete(s.subscribers, ch)
		close(ch)
	}
}

func (s *SimpleStatsNotify) NotifyAll(stats *pb.Stat) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ch := range s.subscribers {
		select {
		case ch <- stats:
		default:
		}
	}
}
