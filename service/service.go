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

// AdminHandler тут вы пишете код
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

	adminServer := &grpc2.AdminServer{
		Logger:  logger,
		Counter: counter,
		Acl:     acl,
	}

	bizServer := &grpc2.BizServer{
		Logger:  logger,
		Counter: counter,
		Acl:     acl,
	}

	reflection.Register(srv)
	pb.RegisterBizServer(srv, bizServer)
	pb.RegisterAdminServer(srv, adminServer)

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
