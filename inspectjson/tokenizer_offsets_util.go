package inspectjson

import (
	"github.com/dpb587/cursorio-go/cursorio"
)

func (t *Tokenizer) newOffsetError(err error, readUncomitted, readIgnored cursorio.DecodedRunes) error {
	werr := cursorio.OffsetError{
		Err: err,
	}

	if t.doc == nil {
		werr.Offset = t.buf.GetByteOffset() - cursorio.ByteOffset(readIgnored.Size)
	} else {
		if readUncomitted.Size > 0 {
			werr.Offset = *t.uncommittedTextOffset(readUncomitted)
		} else {
			werr.Offset = t.doc.GetTextOffset()
		}
	}

	return werr
}

func (t *Tokenizer) getTextOffset() *cursorio.TextOffset {
	if t.doc == nil {
		return nil
	}

	v := t.doc.GetTextOffset()

	return &v
}

func (t *Tokenizer) commit(runes cursorio.DecodedRunes) {
	if t.doc == nil {
		return
	}

	t.doc.WriteRunes(runes.Runes, runes.Size)
}

func (t *Tokenizer) commitForTextOffsetRange(runes cursorio.DecodedRunes) *cursorio.TextOffsetRange {
	if t.doc == nil {
		return nil
	}

	v := t.doc.WriteRunesForOffsetRange(runes.Runes, runes.Size)

	return &v
}

func (t *Tokenizer) uncommittedTextOffset(runes cursorio.DecodedRunes) *cursorio.TextOffset {
	if t.doc == nil {
		return nil
	}

	clone := t.doc.Clone()
	v := clone.WriteRunesForOffset(runes.Runes, runes.Size)

	return &v
}

func (t *Tokenizer) uncommittedTextOffsetRange(prefixIgnored, runes cursorio.DecodedRunes) *cursorio.TextOffsetRange {
	if t.doc == nil {
		return nil
	}

	clone := t.doc.Clone()

	if prefixIgnored.Size > 0 {
		clone.WriteRunes(prefixIgnored.Runes, prefixIgnored.Size)
	}

	v := clone.WriteRunesForOffsetRange(runes.Runes, runes.Size)

	return &v
}
