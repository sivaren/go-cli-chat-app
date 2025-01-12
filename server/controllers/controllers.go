package controllers

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sivaren/go-cli-chat-app/database/models"
)

// handling login success case
func LoginSuccess(
	cMessage *models.Message,
	sMessage *models.Message,
	connection *websocket.Conn,
	sockConnections map[*websocket.Conn]int,
) {
	sMessage.Text = "Login successful!"
	fmt.Printf("[ID=%v][LOGIN] @%s successful!\n", sockConnections[connection], cMessage.Username)

	err := connection.WriteJSON(sMessage)
	if err != nil {
		fmt.Println("[ERROR] Sending message, closing connection.", err)
		connection.Close()
		delete(sockConnections, connection)
	}
}

// handling login failed case
func LoginFailed(
	cMessage *models.Message,
	sMessage *models.Message,
	connection *websocket.Conn,
	sockConnections map[*websocket.Conn]int,
) {
	sMessage.Text = "Login invalid!"
	fmt.Printf("[LOGIN] @%s invalid!\n", cMessage.Username)

	err := connection.WriteJSON(sMessage)
	if err != nil {
		fmt.Println("[ERROR] Sending message, closing connection.", err)
	}
	connection.Close()
	delete(sockConnections, connection)
}

// handling user registration case
func UserRegistration(
	cMessage *models.Message,
	sMessage *models.Message,
	connection *websocket.Conn,
	sockConnections map[*websocket.Conn]int,
) {
	sMessage.Text = "Account registered!"
	fmt.Printf("[ID=%v][REGISTER] @%s account registered!\n", sockConnections[connection], cMessage.Username)

	err := connection.WriteJSON(sMessage)
	if err != nil {
		fmt.Println("[ERROR] Sending message, closing connection.", err)
		connection.Close()
		delete(sockConnections, connection)
	}
}

// handling direct message case
func SendDM(
	cMessage *models.Message,
	sMessage *models.Message,
	connection *websocket.Conn,
	sockConnections map[*websocket.Conn]int,
	sockConnUsernameBased map[string]*websocket.Conn,
) {
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
}

// handling user exit the program case
func ExitProgram(
	cMessage *models.Message,
	connection *websocket.Conn,
	sockConnections map[*websocket.Conn]int,
	sockConnUsernameBased map[string]*websocket.Conn,
) {
	fmt.Printf("[ID=%v][CH] @%s Leaving chat room.\n", sockConnections[connection], cMessage.Username)
	connection.Close()
	delete(sockConnections, connection)
	delete(sockConnUsernameBased, cMessage.Username)
}

// handling broadcast from server to users case
func SendBroadcast(
	sockConnections map[*websocket.Conn]int,
	connection *websocket.Conn,
	message models.Message,
) {
	if message.Type == "ROOM_CHAT" { // user send msg to room chat
		for conn := range sockConnections {
			// prevent echo the msg to the origin (msg sender)
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
		// except room chat msg, broadcast will be sent to all users (including the origin)
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
