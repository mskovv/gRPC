package grpc_service

import (
	pb "coursera/hw7_microservice/gen"
	"fmt"
	"time"
)

type EventLogger interface {
	LogEvent(consumer, method, host string)
}

type SimpleEventLogger struct {
	logCh chan *pb.Event
}

func NewSimpleEventLogger() *SimpleEventLogger {
	return &SimpleEventLogger{
		logCh: make(chan *pb.Event, 100),
	}
}

func (l *SimpleEventLogger) LogEvent(consumer, method, host string) {
	event := &pb.Event{
		Timestamp: time.Now().Unix(),
		Consumer:  consumer,
		Method:    method,
		Host:      host,
	}
	l.logCh <- event
	fmt.Printf("logEvent: {consumer: %s, method: %s }\n", consumer, method)
}

func (l *SimpleEventLogger) GetLogChannel() <-chan *pb.Event {
	return l.logCh
}
