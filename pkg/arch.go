package xdeb

import (
	"fmt"
	"runtime"
)

var ARCHITECTURE_MAP = map[string]string{
	"amd64": "x86_64",
	//"386":   "i386",
}

func FindArchitecture() (string, error) {
	arch, ok := ARCHITECTURE_MAP[runtime.GOARCH]

	if !ok {
		return "", fmt.Errorf("Architecture %s not supported (yet).", runtime.GOARCH)
	}

	return arch, nil
}
