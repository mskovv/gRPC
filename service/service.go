package service

import (
	"coursera/hw7_microservice/gen"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type AdminService struct {
	gen.UnimplementedAdminServer
}

func (s AdminService) Logging(*gen.Nothing, gen.Admin_LoggingServer) error {
	return nil
}

func (s AdminService) Statistics(*gen.StatInterval, gen.Admin_StatisticsServer) error {
	return nil
}

func (s AdminService) mustEmbedUnimplementedAdminServer() {
	return
}

// AdminHandler тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, listenAddr string, aclData string) error {
	//_, err := parseACL(aclData)
	//if err != nil {
	//	return err
	//}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterAdminServer(srv, &AdminService{})
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}

	return nil
}

func parseACL(aclData string) (string, error) {
	return aclData, nil
}
