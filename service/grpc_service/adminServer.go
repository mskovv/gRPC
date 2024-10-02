package grpc_service

import (
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
	consumer := GetConsumerFromContext(stream.Context())

	if !s.Acl.CheckAccess(consumer, "/main.Admin/Logging") {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	logCh := s.Logger.(*SimpleEventLogger).GetLogChannel()

	for event := range logCh {
		if err := stream.Send(event); err != nil {
			return err
		}
	}
	stream.Context().Done()
	return nil
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
