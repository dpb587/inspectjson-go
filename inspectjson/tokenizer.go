package inspectjson

import (
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf16"

	"github.com/dpb587/cursorio-go/cursorio"
	"github.com/dpb587/cursorio-go/x/cursorioutil"
)

type tokenizerOptions struct {
	sourceOffsets       bool
	sourceInitialOffset cursorio.TextOffset
	emitWhitespace      bool
	laxBehaviors        SyntaxBehavior
	laxListener         SyntaxRecoveryHookFunc
	multistream         bool
}

type Tokenizer struct {
	buf *cursorioutil.RuneBuffer
	doc *cursorio.TextWriter

	emitWhitespace bool
	laxBehaviors   SyntaxBehavior
	laxListener    SyntaxRecoveryHookFunc
	multistream    bool

	openValueSeparator *ValueSeparatorToken

	stack  []lexFunc
	tokens []Token
	err    error
}

type lexFunc func(t *Tokenizer, r0 rune, err error) (lexFunc, error)

func NewTokenizer(r io.Reader, opts ...TokenizerOption) *Tokenizer {
	compiledOpts := &tokenizerOptions{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt.applyTokenizer(compiledOpts)
	}

	t := &Tokenizer{
		buf:            cursorioutil.NewRuneBuffer(r),
		emitWhitespace: compiledOpts.emitWhitespace,
		laxBehaviors:   compiledOpts.laxBehaviors,
		laxListener:    compiledOpts.laxListener,
		multistream:    compiledOpts.multistream,
		stack:          []lexFunc{tokenizer_lexInit},
	}

	if compiledOpts.sourceOffsets {
		t.doc = cursorio.NewTextWriter(compiledOpts.sourceInitialOffset)
	}

	return t
}

func (t *Tokenizer) Next() (Token, error) {
	if len(t.tokens) > 0 {
		t.tokens = t.tokens[1:]
	}

	var fnNext lexFunc

	for {
		if len(t.tokens) > 0 {
			if fnNext != nil {
				t.stackShift(fnNext)
			}

			return t.tokens[0], nil
		} else if t.err != nil {
			if fnNext != nil {
				t.stackShift(fnNext)
			}

			return nil, t.err
		} else if fnNext == nil {
			if len(t.stack) == 0 {
				fnNext = tokenizer_lexEnded
			} else {
				fnNext = t.stack[len(t.stack)-1]
				t.stack = t.stack[:len(t.stack)-1]
			}
		}

		fnNext, t.err = t.lex(fnNext)
	}
}

func (t *Tokenizer) stackShift(fn lexFunc) error {
	t.stack = append(t.stack, fn)

	return nil
}

