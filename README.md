# shellai

Translate natural language into shell commands using a local LLM.

```
shellai "find all files larger than 100MB"
  Command:  find . -size +100M -type f
  Reason:   find files over 100MB in current dir
```

No data leaves your machine — shellai runs entirely against a local model via [Ollama](https://ollama.com) or [llama.cpp](https://github.com/ggerganov/llama.cpp).

---

## What made me do this?

I saw a video recently by Some Ordinary Gamers working on something similar, see [this video](https://youtu.be/IOf1vjIhr6Y?t=534), and thought: why not? I can make one myself as well!


## Requirements

- [Ollama](https://ollama.com) running locally (default), **or** a [llama.cpp](https://github.com/ggerganov/llama.cpp) server
- A pulled model — e.g. `ollama pull llama3.2`
- Go 1.21+ if installing via `go install`

---

## Installation

### Option 1 — go install

```bash
go install github.com/SachinVenugopalan30/shell-ai/cmd/shellai@latest
```

The binary lands in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your `$PATH`.

### Option 2 — pre-built binary

Download the latest release for your OS and architecture from [GitHub Releases](https://github.com/SachinVenugopalan30/shell-ai/releases), then move the binary to somewhere on your `$PATH`:

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
| `-v, --version` | Print version info and exit |
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
| `cache_ttl` | `30m` | How long a past command stays eligible for cache reuse |

**Example:**
```bash
shellai set provider ollama
shellai set model llama3.2
shellai set endpoint http://localhost:11434
shellai set cache_ttl 30m
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

## Recent-Command Cache

If you ask for something similar to what you ran a few minutes ago, shellai skips the LLM call and offers the prior command back:

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

- **Run cached command** — executes immediately, no LLM round-trip.
- **Send to LLM again** — falls through to the normal flow if the cached command isn't quite right.
- **Abort** — cancel and exit.

For destructive commands the first option becomes `Yes, run cached (destructive)` and a red warning is shown.

### How the lookup works

1. Loads `~/.config/shellai/history.json` (the same file used by `shellai history`).
2. Filters entries that:
   - actually executed,
   - exited cleanly (`exit_code <= 1` — `1` is treated as "no results", consistent with how grep / lsof report misses),
   - happened within `cache_ttl` (default 30 minutes).
3. For each remaining entry, normalizes both intents (lowercase, split on whitespace) and computes **Jaccard similarity** — the fraction of shared tokens over total unique tokens.
4. Picks the entry with the highest score, requiring `>= 0.6` to count as a match. Ties go to the most recent run.

Because the cache is just a view over `history.json`, it persists across shell sessions and survives reboots. There is no separate cache file. To invalidate, delete history (`shellai clear`) or wait for the TTL to expire.

Tune the window to your workflow:

```bash
shellai set cache_ttl 5m    # short-lived, only catches back-to-back retries
shellai set cache_ttl 2h    # longer memory across a work session
```

---

## How It Works

1. Detects your environment (OS, distro, shell, package manager, working directory)
2. Looks up recent history for a similar prior intent — if found, prompts to reuse it (skipping the LLM)
3. Sends environment + intent to the local LLM
4. Parses the JSON response `{"command": "...", "reason": "..."}`
5. Runs safety checks (destructive patterns, package manager mismatch)
6. Shows the command + asks for confirmation
7. Executes and logs to `~/.config/shellai/history.json`

> [!NOTE]
> The responses you get from running these commands will be very dependant on the model you are running. The better model you run, the better and more accurate commands you are going to get.


> [!CAUTION]
> I will preface this by saying that this something to help people with basic commands and not to be used with complicated tasks through the terminal. I am not responsible for you running commands generated by these models that may cause harm in some way to your system, it is your responsibility to research what the commands do to verify them before you run them. I am also not responsible if somehow your request to the model starts a thermonuclear war.


All LLM traffic stays local — shellai only talks to `localhost` by default.

---

## Supported Providers

| Provider | Default endpoint | Notes |
|----------|-----------------|-------|
| Ollama | `http://localhost:11434` | Auto-selects first available model |
| llama.cpp | `http://localhost:8080` | Uses OpenAI-compatible `/v1/` API |
