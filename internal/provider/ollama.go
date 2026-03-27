package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Ollama struct {
	endpoint string
	model    string
	timeout  time.Duration
}

func (o *Ollama) client() *http.Client {
	return &http.Client{Timeout: o.timeout}
}

func (o *Ollama) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", o.endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := o.client().Do(req)
	if err != nil {
		return fmt.Errorf("Ollama not found at %s — is it running?", o.endpoint)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}
	return nil
}

func (o *Ollama) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", o.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := o.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range result.Models {
		names = append(names, m.Name)
	}
	return names, nil
}

func (o *Ollama) Complete(ctx context.Context, msgs []Message) (string, error) {
	model := o.model
	if model == "" {
		// use first available model
		models, err := o.ListModels(ctx)
		if err != nil || len(models) == 0 {
			return "", fmt.Errorf("No models found — run: ollama pull <model>")
		}
		model = models[0]
	}

	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"messages": msgs,
		"stream":   false,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Message.Content, nil
}
