package util

import "io"

type multiWriteCloser struct {
	writers []io.WriteCloser
}

func (m multiWriteCloser) Write(p []byte) (n int, err error) {
	for _, writer := range m.writers {
		n, err := writer.Write(p)
		if err != nil {
			return n, err
		}
	}

	return len(p), nil
}

func (m multiWriteCloser) Close() error {
	var err error
	for _, closer := range m.writers {
		err1 := closer.Close()
		if err1 != nil {
			err = err1
		}
	}

	return err
}

func MultiWriteCloser(writeClosers ...io.WriteCloser) io.WriteCloser {
	return multiWriteCloser{writers: writeClosers}
}

var Discard discardWriteCloser

type discardWriteCloser struct {
}

func (d discardWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (d discardWriteCloser) Close() error {
	return nil
}
