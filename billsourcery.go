package main

import (
	"log"
	"os"
	"os/user"
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

	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dir := filepath.Join(user.HomeDir, "work/uw-bill-source-history")

	sourceRoot := app.StringOpt("source-root", dir, "Root directory for equinox source. Subdirs Methods/ Forms/ etc are expected")

	app.Command("stats", "Provide basic stats about the source code", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &statsProcessor{})
		}
	})

	app.Command("timestats-image", "Provide stats over time about the source code in a png/jpg/svg", func(cmd *cli.Cmd) {
		cacheDB := cmd.StringOpt("cache-db", os.Getenv("HOME")+"/.billsourcery_timestats_cache", "timestats cache")
		notBefore := cmd.StringOpt("earliest", "c7937fbe95bbef245d627dccad0dfc4baad35b7c", "Do not include data from before this revision")
		branchesCs := cmd.StringOpt("branches", "master", "which branches to cover (comma separated, no spaces")
		output := cmd.StringOpt("output", "/tmp/billtimestats.png", "output graph for stats over time")
		cmd.Action = func() {
			branches := strings.Split(*branchesCs, ",")
			doProcessAll(*sourceRoot, newTimeStatsImageProcessor(*cacheDB, *notBefore, branches, *output))
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

	app.Command("all-modules", "List all modules (not procedures)", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, &allModules{})
		}
	})

	app.Command("calls-neo", "Produce neo4j cypher statements to create bill call graph. (Procedures not supported properly yet)", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, newCalls())
		}
	})

	app.Command("calls-dot", "Produce a .dot file of calls", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			doProcessAll(*sourceRoot, newGVCalls())
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

	app.Command("calls-stats-table", "Produce a table of module call counts", func(cmd *cli.Cmd) {
		dsn := cmd.String(cli.StringOpt{
			Name:      "dsn",
			Desc:      "bill pg mirror data source name",
			Value:     "postgres://root:xxxxxxxx@hlsv0pgrs01.tp.private:5432/bill?sslmode=disable",
			EnvVar:    "BILL_PG_MIRROR_DSN",
			HideValue: true,
		})
		cmd.Action = func() {
			callStatsTable(*sourceRoot, *dsn)
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
