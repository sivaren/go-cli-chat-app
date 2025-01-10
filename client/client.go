package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)


func main() {	
	// try to connect to websocket server 
	serverURL := "ws://localhost:8080/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		fmt.Println("Failed to connect to server: ", err)	
	}
	defer conn.Close()

	fmt.Println("Connected to WebSocket server!")

	ch := make(chan string)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading from server: ", err)
				close(ch)
				return
			}
			ch <- string(message)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		serverMsg := <-ch
		fmt.Println("Message from server:", serverMsg)
		
		fmt.Println("Type your message and press Enter to send: ")

		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		text := scanner.Text()
		if text == "exit" {
			fmt.Println("Exiting...")
			break
		}

		err := conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			fmt.Println("Error sending message: ", err)
			break
		}
	}
}
