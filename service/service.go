package service

import (
	pb "coursera/hw7_microservice/gen"
	grpc2 "coursera/hw7_microservice/service/grpc_service"
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

// StartMyMicroservice AdminHandler тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, listenAddr string, aclData string) error {
	aclMap := grpc2.ACLData{}
	err := json.Unmarshal([]byte(aclData), &aclMap)
	if err != nil {
		//log.Fatalf("failed to Unmarshal: %v", err)
		return err
	}

	acl := grpc2.NewSimpleACLChecker(aclMap)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		//log.Fatalf("failed to listen: %v", err)
		return err
	}

	srv := grpc.NewServer()

	logger := grpc2.NewSimpleEventLogger()
	counter := grpc2.NewSimpleStatsCounter()
	notifier := grpc2.NewStatsNotifier()

	adminServer := &grpc2.AdminServer{
		Logger:        logger,
		Counter:       counter,
		StatsNotifier: notifier,
		Acl:           acl,
		Ctx:           ctx,
	}

	bizServer := &grpc2.BizServer{
		Logger:  logger,
		Counter: counter,
		Acl:     acl,
	}

	reflection.Register(srv)
	pb.RegisterAdminServer(srv, adminServer)
	pb.RegisterBizServer(srv, bizServer)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
		lis.Close()
	}()

	return nil
}
