package main

import (
	"fmt"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type pubProcs struct {
	stmts []*statement
}

func (ex *pubProcs) end() error {
	for _, stmt := range ex.stmts {
		fmt.Println(stmt.String())
	}
	return nil
}

func (lp *pubProcs) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, lp)
}

func (ex *pubProcs) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	var stmt *statement

	atStart := true

	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.EOF:
			if stmt != nil {
				ex.stmts = append(ex.stmts, stmt)
			}
			return nil
		case equilex.WS:
			if stmt != nil {
				stmt.add(tok, lit)
			}
		case equilex.Public:
			if atStart {
				stmt = &statement{}
				stmt.add(tok, lit)
				atStart = false
			}
		case equilex.NewLine:
			if stmt != nil {
				ex.stmts = append(ex.stmts, stmt)
			}
			stmt = nil
			atStart = true
		default:
			if stmt != nil {
				stmt.add(tok, lit)
			}
			atStart = false
		}
	}
}
