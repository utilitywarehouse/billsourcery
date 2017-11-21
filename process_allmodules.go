package main

import (
	"fmt"
	"path/filepath"
)

type module struct {
	moduleName string
	moduleType string
}

type allModules struct {
	modules []module
}

func (m *allModules) end() error {
	for _, method := range m.modules {
		fmt.Println(method)
	}
	return nil
}

func (m *allModules) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, m)
}

func (m *allModules) process(path string) error {
	dir, file := filepath.Split(path)
	_, mType := filepath.Split(dir[0 : len(dir)-1])
	m.modules = append(m.modules, module{filenameToIdentifier(file), mType})
	return nil
}
