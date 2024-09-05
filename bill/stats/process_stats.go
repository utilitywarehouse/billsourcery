package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type statsProcessor struct {
	FileCount    int `json:"file-count"`
	CommentCount int `json:"comment-chars"`
	OtherCount   int `json:"other-chars"`
}

func (lp *statsProcessor) print() error {
	fmt.Printf("files : %d\n", lp.FileCount)
	fmt.Printf("code bytes  (non-comments) : %d\n", lp.OtherCount)
	fmt.Printf("comment bytes : %d\n", lp.CommentCount)
	fmt.Printf("total bytes : %d\n", lp.CommentCount+lp.OtherCount)
	return nil
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

func walkSource(sourceRoot string, proc fileProcessor) error {
	inSourceDir := func(root, path string) bool {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			panic(err)
		}
		switch filepath.Dir(relative) {
		case "Exports", "Forms", "Imports", "Methods", "Procedures", "Processes", "Queries", "Reports":
			return true
		default:
			return false
		}
	}

	return filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".txt") && inSourceDir(sourceRoot, path) {
			//	log.Println(path)
			err := proc.process(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type fileProcessor interface {
	process(path string) error
}
