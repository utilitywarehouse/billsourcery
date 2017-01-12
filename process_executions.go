package main

import (
	"log"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type executions struct{}

func (ex *executions) end() {}

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
				log.Println(stmt.String())
			}
			return nil
		case equilex.Execute:
			stmt = &statement{}
			stmt.add(tok, lit)
		case equilex.NewLine:
			if stmt != nil {
				log.Println(stmt.String())
			}
			stmt = nil
		default:
			if stmt != nil {
				stmt.add(tok, lit)
			}
		}
	}
}
