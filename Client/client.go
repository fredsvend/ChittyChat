package main

import (
	"bufio"
	"context"
	"example/service"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	gg "github.com/thecodeteam/goodbye"
)

var (
	username string
	clock    uint64
)

func main() {
	
	
	port := "localhost:8080"

	fmt.Println("Trying to connect to ShittyChat")
	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect %v", err)
	}
	defer conn.Close()

	client := service.NewChittyChatClient(conn)
	context := context.Background()
	
	defer gg.Exit(context, -1)

	out, err := client.Publish(context)
	if err != nil {
		log.Fatal("Failed to open up for publishing chat messages")
	}

	in, err := client.Broadcast(context, &emptypb.Empty{})
	if err != nil {
		log.Fatal("Failed to receive chat message broadcast")
	}
	fmt.Println("<<<<<Hello there, and welcome to ShittyChat!>>>>>")
	fmt.Println("<<<Enter username!>>>")
	reader := bufio.NewReader(os.Stdin)
	username, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal("DIDNT READ")
	}
	username = strings.TrimRight(username, "\r\n")
	fmt.Println("<<<Now you are ready to chat! Simply type a message and press Enter>>>")
	clock = 0
	
	channel := make(chan (*service.UserMessage), 1000)
	stream := make(chan (string))
	go messageReceiver(in, channel)
	go messageSender(out, stream)

	bl := make(chan bool)
	<-bl
}

func messageReceiver(stream service.ChittyChat_BroadcastClient, channel chan<- *service.UserMessage) {
	for {
		msg, err := stream.Recv()
		
		if err != nil {
			log.Fatal("Failed to receive message")
		}
		if msg.Message == "" {
			log.Printf("User %s just joined! Current clock: [%d]", msg.GetUsername(), msg.GetClock())
		}else{
		log.Printf("%s has sent message %s", msg.GetUsername(), msg.GetMessage())
		}

		clock = computeMax(clock, (msg.GetClock()))
		clock++
		channel <- msg
	}
}

func messageSender(stream service.ChittyChat_PublishClient, mstream <-chan string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if(len(message)>128){
			fmt.Println("<<<Your message has to be contained in less than 129 letters!>>>")
		
		}else{
		if err != nil {
			log.Fatal("DIDNT READ")
		}
		message = strings.TrimSuffix(message, "\n")

		clock++
		userMessage := service.UserMessage{Username: username, Message: message, Clock: clock}
		stream.Send(&userMessage)
	}
	}
}

func computeMax(x uint64, y uint64) uint64 {
	if (x > y) {
		return x
	} else {
		return y
	}
}
