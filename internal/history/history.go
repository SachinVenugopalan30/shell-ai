package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Intent    string    `json:"intent"`
	Command   string    `json:"command"`
	Reason    string    `json:"reason"`
	ExitCode  int       `json:"exit_code"`
	Executed  bool      `json:"executed"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "shellai", "history.json")
}

func Load(path string, limit int) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	// return last N entries
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	return entries, nil
}

func Append(path string, e Entry) error {
	entries, _ := Load(path, 0)
	entries = append(entries, e)
	if len(entries) > 1000 {
		entries = entries[len(entries)-1000:]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	// atomic write
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func Clear(path string) error {
	return os.Remove(path)
}
