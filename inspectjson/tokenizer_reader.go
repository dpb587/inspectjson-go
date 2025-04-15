package inspectjson

import "io"

type tokenizerReader struct {
	t *Tokenizer

	err    error
	buf    []byte
	bufidx int
}

var _ io.Reader = (*tokenizerReader)(nil)

func NewTokenizerReader(t *Tokenizer) io.Reader {
	return &tokenizerReader{
		t: t,
	}
}

func (r *tokenizerReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	} else if r.bufidx >= len(r.buf) {
		t, err := r.t.Next()
		if err != nil {
			r.err = err

			return 0, err
		}

		var b []byte

		r.bufidx = 0
		r.buf, r.err = t.AppendText(b)
		if r.err != nil {
			return 0, r.err
		}
	}

	n = copy(p, r.buf[r.bufidx:])
	r.bufidx += n

	if r.bufidx >= len(r.buf) {
		r.buf = nil
	}

	return n, nil
}
