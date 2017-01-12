package main

import (
	"log"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type lexCheck struct {
	anyErrors bool
}

func (lp *lexCheck) end() {
	if !lp.anyErrors {
		log.Println("no lexer errors.")
	}
}

func (lp *lexCheck) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	for {
		tok, lit := l.Scan()

		switch tok {
		case equilex.EOF:
			return nil
		case equilex.Illegal:
			lp.anyErrors = true
			log.Printf("illegal token in file '%s' : '%v'\n", path, lit)
		}
	}
}
