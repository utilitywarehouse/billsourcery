package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jawher/mow.cli"
	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type processor interface {
	end()
	process(path string) error
}

func main() {

	log.SetFlags(0)

	app := cli.App("billsourcery", "Bill source code attempted wizardry")

	sourceRoot := app.StringOpt("source-root", "/home/mgarton/work/uw-bill-source-history", "Root directory for equinox source. Subdirs Methods/ Forms/ etc are expected")

	var cmdErr error

	app.Command("comment-stats", "Provide stats about comments", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &statsProcessor{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("strip-comments", "Remove comments from the source files", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &commentStripper{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("string-constants", "Dump all \" delimited string constants found in the source, one per line, to stdout (multi-line strings not included)", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &stringConsts{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("executes", "List execute statements. Incomplete", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &executions{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("lexer-check", "Ensure the lexer can correctly scan all source. This is mostly for debugging the lexer", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &lexCheck{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("identifiers", "List identifier tokens, one per line.  This is mostly for debugging the lexer", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &identFreq{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	if cmdErr != nil {
		log.Fatal(cmdErr)
	}
}

func walkSource(sourceRoot string, proc processor) error {

	inSourceDir := func(root, path string) bool {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			panic(err)
		}
		switch filepath.Dir(relative) {
		case "Forms", "Methods", "Procedures":
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

type identFreq struct{}

func (ifr *identFreq) end() {}

func (ifr *identFreq) process(path string) error {
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
		case equilex.Identifier:
			fmt.Println(lit)
		}
	}
}

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

type token struct {
	tok equilex.Token
	lit string
}

type statement struct {
	tokens []token
}

func (stmt *statement) String() string {
	var buf bytes.Buffer
	for _, t := range stmt.tokens {
		buf.WriteString(t.lit)
	}
	return buf.String()
}

func (stmt *statement) add(tok equilex.Token, lit string) {
	stmt.tokens = append(stmt.tokens, token{tok, lit})
}
