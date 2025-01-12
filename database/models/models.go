package models

import "time"

type ChatRoom struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	CreateBy     string   `json:"createdBy"`
	Participants []string `json:"participants"`
}

type Message struct {
	Username  string    `json:"username"`
	Receiver  string    `json:"receiver"`
	Text      string    `json:"text"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
