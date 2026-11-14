package excommand

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/ge-editor/utils"
)

// Complete returns command names and aliases containing the input characters
// in order. Matching is case-insensitive. Results are sorted by value
// so completion order is stable even when global commands come from a map.
func Complete(commands []ExCommand, input string) []Candidate {
	seen := make(map[string]struct{})
	var candidates []Candidate

	for _, cmd := range Candidates(commands) {
		if cmd == nil {
			continue
		}

		// value := strings.Join(append([]string{cmd.Name()}, cmd.Aliases()...), "/")
		// value := cmd.Name()
		if cmd.Name() == "" {
			continue
		}

		key := strings.ToLower(cmd.Name())
		if _, ok := seen[key]; ok {
			continue
		}
		if !utils.ContainsAllCharacters(cmd.Name(), input) {
			continue
		}

		seen[key] = struct{}{}
		candidates = append(candidates, Candidate{
			Value: cmd.Name(),
			// Description:    cmd.Description(),
			Command:        cmd,
			MatchedIndexes: matchingIndexes(cmd.Name(), input),
			Score:          utf8.RuneCountInString(input),
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return strings.ToLower(candidates[i].Value) < strings.ToLower(candidates[j].Value)
	})
	return candidates
}

/*
func Complete(commands []ExCommand, input string) []Candidate {
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
			if !utils.ContainsAllCharacters(value, input) {
				continue
			}

			seen[key] = struct{}{}
			candidates = append(candidates, Candidate{
				Value: value,
				// Description:    cmd.Description(),
				Command:        cmd,
				MatchedIndexes: matchingIndexes(value, input),
				Score:          utf8.RuneCountInString(input),
			})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return strings.ToLower(candidates[i].Value) < strings.ToLower(candidates[j].Value)
	})
	return candidates
}
*/

func matchingIndexes(value, input string) []int {
	valueRunes := []rune(strings.ToLower(value))
	inputRunes := []rune(strings.ToLower(input))
	indexes := make([]int, 0, len(inputRunes))
	next := 0
	for _, needle := range inputRunes {
		for next < len(valueRunes) && valueRunes[next] != needle {
			next++
		}
		if next == len(valueRunes) {
			return nil
		}
		indexes = append(indexes, next)
		next++
	}
	return indexes
}
