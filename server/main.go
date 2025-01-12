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
)

// declare Message struct
type Message struct {
	Connection *websocket.Conn `json:"connection"`
	Username   string          `json:"username"`
	Text       string          `json:"text"`
	Type       string          `json:"type"`
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
var chatRoom = make(chan Message)

// declare variables for database
var users map[string]string

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
		var cMessage Message
		err := conn.ReadJSON(&cMessage)
		if err != nil {
			fmt.Printf("[CLOSING] Client ID=%v closed, closing connection.\n", sockConnections[conn])
			delete(sockConnections, conn)
			break
		}

		// send message to the chat room to be handled concurrently
		cMessage.Connection = conn
		chatRoom <- cMessage
	}
}

func handleMessage() {
	for {
		cMessage := <-chatRoom
		if cMessage.Type == "Login" {
			isAuth := auth.IsPasswordValid(users[cMessage.Username], cMessage.Text)
			if isAuth {
				sMessage := Message{
					Username: cMessage.Username,
					Text:     "Login successful!",
					Type:     "Login",
				}

				fmt.Printf("[LOGIN] %s successful!\n", cMessage.Username)
				err := cMessage.Connection.WriteJSON(sMessage)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					cMessage.Connection.Close()
					delete(sockConnections, cMessage.Connection)
				}
			}
		} else {
			fmt.Printf("[CH][%s] %s\n", cMessage.Username, cMessage.Text)
	
			for conn := range sockConnections {
				if conn != cMessage.Connection {
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
	usersFilePath := filepath.Join(cwd, "database", "data", "users.json")
	users = database.ReadUsersFromFile(usersFilePath)

	fmt.Println("WebSocket server started on port", *port)

	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
