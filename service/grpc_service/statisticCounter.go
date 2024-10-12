package grpc_service

import (
	"sync"
)

type SimpleStatsCounter struct {
	statsLock     sync.Mutex
	methodCount   map[string]uint64
	consumerCount map[string]uint64
}

func NewSimpleStatsCounter() *SimpleStatsCounter {
	return &SimpleStatsCounter{
		methodCount:   make(map[string]uint64),
		consumerCount: make(map[string]uint64),
	}
}

func (s *SimpleStatsCounter) AddInStatistics(consumer, method string) {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	s.methodCount[method]++
	s.consumerCount[consumer]++
}

func (s *SimpleStatsCounter) GetStats() *SimpleStatsCounter {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	return s
}

func (s *SimpleStatsCounter) ClearStat() {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	clear(s.methodCount)
	clear(s.consumerCount)
}
