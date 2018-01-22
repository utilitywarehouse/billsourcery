package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type methods struct {
	methods []module
}

func (m *methods) end() error {
	for _, method := range m.methods {
		fmt.Println(method)
	}
	return nil
}

func (m *methods) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, m)
}

func (m *methods) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Methods/") {
		m.methods = append(m.methods, moduleFromFullFilename(file))
	}
	return nil
}
