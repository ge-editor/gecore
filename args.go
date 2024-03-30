package gecore

import (
	"os"
	"strings"
)

// Arguments groups
var (
	WorkSpaces []string // directory
	Switches   []string // prefix -
	Files      []string
)

func init() {
	// Parese arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if info, err := os.Stat(arg); err == nil && info.IsDir() {
			WorkSpaces = append(WorkSpaces, arg)
		} else if strings.HasPrefix(arg, "-") {
			Switches = append(Switches, arg)
		} else {
			Files = append(Files, arg)
		}
	}
}
