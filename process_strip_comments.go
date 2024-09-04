package main

import (
	"bytes"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type commentStripper struct{}

func (lp *commentStripper) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	var out bytes.Buffer

	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.Comment:
		case equilex.EOF:
			cp1252Bytes, _, err := transform.Bytes(charmap.Windows1252.NewEncoder(), out.Bytes())
			if err != nil {
				return err
			}
			if err := os.WriteFile(path, cp1252Bytes, 0o644); err != nil {
				return err
			}
			return nil
		default:
			out.WriteString(lit)
		}
	}
}
