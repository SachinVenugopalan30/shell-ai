package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/SachinVenugopalan30/shell-ai/internal/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	Ping(ctx context.Context) error
	Complete(ctx context.Context, msgs []Message) (string, error)
	ListModels(ctx context.Context) ([]string, error)
}

func New(cfg *config.Config) (Provider, error) {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid endpoint URL: %s", cfg.Endpoint)
	}
	if u.Scheme == "http" && !strings.HasPrefix(u.Host, "localhost") && !strings.HasPrefix(u.Host, "127.0.0.1") {
		fmt.Fprintln(os.Stderr, "Warning: using non-localhost HTTP endpoint — traffic is not encrypted")
	}

	switch cfg.Provider {
	case "llamacpp":
		return &LlamaCpp{endpoint: cfg.Endpoint, model: cfg.Model, timeout: cfg.Timeout}, nil
	default:
		return &Ollama{endpoint: cfg.Endpoint, model: cfg.Model, timeout: cfg.Timeout}, nil
	}
}
