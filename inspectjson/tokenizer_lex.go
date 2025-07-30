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

		if r0.Rune != '-' {
			goto NUMBER_INT_CONT
		}

		{
			r0, err := t.buf.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil, io.ErrUnexpectedEOF
				}

				return nil, err
			}

			switch r0.Rune {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
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
			Content: uncommitted.AsDecodedRunes().String(),
		}

		{
			// easier to handle leading zeroes once we have the full number
			// * allows grouping syntax recovery hook for multiple zeros
			// * avoids scan logic for allowing zeros with exponents/decimals

			startPos := 0

			if uncommitted[0].Rune == '-' {
				startPos = 1
			}

			if uncommitted[startPos].Rune == '0' && len(uncommitted) > startPos+1 {
				untilPos := len(uncommitted) - 1

				for i := startPos + 1; i < len(uncommitted); i++ {
					switch uncommitted[i].Rune {
					case '0':
						continue
					case 'e', 'E', '.':
						if i == startPos+1 {
							// allowed
							goto NUMBER_LEADING_ZERO_DONE
						} else {
							i -= 1 // offset to retain one zero
						}
					}

					untilPos = i

					goto NUMBER_LEADING_ZERO_VALIDATE
				}

			NUMBER_LEADING_ZERO_VALIDATE:

				if untilPos-startPos > 0 {
					if t.laxBehaviors&LaxNumberTrimLeadingZero == 0 {
						return nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
							Rune: uncommitted[startPos].Rune,
						}, uncommitted[0:startPos].AsDecodedRunes(), uncommitted[startPos:].AsDecodedRunes())
					} else if t.laxListener != nil {
						t.laxListener(SyntaxRecovery{
							Behavior:      LaxNumberTrimLeadingZero,
							SourceOffsets: t.uncommittedTextOffsetRange(uncommitted[0:startPos].AsDecodedRunes(), uncommitted[startPos:untilPos].AsDecodedRunes()),
							SourceRunes:   uncommitted[startPos:untilPos].AsDecodedRunes().Runes,
							ValueStart:    t.getTextOffset(),
						})
					}

					tn.Content = append(uncommitted[0:startPos], uncommitted[untilPos:]...).String()
				}
			}
		}

	NUMBER_LEADING_ZERO_DONE:

		tn.SourceOffsets = t.commitForTextOffsetRange(uncommitted.AsDecodedRunes())

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
