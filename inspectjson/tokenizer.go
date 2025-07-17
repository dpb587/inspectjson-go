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

type lexFunc func(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error)

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
	var uncommitted cursorio.DecodedRuneList

	for {
		r0, err := t.buf.NextRune()
		if err != nil {
			return fn(t, r0, err)
		}

		switch r0.Rune {
		case '/':
			if len(uncommitted) > 0 {
				uncommittedSquashed := uncommitted.AsDecodedRunes()

				if t.emitWhitespace {
					t.emit(WhitespaceToken{
						SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
						Content:       uncommittedSquashed.String(),
					})
				} else {
					t.commit(uncommittedSquashed)
				}

				uncommitted = nil
			}

			if t.laxBehaviors&(LaxIgnoreBlockComment|LaxIgnoreLineComment) == 0 {
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
			}

			uncommitted = append(uncommitted, r0)

			r1, err := t.buf.NextRune()
			if err != nil {
				return nil, err
			}

			switch r1.Rune {
			case '/':
				if t.laxBehaviors&LaxIgnoreLineComment == 0 {
					return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r1.Rune,
					}, uncommitted.AsDecodedRunes(), r1.AsDecodedRunes())
				}

				uncommitted = append(uncommitted, r1)

				for {
					r0, err := t.buf.NextRune()
					if err != nil {
						if errors.Is(err, io.EOF) {
							if t.laxListener != nil {
								uncommittedSquashed := uncommitted.AsDecodedRunes()

								t.laxListener(SyntaxRecovery{
									Behavior:      LaxIgnoreLineComment,
									SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
									SourceRunes:   uncommittedSquashed.Runes,
								})
							}
						}

						return nil, err
					}

					// TODO \r\n?
					if r0.Rune == '\n' {
						t.buf.BacktrackRunes(r0)

						goto LINE_DONE
					}

					uncommitted = append(uncommitted, r0)
				}

			LINE_DONE:

				if t.laxListener != nil {
					uncommittedSquashed := uncommitted.AsDecodedRunes()

					t.laxListener(SyntaxRecovery{
						Behavior:      LaxIgnoreLineComment,
						SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
						SourceRunes:   uncommittedSquashed.Runes,
					})
				}

				return fn, nil
			case '*':
				if t.laxBehaviors&LaxIgnoreBlockComment == 0 {
					return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r1.Rune,
					}, uncommitted.AsDecodedRunes(), r1.AsDecodedRunes())
				}

				uncommitted = append(uncommitted, r1)

				for {
					r0, err := t.buf.NextRune()
					if err != nil {
						if errors.Is(err, io.EOF) {
							if t.laxListener != nil {
								uncommittedSquashed := uncommitted.AsDecodedRunes()

								t.laxListener(SyntaxRecovery{
									Behavior:      LaxIgnoreBlockComment,
									SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
									SourceRunes:   uncommittedSquashed.Runes,
								})
							}
						}

						return nil, err
					}

					uncommitted = append(uncommitted, r0)

					if r0.Rune == '*' {
						r1, err := t.buf.NextRune()
						if err != nil {
							if errors.Is(err, io.EOF) {
								if t.laxListener != nil {
									uncommittedSquashed := uncommitted.AsDecodedRunes()

									t.laxListener(SyntaxRecovery{
										Behavior:      LaxIgnoreBlockComment,
										SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
										SourceRunes:   uncommittedSquashed.Runes,
									})
								}
							}

							return nil, err
						}

						uncommitted = append(uncommitted, r1)

						if r1.Rune == '/' {
							goto BLOCK_DONE
						}
					}
				}

			BLOCK_DONE:

				if t.laxListener != nil {
					uncommittedSquashed := uncommitted.AsDecodedRunes()

					t.laxListener(SyntaxRecovery{
						Behavior:      LaxIgnoreBlockComment,
						SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
						SourceRunes:   uncommittedSquashed.Runes,
					})
				}

				return fn, nil
			default:
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: uncommitted[0].Rune,
				}, cursorio.NewDecodedRunes(uncommitted[0:1]...), r1.AsDecodedRunes())
			}
		case 0x20, 0x09, 0x0A, 0x0D:
			uncommitted = append(uncommitted, r0)
		default:
			uncommittedSquashed := uncommitted.AsDecodedRunes()

			if len(uncommitted) > 0 {
				if t.emitWhitespace {
					t.emit(WhitespaceToken{
						SourceOffsets: t.commitForTextOffsetRange(uncommittedSquashed),
						Content:       uncommittedSquashed.String(),
					})
				} else {
					t.commit(uncommittedSquashed)
				}
			}

			return fn(t, r0, err)
		}
	}
}