func (t *Tokenizer) lex(fn lexFunc) (lexFunc, error) {
	var uncommitted []rune

	for {
		r0, err := t.buf.NextRune()
		if err != nil {
			return fn(t, r0, err)
		}

		switch r0 {
		case '/':
			if len(uncommitted) > 0 {
				if t.emitWhitespace {
					t.emit(WhitespaceToken{
						SourceOffsets: t.commitForTextOffsetRange(uncommitted),
						Content:       string(uncommitted),
					})
				} else {
					t.commit(uncommitted)
				}

				uncommitted = nil
			}

			if t.laxBehaviors&(LaxIgnoreBlockComment|LaxIgnoreLineComment) == 0 {
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0,
				}, nil, []rune{r0})
			}

			uncommitted = append(uncommitted, r0)

			r1, err := t.buf.NextRune()
			if err != nil {
				return nil, err
			}

			switch r1 {
			case '/':
				if t.laxBehaviors&LaxIgnoreLineComment == 0 {
					return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r1,
					}, uncommitted, []rune{r1})
				}

				uncommitted = append(uncommitted, r1)

				for {
					r0, err := t.buf.NextRune()
					if err != nil {
						if errors.Is(err, io.EOF) {
							if t.laxListener != nil {
								t.laxListener(SyntaxRecovery{
									Behavior:      LaxIgnoreLineComment,
									SourceOffsets: t.commitForTextOffsetRange(uncommitted),
									SourceRunes:   uncommitted,
								})
							}
						}

						return nil, err
					}

					// TODO \r\n?
					if r0 == '\n' {
						t.buf.BacktrackRunes(r0)

						goto LINE_DONE
					}

					uncommitted = append(uncommitted, r0)
				}

			LINE_DONE:

				if t.laxListener != nil {
					t.laxListener(SyntaxRecovery{
						Behavior:      LaxIgnoreLineComment,
						SourceOffsets: t.commitForTextOffsetRange(uncommitted),
						SourceRunes:   uncommitted,
					})
				}

				return fn, nil
			case '*':
				if t.laxBehaviors&LaxIgnoreBlockComment == 0 {
					return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r1,
					}, uncommitted, []rune{r1})
				}

				uncommitted = append(uncommitted, r1)

				for {
					r0, err := t.buf.NextRune()
					if err != nil {
						if errors.Is(err, io.EOF) {
							if t.laxListener != nil {
								t.laxListener(SyntaxRecovery{
									Behavior:      LaxIgnoreBlockComment,
									SourceOffsets: t.commitForTextOffsetRange(uncommitted),
									SourceRunes:   uncommitted,
								})
							}
						}

						return nil, err
					}

					uncommitted = append(uncommitted, r0)

					if r0 == '*' {
						r1, err := t.buf.NextRune()
						if err != nil {
							if errors.Is(err, io.EOF) {
								if t.laxListener != nil {
									t.laxListener(SyntaxRecovery{
										Behavior:      LaxIgnoreBlockComment,
										SourceOffsets: t.commitForTextOffsetRange(uncommitted),
										SourceRunes:   uncommitted,
									})
								}
							}

							return nil, err
						}

						uncommitted = append(uncommitted, r1)

						if r1 == '/' {
							goto BLOCK_DONE
						}
					}
				}

			BLOCK_DONE:

				if t.laxListener != nil {
					t.laxListener(SyntaxRecovery{
						Behavior:      LaxIgnoreBlockComment,
						SourceOffsets: t.commitForTextOffsetRange(uncommitted),
						SourceRunes:   uncommitted,
					})
				}

				return fn, nil
			default:
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: uncommitted[0],
				}, uncommitted[0:1], []rune{r1})
			}
		case 0x20, 0x09, 0x0A, 0x0D:
			uncommitted = append(uncommitted, r0)
		default:
			if len(uncommitted) > 0 {
				if t.emitWhitespace {
					t.emit(WhitespaceToken{
						SourceOffsets: t.commitForTextOffsetRange(uncommitted),
						Content:       string(uncommitted),
					})
				} else {
					t.commit(uncommitted)
				}
			}

			return fn(t, r0, err)
		}
	}
}

func (r *Tokenizer) emit(tokens ...Token) {
	r.tokens = append(r.tokens, tokens...)
}

