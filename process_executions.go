package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func newExecutions() *executions {
	return &executions{make(map[string]([]*statement))}
}

type executions struct {
	stmts map[string]([]*statement)
}

func (ex *executions) end() error {
	for _, stmts := range ex.stmts {
		for _, stmt := range stmts {
			fmt.Println(stmt.String())
		}
	}
	return nil
}

func (m *executions) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, m)
}
func (ex *executions) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	fn := filename(path)

	var stmt *statement

	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.EOF:
			if stmt != nil {
				s := ex.stmts[fn]
				s = append(s, stmt)
				ex.stmts[fn] = s
			}
			return nil
		case equilex.Execute:
			stmt = &statement{}
			stmt.add(tok, lit)
		case equilex.NewLine:
			if stmt != nil {
				s := ex.stmts[fn]
				s = append(s, stmt)
				ex.stmts[fn] = s
			}
			stmt = nil
		default:
			if stmt != nil {
				stmt.add(tok, lit)
			}
		}
	}
}

func filename(path string) string {
	_, file := filepath.Split(path)
	return file
}
