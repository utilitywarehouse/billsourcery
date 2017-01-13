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
	file = strings.ToLower(file)
	if strings.HasSuffix(dir, "/Methods/") && strings.HasSuffix(file, ".jc@.txt") {
		m.methods = append(m.methods, file[0:len(file)-8])
	}
	return nil
}
