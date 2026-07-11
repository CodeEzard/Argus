package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/codeezard/argus/internal/sysinfo"
)

// Provider is the interface.
type Provider interface {
	Complete(prompt string) (string, error)
}

// Client wraps a Provider and handles parsing the response.
type Client struct {
	provider Provider
}

func NewClient(provider Provider) *Client {
	return &Client{provider: provider}
}

// Diagnose takes a snapshot, builds a prompt, calls the LLM, parses the result.
func (c *Client) Diagnose(snap sysinfo.ContextSnapshot) (sysinfo.Suggestion, error) {
	prompt := sysinfo.BuildPrompt(snap)

	response, err := c.provider.Complete(prompt)
	if err != nil {
		return sysinfo.Suggestion{}, err
	}

	type rawSuggestion struct {
		Severity    string          `json:"severity"`
		Diagnosis   string          `json:"diagnosis"`
		Commands    json.RawMessage `json:"commands"`
		LongTermFix json.RawMessage `json:"long_term_fix"`
		Confidence  float64         `json:"confidence"`
	}

	var raw rawSuggestion
	if err := json.Unmarshal([]byte(response), &raw); err != nil {
		return sysinfo.Suggestion{}, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	var commands []string
	if err := json.Unmarshal(raw.Commands, &commands); err != nil {
		var single string
		if err := json.Unmarshal(raw.Commands, &single); err == nil {
			commands = []string{single}
		}
	}

	var longTermFix string
	if err := json.Unmarshal(raw.LongTermFix, &longTermFix); err != nil {
		longTermFix = string(raw.LongTermFix)
	}

	return sysinfo.Suggestion{
		Severity:    raw.Severity,
		Diagnosis:   raw.Diagnosis,
		Commands:    commands,
		LongTermFix: longTermFix,
		Confidence:  raw.Confidence,
	}, nil

}

// OllamaProvider calls a local Ollama instance.
type OllamaProvider struct {
	BaseURL string // default: http://localhost:11434
	Model   string // default: llama3
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (o *OllamaProvider) Complete(prompt string) (string, error) {
	baseURL := o.BaseURL
	client := &http.Client{Timeout: 60 * time.Second}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := o.Model
	if model == "" {
		model = "llama3"
	}

	reqBody := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := client.Post(baseURL+"/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to call ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(data))
	}

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	fmt.Println("DEBUG RAW:", result.Response)

	return result.Response, nil
}
