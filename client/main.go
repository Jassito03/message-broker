package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	proto "github.com/Jassito03/message-broker/proto"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "26.103.63.45:5501", "The server address in the format of host:port")
	//serverAddr = flag.String("server_addr", "localhost:5501", "The server address in the format of host:port")
)

func subscribeToTopic(service proto.ForumServiceClient, name string, topic proto.Topics) {
	subscribeClient, err := service.SubscribeToTopic(context.Background(), &proto.SubscribeRequest{
		Topic: topic,
		Client: &proto.Client{
			Id: name,
		},
	})
	if err != nil {
		log.Printf("Failed to subscribe to topic: %v", err)
		return
	}

	for {
		message, err := subscribeClient.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}
		log.Printf("Received message: %s of topic %s", message.Content,message.Topic.String())
	}
}

func unSubscribeToTopic(service proto.ForumServiceClient, name string, topic proto.Topics) {
	_, err := service.UnsubscribeFromTopic(context.Background(), &proto.UnsubscribeRequest{
		Topic: topic,
		Client: &proto.Client{
			Id: name,
		},
	})
	if err != nil {
		log.Printf("Failed to subscribe to topic: %v", err)
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
	publishResponse, err := service.PublishMessage(context.Background(), &proto.PublishRequest{
		Topic:   topic,
		Message: message,
	})
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return
	}
	log.Printf("Publish Response: %v", publishResponse)
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear") // Comando para Linux y MacOS
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls") // Comando para Windows
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	flag.Parse()
	setupLogger()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	log.Printf("Client connect to: %v", *serverAddr)

	service := proto.NewForumServiceClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	var name string
	fmt.Println("¿Cúal es su nombre?")
	scanner.Scan()
	name = scanner.Text()
	fmt.Printf("¡Bienvenido, %s! \n", name)
	for {
		fmt.Println("Selecciona una opción: \n 1) Suscribirse a un tema \n 2) Publicar un mensaje \n 3) Desuscribirse de un tema \n 4) Salir")
		var choice int
		fmt.Scan(&choice)
		scanner.Scan()
		switch choice {
		case 1:
			clearScreen()
			var topics int
			fmt.Println("¿A qué temas te gustaría suscribirte?")
			fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
			fmt.Println("Digita el número del tema al que quieres suscribirte")
			fmt.Scan(&topics)
			scanner.Scan()
			switch topics {
			case 1:
				go subscribeToTopic(service, name, proto.Topics_Tecnologia)
			case 2:
				go subscribeToTopic(service, name, proto.Topics_Entretenimiento)
			case 3:
				go subscribeToTopic(service, name, proto.Topics_Cocina)
			default:
				fmt.Println("Error: Lo que ingresaste no es correcto")
			}
		case 2:
			clearScreen()
			var topic int
			var message string
			fmt.Println("¿Qué gustas compartir hoy?")
			scanner.Scan()
			message = scanner.Text()

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
			}
		case 3:
			clearScreen()
			var topics int
			fmt.Println("¿A qué temas te gustaría desuscribirte?")
			fmt.Println(" 1) Tecnología \n 2) Entretenimiento \n 3) Cocina")
			fmt.Println("Digita el número del tema al que quieres desuscribirte")
			fmt.Scan(&topics)
			scanner.Scan()

			switch topics {
			case 1:
				go unSubscribeToTopic(service, name, proto.Topics_Tecnologia)
			case 2:
				go unSubscribeToTopic(service, name, proto.Topics_Entretenimiento)
			case 3:
				go unSubscribeToTopic(service, name, proto.Topics_Cocina)
			default:
				fmt.Println("Error: Lo que ingresaste no es correcto")
			}
		case 4:
			clearScreen()
			fmt.Println("Saliendo...")
			os.Exit(0)
		default:
			clearScreen()
			fmt.Println("Opción no válida, por favor intenta de nuevo.")
		}
		clearScreen()
	}
}
