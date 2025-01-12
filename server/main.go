package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
var sockConnUsernameBased = make(map[string]*websocket.Conn)  // for DM purpose 

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
	fmt.Printf("[ID=%v][HANDSHAKE] New connection with ID=%v established!\n", sockConnections[conn], sockConnections[conn])

	for {
		// read message from client
		var cMessage models.Message
		err := conn.ReadJSON(&cMessage)
		if err != nil {
			fmt.Printf("[ID=%v][CLOSING] Client ID=%v closed, closing connection.\n", sockConnections[conn], sockConnections[conn])
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
		cMessage.Timestamp = time.Now()

		messages = append(messages, cMessage)
		database.WriteMessagesToFile(messagesFilePath, messages)

		var sMessage models.Message
		sMessage.Username = cMessage.Username
		sMessage.Type = cMessage.Type
		sMessage.Timestamp = time.Now()

		if cMessage.Type == "LOGIN" {
			isAuth := auth.IsPasswordValid(users[cMessage.Username], cMessage.Text)
			if isAuth {
				// add connection based on username
				sockConnUsernameBased[cMessage.Username] = connection

				sMessage.Text = "Login successful!"
				fmt.Printf("[ID=%v][LOGIN] @%s successful!\n", sockConnections[connection], cMessage.Username)

				err := connection.WriteJSON(sMessage)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					connection.Close()
					delete(sockConnections, connection)
				}

				// broadcast to room chat user has joined
				broadcastMsg := models.Message{
					Text:      fmt.Sprintf("@%s has joined the chat!", cMessage.Username),
					Type:      "BROADCAST",
					Timestamp: time.Now(),
				}
				sendBroadcast(connection, broadcastMsg)
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
		} else if cMessage.Type == "REGISTER" {
			// add connection based on username
			sockConnUsernameBased[cMessage.Username] = connection

			users[cMessage.Username] = cMessage.Text
			database.WriteUsersToFile(usersFilePath, users)

			sMessage.Text = "Account registered!"
			fmt.Printf("[ID=%v][REGISTER] @%s account registered!\n", sockConnections[connection], cMessage.Username)

			err := connection.WriteJSON(sMessage)
			if err != nil {
				fmt.Println("[ERROR] Sending message, closing connection.", err)
				connection.Close()
				delete(sockConnections, connection)
			}

			// broadcast to room chat user has joined
			broadcastMsg := models.Message{
				Text:      fmt.Sprintf("@%s has joined the chat!", cMessage.Username),
				Type:      "BROADCAST",
				Timestamp: time.Now(),
			}
			sendBroadcast(connection, broadcastMsg)
		} else if cMessage.Type == "ROOM_CHAT" {
			fmt.Printf("[ID=%v][CH][@%s] %s\n", sockConnections[connection], cMessage.Username, cMessage.Text)

			// broadcast message to room chat
			sendBroadcast(connection, cMessage)
		} else if cMessage.Type == "DM" {
			fmt.Printf("[ID=%v][DM][from:@%s][to:@%s] %s\n", sockConnections[connection], cMessage.Username, cMessage.Receiver, cMessage.Text)

			receiverConn := sockConnUsernameBased[cMessage.Receiver]
			sMessage.Receiver = cMessage.Receiver
			sMessage.Text = cMessage.Text

			err := receiverConn.WriteJSON(sMessage)
			if err != nil {
				fmt.Println("[ERROR] Sending message, closing connection.", err)
				receiverConn.Close()
				delete(sockConnections, receiverConn)
				delete(sockConnUsernameBased, cMessage.Receiver)
			}
		} else if cMessage.Type == "EXIT" {
			fmt.Printf("[ID=%v][CH] @%s Leaving chat room.\n", sockConnections[connection], cMessage.Username)
			connection.Close()
			delete(sockConnections, connection)

			// broadcast message to room chat
			broadcastMsg := models.Message{
				Text:      fmt.Sprintf("@%s has left the chat!", cMessage.Username),
				Type:      "BROADCAST",
				Timestamp: time.Now(),
			}
			sendBroadcast(connection, broadcastMsg)
		}
	}
}

func sendBroadcast(connection *websocket.Conn, message models.Message) {
	if message.Type == "ROOM_CHAT" {
		for conn := range sockConnections {
			if conn != connection {
				err := conn.WriteJSON(message)
				if err != nil {
					fmt.Println("[ERROR] Sending message, closing connection.", err)
					conn.Close()
					delete(sockConnections, conn)
				}
			}
		}
	} else {
		for conn := range sockConnections {
			err := conn.WriteJSON(message)
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
	usersFilePath = filepath.Join(cwd, "database", "data", "users.json")
	messagesFilePath = filepath.Join(cwd, "database", "data", "messages.json")
	users = database.ReadUsersFromFile(usersFilePath)
	messages = database.ReadMessagesFromFile(messagesFilePath)

	fmt.Println("WebSocket server started on port", *port)

	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
