package prompt

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/SachinVenugopalan30/shell-ai/internal/context"
	"github.com/SachinVenugopalan30/shell-ai/internal/provider"
)

type LLMResponse struct {
	Command string
	Reason  string
	Raw     string // set when parsing fails
}

var sysTmpl = template.Must(template.New("sys").Parse(`You are a command-line assistant. The user will describe what they want to do.
Return exactly one shell command that accomplishes their goal.

Environment:
- OS: {{.OS}}{{if .Distro}}
- Distro: {{.Distro}}{{end}}
- Architecture: {{.Arch}}{{if .PackageManager}}
- Package manager: {{.PackageManager}}{{end}}{{if .Shell}}
- Shell: {{.Shell}}{{end}}
- Working directory: {{.CWD}}

Rules:
1. Return JSON only: {"command": "...", "reason": "..."}
2. "command" MUST be compatible with the detected OS and shell — do not use Linux flags on macOS or vice versa.
3. "reason" is a short phrase (max 8 words) explaining what the command does.
4. No markdown, no code fences, no alternatives.
5. Include sudo if elevated privileges are needed.
6. Always use finite/bounded versions of commands (e.g. "ping -c 5" not "ping", "head -n 50" not "cat" for large files).`))

var explainTmpl = template.Must(template.New("exp").Parse(
	`You are a command-line assistant. Explain what this shell command does in plain English, step by step. Be concise. Do not suggest alternatives.`,
))

func BuildMessages(ctx *context.EnvContext, intent string) []provider.Message {
	var buf bytes.Buffer
	sysTmpl.Execute(&buf, ctx)
	return []provider.Message{
		{Role: "system", Content: buf.String()},
		{Role: "user", Content: intent},
	}
}

func BuildExplainMessages(command string) []provider.Message {
	var buf bytes.Buffer
	explainTmpl.Execute(&buf, nil)
	return []provider.Message{
		{Role: "system", Content: buf.String()},
		{Role: "user", Content: "Explain this command: " + command},
	}
}

func ParseResponse(raw string) *LLMResponse {
	raw = strings.TrimSpace(raw)
	cleaned := fixEscapes(stripFences(raw))

	var parsed struct {
		Command string `json:"command"`
		Reason  string `json:"reason"`
	}

	// 1. try parsing the cleaned string directly
	if err := json.Unmarshal([]byte(cleaned), &parsed); err == nil && parsed.Command != "" {
		return &LLMResponse{Command: parsed.Command, Reason: parsed.Reason, Raw: raw}
	}

	// 2. find first { and decode the first JSON object from there
	if idx := strings.Index(cleaned, "{"); idx >= 0 {
		dec := json.NewDecoder(strings.NewReader(cleaned[idx:]))
		if err := dec.Decode(&parsed); err == nil && parsed.Command != "" {
			return &LLMResponse{Command: parsed.Command, Reason: parsed.Reason, Raw: raw}
		}
	}

	// 3. raw fallback — show the response and let the user decide
	return &LLMResponse{Raw: raw}
}

// stripFences removes ```json ... ``` or ``` ... ``` wrappers
func stripFences(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// fixEscapes drops invalid JSON escape sequences (e.g. \$ from awk commands)
func fixEscapes(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
				out.WriteByte(s[i])
			default:
				// invalid escape — drop the backslash, keep the character
			}
		} else {
			out.WriteByte(s[i])
		}
	}
	return out.String()
}
