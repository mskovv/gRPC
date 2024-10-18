package grpc_service

import (
	"context"
	pb "coursera/hw7_microservice/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type EventLogger interface {
	LogEvent(event *pb.Event)
	Subscribe() chan *pb.Event
	Unsubscribe(chan *pb.Event)
}

type ACLChecker interface {
	CheckAccess(consumer, method string) bool
}

type AdminServer struct {
	pb.UnimplementedAdminServer
	Logger EventLogger
	Acl    ACLChecker
	Ctx    context.Context

	BroadcastStatCh   chan *pb.Stat
	AddStatListenerCh chan chan *pb.Stat
	broadcastLogCh    chan *pb.Event
	StatListeners     []chan *pb.Stat
}

func (s *AdminServer) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	consumer := GetConsumerFromContext(ctx)

	s.BroadcastStatCh <- &pb.Stat{
		ByConsumer: map[string]uint64{consumer: 1},
		ByMethod:   map[string]uint64{info.FullMethod: 1},
	}
	return handler(ctx, req)
}

func (s *AdminServer) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	consumer := GetConsumerFromContext(ss.Context())

	s.BroadcastStatCh <- &pb.Stat{
		ByConsumer: map[string]uint64{consumer: 1},
		ByMethod:   map[string]uint64{info.FullMethod: 1},
	}
	return handler(srv, ss)
}

func (s *AdminServer) Logging(_ *pb.Nothing, stream pb.Admin_LoggingServer) error {
	method := "/main.Admin/Logging"
	consumer := GetConsumerFromContext(stream.Context())

	if !s.Acl.CheckAccess(consumer, method) {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}
	s.Logger.LogEvent(
		&pb.Event{
			Consumer: consumer,
			Method:   method,
			Host:     "127.0.0.1:8082",
		})

	logCh := s.Logger.Subscribe()
	defer s.Logger.Unsubscribe(logCh)

	for {
		select {
		case event := <-logCh:
			if err := stream.Send(event); err != nil {
				return err
			}
		case <-s.Ctx.Done():
			s.Logger.Unsubscribe(logCh)
			return stream.Context().Err()
		}
	}
}

func (s *AdminServer) Statistics(interval *pb.StatInterval, stream pb.Admin_StatisticsServer) error {
	method := "/main.Admin/Statistics"
	var consumer string
	consumer = GetConsumerFromContext(stream.Context())

	if !s.Acl.CheckAccess(consumer, method) {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	ch := make(chan *pb.Stat)
	s.AddStatListenerCh <- ch

	ticker := time.NewTicker(time.Second * time.Duration(interval.IntervalSeconds))
	sum := &pb.Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
	for {
		select {
		case stat := <-ch:
			for k, v := range stat.ByMethod {
				sum.ByMethod[k] += v
			}
			for k, v := range stat.ByConsumer {
				sum.ByConsumer[k] += v
			}

		case <-ticker.C:
			err := stream.Send(sum)
			if err != nil {
				return err
			}
			sum = &pb.Stat{
				ByMethod:   make(map[string]uint64),
				ByConsumer: make(map[string]uint64),
			}

		case <-s.Ctx.Done():
			return nil
		}
	}
}

func (s *AdminServer) mustEmbedUnimplementedAdminServer() {
	return
}
