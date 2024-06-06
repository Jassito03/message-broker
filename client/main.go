package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	proto "github.com/Jassito03/message-broker/proto"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "26.103.63.45:5501", "The server address in the format of host:port")
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

func setupLogger() {
	logFile, err := os.OpenFile("Clientlogs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
}

func sendMessage(service proto.ForumServiceClient, message string, topic proto.Topics) {
	_, err := service.PublishMessage(context.Background(), &proto.PublishRequest{
		Topic:   topic,
		Message: message,
	})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	//log.Printf("Publish Response: %v", publishResponse)
}

func main() {
	flag.Parse()
	setupLogger()
	var wg sync.WaitGroup

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	service := proto.NewForumServiceClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	var name string
	fmt.Println("¿Cúal es su nombre?")
	scanner.Scan()
	name = scanner.Text()
	fmt.Printf("¡Bienvenido, %s! \n", name)

	var topics [3]int
	for i := 0; i < 3; i++ {
		var aux string
		fmt.Println("¿A qué temas te gustaría suscribirte?")
		fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
		fmt.Println("Digita el número del tema al que quieres suscribirte")
		fmt.Scan(&topics[i])
		scanner.Scan()

		fmt.Println("¿Quieres suscribirte a otro tema? s/n")
		fmt.Scan(&aux)
		scanner.Scan()

		switch topics[i] {
		case 1:
			wg.Add(1)
			go func() {
				defer wg.Done()
				subscribeToTopic(service, name, proto.Topics_Tecnologia)
			}()
		case 2:
			wg.Add(1)
			go func() {
				defer wg.Done()
				subscribeToTopic(service, name, proto.Topics_Entretenimiento)
			}()
		case 3:
			wg.Add(1)
			go func() {
				defer wg.Done()
				subscribeToTopic(service, name, proto.Topics_Cocina)
			}()
		default:
			fmt.Println("Error: Lo que ingresaste no es correcto")
			i--
		}
		if aux == "n" {
			break
		}
	}

	/*var message string
	fmt.Println("¿Qué gustas compartir el día de hoy?")
	scanner.Scan()
	message = scanner.Text()

	var topic int
	fmt.Println("¿A qué tema te gustaría publicarlo?")
	fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
	fmt.Println("Digita el número del tema al que quieres enviarlo")
	fmt.Scan(&topic)
	scanner.Scan()
	switch topic {
	case 1:
		go sendMessage(service, message, proto.Topics_Tecnologia)
	case 2:
		go sendMessage(service, message, proto.Topics_Entretenimiento)
	case 3:
		go sendMessage(service, message, proto.Topics_Cocina)
	default:
		fmt.Println("Error: Lo que ingresaste no es correcto")
	}*/

	go func() {
		var topic int
		for {
			fmt.Println("¿Qué gustas compartir hoy? (Escribe 'exit' para salir)")
			scanner.Scan()
			message := scanner.Text()
			if message == "exit" {
				break
			}
			fmt.Println("¿A qué tema te gustaría publicarlo?")
			fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
			fmt.Println("Digita el número del tema al que quieres enviarlo")

			fmt.Scan(&topic)
			scanner.Scan() // Limpiar el buffer

			switch topic {
			case 1:
				go sendMessage(service, message, proto.Topics_Tecnologia)
			case 2:
				go sendMessage(service, message, proto.Topics_Entretenimiento)
			case 3:
				go sendMessage(service, message, proto.Topics_Cocina)
			default:
				fmt.Println("Error: Lo que ingresaste no es correcto")
			}
		}
		wg.Wait()
		os.Exit(0)
	}()

	wg.Wait()
}
