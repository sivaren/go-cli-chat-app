package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sivaren/go-cli-chat-app/auth"
	"github.com/sivaren/go-cli-chat-app/database/models"
)

// interface for websocket connnection reading process
type ConnectionReader interface {
	ReadJSON(v interface{}) error
}

// interface for websocket connnection writing process
type ConnectionWriter interface {
	WriteJSON(v interface{}) error
	Close() error
}

// interface for input scanner
type Scanner interface {
	Scan() bool
	Text() string
}

// declare menu option for user input
var menuOption string

func main() {
	var username string // declare username given by user
	var password string // declare username given by user

	// input scanner initialization
	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

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
	fmt.Printf("[>] Connecting to the server @%s.\n", *server)
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

	if menuOption == "1" { // user login
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
	} else if menuOption == "2" { // user registration
		fmt.Print("[>][INPUT] Username: ")
		scanner.Scan()
		username = scanner.Text()
		fmt.Print("[>][INPUT] Password: ")
		scanner.Scan()
		password = scanner.Text()

		// hashing user's password
		hashedPassword, _ := auth.HashPassword(password)

		// send user registration to server
		cMessage := models.Message{
			Username: username,
			Text:     hashedPassword,
			Type:     "REGISTER",
		}
		conn.WriteJSON(cMessage)
	} else {
		fmt.Println("[>][ERROR] There's no such option, closing program.")
		os.Exit(0)
	}

	// welcoming app interface
	fmt.Printf("[>] Welcome to CLI Chat App @%s!\n", username)
	fmt.Println("[>] 1. Type 'dm@<username>:<your-message>' to send a DM")
	fmt.Println("[>] 2. Type 'exit' to close the program")

	// handling incoming message concurrently
	go handleReceiveMessage(conn)

	// handling send message to the server
	handleSendMessage(conn, scanner, username)
}

// handling incoming message concurrently
func handleReceiveMessage(conn ConnectionReader) {
	// infinite loop to listen messages from server
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
		} else if sMessage.Type == "DM" {
			fmt.Printf("[DM][from:@%s] %s\n", sMessage.Username, sMessage.Text)
		}
	}
}

// handling send message to the server
func handleSendMessage(conn ConnectionWriter, scanner Scanner, username string) {
	var cMessage models.Message
	cMessage.Username = username // set message's username

	// infinite loop to get user input (to send message to the server)
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

			parsedInput := strings.Split(cMessage.Text, "@")
			if len(parsedInput) > 1 { // sending DM to @<username>
				// parsing to get receiver's username and text message
				parsedUserText := strings.Split(parsedInput[1], ":")
				receiver := parsedUserText[0] // get receiver's username'
				text := parsedUserText[1]     // get text message

				cMessage.Receiver = receiver // set message's receiver (using @username)
				cMessage.Text = text         // set message's text
				cMessage.Type = "DM"         // set message's type' to 'DM'

				fmt.Printf("[DM][to:@%s] %s\n", receiver, text)
				err := conn.WriteJSON(cMessage)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					break
				}
			} else {
				cMessage.Type = "ROOM_CHAT"

				fmt.Printf("[CH][@%s] %s\n", cMessage.Username, cMessage.Text)
				err := conn.WriteJSON(cMessage)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					break
				}
			}
		}
	}
}
