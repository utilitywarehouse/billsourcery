package main

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type commentStripper struct{}

func (lp *commentStripper) end() {}

func (lp *commentStripper) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	var out bytes.Buffer

	for {
		tok, lit := l.Scan()

		switch tok {
		case equilex.Comment:
		case equilex.EOF:
			cp1252Bytes, _, err := transform.Bytes(charmap.Windows1252.NewEncoder(), out.Bytes())
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, cp1252Bytes, 0644); err != nil {
				return err
			}
			return nil
		default:
			out.WriteString(lit)
		}
	}
}
