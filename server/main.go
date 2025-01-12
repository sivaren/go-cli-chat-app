package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/sivaren/go-cli-chat-app/auth"
	"github.com/sivaren/go-cli-chat-app/database"
	"github.com/sivaren/go-cli-chat-app/database/models"
)

type ChatRoomMessage struct {
	Connection *websocket.Conn
	Message    models.Message
}

// initialize websocket upgrader : upgrade http to websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// declare mapping(connection => bool) to track open connection
var sockConnections = make(map[*websocket.Conn]int)

// declare channel (chat room) to digest messages concurrently
var chatRoom = make(chan ChatRoomMessage)

// declare variables for database
var users map[string]string
var messages []models.Message
var usersFilePath string
var messagesFilePath string

// handle websocket connection
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// upgrade initial HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("[ERROR] Upgrading connection:", err)
		return
	}
	defer conn.Close() // ensure connection is closed when function exits

	// new connection established
	sockConnections[conn] = len(sockConnections) + 1
	fmt.Printf("[HANDSHAKE] New connection with ID=%v established!\n", sockConnections[conn])

	for {
		// read message from client
		var cMessage models.Message
		err := conn.ReadJSON(&cMessage)
		if err != nil {
			fmt.Printf("[CLOSING] Client ID=%v closed, closing connection.\n", sockConnections[conn])
			delete(sockConnections, conn)
			break
		}

		// send message to the chat room to be handled concurrently
		chatRoom <- ChatRoomMessage{
			Connection: conn,
			Message:    cMessage,
		}
	}
}

func handleMessage() {
	for {
		chatRoomMessage := <-chatRoom

		connection := chatRoomMessage.Connection
		cMessage := chatRoomMessage.Message

		messages = append(messages, cMessage)
		database.WriteMessagesToFile(messagesFilePath, messages)

		var sMessage models.Message
		sMessage.Username = cMessage.Username
		sMessage.Type = cMessage.Type

		if cMessage.Type == "Login" {
			isAuth := auth.IsPasswordValid(users[cMessage.Username], cMessage.Text)
			if isAuth {
				sMessage.Text = "Login successful!"
				fmt.Printf("[LOGIN] @%s successful!\n", cMessage.Username)

				err := connection.WriteJSON(sMessage)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					connection.Close()
					delete(sockConnections, connection)
				}
			} else {
				sMessage.Text = "Login invalid!"
				fmt.Printf("[LOGIN] @%s invalid!\n", cMessage.Username)

				err := connection.WriteJSON(sMessage)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					connection.Close()
					delete(sockConnections, connection)
				}
				connection.Close()
				delete(sockConnections, connection)
			}
		} else if cMessage.Type == "Register" {
			users[cMessage.Username] = cMessage.Text
			database.WriteUsersToFile(usersFilePath, users)

			sMessage.Text = "Account registered!"
			fmt.Printf("[REGISTER] Account @%s registered!\n", cMessage.Username)

			err := connection.WriteJSON(sMessage)
			if err != nil {
				fmt.Println("[ERROR] Sending message, closing connection.", err)
				connection.Close()
				delete(sockConnections, connection)
			}
		} else {
			fmt.Printf("[CH][@%s] %s\n", cMessage.Username, cMessage.Text)

			for conn := range sockConnections {
				if conn != connection {
					err := conn.WriteJSON(cMessage)
					if err != nil {
						fmt.Println("[ERROR] Sending message, closing connection.", err)
						conn.Close()
						delete(sockConnections, conn)
					}
				}
			}
		}
	}
}

func main() {
	// provide custom port for CLI
	port := flag.String("port", ":8080", "server open port")
	flag.Parse()

	http.HandleFunc("/ws", handleConnections) // endpoint for WebSocket connections

	go handleMessage() // handle incoming messages concurrently

	// loading data
	cwd, errCWD := os.Getwd()
	if errCWD != nil {
		fmt.Println("Error getting CWD: ", errCWD)
	}
	usersFilePath = filepath.Join(cwd, "database", "data", "users.json")
	messagesFilePath = filepath.Join(cwd, "database", "data", "messages.json")
	users = database.ReadUsersFromFile(usersFilePath)
	messages = database.ReadMessagesFromFile(messagesFilePath)

	fmt.Println("WebSocket server started on port", *port)

	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
