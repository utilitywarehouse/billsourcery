package main

import (
	"fmt"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type stringConsts struct{}

func (lp *stringConsts) end() error { return nil }

func (lp *stringConsts) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.EOF:
			return nil
		case equilex.StringConstant:
			fmt.Println(lit[1 : len(lit)-1])
		}
	}
}
