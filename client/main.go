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
	Connection *websocket.Conn `json:"connection"`
	Username   string          `json:"username"`
	Text       string          `json:"text"`
	Type       string          `json:"type"`
}

func main() {
	var username string        // declare username given by user
	var password string        // declare username given by user
	var scanner *bufio.Scanner // declare scanner to read user input
	scanner = bufio.NewScanner(os.Stdin)

	// provider custom address and port CLI
	server := flag.String("server", "localhost:8080", "server network address")
	path := flag.String("path", "/ws", "websocket path")
	flag.Parse()
	serverURL := url.URL{
		Scheme: "ws",
		Host:   *server,
		Path:   *path,
	}

	// connecting to the server
	fmt.Println("[>] Connecting to the server...")
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()
	fmt.Println("[>] Connected to the server.")

	// login & register
	fmt.Print("[>][INPUT] Username: ")
	scanner.Scan()
	username = scanner.Text()
	fmt.Print("[>][INPUT] Password: ")
	scanner.Scan()
	password = scanner.Text()

	// send user validation to server
	cMessage := Message{
		Username: username,
		Text:     password,
		Type:     "Login",
	}
	conn.WriteJSON(cMessage)

	// app interface
	// fmt.Printf("Welcome to Chat App %s!\n", username)
	// fmt.Printf("Connecting to server @ %s\n", *server)

	go handleReceiveMessage(conn)
	handleSendMessage(conn, scanner, username)
}

func handleReceiveMessage(conn ConnectionReader) {
	for {
		var sMessage Message

		err := conn.ReadJSON(&sMessage)
		if err != nil {
			fmt.Println("[SERVER] Server closed, exiting.")
			os.Exit(0)
		}

		if sMessage.Type == "Login" {
			fmt.Printf("[SERVER] %s\n", sMessage.Text)
		} else {
			fmt.Printf("[CH][%s] %s\n", sMessage.Username, sMessage.Text)
		}
	}
}

func handleSendMessage(conn ConnectionWriter, scanner Scanner, uname string) {
	var cMessage Message
	cMessage.Username = uname

	for {
		if scanner.Scan() {
			cMessage.Text = scanner.Text()

			if cMessage.Text == "exit" {
				fmt.Println("You're leaving chat room.")
				conn.WriteJSON(Message{
					Username: uname,
					Text:     "has disconnected.",
				})
				conn.Close()
				os.Exit(0)
			}

			fmt.Printf("[CH][%s] %s\n", cMessage.Username, cMessage.Text)

			err := conn.WriteJSON(cMessage)
			if err != nil {
				fmt.Println("[ERROR] Sending message, clossing connection.", err)
				break
			}
		}
	}
}
