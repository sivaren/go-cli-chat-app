package models

type ChatRoom struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	CreateBy     string   `json:"createdBy"`
	Participants []string `json:"participants"`
}

type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}
