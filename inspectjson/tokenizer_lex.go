package inspectjson

import (
	"errors"
	"fmt"
	"io"

	"github.com/dpb587/cursorio-go/cursorio"
	"github.com/dpb587/cursorio-go/x/cursorioutil"
)

func tokenizer_lexInit(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	return tokenizer_lexValue(t, r0, err)
}

func tokenizer_lexEnded(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		return nil, err
	}

	return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
		Rune: r0.Rune,
	}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
}

func tokenizer_lexObjectMemberEnded(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	switch r0.Rune {
	case '}':
		t.emit(EndObjectToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return nil, nil
	case ',':
		t.openValueSeparator = &ValueSeparatorToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		}

		return tokenizer_lexObjectMember, nil
	}

	return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
		Rune: r0.Rune,
	}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
}

func tokenizer_lexObjectMemberNameSeparator(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	if r0.Rune == ':' {
		t.emit(NameSeparatorToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		t.stackShift(tokenizer_lexObjectMemberEnded)

		return tokenizer_lexValue, nil
	}

	return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
		Rune: r0.Rune,
	}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
}

func tokenizer_lexObjectMember(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	switch r0.Rune {
	case '}':
		if t.laxBehaviors&LaxIgnoreExtraComma == 0 {
			return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
				Rune: r0.Rune,
			}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
		} else if t.laxListener != nil {
			t.laxListener(SyntaxRecovery{
				Behavior:      LaxIgnoreExtraComma,
				SourceOffsets: t.openValueSeparator.SourceOffsets,
				SourceRunes:   []rune{','},
			})
		}

		t.emit(EndObjectToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return nil, nil
	case ',':
		if t.laxBehaviors&LaxIgnoreExtraComma == 0 {
			return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
				Rune: r0.Rune,
			}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
		} else if t.laxListener != nil {
			t.laxListener(SyntaxRecovery{
				Behavior:      LaxIgnoreExtraComma,
				SourceOffsets: t.openValueSeparator.SourceOffsets,
				SourceRunes:   []rune{','},
			})
		}

		t.openValueSeparator = &ValueSeparatorToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		}

		return tokenizer_lexObjectMember, nil
	case '"':
		if t.openValueSeparator != nil {
			t.emit(ValueSeparatorToken{
				SourceOffsets: t.openValueSeparator.SourceOffsets,
			})

			t.openValueSeparator = nil
		}

		return tokenizer_lexObjectMemberNameSeparator, t.emitString(r0)
	}

	return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
		Rune: r0.Rune,
	}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
}

func tokenizer_lexArrayValue(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	switch r0.Rune {
	case ']':
		if t.laxBehaviors&LaxIgnoreExtraComma == 0 {
			return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
				Rune: r0.Rune,
			}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
		} else if t.laxListener != nil {
			t.laxListener(SyntaxRecovery{
				Behavior:      LaxIgnoreExtraComma,
				SourceOffsets: t.openValueSeparator.SourceOffsets,
				SourceRunes:   []rune{','},
			})
		}

		t.emit(EndArrayToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return nil, nil
	case ',':
		if t.laxBehaviors&LaxIgnoreExtraComma == 0 {
			return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
				Rune: r0.Rune,
			}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
		} else if t.laxListener != nil {
			t.laxListener(SyntaxRecovery{
				Behavior:      LaxIgnoreExtraComma,
				SourceOffsets: t.openValueSeparator.SourceOffsets,
				SourceRunes:   []rune{','},
			})
		}

		t.openValueSeparator = &ValueSeparatorToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		}

		return tokenizer_lexArrayValue, nil
	}

	if t.openValueSeparator != nil {
		t.emit(ValueSeparatorToken{
			SourceOffsets: t.openValueSeparator.SourceOffsets,
		})

		t.openValueSeparator = nil
	}

	t.stackShift(tokenizer_lexArrayValueEnded)

	return tokenizer_lexValue(t, r0, err)
}

func tokenizer_lexArrayValueEnded(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	switch r0.Rune {
	case ']':
		t.emit(EndArrayToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return nil, nil
	case ',':
		t.openValueSeparator = &ValueSeparatorToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		}

		return tokenizer_lexArrayValue, nil
	}

	return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
		Rune: r0.Rune,
	}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
}

func tokenizer_lexObject(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	if r0.Rune == '}' {
		t.emit(EndObjectToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return nil, nil
	} else if r0.Rune == ',' {
		if t.laxBehaviors&LaxIgnoreExtraComma == 0 {
			return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
				Rune: r0.Rune,
			}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
		} else if t.laxListener != nil {
			t.laxListener(SyntaxRecovery{
				Behavior:      LaxIgnoreExtraComma,
				SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
				SourceRunes:   []rune{r0.Rune},
			})
		} else {
			t.commit(r0.AsDecodedRunes())
		}

		return tokenizer_lexObject, nil
	}

	return tokenizer_lexObjectMember(t, r0, nil)
}

