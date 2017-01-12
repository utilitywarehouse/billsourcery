package main

import (
	"fmt"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type pubProcs struct{}

func (ex *pubProcs) end() {}

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
		tok, lit := l.Scan()

		switch tok {
		case equilex.EOF:
			if stmt != nil {
				fmt.Println(stmt.String())
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
				fmt.Println(stmt.String())
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
