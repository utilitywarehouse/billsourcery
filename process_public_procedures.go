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

	stmt := &statement{}

	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch {
		case tok == equilex.EOF:
			if !stmt.empty() {
				ex.stmts = append(ex.stmts, stmt)
			}
			return nil
		case stmt.empty() && tok == equilex.Public:
			stmt.add(tok, lit)
		case tok == equilex.NewLine && !stmt.empty():
			ex.stmts = append(ex.stmts, stmt)
			stmt = &statement{}
		case !stmt.empty():
			stmt.add(tok, lit)
		}
	}
}
