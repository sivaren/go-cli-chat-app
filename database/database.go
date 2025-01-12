package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sivaren/go-cli-chat-app/database/models"
)

func ReadMessagesFromFile(filePath string) []models.Message {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]models.Message, 0)
		}
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	var messages []models.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		fmt.Println("Error parsing JSON data:", err)
	}

	return messages
}

func ReadUsersFromFile(filePath string) map[string]string {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string)
		}
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	var users map[string]string
	if err := json.Unmarshal(data, &users); err != nil {
		fmt.Println("Error parsing JSON data:", err)
	}

	return users
}

func WriteMessagesToFile(filePath string, messages []models.Message) {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON data:", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func WriteUsersToFile(filePath string, users map[string]string) {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON data:", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Println("Error writing file:", err)
	}
}
