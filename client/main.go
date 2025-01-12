package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/sivaren/go-cli-chat-app/auth"
	"github.com/sivaren/go-cli-chat-app/database/models"
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

var menuOption string

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

	// menu options
	fmt.Println("[>] Choose Menu | type the option number:")
	fmt.Println("[>] 1. Login")
	fmt.Println("[>] 2. Register")
	fmt.Print("[>][INPUT] Menu Option: ")
	scanner.Scan()
	menuOption = scanner.Text()

	if menuOption == "1" {
		// login
		fmt.Print("[>][INPUT] Username: ")
		scanner.Scan()
		username = scanner.Text()
		fmt.Print("[>][INPUT] Password: ")
		scanner.Scan()
		password = scanner.Text()

		// send user validation to server
		cMessage := models.Message{
			Username: username,
			Text:     password,
			Type:     "LOGIN",
		}
		conn.WriteJSON(cMessage)
	} else {
		// register
		fmt.Print("[>][INPUT] Username: ")
		scanner.Scan()
		username = scanner.Text()
		fmt.Print("[>][INPUT] Password: ")
		scanner.Scan()
		password = scanner.Text()

		hashedPassword, _ := auth.HashPassword(password)

		// send user validation to server
		cMessage := models.Message{
			Username: username,
			Text:     hashedPassword,
			Type:     "REGISTER",
		}
		conn.WriteJSON(cMessage)
	}

	// app interface
	// fmt.Printf("Welcome to Chat App %s!\n", username)
	// fmt.Printf("Connecting to server @ %s\n", *server)

	go handleReceiveMessage(conn)
	handleSendMessage(conn, scanner, username)
}

func handleReceiveMessage(conn ConnectionReader) {
	for {
		var sMessage models.Message

		err := conn.ReadJSON(&sMessage)
		if err != nil {
			fmt.Println("[SERVER] Server closed, exiting.")
			os.Exit(0)
		}

		if sMessage.Type == "LOGIN" {
			fmt.Printf("[SERVER] %s\n", sMessage.Text)
		} else if sMessage.Type == "REGISTER" {
			fmt.Printf("[SERVER] %s\n", sMessage.Text)
		} else if sMessage.Type == "ROOM_CHAT" {
			fmt.Printf("[CH][@%s] %s\n", sMessage.Username, sMessage.Text)
		} else if sMessage.Type == "BROADCAST" {
			fmt.Printf("[SERVER] %s\n", sMessage.Text)
		}
	}
}

func handleSendMessage(conn ConnectionWriter, scanner Scanner, username string) {
	var cMessage models.Message
	cMessage.Username = username

	for {
		if scanner.Scan() {
			cMessage.Text = scanner.Text()

			if cMessage.Text == "exit" {
				fmt.Println("[CH] You're leaving the chat room.")
				conn.WriteJSON(models.Message{
					Username: username,
					Type:     "EXIT",
				})
				conn.Close()
				os.Exit(0)
			}

			cMessage.Type = "ROOM_CHAT"
			fmt.Printf("[CH][@%s] %s\n", cMessage.Username, cMessage.Text)

			err := conn.WriteJSON(cMessage)
			if err != nil {
				fmt.Println("[ERROR] Sending message, clossing connection.", err)
				break
			}
		}
	}
}
