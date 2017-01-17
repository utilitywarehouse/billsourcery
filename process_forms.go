package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type forms struct {
	forms []string
}

func (m *forms) end() {
	for _, method := range m.forms {
		fmt.Println(method)
	}
}

func (m *forms) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Forms/") {
		m.forms = append(m.forms, filenameToIdentifier(file))
	}
	return nil
}
