package main

import (
	"coursera/hw7_microservice/gen"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type AdminService struct {
	gen.UnimplementedAdminServer
}

func (s AdminService) Logging(*gen.Nothing, gen.Admin_LoggingServer) error {
	fmt.Println("call logging")
	return nil
}

func (s AdminService) Statistics(*gen.StatInterval, gen.Admin_StatisticsServer) error {
	return nil
}

func (s AdminService) mustEmbedUnimplementedAdminServer() {
	return
}

//func main() {
//	println("usage: make test")
//}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:8083")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}

	srv := grpc.NewServer()

	fmt.Println("Server start at port: " + lis.Addr().String())
	defer srv.GracefulStop()

	reflection.Register(srv)
	gen.RegisterAdminServer(srv, &AdminService{})
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return
	}

	return
}
