package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type reports struct {
	reports []module
}

func (m *reports) end() error {
	for _, method := range m.reports {
		fmt.Println(method)
	}
	return nil
}

func (m *reports) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Reports/") {
		m.reports = append(m.reports, moduleFromFullFilename(file))
	}
	return nil
}
