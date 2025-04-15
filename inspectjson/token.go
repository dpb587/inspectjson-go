package inspectjson

import (
	"encoding"
	"strconv"
	"unicode/utf16"

	"github.com/dpb587/cursorio-go/cursorio"
)

type Token interface {
	encoding.TextAppender

	GetGrammarName() GrammarName
	GetSourceOffsets() *cursorio.TextOffsetRange
}

//

type BeginArrayToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = BeginArrayToken{}

func (BeginArrayToken) GetGrammarName() GrammarName {
	return grammarName_BeginArray
}

func (t BeginArrayToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t BeginArrayToken) AppendText(b []byte) ([]byte, error) {
	return append(b, '['), nil
}

//

type BeginObjectToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = BeginObjectToken{}

func (BeginObjectToken) GetGrammarName() GrammarName {
	return grammarName_BeginObject
}

func (t BeginObjectToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t BeginObjectToken) AppendText(b []byte) ([]byte, error) {
	return append(b, '{'), nil
}

//

type EndArrayToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = EndArrayToken{}

func (EndArrayToken) GetGrammarName() GrammarName {
	return grammarName_EndArray
}

func (t EndArrayToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t EndArrayToken) AppendText(b []byte) ([]byte, error) {
	return append(b, ']'), nil
}

//

type EndObjectToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = EndObjectToken{}

func (EndObjectToken) GetGrammarName() GrammarName {
	return grammarName_EndObject
}

func (t EndObjectToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t EndObjectToken) AppendText(b []byte) ([]byte, error) {
	return append(b, '}'), nil
}

//

type NameSeparatorToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = NameSeparatorToken{}

func (NameSeparatorToken) GetGrammarName() GrammarName {
	return grammarName_NameSeparator
}

func (t NameSeparatorToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t NameSeparatorToken) AppendText(b []byte) ([]byte, error) {
	return append(b, ':'), nil
}

//

type ValueSeparatorToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = ValueSeparatorToken{}

func (ValueSeparatorToken) GetGrammarName() GrammarName {
	return grammarName_ValueSeparator
}

func (t ValueSeparatorToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t ValueSeparatorToken) AppendText(b []byte) ([]byte, error) {
	return append(b, ','), nil
}

//

type FalseToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = FalseToken{}

func (FalseToken) GetGrammarName() GrammarName {
	return grammarName_False
}

func (t FalseToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t FalseToken) AppendText(b []byte) ([]byte, error) {
	return append(b, "false"...), nil
}

//

type NullToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = NullToken{}

func (NullToken) GetGrammarName() GrammarName {
	return grammarName_Null
}

func (t NullToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t NullToken) AppendText(b []byte) ([]byte, error) {
	return append(b, "null"...), nil
}

//

type TrueToken struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Token = TrueToken{}

func (TrueToken) GetGrammarName() GrammarName {
	return grammarName_True
}

func (t TrueToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t TrueToken) AppendText(b []byte) ([]byte, error) {
	return append(b, "true"...), nil
}

//

type StringToken struct {
	SourceOffsets *cursorio.TextOffsetRange
	Content       string
}

var _ Token = StringToken{}

func (StringToken) GetGrammarName() GrammarName {
	return grammarName_String
}

func (t StringToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t StringToken) AppendText(b []byte) ([]byte, error) {
	var echar, uchar4, uchar8 int

	tr := []rune(t.Content)

	for i := 0; i < len(tr); i++ {
		switch tr[i] {
		case '\b', '\f', '\n', '\r', '\t', '\\', '"':
			echar++
		default:
			if tr[i] > 0xffff {
				uchar8++
			} else if tr[i] > 0xff || tr[i] <= 0x1f || tr[i] >= 0x7f && tr[i] <= 0x9f {
				uchar4++
			}
		}
	}

	if echar == 0 && uchar4 == 0 && uchar8 == 0 {
		b = append(b, '"')
		b = append(b, t.Content...)
		b = append(b, '"')

		return b, nil
	}

	buf := make([]rune, len(tr)+echar+uchar4*5+uchar8*11)
	widx := 0

	for ridx := range tr {
		rr := tr[ridx]

		switch rr {
		case '\b':
			buf[widx] = '\\'
			buf[widx+1] = 'b'
			widx += 2
		case '\f':
			buf[widx] = '\\'
			buf[widx+1] = 'f'
			widx += 2
		case '\n':
			buf[widx] = '\\'
			buf[widx+1] = 'n'
			widx += 2
		case '\r':
			buf[widx] = '\\'
			buf[widx+1] = 'r'
			widx += 2
		case '\t':
			buf[widx] = '\\'
			buf[widx+1] = 't'
			widx += 2
		case '\\':
			buf[widx] = '\\'
			buf[widx+1] = '\\'
			widx += 2
		case '"':
			buf[widx] = '\\'
			buf[widx+1] = '"'
			widx += 2
		default:
			if rr > 0xffff {
				rr1, rr2 := utf16.EncodeRune(rr)
				buf[widx] = '\\'
				buf[widx+1] = 'u'
				buf[widx+2] = rune(hexUpper[rr1&0xf000>>12])
				buf[widx+3] = rune(hexUpper[rr1&0x0f00>>8])
				buf[widx+4] = rune(hexUpper[rr1&0x00f0>>4])
				buf[widx+5] = rune(hexUpper[rr1&0x000f])
				buf[widx+6] = '\\'
				buf[widx+7] = 'u'
				buf[widx+8] = rune(hexUpper[rr2&0xf000>>12])
				buf[widx+9] = rune(hexUpper[rr2&0x0f00>>8])
				buf[widx+10] = rune(hexUpper[rr2&0x00f0>>4])
				buf[widx+11] = rune(hexUpper[rr2&0x000f])
				widx += 12
			} else if rr > 0xff || rr <= 0x1f || rr >= 0x7f && rr <= 0x9f {
				buf[widx] = '\\'
				buf[widx+1] = 'u'
				buf[widx+2] = rune(hexUpper[rr&0xf000>>12])
				buf[widx+3] = rune(hexUpper[rr&0x0f00>>8])
				buf[widx+4] = rune(hexUpper[rr&0x00f0>>4])
				buf[widx+5] = rune(hexUpper[rr&0x000f])
				widx += 6
			} else {
				buf[widx] = rr
				widx++
			}
		}
	}

	b = append(b, '"')
	b = append(b, []byte(string(buf))...)
	b = append(b, '"')

	return b, nil
}

//

type NumberToken struct {
	SourceOffsets *cursorio.TextOffsetRange
	Content       string
}

var _ Token = NumberToken{}

func (NumberToken) GetGrammarName() GrammarName {
	return grammarName_Number
}

func (t NumberToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t NumberToken) Float64() (float64, error) {
	return strconv.ParseFloat(t.Content, 64)
}

func (t NumberToken) Int64() (int64, error) {
	return strconv.ParseInt(t.Content, 10, 64)
}

func (t NumberToken) AppendText(b []byte) ([]byte, error) {
	return append(b, t.Content...), nil
}

//

type WhitespaceToken struct {
	SourceOffsets *cursorio.TextOffsetRange
	Content       string
}

var _ Token = WhitespaceToken{}

func (WhitespaceToken) GetGrammarName() GrammarName {
	return grammarName_Ws
}

func (t WhitespaceToken) GetSourceOffsets() *cursorio.TextOffsetRange {
	return t.SourceOffsets
}

func (t WhitespaceToken) AppendText(b []byte) ([]byte, error) {
	return append(b, t.Content...), nil
}
