package cmdutil

import "io"

type ErrReader struct {
	ch chan error
}

func NewErrReader() ErrReader {
	errReader := ErrReader{
		ch: make(chan error, 1),
	}

	return errReader
}

func (e ErrReader) Read(_ []byte) (n int, err error) {
	return 0, <-e.ch
}

func (e ErrReader) Close() {
	e.ch <- io.EOF
}

func (e ErrReader) CloseWithError(err error) {
	e.ch <- err
}
