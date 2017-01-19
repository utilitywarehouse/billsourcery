package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type processes struct {
	procs []string
}

func (m *processes) end() error {
	for _, method := range m.procs {
		fmt.Println(method)
	}
	return nil
}

func (m *processes) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, m)
}

func (m *processes) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Processes/") {
		m.procs = append(m.procs, filenameToIdentifier(file))
	}
	return nil
}
