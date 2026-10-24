// gecore/excommand/resolve.go
package excommand

import (
	"context"
	"strings"
)

// Candidates removes shadowed command names and aliases from an already
// ordered list. Callers provide commands in resolution order, nearest first.
func Candidates(commands []ExCommand) []ExCommand {
	seen := make(map[string]bool)
	result := make([]ExCommand, 0, len(commands))
	for _, cmd := range commands {
		if cmd == nil {
			continue
		}

		// Keep the first command for each name/alias: callers order nearest first.
		names := append([]string{cmd.Name()}, cmd.Aliases()...)
		shadowed := false
		for _, name := range names {
			if seen[name] {
				shadowed = true
				break
			}
		}
		if shadowed {
			continue
		}

		result = append(result, cmd)
		for _, name := range names {
			seen[name] = true
		}
	}
	return result
}

// Execute resolves a command from an already ordered command list. A command
// returning handled=false allows resolution to continue.
func Execute(commands []ExCommand, ctx context.Context, name string, args []string) (bool, error) {
	for _, cmd := range commands {
		if !matchesName(cmd, name) {
			continue
		}
		handled, err := cmd.Execute(ctx, args)
		if handled || err != nil {
			return handled, err
		}
	}
	return false, nil
}

// Find returns the first command in an already ordered list (nearest first)
// whose name or alias equals name, ignoring case. It returns nil if none does.
func Find(commands []ExCommand, name string) ExCommand {
	for _, cmd := range commands {
		if matchesName(cmd, name) {
			return cmd
		}
	}
	return nil
}

func matchesName(cmd ExCommand, name string) bool {
	if cmd == nil {
		return false
	}
	if strings.EqualFold(cmd.Name(), name) {
		return true
	}
	for _, alias := range cmd.Aliases() {
		if strings.EqualFold(alias, name) {
			return true
		}
	}
	return false
}
