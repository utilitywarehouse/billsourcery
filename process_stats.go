package main

import (
	"fmt"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type statsProcessor struct {
	filecount    int
	commentcount int
	othercount   int
}

func (lp *statsProcessor) end() error {
	fmt.Printf("files : %d\n", lp.filecount)
	fmt.Printf("code bytes  (non-comments) : %d\n", lp.othercount)
	fmt.Printf("comment bytes : %d\n", lp.commentcount)
	fmt.Printf("total bytes : %d\n", lp.commentcount+lp.othercount)
	return nil
}

func (lp *statsProcessor) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, lp)
}

func (lp *statsProcessor) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	lp.filecount++

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.EOF:
			return nil
		case equilex.Comment:
			lp.commentcount += len(lit)
		default:
			lp.othercount += len(lit)
		}

		switch tok {
		case equilex.EOF:
			return nil
		}
	}
}