func (r *Tokenizer) emit(tokens ...Token) {
	r.tokens = append(r.tokens, tokens...)
}

func (t *Tokenizer) emitString(r0 cursorio.DecodedRune) error {
	// assert(r0 == '"')

	var uncommitted = cursorio.DecodedRuneList{r0}
	var decoded []rune

	for {
		r0, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}

			return err
		}

		switch r0.Rune {
		case '\\':
			r1, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return io.ErrUnexpectedEOF
				}

				return err
			}

			switch r1.Rune {
			case '"', '\\', '/':
				decoded = append(decoded, r1.Rune)
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
					} else if r2.Rune != '\\' {
						if t.laxListener != nil {
							t.laxListener(SyntaxRecovery{
								Behavior:         WarnStringUnicodeReplacementChar,
								SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-6].AsDecodedRunes(), uncommitted[len(uncommitted)-6:].AsDecodedRunes()),
								SourceRunes:      uncommitted[len(uncommitted)-6:].AsDecodedRunes().Runes,
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
						} else if r3.Rune != 'u' {
							if t.laxListener != nil {
								t.laxListener(SyntaxRecovery{
									Behavior:         WarnStringUnicodeReplacementChar,
									SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-6].AsDecodedRunes(), uncommitted[len(uncommitted)-6:].AsDecodedRunes()),
									SourceRunes:      uncommitted[len(uncommitted)-6:].AsDecodedRunes().Runes,
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
									SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-12].AsDecodedRunes(), uncommitted[len(uncommitted)-12:].AsDecodedRunes()),
									SourceRunes:      uncommitted[len(uncommitted)-12:].AsDecodedRunes().Runes,
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
				rawRunes := cursorio.NewDecodedRunes(r0, r1)

				if t.laxBehaviors&LaxStringEscapeInvalidEscape == 0 {
					return t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r1.Rune,
					}, uncommitted.AsDecodedRunes(), rawRunes)
				} else if t.laxListener != nil {
					t.laxListener(SyntaxRecovery{
						Behavior:         LaxStringEscapeInvalidEscape,
						ValueStart:       t.getTextOffset(),
						SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted.AsDecodedRunes(), rawRunes),
						SourceRunes:      rawRunes.Runes,
						ReplacementRunes: rawRunes.Runes,
					})
				}

				decoded = append(decoded, rawRunes.Runes...)
				uncommitted = append(uncommitted, r0, r1)
			}
		case '\b', '\f', '\n', '\r', '\t':
			if t.laxBehaviors&LaxStringEscapeMissingEscape == 0 {
				return t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, uncommitted.AsDecodedRunes(), r0.AsDecodedRunes())
			}

			if t.laxListener != nil {
				lt := SyntaxRecovery{
					Behavior:      LaxStringEscapeMissingEscape,
					SourceOffsets: t.uncommittedTextOffsetRange(uncommitted.AsDecodedRunes(), r0.AsDecodedRunes()),
					SourceRunes:   []rune{r0.Rune},
					ValueStart:    t.getTextOffset(),
				}

				switch r0.Rune {
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
			decoded = append(decoded, r0.Rune)
		case '"':
			uncommitted = append(uncommitted, r0)

			t.emit(StringToken{
				SourceOffsets: t.commitForTextOffsetRange(uncommitted.AsDecodedRunes()),
				Content:       string(decoded),
			})

			return nil
		default:
			if r0.Rune < 0x1F || (r0.Rune >= 0x80 && r0.Rune <= 0x9F) {
				if t.laxBehaviors&LaxStringEscapeMissingEscape == 0 {
					return t.newOffsetError(cursorioutil.UnexpectedRuneError{
						Rune: r0.Rune,
					}, uncommitted.AsDecodedRunes(), r0.AsDecodedRunes())
				} else if t.laxListener != nil {
					t.laxListener(SyntaxRecovery{
						Behavior:         LaxStringEscapeMissingEscape,
						SourceOffsets:    t.uncommittedTextOffsetRange(uncommitted.AsDecodedRunes(), r0.AsDecodedRunes()),
						SourceRunes:      []rune{r0.Rune},
						ValueStart:       t.getTextOffset(),
						ReplacementRunes: []rune(fmt.Sprintf("\\u%04x", r0)),
					})
				}
			}

			uncommitted = append(uncommitted, r0)
			decoded = append(decoded, r0.Rune)
		}
	}
}
