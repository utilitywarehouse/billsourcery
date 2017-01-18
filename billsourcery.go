package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jawher/mow.cli"
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
			proc := newExecutions()
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("public-procs", "List public procedures and public externals", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &pubProcs{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("methods", "List method names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &methods{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("forms", "List form names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &forms{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("processes", "List process names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &processes{}
			cmdErr = walkSource(*sourceRoot, proc)
			if cmdErr != nil {
				return
			}

			proc.end()
		}
	})

	app.Command("reports", "List report names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			proc := &reports{}
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
			proc := &identifiers{}
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
