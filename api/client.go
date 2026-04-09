package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

type TaskResult struct {
	FbxFiles    []string `json:"fbx_files,omitempty"`
	HtmlContent string   `json:"html_content,omitempty"`
}

type Task struct {
	TaskID      string     `json:"task_id"`
	Status      string     `json:"status"`
	Text        string     `json:"text"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	CompletedAt time.Time  `json:"completed_at,omitempty"`
	Result      *TaskResult `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type SubmitRequest struct {
	Text         string  `json:"text"`
	Duration     float64 `json:"duration,omitempty"`
	Seeds        []int   `json:"seeds,omitempty"`
	CfgScale     float64 `json:"cfg_scale,omitempty"`
	OutputFormat string  `json:"output_format,omitempty"`
}

type SubmitResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type QueueResponse struct {
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type HealthResponse struct {
	Status       string `json:"status"`
	GPUAvailable bool   `json:"gpu_available"`
	ModelLoaded  bool   `json:"model_loaded"`
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
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Id", c.auth.userID)
	req.Header.Set("X-Token", c.auth.token)

	return c.httpClient.Do(req)
}

func (c *Client) SubmitTask(text string, duration float64, seeds []int, cfgScale float64, outputFormat string) (*SubmitResponse, error) {
	req := SubmitRequest{
		Text:         text,
		Duration:     duration,
		Seeds:        seeds,
		CfgScale:     cfgScale,
		OutputFormat: outputFormat,
	}
	resp, err := c.doRequest("POST", "/tasks", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	var result SubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
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
		return nil, fmt.Errorf("任务未找到: %s", taskID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
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
		return nil, fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	var queue QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queue); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &queue, nil
}

func (c *Client) GetHealth() (*HealthResponse, error) {
	resp, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &health, nil
}

func (c *Client) DownloadTask(taskID, format, outputPath string, version int) error {
	url := fmt.Sprintf("%s/download/%s?format=%s&version=%d", c.baseURL, taskID, format, version)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-Id", c.auth.userID)
	req.Header.Set("X-Token", c.auth.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("任务或文件未找到")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("任务未完成，无法下载")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("认证失败")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}
