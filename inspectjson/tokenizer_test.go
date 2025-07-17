package inspectjson

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/dpb587/cursorio-go/cursorio"
)

type testCaseLaxBehavior struct {
	Name              string
	Input             string
	LaxSanitized      string
	LaxSyntaxRecovery SyntaxRecovery
}

func testLaxBehavior(t *testing.T, tcl ...testCaseLaxBehavior) {
	for _, tc := range tcl {
		t.Run(fmt.Sprintf("%s/%s", tc.LaxSyntaxRecovery.Behavior, tc.Name), func(t *testing.T) {
			var syntaxRecoveryList []SyntaxRecovery

			sanitizedBytes, err := io.ReadAll(
				NewTokenizerReader(
					NewTokenizer(
						strings.NewReader(tc.Input),
						TokenizerConfig{}.
							SetLax(true).
							SetSourceOffsets(true).
							SetSyntaxRecoveryHook(func(event SyntaxRecovery) {
								syntaxRecoveryList = append(syntaxRecoveryList, event)
							}),
					),
				),
			)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			} else if _a, _e := string(sanitizedBytes), tc.LaxSanitized; _a != _e {
				t.Fatalf("sanitized: expected %q, got %q", _e, _a)
			} else if _a, _e := len(syntaxRecoveryList), 1; _a != _e {
				t.Fatalf("syntax recovery length: expected %d, got %d", _e, _a)
			}

			actualSyntaxRecovery := syntaxRecoveryList[0]
			if _a, _e := actualSyntaxRecovery.Behavior, tc.LaxSyntaxRecovery.Behavior; _a != _e {
				t.Fatalf("behavior: expected %q, got %q", _e, _a)
			} else if _a, _e := actualSyntaxRecovery.SourceOffsets.OffsetRangeString(), tc.LaxSyntaxRecovery.SourceOffsets.OffsetRangeString(); _a != _e {
				t.Fatalf("source offsets: expected %q, got %q", _e, _a)
			} else if _a, _e := string(actualSyntaxRecovery.SourceRunes), string(tc.LaxSyntaxRecovery.SourceRunes); _a != _e {
				t.Fatalf("source runes: expected %q, got %q", _e, _a)
			} else if _a, _e := string(actualSyntaxRecovery.ReplacementRunes), string(tc.LaxSyntaxRecovery.ReplacementRunes); _a != _e {
				t.Fatalf("replacement runes: expected %q, got %q", _e, _a)
			}

			if actualValueStart, expectedValueStart := actualSyntaxRecovery.ValueStart, tc.LaxSyntaxRecovery.ValueStart; actualValueStart != nil || expectedValueStart != nil {
				if actualValueStart == nil || expectedValueStart == nil {
					t.Fatalf("value start: expected %q, got %q", expectedValueStart, actualValueStart)
				} else if _a, _e := actualValueStart.OffsetString(), expectedValueStart.OffsetString(); _a != _e {
					t.Fatalf("value start: expected %q, got %q", _e, _a)
				}
			}
		})
	}
}

