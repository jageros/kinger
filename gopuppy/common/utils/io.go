package utils

import (
	"io"
)

func WriteAll(rw io.Writer, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := rw.Write(data)
		if n == left && err == nil {
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil {
			return err
		}
	}
	return nil
}
