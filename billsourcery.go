package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jawher/mow.cli"
)

type processor interface {
	end() error
	processAll(sourceRoot string) error
}

type fileProcessor interface {
	process(path string) error
}

func main() {

	log.SetFlags(0)

	app := cli.App("billsourcery", "Bill source code attempted wizardry")

	sourceRoot := app.StringOpt("source-root", "/home/mgarton/work/uw-bill-source-history", "Root directory for equinox source. Subdirs Methods/ Forms/ etc are expected")

	app.Command("stats", "Provide basic stats about the source code", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &statsProcessor{})
		}
	})

	app.Command("strip-comments", "Remove comments from the source files", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &commentStripper{})
		}
	})

	app.Command("string-constants", "Dump all \" delimited string constants found in the source, one per line, to stdout (multi-line strings not included)", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &stringConsts{})
		}
	})

	app.Command("executes", "List execute statements. Incomplete", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, newExecutions())
		}
	})

	app.Command("public-procs", "List public procedures and public externals", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &pubProcs{})
		}
	})

	app.Command("methods", "List method names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &methods{})
		}
	})

	app.Command("forms", "List form names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &forms{})
		}
	})

	app.Command("processes", "List process names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &processes{})
		}
	})

	app.Command("reports", "List report names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &reports{})
		}
	})

	app.Command("lexer-check", "Ensure the lexer can correctly scan all source. This is mostly for debugging the lexer", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &lexCheck{})
		}
	})

	app.Command("identifiers", "List identifier tokens, one per line.  This is mostly for debugging the lexer", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &identifiers{})
		}
	})

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func doProcessAll(sourceRoot string, proc processor) {
	err := proc.processAll(sourceRoot)

	if err == nil {
		err = proc.end()
	}

	if err != nil {
		log.Fatal(err)
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
