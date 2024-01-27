package main

import (
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
)

var PORT string = "5000"

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Printf("Failed to load .env file :%s", err)
	} else {
		PORT = os.Getenv("PORT")
	}
}

type MessageType int

const (
	NewClient MessageType = iota
	NewMessage
	ClientLeft
)

type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}


// serving thread
func server(messages chan Message) {
	conns := map[string]net.Conn{}

	for {
		msg := <-messages

		switch msg.Type {
		case NewClient:
			log.Printf("Client [%s] joined\n", msg.Conn.RemoteAddr())
			for _, conn := range conns {
				_, err := conn.Write([]byte(msg.Conn.RemoteAddr().String() + " joined the chatroom!\n"))
				if err != nil {
					log.Printf("Failed to notify users: %s", err)
				}

			}
			conns[msg.Conn.RemoteAddr().String()] = msg.Conn
			msg.Conn.Write([]byte("Welcome to chatroom!\n"))
		case NewMessage:
			log.Printf("Client [%s] says: %s", msg.Conn.RemoteAddr(), msg.Text)

			// send to all clients!
			for _, conn := range conns {
				if conn.RemoteAddr() != msg.Conn.RemoteAddr() {
					_, err := conn.Write([]byte(msg.Text))
					if err != nil {
						log.Printf("Couldn't send data to %s: %s", conn.RemoteAddr(), err)
					}
				}
			}
		case ClientLeft:
			delete(conns, msg.Conn.RemoteAddr().String())
			msg.Conn.Close()
			log.Printf("Client [%s] disconnected", msg.Conn.RemoteAddr())
			for _, conn := range conns {
				_, err := conn.Write([]byte(msg.Conn.RemoteAddr().String() + " disconnected from the chatroom!\n"))
				if err != nil {
					log.Printf("Failed to notify users: %s", err)
				}

			}
		}
	}
}

// client thread
func client(conn net.Conn, messages chan Message) {

	log.Printf("New go-routine spawned for %s\n", conn.RemoteAddr())
	buff := make([]byte, 512)

	for {
		n, err := conn.Read(buff)
		if err != nil {
			messages <- Message{
				Type: ClientLeft,
				Conn: conn,
			}
      break
		} else {

			message := string(buff[:n])
			messages <- Message{
				Type: NewMessage,
				Text: message,
				Conn: conn,
			}
		}

	}

	log.Printf("go-routine stopped for %s\n", conn.RemoteAddr())
}

func main() {
	ln, err := net.Listen("tcp", ":"+PORT)

	if err != nil {
		//handle err
		log.Fatalf("ERROR: could not listen to port %s: %s\n", PORT, err)
	}

	log.Printf("LISTENING to TCP connection on port %s ...\n", PORT)
	messages := make(chan Message)

	// basically mpsc
	// single consumer
	go server(messages)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle err
			log.Printf("ERROR: couldnt accept a connection: %s\n", err)

		}

		// add to channel
		messages <- Message{
			Type: NewClient,
			Conn: conn,
		}

		// multiple producers
		go client(conn, messages)
	}
}
