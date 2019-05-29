package io_

import (
	"os"
)

// Stater is the interface that wraps the basic Stat method.
// Stat returns the FileInfo structure describing file.
type Stater interface {
	Stat() (os.FileInfo, error)
}
