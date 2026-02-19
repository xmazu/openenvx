package envfile

type LineType int

const (
	LineTypeEmpty LineType = iota
	LineTypeComment
	LineTypeCommentedAssignment
	LineTypeVariable
	LineTypeInvalid
)

type Line struct {
	Type          LineType
	Num           int
	Raw           string
	Key           string
	Value         string
	InlineComment string
}