func (t *Tokenizer) emitString(r0 rune) error {
	// assert(r0 == '"')

	var uncommitted = []rune{r0}
	var decoded []rune

	for {
		r0, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}

			return err
		}

		switch r0 {
		case '\\':
			r1, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return io.ErrUnexpectedEOF
				}

				return err
			}

			switch r1 {
			case '"', '\\', '/':
				decoded = append(decoded, r1)
				uncommitted = append(uncommitted, r0, r1)
			case 'b':
				decoded = append(decoded, '\b')
				uncommitted = append(uncommitted, r0, r1)
			case 'f':
				decoded = append(decoded, '\f')
				uncommitted = append(uncommitted, r0, r1)
			case 'n':
				decoded = append(decoded, '\n')
				uncommitted = append(uncommitted, r0, r1)
			case 'r':
				decoded = append(decoded, '\r')
				uncommitted = append(uncommitted, r0, r1)
			case 't':
				decoded = append(decoded, '\t')
				uncommitted = append(uncommitted, r0, r1)
			case 'u':
				uncommitted = append(uncommitted, r0, r1)

				decodedRune, nextUncommitted, err := scanUnicode(t, uncommitted)
				if err != nil {
					if errors.Is(err, io.EOF) {
						return io.ErrUnexpectedEOF
					}

					return err
				}

				uncommitted = nextUncommitted

				if utf16.IsSurrogate(decodedRune) {
					r2, err := t.buf.NextRune()
					if err != nil {
						if errors.Is(err, io.EOF) {
							return io.ErrUnexpectedEOF
						}

						return err
					} else if r2 != '\\' {
						if t.laxListener != nil {
							t.laxListener(SyntaxRecovery{
								Behavior:         WarnStringUnicodeReplacementChar,
								SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-6], uncommitted[len(uncommitted)-6:]),
								SourceRunes:      uncommitted[len(uncommitted)-6:],
								ValueStart:       t.getTextOffset(),
								ReplacementRunes: []rune{'\\', 'u', 'F', 'F', 'F', 'D'},
							})
						}

						t.buf.BacktrackRunes(r2)
						decodedRune = unicode.ReplacementChar
					} else {
						r3, err := t.buf.NextRune()
						if err != nil {
							if errors.Is(err, io.EOF) {
								return io.ErrUnexpectedEOF
							}

							return err
						} else if r3 != 'u' {
							if t.laxListener != nil {
								t.laxListener(SyntaxRecovery{
									Behavior:         WarnStringUnicodeReplacementChar,
									SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-6], uncommitted[len(uncommitted)-6:]),
									SourceRunes:      uncommitted[len(uncommitted)-6:],
									ValueStart:       t.getTextOffset(),
									ReplacementRunes: []rune{'\\', 'u', 'F', 'F', 'F', 'D'},
								})
							}

							t.buf.BacktrackRunes(r2, r3)
							decodedRune = unicode.ReplacementChar
						} else {
							uncommitted = append(uncommitted, r2, r3)

							decodedRunePair, nextUncommitted, err := scanUnicode(t, uncommitted)
							if err != nil {
								if errors.Is(err, io.EOF) {
									return io.ErrUnexpectedEOF
								}

								return err
							}

							uncommitted = nextUncommitted

							decodedPair := utf16.DecodeRune(decodedRune, decodedRunePair)
							if decodedPair == unicode.ReplacementChar && t.laxListener != nil {
								t.laxListener(SyntaxRecovery{
									Behavior:         WarnStringUnicodeReplacementChar,
									SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-12], uncommitted[len(uncommitted)-12:]),
									SourceRunes:      uncommitted[len(uncommitted)-12:],
									ValueStart:       t.getTextOffset(),
									ReplacementRunes: []rune{'\\', 'u', 'F', 'F', 'F', 'D'},
								})
							}

							decodedRune = decodedPair
						}
					}
				}

				decoded = append(decoded, decodedRune)
			default:
				if t.laxBehaviors&LaxStringEscapeInvalidEscape == 0 {
					return t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r1,
					}, uncommitted, []rune{r0, r1})
				} else if t.laxListener != nil {
					t.laxListener(SyntaxRecovery{
						Behavior:         LaxStringEscapeInvalidEscape,
						ValueStart:       t.getTextOffset(),
						SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted, []rune{r0, r1}),
						SourceRunes:      []rune{r0, r1},
						ReplacementRunes: []rune{r0, r1},
					})
				}

				decoded = append(decoded, r0, r1)
				uncommitted = append(uncommitted, r0, r1)
			}
		case '\b', '\f', '\n', '\r', '\t':
			if t.laxBehaviors&LaxStringEscapeMissingEscape == 0 {
				return t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0,
				}, uncommitted, []rune{r0})
			}

			if t.laxListener != nil {
				lt := SyntaxRecovery{
					Behavior:      LaxStringEscapeMissingEscape,
					SourceOffsets: t.uncommittedTextOffsetRange(uncommitted, []rune{r0}),
					SourceRunes:   []rune{r0},
					ValueStart:    t.getTextOffset(),
				}

				switch r0 {
				case '\b':
					lt.ReplacementRunes = []rune("\\b")
				case '\f':
					lt.ReplacementRunes = []rune("\\f")
				case '\n':
					lt.ReplacementRunes = []rune("\\n")
				case '\r':
					lt.ReplacementRunes = []rune("\\r")
				case '\t':
					lt.ReplacementRunes = []rune("\\t")
				}

				t.laxListener(lt)
			}

			uncommitted = append(uncommitted, r0)
			decoded = append(decoded, r0)
		case '"':
			uncommitted = append(uncommitted, r0)

			t.emit(StringToken{
				SourceOffsets: t.commitForTextOffsetRange(uncommitted),
				Content:       string(decoded),
			})

			return nil
		default:
			if r0 < 0x1F || (r0 >= 0x80 && r0 <= 0x9F) {
				if t.laxBehaviors&LaxStringEscapeMissingEscape == 0 {
					return t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r0,
					}, nil, []rune{r0})
				} else if t.laxListener != nil {
					t.laxListener(SyntaxRecovery{
						Behavior:         LaxStringEscapeMissingEscape,
						SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted, []rune{r0}),
						SourceRunes:      []rune(string(r0)),
						ValueStart:       t.getTextOffset(),
						ReplacementRunes: []rune(fmt.Sprintf("\\u%04x", r0)),
					})
				}
			}

			uncommitted = append(uncommitted, r0)
			decoded = append(decoded, r0)
		}
	}
}
