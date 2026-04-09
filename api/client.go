package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hy-motion-cli/config"
)

type Client struct {
	baseURL string
	timeout time.Duration
	auth    struct {
		userID string
		token  string
	}
	httpClient *http.Client
}

type Task struct {
	TaskID      string    `json:"task_id"`
	Status      string    `json:"status"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type SubmitRequest struct {
	Text string `json:"text"`
}

type SubmitResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type QueueResponse struct {
	Pending int `json:"pending"`
	Running int `json:"running"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.API.URL,
		timeout: time.Duration(cfg.API.Timeout) * time.Second,
		auth: struct {
			userID string
			token  string
		}{
			userID: cfg.Auth.UserID,
			token:  cfg.Auth.Token,
		},
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.API.Timeout) * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = nil
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Id", c.auth.userID)
	req.Header.Set("X-Token", c.auth.token)

	return c.httpClient.Do(req)
}

func (c *Client) SubmitTask(text string) (*SubmitResponse, error) {
	resp, err := c.doRequest("POST", "/tasks", SubmitRequest{Text: text})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result SubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetTaskStatus(taskID string) (*Task, error) {
	resp, err := c.doRequest("GET", "/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &task, nil
}

func (c *Client) GetQueue() (*QueueResponse, error) {
	resp, err := c.doRequest("GET", "/queue", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var queue QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &queue, nil
}
