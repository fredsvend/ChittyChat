package main

import (
	pb "example/service"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChittyChatServer struct {
	pb.UnimplementedChittyChatServer
	users    map[string]pb.ChittyChat_BroadcastServer
	clock    map[string]*pb.Clock
	messages map[string]chan *pb.UserMessage

	serverClock *pb.Clock
}

var (
	userCount = 0
	lock sync.Mutex
)

const (
	port = "localhost:8080"
)

func main() {

	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("This is the start of the ShittyChat log")

	lis, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	fmt.Println("Starting a ShittyChat server")
	s := grpc.NewServer()

	s1 := ChittyChatServer{
		clock: make(map[string](*pb.Clock)),
		messages:    make(map[string](chan *pb.UserMessage)),
		users:       make(map[string]pb.ChittyChat_BroadcastServer),
		serverClock: pb.NewClock()}

	var i,j int
    for i = 0; i < 10; i++ {
        for j = 0; j < 10; j++ {
            fmt.Print("* ")
        }
        fmt.Println()
    }
	fmt.Println("ShittyChat server has started successfully :----)")
	
	pb.RegisterChittyChatServer(s, &s1)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("ShittyChat server has not started successfully :( %v", err)

	}
	

}
func (server *ChittyChatServer) Broadcast(_ *emptypb.Empty, stream pb.ChittyChat_BroadcastServer) error {

	username := uuid.Must(uuid.NewRandom()).String()[0:4]

	server.messages[username] = make(chan *pb.UserMessage, 10)

	server.users[username] = stream
	server.clock[username] = pb.NewClock()
	server.UserJoinMessage("", username, *server.clock[username])
	
	lock.Lock()
	server.serverClock.Increment()
	lock.Unlock()

	for {
		message := <-server.messages[username]
		maxClock := computeMax(message.GetClock(), server.serverClock.Counter)
		message.Clock = maxClock
		lock.Lock()
		server.serverClock.Increment()
		lock.Unlock()
		if message.GetMessage() == "" {
			log.Printf("BROADCASTING: User %s just joined! Current clock: serverClock: [%d]", message.GetUsername(), server.serverClock.Counter)
		} else {
			log.Printf("BROADCASTING: User %s with clock: [%d] and just wrote %s", message.GetUsername(), message.GetClock(), message.GetMessage())
			log.Printf("serverClock: [%d]", server.serverClock.Counter)
		}
	
		err := server.users[username].Send(message)
		
		if err != nil {
			break
		}
	}

	return nil
}

func (server *ChittyChatServer) Publish(stream pb.ChittyChat_PublishServer) error {

	for {
		message, err := stream.Recv()

		if err != nil {
			return err
		}

		if userCount <= len(server.messages) {
			log.Printf("%s just joined", message.GetUsername())
			userCount++
		}

		if message.GetMessage() == "" {
			
		} else {
			log.Printf("PUBLISHING: User %s with clock: [%d] just wrote %s", message.GetUsername(), message.GetClock(), message.GetMessage())
		}

		server.serverClock.Counter = computeMax(server.serverClock.Counter, message.GetClock())
		lock.Lock()
		server.serverClock.Increment()
		lock.Unlock()

		for _, user := range server.messages {
			user <- message
		}
	}

	return nil
}


func (server *ChittyChatServer) UserJoinMessage(emptyString string, username string, clock pb.Clock) {
	toSend := pb.UserMessage{Message: emptyString, Username: username, Clock: uint64(clock.Time())}

	for _, user := range server.messages {
		user <- &toSend
	}
}

func computeMax(x uint64, y uint64) uint64 {
	if x > y {
		return x
	} else {
		return y
	}
}
