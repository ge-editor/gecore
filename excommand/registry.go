// gecore/excommand/registry.go

package excommand

import "sync"

type Registry struct {
	mu       sync.RWMutex
	commands map[string]ExCommand
}

func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]ExCommand)}
}

func (r *Registry) Register(cmd ExCommand) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands[cmd.Name()] = cmd
	for _, alias := range cmd.Aliases() {
		r.commands[alias] = cmd
	}
}
