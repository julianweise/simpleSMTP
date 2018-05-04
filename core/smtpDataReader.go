package core

import (
	"io"
	"net/textproto"
)

type SMTPDataReader struct {
	r			io.Reader
	limit 		int
	fillLevel 	int
}

func newSMTPDataReader(reader *textproto.Reader, limit int) io.Reader {
	return &SMTPDataReader{
		r: reader.DotReader(),
		limit: limit,
		fillLevel: 0,
	}
}

func (r *SMTPDataReader) Read(b []byte) (n int, err error) {
	if r.fillLevel <= 0 {
		return 0, &SMTPError{code: 552, message: "Maximum message size exceeded"}
	}
	if int(len(b)) > r.fillLevel {
		b = b[0:r.fillLevel]
	}
	n, err = r.Read(b)
	r.fillLevel -= n
	return n, err
}
