package styleresolver

type RegionKind uint16

const (
	RegionUnknown RegionKind = iota

	// ---- interactive ----

	RegionURL
	RegionFilePath
	RegionEmail

	// markdown
	RegionMarkdownLink
	RegionMarkdownImage
	RegionMarkdownCode
	RegionMarkdownCodeBlock

	// ---- syntax ----

	RegionKeyword
	RegionString
	RegionComment
	RegionNumber
	RegionOperator
	RegionType
	RegionFunction
	RegionVariable
	RegionProperty

	// ---- semantic token ----

	RegionLSPFunction
	RegionLSPMethod
	RegionLSPParameter
	RegionLSPVariable
	RegionLSPType
	RegionLSPNamespace

	// ---- diagnostics ----

	RegionError
	RegionWarning
	RegionInfo
	RegionHint

	// ---- editor ----

	RegionSearch
	RegionSearchCurrent

	RegionSelection
	RegionCursorLine
	RegionCurrentLine

	RegionDiffAdded
	RegionDiffModified
	RegionDiffDeleted

	RegionReadonly

	// ---- ui ----

	RegionFolded
	RegionTrailingWhitespace
	RegionIndentGuide
	RegionMatchingBracket
)

type InteractiveRegion struct {
	Start int
	End   int

	Kind    RegionKind
	Payload any
}
