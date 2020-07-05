package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"grng.dev/guff/database"
	"grng.dev/guff/pb"
	"grng.dev/guff/services/post"
	"grng.dev/guff/utils"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("%s", err)
	}
	log.Print("database ok")
	lis, err := net.Listen("tcp", ":"+os.Getenv("SERVE_PORT"))
	if err != nil {
		log.Fatalf("failed to listen on port.  %v", err)
	}
	store := utils.NewGCPStorage()
	postService := post.Service{DB: database.GetDB(), Storage: store}
	grpcServer := grpc.NewServer()
	pb.RegisterPostServiceServer(grpcServer, &postService)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve %s", err)
	}
}
