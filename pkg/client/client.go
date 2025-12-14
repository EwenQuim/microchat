package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendMessage(room, user, message string) error {
	url := fmt.Sprintf("%s/api/rooms/%s/messages", c.baseURL, room)

	payload := models.SendMessageRequest{
		User:    user,
		Content: message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}

	return nil
}

func (c *Client) GetMessages(room string) ([]models.Message, error) {
	url := fmt.Sprintf("%s/api/rooms/%s/messages", c.baseURL, room)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get messages: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (c *Client) GetRooms() ([]string, error) {
	url := fmt.Sprintf("%s/api/rooms", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get rooms: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rooms []models.Room
	if err := json.Unmarshal(body, &rooms); err != nil {
		return nil, err
	}

	names := make([]string, len(rooms))
	for i, room := range rooms {
		names[i] = room.Name
	}

	return names, nil
}
