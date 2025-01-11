package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/sivaren/go-cli-chat-app/database"
)

// declare Message struct
type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
	Type     string `json:"type"`
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
var sockConnections = make(map[*websocket.Conn]bool)

// declare channel (chat room) to digest messages concurrently
var chatRoom = make(chan Message)

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
	sockConnections[conn] = true
	fmt.Println("[HANDSHAKE] New websocket connection established!")

	for {
		// read message from client
		var cMessage Message
		err := conn.ReadJSON(&cMessage)
		if err != nil {
			fmt.Println("[CLOSING] Client closed, closing connection.")
			delete(sockConnections, conn)
			break
		}

		// send message to the chat room to be handled concurrently
		chatRoom <- cMessage
	}
}

func handleMessage() {
	for {
		cMessage := <-chatRoom
		fmt.Printf("[CH][%s] %s\n", cMessage.Username, cMessage.Text)

		for conn := range sockConnections {
			err := conn.WriteJSON(cMessage)
			if err != nil {
				fmt.Println("[ERROR] Sending message, closing connection.", err)
				conn.Close()
				delete(sockConnections, conn)
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
	users := database.ReadUsersFromFile(usersFilePath)
	database.WriteUsersToFile(usersFilePath, users)

	fmt.Println("WebSocket server started on port", *port)

	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
