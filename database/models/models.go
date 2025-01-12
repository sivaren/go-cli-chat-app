package models

import "github.com/gorilla/websocket"

type ChatRoom struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	CreateBy     string   `json:"createdBy"`
	Participants []string `json:"participants"`
}

type Message struct {
	Connection *websocket.Conn `json:"connection"`
	Username   string          `json:"username"`
	Text       string          `json:"text"`
	Type       string          `json:"type"`
}
