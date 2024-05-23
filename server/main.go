package main

import (
	"context"
	"flag"
	"fmt"

	//"io"
	"log"
	"net"

	proto "github.com/Jassito03/message-broker/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 5501, "The server port")
)

type service struct {
	proto.UnimplementedForumServiceServer
}

func (service *service) PublishMessage (ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error){
	log.Printf("Publish message of Topic: %s", req.Topic);
	return nil, nil
}

func main() {

	flag.Parse()
	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	
	log.Printf("Server listening at %v", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}