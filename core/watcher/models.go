package watcher

import (
	"os"
)

type Object struct {
	path string
	obj  os.FileInfo
}
