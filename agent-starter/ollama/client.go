package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Message is one turn in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Think    bool      `json:"think"`
}

type ChatResponse struct {
	Message    Message `json:"message"`
	Model      string  `json:"model"`
	TokenCount int     `json:"prompt_eval_count"`
	Duration   float64 `json:"total_duration"`
}

// APIError is returned when Ollama answers with a non-200 status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ollama: unexpected status %d: %s", e.StatusCode, e.Body)
}

// Client talks to an Ollama server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a Client for the server at baseURL, e.g. "http://localhost:11434".
func CreateClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Chat sends req to /api/chat and returns the decoded reply.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("encoding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("building request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return ChatResponse{}, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ChatResponse{}, fmt.Errorf("decoding response: %w", err)
	}
	return out, nil
}

func WgSimulation() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("Goroutine 1 is running")
		time.Sleep(2 * time.Second)
		fmt.Println("Goroutine 1 is done")
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Goroutine 2 is running")
		time.Sleep(1 * time.Second)
		fmt.Println("Goroutine 2 is done")
	}()

	wg.Wait()
	fmt.Println("All goroutines are done")
}

// stage 1: numbers in -> channel out
func Gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

// stage 2: square each number
func Sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

// FAN-IN: merge many channels into one
func Merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	wg.Add(len(cs))

	copy := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	for _, c := range cs {
		go copy(c)
	}

	// close out once every copier has finished
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
