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
	process(path string) error
}

func main() {

	log.SetFlags(0)

	app := cli.App("billsourcery", "Bill source code attempted wizardry")

	sourceRoot := app.StringOpt("source-root", "/home/mgarton/work/uw-bill-source-history", "Root directory for equinox source. Subdirs Methods/ Forms/ etc are expected")

	app.Command("stats", "Provide basic stats about the source code", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &statsProcessor{})
		}
	})

	app.Command("strip-comments", "Remove comments from the source files", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &commentStripper{})
		}
	})

	app.Command("string-constants", "Dump all \" delimited string constants found in the source, one per line, to stdout (multi-line strings not included)", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &stringConsts{})
		}
	})

	app.Command("executes", "List execute statements. Incomplete", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, newExecutions())
		}
	})

	app.Command("public-procs", "List public procedures and public externals", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &pubProcs{})
		}
	})

	app.Command("methods", "List method names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &methods{})
		}
	})

	app.Command("forms", "List form names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &forms{})
		}
	})

	app.Command("processes", "List process names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &processes{})
		}
	})

	app.Command("reports", "List report names", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &reports{})
		}
	})

	app.Command("lexer-check", "Ensure the lexer can correctly scan all source. This is mostly for debugging the lexer", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &lexCheck{})
		}
	})

	app.Command("identifiers", "List identifier tokens, one per line.  This is mostly for debugging the lexer", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcess(*sourceRoot, &identifiers{})
		}
	})

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func doProcess(sourceRoot string, proc processor) {
	cmdErr := walkSource(sourceRoot, proc)
	if cmdErr != nil {
		log.Fatal(cmdErr)
	}

	proc.end()
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
