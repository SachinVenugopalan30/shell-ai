# shellai

Translate natural language into shell commands using a local LLM.

```
shellai "find all files larger than 100MB"
  Command:  find . -size +100M -type f
  Reason:   find files over 100MB in current dir
```

No data leaves your machine — shellai runs entirely against a local model via [Ollama](https://ollama.com) or [llama.cpp](https://github.com/ggerganov/llama.cpp).

---

## Requirements

- [Ollama](https://ollama.com) running locally (default), **or** a [llama.cpp](https://github.com/ggerganov/llama.cpp) server
- A pulled model — e.g. `ollama pull llama3.2`
- Go 1.21+ if installing via `go install`

---

## Installation

### Option 1 — go install

```bash
go install github.com/SachinVenugopalan30/shellai@latest
```

The binary lands in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your `$PATH`.

### Option 2 — pre-built binary

Download the latest release for your OS and architecture from [GitHub Releases](https://github.com/SachinVenugopalan30/shellai/releases), then move the binary to somewhere on your `$PATH`:

```bash
# example for macOS arm64
mv shellai_darwin_arm64 /usr/local/bin/shellai
chmod +x /usr/local/bin/shellai
```

---

## Quick Start

```bash
# 1. Pull a model
ollama pull llama3.2

# 2. Configure shellai (one-time setup)
shellai set

# 3. Ask for a command
shellai "show disk usage for each folder here"
```

shellai will display the generated command and a short reason, then ask you to confirm before running anything.

---

## Commands

| Command | Description |
|---------|-------------|
| `shellai "intent"` | Generate and run a shell command |
| `shellai set` | Interactive setup wizard |
| `shellai set <key> <value>` | Set a single config value |
| `shellai models` | List available models from your provider |
| `shellai context` | Show detected environment (OS, shell, CWD, etc.) |
| `shellai explain "cmd"` | Explain what a shell command does |
| `shellai history` | Show past commands (use `--limit N` to control count) |
| `shellai clear` | Reset config and wipe history |
| `shellai version` | Print version info |

### Flags

| Flag | Description |
|------|-------------|
| `-y, --yes` | Auto-confirm non-destructive, non-risky commands |
| `--model` | Override the model for this run |
| `--provider` | Override the provider (`ollama` or `llamacpp`) |
| `--endpoint` | Override the API endpoint URL |

---

## Configuration

Config is stored at `~/.config/shellai/config.yaml`. Run `shellai set` to configure interactively, or edit the file directly.

| Key | Default | Description |
|-----|---------|-------------|
| `provider` | `ollama` | `ollama` or `llamacpp` |
| `endpoint` | `http://localhost:11434` | Provider API URL |
| `model` | _(first available)_ | Model name to use |
| `timeout` | `60s` | Request timeout |

**Example:**
```bash
shellai set provider ollama
shellai set model llama3.2
shellai set endpoint http://localhost:11434
```

---

## Tips

**shellai knows your working directory.** The LLM receives your current path as context, so you can reference files relatively:

```bash
shellai "list files in the parent directory"
# → ls ..

shellai "find logs from yesterday in ~/Downloads"
# → find ~/Downloads -name "*.log" -newer ~/Downloads -mtime -1
```

**Use `-y` to skip confirmation** for routine commands you trust:

```bash
shellai -y "show running processes sorted by memory"
```

Commands involving pipes to a shell, `eval`, `exec`, or downloads are always confirmed regardless of `-y`.

**Destructive commands get an extra warning.** Anything that looks like it could delete, format, or overwrite data shows a red warning and requires explicit confirmation.

**Switch models on the fly:**

```bash
shellai --model llama3.1:70b "optimise this directory's git history"
```

---

## How It Works

1. Detects your environment (OS, distro, shell, package manager, working directory)
2. Sends that context + your intent to the local LLM
3. Parses the JSON response `{"command": "...", "reason": "..."}`
4. Runs safety checks (destructive patterns, package manager mismatch)
5. Shows the command + asks for confirmation
6. Executes and logs to `~/.config/shellai/history.json`

All LLM traffic stays local — shellai only talks to `localhost` by default.

---

## Supported Providers

| Provider | Default endpoint | Notes |
|----------|-----------------|-------|
| Ollama | `http://localhost:11434` | Auto-selects first available model |
| llama.cpp | `http://localhost:8080` | Uses OpenAI-compatible `/v1/` API |
