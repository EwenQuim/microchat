package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/EwenQuim/microchat/client/sdk/generated"
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

	c, err := generated.NewClientWithResponses(*apiURL)
	if err != nil {
		slog.Error("Failed to create client", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch *command {
	case "send":
		if *message == "" || *user == "" {
			slog.Error("message and user are required for send command")
			os.Exit(1)
		}
		resp, err := c.POSTapiroomsRoommessagesWithResponse(ctx, *room, &generated.POSTapiroomsRoommessagesParams{}, generated.SendMessageRequest{Content: *message, User: *user})
		if err != nil {
			slog.Error("Failed to send message", "error", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 201 {
			slog.Error("Failed to send message", "status", resp.StatusCode())
			os.Exit(1)
		}
		fmt.Println("Message sent successfully")

	case "list":
		resp, err := c.GETapiroomsRoommessagesWithResponse(ctx, *room, &generated.GETapiroomsRoommessagesParams{})
		if err != nil {
			slog.Error("Failed to get messages", "error", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 200 {
			slog.Error("Failed to get messages", "status", resp.StatusCode())
			os.Exit(1)
		}
		for _, msg := range *resp.JSON200 {
			ts := ""
			if msg.Timestamp != nil {
				ts = msg.Timestamp.String()
			}
			u := ""
			if msg.User != nil {
				u = *msg.User
			}
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			fmt.Printf("[%s] %s: %s\n", ts, u, content)
		}

	case "rooms":
		resp, err := c.GETapiroomsWithResponse(ctx, &generated.GETapiroomsParams{})
		if err != nil {
			slog.Error("Failed to get rooms", "error", err)
			os.Exit(1)
		}
		if resp.StatusCode() != 200 {
			slog.Error("Failed to get rooms", "status", resp.StatusCode())
			os.Exit(1)
		}
		fmt.Println("Available rooms:")
		for _, r := range *resp.JSON200 {
			if r.Name == nil {
				continue
			}
			fmt.Printf("  - %s\n", *r.Name)
		}

	default:
		slog.Error("Unknown command", "command", *command)
		os.Exit(1)
	}
}
