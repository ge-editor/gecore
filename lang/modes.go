package lang

import (
	"github.com/ge-editor/gecore/verb"
)

var Modes = newModes()

func newModes() *ModesStruct {
	return &ModesStruct{
		names: map[string]int{},
		modes: []Mode{},
	}
}

type ModesStruct struct {
	modes []Mode
	names map[string]int
}

// Register Mode
func (ms *ModesStruct) Register(mode Mode) {
	verb.PP("mode %v", mode)
	index := len(ms.modes)
	ms.modes = append(ms.modes, mode)
	ms.names[mode.Name()] = index // Allow mode to be retrieved from mode name
}

func (ms ModesStruct) GetModes() []Mode {
	return ms.modes
}

// Return *Mode by Mode Name
func (ms *ModesStruct) GetModeByName(name string) *Mode {
	return &(ms.modes[ms.names[name]])
}

// Return the first registered *Mode as default mode
func (ms *ModesStruct) GetDefaultMode() *Mode {
	return &(ms.modes[0])
}

// Return the first registered *Mode as default mode
func (ms *ModesStruct) GetMode(filePath string) *Mode {
	for i := 1; i < len(ms.modes); i++ {
		if ms.modes[i].HasMatchingExtension(filePath) {
			return &ms.modes[i]
		}
	}
	return &(ms.modes[0]) // file type is Fundamental
}
