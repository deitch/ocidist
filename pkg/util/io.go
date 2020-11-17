package util

import (
	"io"
)

type GetReadCloser func() (io.ReadCloser, error)
