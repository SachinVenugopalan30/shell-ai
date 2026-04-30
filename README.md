# shellai

Turn plain English into shell commands, all locally.

```
shellai "find all files larger than 100MB"
  Command:  find . -size +100M -type f
  Reason:   find files over 100MB in current dir
```

Nothing leaves your machine. shellai runs entirely against a local model via [Ollama](https://ollama.com) or [llama.cpp](https://github.com/ggerganov/llama.cpp).

---

## What made me do this?

Saw a video by Some Ordinary Gamers working on something similar ([timestamped link here](https://youtu.be/IOf1vjIhr6Y?t=534)) and thought: hey, why not? I can build one too.


## Requirements

- [Ollama](https://ollama.com) running locally (default), **or** a [llama.cpp](https://github.com/ggerganov/llama.cpp) server
- A pulled model, e.g. `ollama pull gemma4:e4b`
- Go 1.21+ if you're installing via `go install`

---

## Installation

### Option 1: go install

```bash
go install github.com/SachinVenugopalan30/shell-ai/cmd/shellai@latest
```

The binary lands in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your `$PATH`.

### Option 2: pre-built binary

Grab the latest release for your OS and architecture from [GitHub Releases](https://github.com/SachinVenugopalan30/shell-ai/releases), then drop the binary somewhere on your `$PATH`:

```bash
# example for macOS arm64
mv shellai_darwin_arm64 /usr/local/bin/shellai
chmod +x /usr/local/bin/shellai
```

---

## Quick Start

```bash
# 1. Pull a model
ollama pull gemma4:e4b  # recommend this model for fast inference and **mostly** accurate command results

# 2. Configure shellai (one-time setup)
shellai set

# 3. Ask for a command
shellai "show disk usage for each folder here"
```

shellai shows you the generated command and a short reason, then asks you to confirm before running anything.

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
| `-v, --version` | Print version info and exit |
| `--model` | Override the model for this run |
| `--provider` | Override the provider (`ollama` or `llamacpp`) |
| `--endpoint` | Override the API endpoint URL |

---

## Configuration

Config lives at `~/.config/shellai/config.yaml`. Run `shellai set` to configure it interactively, or just edit the file directly.

| Key | Default | Description |
|-----|---------|-------------|
| `provider` | `ollama` | `ollama` or `llamacpp` |
| `endpoint` | `http://localhost:11434` | Provider API URL |
| `model` | _(first available)_ | Model name to use |
| `timeout` | `60s` | Request timeout |
| `cache_ttl` | `30m` | How long a past command stays eligible for cache reuse |

**Example:**
```bash
shellai set provider ollama
shellai set model gemma4:e4b
shellai set endpoint http://localhost:11434
shellai set cache_ttl 30m
```

---

## Tips

**shellai knows your working directory.** The LLM gets your current path as context, so you can reference files relatively:

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

Commands with pipes to a shell, `eval`, `exec`, or downloads always ask for confirmation, even with `-y`.

**Destructive commands get an extra warning.** Anything that looks like it could delete, format, or overwrite stuff shows a red warning and forces you to confirm.

**Switch models on the fly:**

```bash
shellai --model qwen3.6:35b "optimise this directory's git history"
```

---

## Recent-Command Cache

If you ask for something close to what you ran a few minutes ago, shellai skips the LLM call and offers the previous command back:

```
Found a recent similar request:
  You asked: list the largest 5 files in current dir  (2m0s ago)
  Command:   du -ah . | sort -rh | head -n 5
  Reason:    show top 5 biggest files by size

? What would you like to do?  [Use arrows to move, enter to select]
> Run cached command
  Send to LLM again
  Abort
```

- **Run cached command**: runs it right away, no LLM round-trip.
- **Send to LLM again**: falls through to the normal flow if the cached command isn't quite right.
- **Abort**: cancel and bail out.

For destructive commands the first option becomes `Yes, run cached (destructive)` and you get the usual red warning.

### How the lookup works

1. Loads `~/.config/shellai/history.json` (same file used by `shellai history`).
2. Keeps only entries that:
   - actually executed,
   - exited cleanly (`exit_code <= 1`, since `1` usually just means "no results" from things like grep or lsof),
   - happened within `cache_ttl` (default 30 minutes).
3. For each remaining entry, normalizes both intents (lowercase, split on whitespace) and computes **Jaccard similarity**: the fraction of shared tokens over total unique tokens.
4. Picks the entry with the highest score and requires `>= 0.6` to count as a match. Ties go to the most recent run.

Because the cache is just a view over `history.json`, it sticks around across shell sessions and survives reboots. No separate cache file. To invalidate, wipe history with `shellai clear` or wait for the TTL to lapse.

Tune the window to your workflow:

```bash
shellai set cache_ttl 5m    # short-lived, only catches back-to-back retries
shellai set cache_ttl 2h    # longer memory across a work session
```

---

## How It Works

1. Detects your environment (OS, distro, shell, package manager, working directory)
2. Checks recent history for a similar prior intent. If something matches, you get a prompt to reuse it (skipping the LLM)
3. Sends the environment + your intent to the local LLM
4. Parses the JSON response `{"command": "...", "reason": "..."}`
5. Runs safety checks (destructive patterns, package manager mismatch)
6. Shows the command and asks for confirmation
7. Executes and logs to `~/.config/shellai/history.json`

> [!NOTE]
> The quality of what you get back depends a lot on the model you're running. Better model in, better commands out.


> [!CAUTION]
> I will preface this by saying that this something to help people with basic commands and not to be used with complicated tasks through the terminal. I am not responsible for you running commands generated by these models that may cause harm in some way to your system, it is your responsibility to research what the commands do to verify them before you run them. I am also not responsible if somehow your request to the model starts a thermonuclear war.

---

## Supported Providers

| Provider | Default endpoint | Notes |
|----------|-----------------|-------|
| Ollama | `http://localhost:11434` | Auto-selects first available model |
| llama.cpp | `http://localhost:8080` | Uses OpenAI-compatible `/v1/` API |
