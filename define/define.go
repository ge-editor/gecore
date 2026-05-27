package define

const (
	LOTUS = '🪷'
	GHOST = '👻'
	EOF   = 0x1a //  26 0x1a ^Z SUB (置換)
	DEL   = 0x7f // 127 0x7f ^? DEL
	// LF    = 0x0a
	// CR    = 0x0d
	// CRLF  = 0x0d0x0a

	NO_BREAK_SPACE = 0xC2A0 // no-break space
)

var (
	NewLineLF   = []byte{'\n'}
	NewLineCR   = []byte{'\r'}
	NewLineCRLF = []byte{'\r', '\n'}
)