func TestTokenizer_SyntaxRecovery_LaxIgnoreBlockComment(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "Leading",
			Input:        `/* comment */[]`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreBlockComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 13, LineColumn: cursorio.TextLineColumn{0, 13}},
				},
				SourceRunes: []rune("/* comment */"),
			},
		},
		testCaseLaxBehavior{
			Name:         "Trailing",
			Input:        `[]/* comment */`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreBlockComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
					Until: cursorio.TextOffset{Byte: 15, LineColumn: cursorio.TextLineColumn{0, 15}},
				},
				SourceRunes: []rune("/* comment */"),
			},
		},
		testCaseLaxBehavior{
			Name:         "UnclosedEOF0",
			Input:        `[]/* comment`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreBlockComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
					Until: cursorio.TextOffset{Byte: 12, LineColumn: cursorio.TextLineColumn{0, 12}},
				},
				SourceRunes: []rune("/* comment"),
			},
		},
		testCaseLaxBehavior{
			Name:         "UnclosedEOF1",
			Input:        `[]/* comment *`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreBlockComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
					Until: cursorio.TextOffset{Byte: 14, LineColumn: cursorio.TextLineColumn{0, 14}},
				},
				SourceRunes: []rune("/* comment *"),
			},
		},
		testCaseLaxBehavior{
			Name:         "Inset",
			Input:        `[/* comment */]`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreBlockComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 14, LineColumn: cursorio.TextLineColumn{0, 14}},
				},
				SourceRunes: []rune("/* comment */"),
			},
		},
		testCaseLaxBehavior{
			Name: "MultilineFalseEnds",
			Input: `/* comment
/ /
* /
*/true`,
			LaxSanitized: `true`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreBlockComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 21, LineColumn: cursorio.TextLineColumn{3, 2}},
				},
				SourceRunes: []rune("/* comment\n/ /\n* /\n*/"),
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_LaxIgnoreLineComment(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name: "Leading",
			Input: `// comment
[]`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreLineComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 10, LineColumn: cursorio.TextLineColumn{0, 10}},
				},
				SourceRunes: []rune("// comment"),
			},
		},
		testCaseLaxBehavior{
			Name:         "Trailing",
			Input:        `[]// comment`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreLineComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
					Until: cursorio.TextOffset{Byte: 12, LineColumn: cursorio.TextLineColumn{0, 12}},
				},
				SourceRunes: []rune("// comment"),
			},
		},
		testCaseLaxBehavior{
			Name: "Inset",
			Input: `[// comment
]`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreLineComment,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 11, LineColumn: cursorio.TextLineColumn{0, 11}},
				},
				SourceRunes: []rune("// comment"),
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_LaxStringEscapeInvalidEscape(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "One",
			Input:        `"\z"`,
			LaxSanitized: `"\\z"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxStringEscapeInvalidEscape,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 3, LineColumn: cursorio.TextLineColumn{0, 3}},
				},
				SourceRunes:      []rune("\\z"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\z"),
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_LaxStringEscapeMissingEscape(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "U+0008",
			Input:        "\"\b\"",
			LaxSanitized: `"\b"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxStringEscapeMissingEscape,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes:      []rune("\b"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\b"),
			},
		},
		testCaseLaxBehavior{
			Name:         "U+000C",
			Input:        "\"\f\"",
			LaxSanitized: `"\f"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxStringEscapeMissingEscape,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes:      []rune("\f"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\f"),
			},
		},
		testCaseLaxBehavior{
			Name:         "U+000A",
			Input:        "\"\n\"",
			LaxSanitized: `"\n"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxStringEscapeMissingEscape,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{1, 0}},
				},
				SourceRunes:      []rune("\n"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\n"),
			},
		},
		testCaseLaxBehavior{
			Name:         "U+000D",
			Input:        "\"\r\"",
			LaxSanitized: `"\r"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxStringEscapeMissingEscape,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 1}},
				},
				SourceRunes:      []rune("\r"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\r"),
			},
		},
		testCaseLaxBehavior{
			Name:         "U+0009",
			Input:        "\"\t\"",
			LaxSanitized: `"\t"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxStringEscapeMissingEscape,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes:      []rune("\t"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\t"),
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_LaxNumberTrimLeadingZero(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "One",
			Input:        `01`,
			LaxSanitized: `1`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxNumberTrimLeadingZero,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
				},
				SourceRunes: []rune("0"),
				ValueStart:  &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
			},
		},
		testCaseLaxBehavior{
			Name:         "NegativeOne",
			Input:        `-01`,
			LaxSanitized: `-1`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxNumberTrimLeadingZero,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes: []rune("0"),
				ValueStart:  &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_LaxLiteralCaseInsensitive(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "Null",
			Input:        `Null`,
			LaxSanitized: `null`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxLiteralCaseInsensitive,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 4, LineColumn: cursorio.TextLineColumn{0, 4}},
				},
				SourceRunes:      []rune("Null"),
				ReplacementRunes: []rune("null"),
			},
		},
		testCaseLaxBehavior{
			Name:         "nuLl",
			Input:        `nuLl`,
			LaxSanitized: `null`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxLiteralCaseInsensitive,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 4, LineColumn: cursorio.TextLineColumn{0, 4}},
				},
				SourceRunes:      []rune("nuLl"),
				ReplacementRunes: []rune("null"),
			},
		},
		testCaseLaxBehavior{
			Name:         "True",
			Input:        `True`,
			LaxSanitized: `true`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxLiteralCaseInsensitive,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 4, LineColumn: cursorio.TextLineColumn{0, 4}},
				},
				SourceRunes:      []rune("True"),
				ReplacementRunes: []rune("true"),
			},
		},
		testCaseLaxBehavior{
			Name:         "trUe",
			Input:        `trUe`,
			LaxSanitized: `true`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxLiteralCaseInsensitive,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 4, LineColumn: cursorio.TextLineColumn{0, 4}},
				},
				SourceRunes:      []rune("trUe"),
				ReplacementRunes: []rune("true"),
			},
		},
		testCaseLaxBehavior{
			Name:         "False",
			Input:        `False`,
			LaxSanitized: `false`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxLiteralCaseInsensitive,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 5, LineColumn: cursorio.TextLineColumn{0, 5}},
				},
				SourceRunes:      []rune("False"),
				ReplacementRunes: []rune("false"),
			},
		},
		testCaseLaxBehavior{
			Name:         "falSe",
			Input:        `falSe`,
			LaxSanitized: `false`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxLiteralCaseInsensitive,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
					Until: cursorio.TextOffset{Byte: 5, LineColumn: cursorio.TextLineColumn{0, 5}},
				},
				SourceRunes:      []rune("falSe"),
				ReplacementRunes: []rune("false"),
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_LaxIgnoreExtraComma(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "ArrayEmpty",
			Input:        `[,]`,
			LaxSanitized: `[]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreExtraComma,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes: []rune(","),
			},
		},
		testCaseLaxBehavior{
			Name:         "ArrayLeading",
			Input:        `[,true]`,
			LaxSanitized: `[true]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreExtraComma,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes: []rune(","),
			},
		},
		testCaseLaxBehavior{
			Name:         "ArrayTrailing",
			Input:        `[true,]`,
			LaxSanitized: `[true]`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreExtraComma,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 5, LineColumn: cursorio.TextLineColumn{0, 5}},
					Until: cursorio.TextOffset{Byte: 6, LineColumn: cursorio.TextLineColumn{0, 6}},
				},
				SourceRunes: []rune(","),
			},
		},
		testCaseLaxBehavior{
			Name:         "ObjectEmpty",
			Input:        `{,}`,
			LaxSanitized: `{}`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreExtraComma,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes: []rune(","),
			},
		},
		testCaseLaxBehavior{
			Name:         "ObjectLeading",
			Input:        `{,"n":true}`,
			LaxSanitized: `{"n":true}`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreExtraComma,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 2, LineColumn: cursorio.TextLineColumn{0, 2}},
				},
				SourceRunes: []rune(","),
			},
		},
		testCaseLaxBehavior{
			Name:         "ObjectTrailing",
			Input:        `{"n":true,}`,
			LaxSanitized: `{"n":true}`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: LaxIgnoreExtraComma,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 9, LineColumn: cursorio.TextLineColumn{0, 9}},
					Until: cursorio.TextOffset{Byte: 10, LineColumn: cursorio.TextLineColumn{0, 10}},
				},
				SourceRunes: []rune(","),
			},
		},
	)
}

