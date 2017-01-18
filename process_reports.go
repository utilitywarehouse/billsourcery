package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type reports struct {
	reports []string
}

func (m *reports) end() {
	for _, method := range m.reports {
		fmt.Println(method)
	}
}

func (m *reports) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Reports/") {
		m.reports = append(m.reports, filenameToIdentifier(file))
	}
	return nil
}
