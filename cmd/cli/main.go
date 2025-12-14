package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/EwenQuim/microchat/pkg/client"
)

func main() {
	var (
		apiURL  = flag.String("url", "http://localhost:8080", "API server URL")
		command = flag.String("cmd", "", "Command to execute (send, list, rooms)")
		room    = flag.String("room", "general", "Chat room name")
		message = flag.String("message", "", "Message to send")
		user    = flag.String("user", "", "Username")
	)

	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: microchat -cmd <send|list|rooms> [options]")
		fmt.Println("\nCommands:")
		fmt.Println("  send    Send a message to a room")
		fmt.Println("  list    List messages in a room")
		fmt.Println("  rooms   List all available rooms")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	c := client.NewClient(*apiURL)

	switch *command {
	case "send":
		if *message == "" || *user == "" {
			log.Fatal("message and user are required for send command")
		}
		if err := c.SendMessage(*room, *user, *message); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Message sent successfully")

	case "list":
		messages, err := c.GetMessages(*room)
		if err != nil {
			log.Fatal(err)
		}
		for _, msg := range messages {
			fmt.Printf("[%s] %s: %s\n", msg.Timestamp, msg.User, msg.Content)
		}

	case "rooms":
		rooms, err := c.GetRooms()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Available rooms:")
		for _, room := range rooms {
			fmt.Printf("  - %s\n", room)
		}

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}
