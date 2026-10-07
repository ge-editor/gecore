package excommand

// Candidate is a completion result for an extended command.
type Candidate struct {
	Value string
	// Description    string
	Command        ExCommand
	MatchedIndexes []int
	Score          int
}