func TestTokenizer_SyntaxRecovery_WarnStringUnicodeReplacementChar(t *testing.T) {
	testLaxBehavior(t,
		testCaseLaxBehavior{
			Name:         "SurrogateIncomplete",
			Input:        `"\uD834"`,
			LaxSanitized: `"\uFFFD"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: WarnStringUnicodeReplacementChar,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 7, LineColumn: cursorio.TextLineColumn{0, 7}},
				},
				SourceRunes:      []rune("\\uD834"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\uFFFD"),
			},
		},
		testCaseLaxBehavior{
			Name:         "SurrogateInvalid",
			Input:        `"\uD834\u0061"`,
			LaxSanitized: `"\uFFFD"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: WarnStringUnicodeReplacementChar,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 13, LineColumn: cursorio.TextLineColumn{0, 13}},
				},
				SourceRunes:      []rune("\\uD834\\u0061"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\uFFFD"),
			},
		},
		testCaseLaxBehavior{
			Name:         "SurrogateIncompleteFalseContinuance",
			Input:        `"\uD834\t"`,
			LaxSanitized: `"\uFFFD\t"`,
			LaxSyntaxRecovery: SyntaxRecovery{
				Behavior: WarnStringUnicodeReplacementChar,
				SourceOffsets: &cursorio.TextOffsetRange{
					From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
					Until: cursorio.TextOffset{Byte: 7, LineColumn: cursorio.TextLineColumn{0, 7}},
				},
				SourceRunes:      []rune("\\uD834"),
				ValueStart:       &cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				ReplacementRunes: []rune("\\uFFFD"),
			},
		},
	)
}
