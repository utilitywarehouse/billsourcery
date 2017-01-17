package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type methods struct {
	methods []string
}

func (m *methods) end() {
	for _, method := range m.methods {
		fmt.Println(method)
	}
}

func (m *methods) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Methods/") {
		m.methods = append(m.methods, filenameToIdentifier(file))
	}
	return nil
}
