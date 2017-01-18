package main

import (
	"fmt"
	"os"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type statsProcessor struct {
	FileCount    int `json:"file-count"`
	CommentCount int `json:"comment-chars"`
	OtherCount   int `json:"other-chars"`
}

func (lp *statsProcessor) end() error {
	fmt.Printf("files : %d\n", lp.FileCount)
	fmt.Printf("code bytes  (non-comments) : %d\n", lp.OtherCount)
	fmt.Printf("comment bytes : %d\n", lp.CommentCount)
	fmt.Printf("total bytes : %d\n", lp.CommentCount+lp.OtherCount)
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

	lp.FileCount++

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
			lp.CommentCount += len(lit)
		default:
			lp.OtherCount += len(lit)
		}

		switch tok {
		case equilex.EOF:
			return nil
		}
	}
}
