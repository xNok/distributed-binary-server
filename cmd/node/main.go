package main

// 1. Server start listing to a port :3030
// 2. Server recieve a message and propagate the message to child server

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)

type Server struct {
	port string

	rightPort *string
	leftPort  *string
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Connection closed:", err)
			return
		}
		log.Printf("Received message: %s", msg)

		if s.rightPort != nil {
			if err := dialChildren(msg, *s.rightPort); err != nil {
				log.Printf("Failed to dial right child: %v", err)
			}
		}

		if s.leftPort != nil {
			if err := dialChildren(msg, *s.leftPort); err != nil {
				log.Printf("Failed to dial left child: %v", err)
			}
		}
	}
}

func dialChildren(msg string, port string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		return err
	}

	log.Println("Message sent!")
	return nil
}

func main() {
	first := flag.String("port", "", "Server port argument")
	second := flag.String("right", "", "Right child server port string argument")
	third := flag.String("left", "", "Left child server port string argument")

	flag.Parse()

	server := Server{
		port:      *first,
		rightPort: second,
		leftPort:  third,
	}

	fmt.Printf("Parsed args: %+v\n", server)

	listener, err := net.Listen("tcp", ":"+server.port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Listening on port %s", server.port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go server.handleConnection(conn)
	}
}
