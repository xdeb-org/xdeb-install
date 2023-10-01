package xdeb

import (
	"path/filepath"
	"strings"
)

func TrimPathExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}
