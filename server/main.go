package main

import (
	"context"
	"flag"
	"fmt"
	"sync"

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
	topics map[proto.Topics]map[string]*clientStream
	mutex sync.RWMutex
}

type clientStream struct {
	clientID string
	stream proto.ForumService_SubscribeToTopicServer
}

func (service *service) PublishMessage (ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error){
	log.Printf("Publish message of Topic: %s", req.Topic.String());
	
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	
	clients, ok := service.topics[req.Topic]
	if !ok {
		return &proto.PublishResponse{Success: false}, fmt.Errorf("topic not found")
	}

	var wg sync.WaitGroup // Crea el grupo
	for _, client := range clients {
		wg.Add(1)
		go func (cs *clientStream) {
			defer wg.Done()
			err := cs.stream.Send(&proto.Message{
				Topic: req.Topic,
				Content: req.Message,
			})
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
			}
		}(client)
	}

	wg.Wait() //Espera a que las rutinas terminen

	return &proto.PublishResponse{Success: true}, nil
}

func (service *service) SubscribeToTopic(req *proto.SubscribeRequest, stream proto.ForumService_SubscribeToTopicServer) error {
	log.Printf("Subscribe to topic: %s by client: %s", req.Topic.String(), req.Client.Id)

	cs := &clientStream{
		clientID: req.Client.Id,
		stream: stream,
	}

	service.mutex.Lock()
  if service.topics[req.Topic] == nil {
		service.topics[req.Topic] = make(map[string]*clientStream)
  }
  service.topics[req.Topic][req.Client.Id] = cs
  service.mutex.Unlock()

  // Mantener la conexión abierta
  <-stream.Context().Done()

  // Cuando el cliente se desconecte, eliminarlo de las suscripciones
  service.mutex.Lock()
  delete(service.topics[req.Topic], req.Client.Id)
	log.Printf("Subscribe to topic: %s by client: %s is eliminated because the client is desconnected", req.Topic.String(), req.Client.Id)
	service.mutex.Unlock()

  return nil
}

func (service *service) UnsubscribeFromTopic(ctx context.Context, req *proto.UnsubscribeRequest) (*proto.UnsubscribeResponse, error) {
	log.Printf("Unsubscribe from topic: %s by client: %s", req.Topic.String(), req.Client.Id)

	service.mutex.Lock()
  defer service.mutex.Unlock()

  clients, ok := service.topics[req.Topic]
  if !ok || len(clients) == 0 {
    return &proto.UnsubscribeResponse{Success: false}, fmt.Errorf("topic not found or no subscribers")
  }

  if _, exists := clients[req.Client.Id]; exists {
  	delete(clients, req.Client.Id)
    return &proto.UnsubscribeResponse{Success: true}, nil
  }

  return &proto.UnsubscribeResponse{Success: false}, fmt.Errorf("client not subscribed to the topic")
}



func main() {

	flag.Parse()
	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	service := &service{
		topics: make(map[proto.Topics]map[string]*clientStream),
	}

	server := grpc.NewServer()
	proto.RegisterForumServiceServer(server, service)
	
	log.Printf("Server listening at %v", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}