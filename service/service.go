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

	logger := grpc2.NewSimpleEventLogger()

	adminServer := &grpc2.AdminServer{
		Logger: logger,
		Acl:    acl,
		Ctx:    ctx,
	}

	bizServer := &grpc2.BizServer{
		Logger: logger,
		Acl:    acl,
	}

	srv := grpc.NewServer(
		grpc.StreamInterceptor(adminServer.StreamInterceptor),
		grpc.UnaryInterceptor(adminServer.UnaryInterceptor),
	)

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

	adminServer.BroadcastStatCh = make(chan *pb.Stat)
	adminServer.AddStatListenerCh = make(chan chan *pb.Stat)
	go func() {
		for {
			select {
			case ch := <-adminServer.AddStatListenerCh:
				adminServer.StatListeners = append(adminServer.StatListeners, ch)
			case stat := <-adminServer.BroadcastStatCh:
				for _, ch := range adminServer.StatListeners {
					ch <- stat
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
