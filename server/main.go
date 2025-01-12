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
	"github.com/sivaren/go-cli-chat-app/server/controllers"
)

// declare struct for chat room channel
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

// declare mapping(connection => int) to track open connection
var sockConnections = make(map[*websocket.Conn]int)
var sockConnUsernameBased = make(map[string]*websocket.Conn) // for sending DM purpose

// declare channel (chat room) to digest messages concurrently
var chatRoom = make(chan ChatRoomMessage)

// declare variables for database (users and messages)
var users map[string]string
var messages []models.Message
var usersFilePath string
var messagesFilePath string

func main() {
	// provide custom port for CLI
	port := flag.String("port", ":8080", "server open port")
	flag.Parse()

	// endpoint for WebSocket connections
	http.HandleFunc("/ws", handleConnections)

	// handle incoming messages concurrently
	go handleMessage()

	// get current working directory
	cwd, errCWD := os.Getwd()
	if errCWD != nil {
		fmt.Println("Error getting CWD: ", errCWD)
	}

	// laoding users and messages data
	usersFilePath = filepath.Join(cwd, "database", "data", "users.json")
	messagesFilePath = filepath.Join(cwd, "database", "data", "messages.json")
	users = database.ReadUsersFromFile(usersFilePath)
	messages = database.ReadMessagesFromFile(messagesFilePath)

	fmt.Println("WebSocket server started on port", *port)
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

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

	// infinite loop to read message from client
	for {
		var cMessage models.Message
		err := conn.ReadJSON(&cMessage)
		if err != nil {
			fmt.Printf("[ID=%v][CLOSING] Client ID=%v closed, closing connection.\n", sockConnections[conn], sockConnections[conn])
			delete(sockConnections, conn)
			break
		}

		// send message to the chat room (channel) to be handled concurrently
		chatRoom <- ChatRoomMessage{
			Connection: conn,
			Message:    cMessage,
		}
	}
}

func handleMessage() {
	for {
		// incoming data from channel chatRoom{ *websocket.Conn, models.Message }
		chatRoomMessage := <-chatRoom

		// unpack incoming chat room message
		connection := chatRoomMessage.Connection // get origin connection
		cMessage := chatRoomMessage.Message      // get client message
		cMessage.Timestamp = time.Now()          // add timestampp to client message

		// write new message to database
		messages = append(messages, cMessage)
		database.WriteMessagesToFile(messagesFilePath, messages)

		// declare message from server to be sent
		var sMessage models.Message
		sMessage.Username = cMessage.Username // add username to server's message
		sMessage.Type = cMessage.Type         // add type to server's message
		sMessage.Timestamp = time.Now()       // add timestamp to server's message

		if cMessage.Type == "LOGIN" {
			// validate user's password
			isAuth := auth.IsPasswordValid(users[cMessage.Username], cMessage.Text)
			if isAuth { // user authenticated
				// add connection based on username
				sockConnUsernameBased[cMessage.Username] = connection
				controllers.LoginSuccess(&cMessage, &sMessage, connection, sockConnections)

				// broadcast to room chat that new user has joined
				broadcastMsg := models.Message{
					Text:      fmt.Sprintf("@%s has joined the chat!", cMessage.Username),
					Type:      "BROADCAST",
					Timestamp: time.Now(),
				}
				controllers.SendBroadcast(sockConnections, connection, broadcastMsg)
			} else { // failed user authentication
				// processing invalid user login
				controllers.LoginFailed(&cMessage, &sMessage, connection, sockConnections)
			}
		} else if cMessage.Type == "REGISTER" {
			// add connection based on username
			sockConnUsernameBased[cMessage.Username] = connection

			// write new user to database
			users[cMessage.Username] = cMessage.Text
			database.WriteUsersToFile(usersFilePath, users)

			controllers.UserRegistration(&cMessage, &sMessage, connection, sockConnections)

			// broadcast to room chat that new user has joined
			broadcastMsg := models.Message{
				Text:      fmt.Sprintf("@%s has joined the chat!", cMessage.Username),
				Type:      "BROADCAST",
				Timestamp: time.Now(),
			}
			controllers.SendBroadcast(sockConnections, connection, broadcastMsg)
		} else if cMessage.Type == "ROOM_CHAT" {
			fmt.Printf("[ID=%v][CH][@%s] %s\n", sockConnections[connection], cMessage.Username, cMessage.Text)

			// broadcast message to room chat
			controllers.SendBroadcast(sockConnections, connection, cMessage)
		} else if cMessage.Type == "DM" {
			// sending DM from @<username> to @<receiver>
			controllers.SendDM(&cMessage, &sMessage, connection, sockConnections, sockConnUsernameBased)
		} else if cMessage.Type == "EXIT" {
			controllers.ExitProgram(&cMessage, connection, sockConnections, sockConnUsernameBased)

			// broadcast message to room chat that someone has left
			broadcastMsg := models.Message{
				Text:      fmt.Sprintf("@%s has left the chat!", cMessage.Username),
				Type:      "BROADCAST",
				Timestamp: time.Now(),
			}
			controllers.SendBroadcast(sockConnections, connection, broadcastMsg)
		}
	}
}
