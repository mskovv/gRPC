package grpc_service

import (
	"context"
	pb "coursera/hw7_microservice/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BizServer struct {
	pb.UnimplementedBizServer
	Logger  EventLogger
	Counter StatsCounter
	Acl     ACLChecker
}

func (s *BizServer) processMethod(ctx context.Context, method string) (*pb.Nothing, error) {
	consumer := GetConsumerFromContext(ctx)

	if !s.Acl.CheckAccess(consumer, method) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	s.Logger.LogEvent(consumer, method, "127.0.0.1:8082")
	s.Counter.UpdateStatistics(consumer, method)

	return &pb.Nothing{}, nil
}

func (s *BizServer) Check(ctx context.Context, in *pb.Nothing) (*pb.Nothing, error) {
	return s.processMethod(ctx, "/main.Biz/Check")
}

func (s *BizServer) Add(ctx context.Context, in *pb.Nothing) (*pb.Nothing, error) {
	return s.processMethod(ctx, "/main.Biz/Add")
}

func (s *BizServer) Test(ctx context.Context, in *pb.Nothing) (*pb.Nothing, error) {
	return s.processMethod(ctx, "/main.Biz/Test")
}

func GetConsumerFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown"
	}
	consumers := md.Get("consumer")
	if len(consumers) > 0 {
		return consumers[0]
	}
	return "unknown"
}
