package excommand

import "sort"

var globalRegistry = NewRegistry()

// RegisterGlobal registers a command available from every leaf.
func RegisterGlobal(cmd ExCommand) {
	globalRegistry.Register(cmd)
}

// GlobalCommands returns each registered global command once.
func GlobalCommands() []ExCommand {
	return globalRegistry.Commands()
}

// Commands returns registered commands once each, regardless of aliases.
func (r *Registry) Commands() []ExCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]struct{}, len(r.commands))
	commands := make([]ExCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		if _, ok := seen[cmd.Name()]; ok {
			continue
		}
		seen[cmd.Name()] = struct{}{}
		commands = append(commands, cmd)
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})
	return commands
}
