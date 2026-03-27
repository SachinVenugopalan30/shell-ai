package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LlamaCpp struct {
	endpoint string
	model    string
	timeout  time.Duration
}

func (l *LlamaCpp) client() *http.Client {
	return &http.Client{Timeout: l.timeout}
}

func (l *LlamaCpp) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", l.endpoint+"/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := l.client().Do(req)
	if err != nil {
		return fmt.Errorf("llama.cpp server not found at %s — is it running?", l.endpoint)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("llama.cpp returned status %d", resp.StatusCode)
	}
	return nil
}

func (l *LlamaCpp) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", l.endpoint+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := l.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range result.Data {
		names = append(names, m.ID)
	}
	return names, nil
}

func (l *LlamaCpp) Complete(ctx context.Context, msgs []Message) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    l.model,
		"messages": msgs,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", l.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from llama.cpp")
	}
	return result.Choices[0].Message.Content, nil
}
