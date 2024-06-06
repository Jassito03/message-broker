package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	proto "github.com/Jassito03/message-broker/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 5501, "The server port")
	//ip   = flag.String("ip", "26.103.63.45", "The IP address to bind the server to")
	ip   = flag.String("ip", "localhost", "The IP address to bind the server to")
)

type service struct {
	proto.UnimplementedForumServiceServer
	topics map[proto.Topics]map[string]*clientStream
	mutex  sync.RWMutex
}

type clientStream struct {
	clientID string
	stream   proto.ForumService_SubscribeToTopicServer
}

func setupLogger() {
	logFile, err := os.OpenFile("Serverlogs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
}

func (service *service) PublishMessage(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	log.Printf("Publish message %s of Topic: %s", req.GetMessage(), req.Topic.String())

	service.mutex.RLock()
	defer service.mutex.RUnlock()

	clients, ok := service.topics[req.Topic]
    if !ok || len(clients) == 0 {
				log.Print("Topic not found or no subscribers")
        return &proto.PublishResponse{Success: false}, fmt.Errorf("topic not found or no subscribers")
    }

	var wg sync.WaitGroup // Crea el grupo
	for _, client := range clients {
		wg.Add(1)
		go func(cs *clientStream) {
			defer wg.Done()
			err := cs.stream.Send(&proto.Message{
				Topic:   req.Topic,
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

func (service *service) isClientSubscribed(clientID string, topic proto.Topics) bool {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	if clients, exists := service.topics[topic]; exists {
		if _, subscribed := clients[clientID]; subscribed {
				return true 
		}
	}
	return false
}

func (service *service) SubscribeToTopic(req *proto.SubscribeRequest, stream proto.ForumService_SubscribeToTopicServer) error {
	log.Printf("Subscribe to topic: %s by client: %s", req.Topic.String(), req.Client.Id)

	if service.isClientSubscribed(req.Client.Id, req.Topic) {
		log.Printf("The client %s is already subscribed to %s", req.Topic.String(), req.Client.Id)
		return nil
	}
	
	cs := &clientStream{
		clientID: req.Client.Id,
		stream:   stream,
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
	log.Printf("Client %s unsubscribed from topic: %s due to disconnection", req.Client.Id, req.Topic.String())
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

func (service *service) cleanEmptySubscriptions(topic proto.Topics) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if clients, exists := service.topics[topic]; exists {
			if len(clients) == 0 {
					delete(service.topics, topic)
					log.Printf("Removed empty subscription for topic: %s", topic.String())
			}
	}
}

func periodicCleanup(service *service) {
	for range time.Tick(1 * time.Minute) {
			for topic := range service.topics {
					service.cleanEmptySubscriptions(topic)
			}
	}
}


func main() {
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ip, *port))
	if err != nil {
		log.Fatalf("Failed to listen on %s:%d: %v", *ip, *port, err)
	}

	service := &service{
		topics: make(map[proto.Topics]map[string]*clientStream),
	}

	setupLogger()
	server := grpc.NewServer()
	proto.RegisterForumServiceServer(server, service)

	log.Printf("Server listening at %v", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	periodicCleanup(service)
}
