package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type forms struct {
	forms []module
}

func (m *forms) end() error {
	for _, method := range m.forms {
		fmt.Println(method)
	}
	return nil
}

func (m *forms) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, m)
}

func (m *forms) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Forms/") {
		m.forms = append(m.forms, moduleFromFullFilename(file))
	}
	return nil
}
