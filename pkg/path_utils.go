package xdeb

import (
	"path/filepath"
	"strings"
)

func TrimPathExtension(path string, count int) string {
	for i := 0; i < count; i++ {
		path = strings.TrimSuffix(path, filepath.Ext(path))
	}

	return path
}
