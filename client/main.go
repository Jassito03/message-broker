package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	proto "github.com/Jassito03/message-broker/proto"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:5501", "The server address in the format of host:port")
)

func subscribeToTopic(service proto.ForumServiceClient, name string, topic proto.Topics) {
	// Suscribirse a un tema para recibir mensajes
	subscribeClient, err := service.SubscribeToTopic(context.Background(), &proto.SubscribeRequest{
		Topic: topic, 
		Client: &proto.Client{
			Id: name, 
		},
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	for {
		message, err := subscribeClient.Recv()
		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}
		log.Printf("Received message: %s", message.Content)
	}
}

func sendMessage(service proto.ForumServiceClient, message string, topic proto.Topics) {
	publishResponse, err := service.PublishMessage(context.Background(), &proto.PublishRequest{
		Topic:   topic,
		Message: message,
	})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Printf("Publish Response: %v", publishResponse)
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	service := proto.NewForumServiceClient(conn)

	var name string
	fmt.Println("¿Cúal es su nombre?")
	fmt.Scan(&name);
	fmt.Printf("¡Bienvenido, %s!", name)

	var topics [3]int
	for i := 0; i < 3; i++ {
		var aux string
		fmt.Println("¿A qué temas te gustaría suscribirte?")
		fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
		fmt.Println("Digita el número del tema al que quieres suscribirte")
		fmt.Scan(&topics[i])
		switch topics[i] {
			case 1:
				go func() {
					subscribeToTopic(service, name, proto.Topics_Tecnologia)
				}()
			case 2:
				go func() {
					subscribeToTopic(service, name, proto.Topics_Entretenimiento)
				}()
			case 3:
				go func() {
					subscribeToTopic(service, name, proto.Topics_Cocina)
				}()
			default:
				fmt.Println("Error: Lo que ingresaste no es correcto")
				i--;
		}
		fmt.Println("¿Quieres suscribirte a otro tema? s/n")
		fmt.Scan(&aux)
		if aux == "n" {
			break
		}
	}

	var message string
	fmt.Println("¿Qué gustas compartir el día de hoy?")
	fmt.Scan(&message);
		
	var topic int
	fmt.Println("¿A qué tema te gustaría publicarlo?")
	fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
	fmt.Println("Digita el número del tema al que quieres enviarlo")
	fmt.Scan(&topic)
	switch topic {
		case 1:
			go func() {
				sendMessage(service, message, proto.Topics_Tecnologia)
			}()
		case 2:
			go func() {
				sendMessage(service, message, proto.Topics_Entretenimiento)
			}()
		case 3:
			go func() {
				sendMessage(service, message, proto.Topics_Cocina)
			}()
		default:
			fmt.Println("Error: Lo que ingresaste no es correcto")
	}
	
}
