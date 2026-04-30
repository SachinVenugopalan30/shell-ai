package cmd

import (
	"strings"
	"time"

	"github.com/SachinVenugopalan30/shell-ai/internal/history"
)

const simThreshold = 0.6

// findCachedMatch scans recent history for an entry with a similar intent.
// Returns the best-scoring eligible entry (recency breaks ties) or false.
func findCachedMatch(intent string, ttl time.Duration) (*history.Entry, bool) {
	entries, err := history.Load(history.DefaultPath(), 0)
	if err != nil {
		return nil, false
	}

	now := time.Now()
	var best *history.Entry
	bestScore := 0.0

	// iterate in reverse so newer entries win on score ties
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if !e.Executed || e.ExitCode < 0 || e.ExitCode > 1 {
			continue
		}
		if now.Sub(e.Timestamp) > ttl {
			continue
		}
		score := similarity(intent, e.Intent)
		if score >= simThreshold && score > bestScore {
			entry := e
			best = &entry
			bestScore = score
		}
	}

	return best, best != nil
}

// similarity returns Jaccard similarity over normalized tokens.
func similarity(a, b string) float64 {
	ta := tokenize(a)
	tb := tokenize(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}

	setA := make(map[string]struct{}, len(ta))
	for _, t := range ta {
		setA[t] = struct{}{}
	}
	setB := make(map[string]struct{}, len(tb))
	for _, t := range tb {
		setB[t] = struct{}{}
	}

	inter := 0
	for t := range setA {
		if _, ok := setB[t]; ok {
			inter++
		}
	}
	union := len(setA) + len(setB) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func tokenize(s string) []string {
	return strings.Fields(strings.ToLower(s))
}
