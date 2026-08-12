package archive

import (
	"errors"
	"io"
)

const progressUnit = 1 << 20

type progressReader struct {
	reader   io.Reader
	current  *int64
	reported int64
	report   func() bool
}

func (r *progressReader) Read(buffer []byte) (int, error) {
	n, err := r.reader.Read(buffer)
	if n <= 0 {
		return n, err
	}
	*r.current += int64(n)
	if r.report != nil && *r.current-r.reported >= progressUnit {
		if !r.report() {
			return n, errors.New("操作被取消")
		}
		r.reported = *r.current
	}
	return n, err
}
