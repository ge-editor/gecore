package excommand

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// Complete returns command names and aliases matching the current input.
// Matching is case-insensitive prefix matching. Results are sorted by value
// so completion order is stable even when global commands come from a map.
func Complete(commands []ExCommand, input string) []Candidate {
	needle := strings.ToLower(input)
	seen := make(map[string]struct{})
	var candidates []Candidate

	for _, cmd := range Candidates(commands) {
		if cmd == nil {
			continue
		}

		values := append([]string{cmd.Name()}, cmd.Aliases()...)
		for _, value := range values {
			if value == "" {
				continue
			}
			key := strings.ToLower(value)
			if _, ok := seen[key]; ok {
				continue
			}
			if !strings.HasPrefix(key, needle) {
				continue
			}

			seen[key] = struct{}{}
			candidates = append(candidates, Candidate{
				Value: value,
				// Description:    cmd.Description(),
				Command:        cmd,
				MatchedIndexes: prefixIndexes(value, input),
				Score:          utf8.RuneCountInString(input),
			})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return strings.ToLower(candidates[i].Value) < strings.ToLower(candidates[j].Value)
	})
	return candidates
}

func prefixIndexes(value, input string) []int {
	indexes := make([]int, 0, utf8.RuneCountInString(input))
	for i := 0; i < utf8.RuneCountInString(input); i++ {
		indexes = append(indexes, i)
	}
	return indexes
}
