package main

import (
	"fmt"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type executions struct {
	stmts []*statement
}

func (ex *executions) end() {
	for _, stmt := range ex.stmts {
		fmt.Println(stmt.String())
	}
}

func (ex *executions) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	var stmt *statement

	for {
		tok, lit := l.Scan()

		switch tok {
		case equilex.EOF:
			if stmt != nil {
				ex.stmts = append(ex.stmts, stmt)
			}
			return nil
		case equilex.Execute:
			stmt = &statement{}
			stmt.add(tok, lit)
		case equilex.NewLine:
			if stmt != nil {
				ex.stmts = append(ex.stmts, stmt)
			}
			stmt = nil
		default:
			if stmt != nil {
				stmt.add(tok, lit)
			}
		}
	}
}