func tokenizer_lexArray(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	if r0.Rune == ']' {
		t.emit(EndArrayToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return nil, nil
	} else if r0.Rune == ',' {
		if t.laxBehaviors&LaxIgnoreExtraComma == 0 {
			return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
				Rune: r0.Rune,
			}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
		} else if t.laxListener != nil {
			t.laxListener(SyntaxRecovery{
				Behavior:      LaxIgnoreExtraComma,
				SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
				SourceRunes:   []rune{r0.Rune},
			})
		} else {
			t.commit(r0.AsDecodedRunes())
		}

		return tokenizer_lexArray, nil
	}

	return tokenizer_lexArrayValue(t, r0, nil)
}

func tokenizer_lexValue(t *Tokenizer, r0 cursorio.DecodedRune, err error) (lexFunc, error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}

		return nil, err
	}

	switch r0.Rune {
	case '{':
		t.emit(BeginObjectToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return tokenizer_lexObject, nil
	case '[':
		t.emit(BeginArrayToken{
			SourceOffsets: t.commitForTextOffsetRange(r0.AsDecodedRunes()),
		})

		return tokenizer_lexArray, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		var derr error
		var uncommitted = cursorio.DecodedRuneList{r0}
		var uncommittedTrailingZero bool
		var hadPrefixedZero bool

		switch r0.Rune {
		case '-':
			goto NUMBER_INT_INIT
		case '0':
			uncommittedTrailingZero = true
			fallthrough
		default:
			goto NUMBER_INT_CONT
		}

	NUMBER_INT_INIT:

		{
			r0, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil, io.ErrUnexpectedEOF
				}

				return nil, err
			}

			switch r0.Rune {
			case '0':
				uncommittedTrailingZero = true
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// good
			default:
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, uncommitted.AsDecodedRunes(), r0.AsDecodedRunes())
			}

			uncommitted = append(uncommitted, r0)
		}

	NUMBER_INT_CONT:

		for {
			r0, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					derr = err

					goto NUMBER_DONE
				}

				return nil, err
			}

			switch r0.Rune {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				if uncommittedTrailingZero {
					if t.laxBehaviors&LaxNumberTrimLeadingZero == 0 {
						return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
							Rune: uncommitted[len(uncommitted)-1].Rune,
						}, uncommitted[0:len(uncommitted)-1].AsDecodedRunes(), append(uncommitted[len(uncommitted)-1:], r0).AsDecodedRunes())
					} else if t.laxListener != nil {
						// TODO batch so multiple 0s are a single recovery
						t.laxListener(SyntaxRecovery{
							Behavior:      LaxNumberTrimLeadingZero,
							ValueStart:    t.getTextOffset(),
							SourceOffsets: t.uncommittedTextOffsetRange(uncommitted[0:len(uncommitted)-1].AsDecodedRunes(), uncommitted[len(uncommitted)-1:].AsDecodedRunes()),
							SourceRunes:   []rune{uncommitted[len(uncommitted)-1].Rune},
						})
					}

					hadPrefixedZero = true
				}

				uncommitted = append(uncommitted, r0)
			case '.':
				uncommitted = append(uncommitted, r0)

				goto NUMBER_FRAC
			case 'e', 'E':
				uncommitted = append(uncommitted, r0)

				goto NUMBER_EXP
			default:
				t.buf.BacktrackRunes(r0)

				goto NUMBER_DONE
			}
		}

	NUMBER_FRAC:

		{
			r0, err := t.buf.NextRune()
			if err != nil {
				return nil, err
			}

			switch r0.Rune {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				uncommitted = append(uncommitted, r0)
			default:
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, uncommitted.AsDecodedRunes(), r0.AsDecodedRunes())
			}
		}

		for {
			r0, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					derr = err

					goto NUMBER_DONE
				}

				return nil, err
			}

			switch r0.Rune {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				uncommitted = append(uncommitted, r0)
			case 'e', 'E':
				uncommitted = append(uncommitted, r0)

				goto NUMBER_EXP
			default:
				t.buf.BacktrackRunes(r0)

				goto NUMBER_DONE
			}
		}

	NUMBER_EXP:

		{
			r0, err := t.buf.NextRune()
			if err != nil {
				return nil, err
			}

			switch r0.Rune {
			case '+', '-':
				uncommitted = append(uncommitted, r0)

				goto NUMBER_EXP_INIT
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				uncommitted = append(uncommitted, r0)

				goto NUMBER_EXP_CONT
			default:
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, uncommitted.AsDecodedRunes(), r0.AsDecodedRunes())
			}
		}

	NUMBER_EXP_INIT:

		{
			r0, err := t.buf.NextRune()
			if err != nil {
				return nil, err
			}

			switch r0.Rune {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				uncommitted = append(uncommitted, r0)

				goto NUMBER_EXP_CONT
			default:
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, uncommitted.AsDecodedRunes(), r0.AsDecodedRunes())
			}
		}

	NUMBER_EXP_CONT:

		for {
			r0, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					derr = err

					goto NUMBER_DONE
				}

				return nil, err
			}

			switch r0.Rune {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				uncommitted = append(uncommitted, r0)
			default:
				t.buf.BacktrackRunes(r0)

				goto NUMBER_DONE
			}
		}

	NUMBER_DONE:

		tn := NumberToken{
			SourceOffsets: t.commitForTextOffsetRange(uncommitted.AsDecodedRunes()),
		}

		if hadPrefixedZero {
			pos := 0

			for _, r := range uncommitted {
				switch r.Rune {
				case '0':
					pos++
				case 'e', 'E', '.':
					pos--

					goto NUMBER_TRIM_DONE
				default:
					goto NUMBER_TRIM_DONE
				}
			}

		NUMBER_TRIM_DONE:

			tn.Content = uncommitted[pos:].String()
		} else {
			tn.Content = uncommitted.AsDecodedRunes().String()
		}

		t.emit(tn)

		if derr != nil {
			if errors.Is(derr, io.EOF) && len(t.stack) > 0 {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, derr
		}

		return nil, nil
	case '"':
		return nil, t.emitString(r0)
	case 'f', 'F':
		r1, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r1.Rune != 'a' && r1.Rune != 'A' {
			return nil, fmt.Errorf("invalid rune: %q", r1)
		}

		r2, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r2.Rune != 'l' && r2.Rune != 'L' {
			return nil, fmt.Errorf("invalid rune: %q", r2)
		}

		r3, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r3.Rune != 's' && r3.Rune != 'S' {
			return nil, fmt.Errorf("invalid rune: %q", r3)
		}

		r4, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r4.Rune != 'e' && r4.Rune != 'E' {
			return nil, fmt.Errorf("invalid rune: %q", r4)
		}

		literalRunes := cursorio.NewDecodedRunes(r0, r1, r2, r3, r4)

		if literalRunes.String() != strLiteralFalse {
			if t.laxBehaviors&LaxLiteralCaseInsensitive == 0 {
				// TODO should be first non-lowercase rune
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, cursorio.DecodedRunes{}, literalRunes)
			} else if t.laxListener != nil {
				t.laxListener(SyntaxRecovery{
					Behavior:         LaxLiteralCaseInsensitive,
					SourceOffsets:    t.uncommittedTextOffsetRange(cursorio.DecodedRunes{}, literalRunes),
					SourceRunes:      literalRunes.Runes,
					ReplacementRunes: []rune(strLiteralFalse),
				})
			}
		}

		t.emit(FalseToken{
			SourceOffsets: t.commitForTextOffsetRange(literalRunes),
		})

		return nil, nil
	case 'n', 'N':
		r1, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r1.Rune != 'u' && r1.Rune != 'U' {
			return nil, fmt.Errorf("invalid rune: %q", r1)
		}

		r2, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r2.Rune != 'l' && r2.Rune != 'L' {
			return nil, fmt.Errorf("invalid rune: %q", r2)
		}

		r3, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r3.Rune != 'l' && r3.Rune != 'L' {
			return nil, fmt.Errorf("invalid rune: %q", r3)
		}

		literalRunes := cursorio.NewDecodedRunes(r0, r1, r2, r3)

		if literalRunes.String() != strLiteralNull {
			if t.laxBehaviors&LaxLiteralCaseInsensitive == 0 {
				// TODO should be first non-lowercase rune
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, cursorio.DecodedRunes{}, literalRunes)
			} else if t.laxListener != nil {
				t.laxListener(SyntaxRecovery{
					Behavior:         LaxLiteralCaseInsensitive,
					SourceOffsets:    t.uncommittedTextOffsetRange(cursorio.DecodedRunes{}, literalRunes),
					SourceRunes:      literalRunes.Runes,
					ReplacementRunes: []rune(strLiteralNull),
				})
			}
		}

		t.emit(NullToken{
			SourceOffsets: t.commitForTextOffsetRange(literalRunes),
		})

		return nil, nil
	case 't', 'T':
		r1, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r1.Rune != 'r' && r1.Rune != 'R' {
			return nil, fmt.Errorf("invalid rune: %q", r1)
		}

		r2, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r2.Rune != 'u' && r2.Rune != 'U' {
			return nil, fmt.Errorf("invalid rune: %q", r2)
		}

		r3, err := t.buf.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		} else if r3.Rune != 'e' && r3.Rune != 'E' {
			return nil, fmt.Errorf("invalid rune: %q", r3)
		}

		literalRunes := cursorio.NewDecodedRunes(r0, r1, r2, r3)

		if literalRunes.String() != strLiteralTrue {
			if t.laxBehaviors&LaxLiteralCaseInsensitive == 0 {
				// TODO should be first non-lowercase rune
				return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
					Rune: r0.Rune,
				}, cursorio.DecodedRunes{}, literalRunes)
			} else if t.laxListener != nil {
				t.laxListener(SyntaxRecovery{
					Behavior:         LaxLiteralCaseInsensitive,
					SourceOffsets:    t.uncommittedTextOffsetRange(cursorio.DecodedRunes{}, literalRunes),
					SourceRunes:      literalRunes.Runes,
					ReplacementRunes: []rune(strLiteralTrue),
				})
			}
		}

		t.emit(TrueToken{
			SourceOffsets: t.commitForTextOffsetRange(literalRunes),
		})

		return nil, nil
	}

	return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
		Rune: r0.Rune,
	}, cursorio.DecodedRunes{}, r0.AsDecodedRunes())
}
