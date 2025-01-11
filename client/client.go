package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

type ConnectionReader interface {
	ReadJSON(v interface{}) error
}

type ConnectionWriter interface {
	WriteJSON(v interface{}) error
	Close() error
}

type Scanner interface {
	Scan() bool
	Text() string
}

type Message struct {
	Username string `json:"username"`
	Text     string `json:"Text"`
}

func main() {
	var uname string           // declare username given by user
	var scanner *bufio.Scanner // declare scanner to read user input

	// provider custom address and port CLI
	server := flag.String("server", "localhost:8080", "server network address")
	path := flag.String("path", "/ws", "websocket path")
	flag.Parse()
	serverURL := url.URL{
		Scheme: "ws",
		Host:   *server,
		Path:   *path,
	}

	fmt.Print("Username: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	uname = scanner.Text()

	// app interface
	fmt.Println("Welcome", uname)
	fmt.Printf("Connecting to server @ %s...\n", *server)

	// connecting to the server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		log.Fatal("Connection error, closing connection...", err)
	}
	defer conn.Close()

	// notify new client connected
	cMessage := Message{
		Username: uname,
		Text:     "has joined the chat.",
	}
	conn.WriteJSON(cMessage)

	go handleReceiveMessage(conn)
	handleSendMessage(conn, scanner, uname)
}

func handleReceiveMessage(conn ConnectionReader) {
	for {
		var sMessage Message

		err := conn.ReadJSON(&sMessage)
		if err != nil {
			fmt.Println("Server closed, exiting...")
			os.Exit(0)
		}

		fmt.Printf("[%s] %s\n", sMessage.Username, sMessage.Text)
	}
}

func handleSendMessage(conn ConnectionWriter, scanner Scanner, uname string) {
	var cMessage Message
	cMessage.Username = uname

	for {
		if scanner.Scan() {
			cMessage.Text = scanner.Text()

			if cMessage.Text == "exit" {
				fmt.Println("You're leaving chat room...")
				conn.WriteJSON(Message{
					Username: uname,
					Text:     "has disconnected.",
				})
				conn.Close()
				os.Exit(0)
			}

			fmt.Printf("[%s] %s\n", cMessage.Username, cMessage.Text)

			err := conn.WriteJSON(cMessage)
			if err != nil {
				log.Fatal("Error sending message, clossing connection...", err)
				break
			}
		}
	}
}
