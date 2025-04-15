package inspectjson

import "github.com/dpb587/cursorio-go/cursorio"

type SyntaxBehavior int

const (
	LaxIgnoreBlockComment            = 0b1
	LaxIgnoreLineComment             = 0b10
	LaxStringEscapeInvalidEscape     = 0b100
	LaxStringEscapeMissingEscape     = 0b1000
	LaxNumberTrimLeadingZero         = 0b10000
	LaxLiteralCaseInsensitive        = 0b100000
	LaxIgnoreExtraComma              = 0b1000000
	LaxIgnoreTrailingSemicolon       = 0b10000000
	WarnStringUnicodeReplacementChar = 0b100000000
)

func (i SyntaxBehavior) String() string {
	switch i {
	case LaxIgnoreBlockComment:
		return "LaxIgnoreBlockComment"
	case LaxIgnoreLineComment:
		return "LaxIgnoreLineComment"
	case LaxStringEscapeInvalidEscape:
		return "LaxStringEscapeInvalidEscape"
	case LaxStringEscapeMissingEscape:
		return "LaxStringEscapeMissingEscape"
	case LaxNumberTrimLeadingZero:
		return "LaxNumberTrimLeadingZero"
	case LaxLiteralCaseInsensitive:
		return "LaxLiteralCaseInsensitive"
	case LaxIgnoreExtraComma:
		return "LaxIgnoreExtraComma"
	case LaxIgnoreTrailingSemicolon:
		return "LaxIgnoreTrailingSemicolon"
	case WarnStringUnicodeReplacementChar:
		return "WarnStringUnicodeReplacementChar"
	default:
		return "unknown"
	}
}

//

type SyntaxRecovery struct {
	Behavior SyntaxBehavior

	SourceOffsets *cursorio.TextOffsetRange
	SourceRunes   []rune

	ValueStart       *cursorio.TextOffset
	ReplacementRunes []rune
}

//

type SyntaxRecoveryHookFunc func(event SyntaxRecovery)
