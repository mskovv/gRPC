package grpc_service

import (
	"context"
	pb "coursera/hw7_microservice/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type AdminServer struct {
	pb.UnimplementedAdminServer
	Logger  EventLogger
	Counter StatsCounter
	Acl     ACLChecker
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
	defer func() {
		s.Logger.Unsubscribe(logCh)
	}()

	ctx, cancel := context.WithTimeout(stream.Context(), 2*time.Second)
	defer cancel()

	for {
		select {
		case event := <-logCh:
			if err := stream.Send(event); err != nil {
				return err
			}
		case <-ctx.Done():
			s.Logger.Unsubscribe(logCh)
			return stream.Context().Err()
		}
	}
}

func (s *AdminServer) Statistics(interval *pb.StatInterval, stream pb.Admin_StatisticsServer) error {
	consumer := GetConsumerFromContext(stream.Context())

	if !s.Acl.CheckAccess(consumer, "/main.Admin/Statistics") {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		methodStats, consumerStats := s.Counter.GetStats()

		stats := &pb.Stat{
			Timestamp:  time.Now().Unix(),
			ByMethod:   methodStats,
			ByConsumer: consumerStats,
		}

		if err := stream.Send(stats); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminServer) mustEmbedUnimplementedAdminServer() {
	return
}
