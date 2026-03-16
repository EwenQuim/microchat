package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/EwenQuim/microchat/client/sdk/generated"
	"github.com/EwenQuim/microchat/internal/tui"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "microchat",
		Usage: "µchat client — runs TUI by default",
		Action: func(c *cli.Context) error {
			return tui.Run()
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Value: "http://localhost:8080",
				Usage: "API server URL",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "send",
				Usage: "Send a message to a room",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "room", Value: "general", Usage: "Chat room name"},
					&cli.StringFlag{Name: "message", Required: true, Usage: "Message to send"},
					&cli.StringFlag{Name: "user", Required: true, Usage: "Username"},
				},
				Action: runSend,
			},
			{
				Name:  "list",
				Usage: "List messages in a room",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "room", Value: "general", Usage: "Chat room name"},
				},
				Action: runList,
			},
			{
				Name:   "rooms",
				Usage:  "List all available rooms",
				Action: runRooms,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}

func newClient(c *cli.Context) (*generated.ClientWithResponses, error) {
	return generated.NewClientWithResponses(c.String("url"))
}

func runSend(c *cli.Context) error {
	client, err := newClient(c)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	room := c.String("room")
	message := c.String("message")
	user := c.String("user")
	resp, err := client.POSTapiroomsRoommessagesWithResponse(context.Background(), room, &generated.POSTapiroomsRoommessagesParams{}, generated.SendMessageRequest{Content: message, User: user})
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	if resp.StatusCode() != 201 {
		return fmt.Errorf("send message: status %d", resp.StatusCode())
	}
	fmt.Println("Message sent successfully")
	return nil
}

func runList(c *cli.Context) error {
	client, err := newClient(c)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	room := c.String("room")
	resp, err := client.GETapiroomsRoommessagesWithResponse(context.Background(), room, &generated.GETapiroomsRoommessagesParams{})
	if err != nil {
		return fmt.Errorf("get messages: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("get messages: status %d", resp.StatusCode())
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
	return nil
}

func runRooms(c *cli.Context) error {
	client, err := newClient(c)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	resp, err := client.GETapiroomsWithResponse(context.Background(), &generated.GETapiroomsParams{})
	if err != nil {
		return fmt.Errorf("get rooms: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("get rooms: status %d", resp.StatusCode())
	}
	fmt.Println("Available rooms:")
	for _, r := range *resp.JSON200 {
		if r.Name == nil {
			continue
		}
		fmt.Printf("  - %s\n", *r.Name)
	}
	return nil
}
