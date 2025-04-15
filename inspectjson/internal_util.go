package inspectjson

import "github.com/dpb587/cursorio-go/x/cursorioutil"

const hexUpper = "0123456789ABCDEF"

const (
	strLiteralFalse = "false"
	strLiteralNull  = "null"
	strLiteralTrue  = "true"
)

func decodeHex(c rune) (rune, bool) {
	switch c {
	case '0':
		return 0, true
	case '1':
		return 1, true
	case '2':
		return 2, true
	case '3':
		return 3, true
	case '4':
		return 4, true
	case '5':
		return 5, true
	case '6':
		return 6, true
	case '7':
		return 7, true
	case '8':
		return 8, true
	case '9':
		return 9, true
	case 'A', 'a':
		return 10, true
	case 'B', 'b':
		return 11, true
	case 'C', 'c':
		return 12, true
	case 'D', 'd':
		return 13, true
	case 'E', 'e':
		return 14, true
	case 'F', 'f':
		return 15, true
	}

	return 0, false
}

func scanUnicode(t *Tokenizer, uncommitted []rune) (rune, []rune, error) {
	r0, err := t.buf.NextRune()
	if err != nil {
		return 0, nil, err
	}

	r0x, ok := decodeHex(r0)
	if !ok {
		return 0, nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
			Rune: r0,
		}, uncommitted, []rune{r0})
	}

	r1, err := t.buf.NextRune()
	if err != nil {
		return 0, nil, err
	}

	r1x, ok := decodeHex(r1)
	if !ok {
		return 0, nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
			Rune: r1,
		}, append(uncommitted, r0), []rune{r1})
	}

	r2, err := t.buf.NextRune()
	if err != nil {
		return 0, nil, err
	}

	r2x, ok := decodeHex(r2)
	if !ok {
		return 0, nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
			Rune: r2,
		}, append(uncommitted, r0, r1), []rune{r2})
	}

	r3, err := t.buf.NextRune()
	if err != nil {
		return 0, nil, err
	}

	r3x, ok := decodeHex(r3)
	if !ok {
		return 0, nil, t.newOffsetError(cursorioutil.UnexpectedRuneError{
			Rune: r3,
		}, append(uncommitted, r0, r1, r2), []rune{r3})
	}

	return rune(r0x<<12 | r1x<<8 | r2x<<4 | r3x),
		append(uncommitted, r0, r1, r2, r3),
		nil
}
