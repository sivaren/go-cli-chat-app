package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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

func WriteUsersToFile(filePath string, users map[string]string) {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON data:", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Println("Error writing file:", err)
	}
}
