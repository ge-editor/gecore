package lang

// var Modes *ModesStruct

type Mode interface {
	Name() string

	HasMatchingExtension(filePath string) bool

	// Format formats source code and restores the cursor position.
	// The cursor position based on byte indices
	// Format(source [][]byte, cursorRow, cursorCol int) ([][]byte, int, int, error)

	// FormatBeforeSave(source [][]byte, cursorRow, cursorCol int) ([][]byte, int, int, error)

	Format(source [][]byte) ([][]byte, error)

	FormatBeforeSave(source [][]byte) ([][]byte, error)

	IndentWidth() int
	IsSoftTAB() bool
}
