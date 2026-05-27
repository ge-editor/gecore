package excommand

type Candidate struct {
	Value       string
	Description string

	MatchedIndexes []int
	Score          int
}
