package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	_ "net/http/pprof"
)

func init() {
	go func() {
		panic(http.ListenAndServe(":6060", nil))
	}()
}

type processor interface {
	end() error
	processAll(sourceRoot string) error
}

type fileProcessor interface {
	process(path string) error
}

func main() {
	log.SetFlags(0)

	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join(user.HomeDir, "work/uw-bill-source-history")

	app := &cli.App{
		Name:  "billsourcery",
		Usage: "Bill source code attempted wizardry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "source-root",
				Value: dir,
				Usage: "Root directory for equinox source. Subdirs Methods/ Forms/ etc are expected",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "stats",
				Usage: "Provide basic stats about the source code",
				Action: func(cCtx *cli.Context) error {
					return doProcessAll(cCtx.String("source-root"), &statsProcessor{})
				},
			},
			{
				Name:  "timestats-image",
				Usage: "Provide stats over time about the source code in a png/jpg/svg",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "cache-db",
						Value: defaultCacheName(),
						Usage: "timestats cache",
					},
					&cli.StringFlag{
						Name:  "earliest",
						Value: "c7937fbe95bbef245d627dccad0dfc4baad35b7c",
						Usage: "Do not include data from before this revision",
					},
					&cli.StringSliceFlag{
						Name:  "branches",
						Usage: "Branches to include in the stats",
						Value: cli.NewStringSlice("master"),
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "output graph for stats over time",
						Value: defaultOutputName(),
					},
				},
				Action: func(cCtx *cli.Context) error {
					processor := newTimeStatsImageProcessor(
						cCtx.String("cache-db"),
						cCtx.String("earliest"),
						cCtx.StringSlice("branches"),
						cCtx.String("output"))
					return doProcessAll(cCtx.String("source-root"), processor)
				},
			},
			{
				Name:  "timestats-bq-raw",
				Usage: "Provide raw (per file) stats over time about the source code and upload to bigquery",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "cache-db",
						Value: defaultCacheName(),
						Usage: "timestats cache",
					},
					&cli.StringFlag{
						Name:  "earliest",
						Value: "c7937fbe95bbef245d627dccad0dfc4baad35b7c",
						Usage: "Do not include data from before this revision",
					},
					&cli.StringSliceFlag{
						Name:  "branches",
						Usage: "which branches to cover (comma separated, no spaces",
						Value: cli.NewStringSlice("master"),
					},
				},
				Action: func(ctx *cli.Context) error {
					processor := newTimeStatsUnaggBQProcessor(
						ctx.String("cache-db"),
						ctx.String("earliest"),
						ctx.StringSlice("branches"),
					)
					return doProcessAll(ctx.String("source-root"), processor)
				},
			},
			{
				Name:  "strip-comments",
				Usage: "Remove comments from the source files",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &commentStripper{})
				},
			},
			{
				Name:  "string-constants",
				Usage: "Dump all \" delimited string constants found in the source, one per line, to stdout (multi-line strings not included)",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &stringConsts{})

				},
			},
			{
				Name:  "executes",
				Usage: "List execute statements. Incomplete",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), newExecutions())
				},
			},
			{
				Name:  "public-procs",
				Usage: "List public procedures and public externals",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &pubProcs{})
				},
			},
			{
				Name:  "methods",
				Usage: "List method names",
				Action: func(ctx *cli.Context) error {
					calls := newCalls()
					if err := walkSource(ctx.String("source-root"), calls); err != nil {
						return err
					}
					for _, method := range calls.methods {
						fmt.Println(method)
					}
					return nil
				},
			},
			{
				Name:  "forms",
				Usage: "List form names",
				Action: func(ctx *cli.Context) error {
					calls := newCalls()
					if err := walkSource(ctx.String("source-root"), calls); err != nil {
						return err
					}
					for _, form := range calls.forms {
						fmt.Println(form)
					}
					return nil
				},
			},
			{
				Name:  "reports",
				Usage: "List report names",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &reports{})
				},
			},
			{
				Name:  "all-modules",
				Usage: "List all modules (not procedures)",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &allModules{})
				},
			},
			{
				Name:  "calls-neo",
				Usage: "Produce neo4j cypher statements to create bill call graph. (Procedures not supported properly yet)",
				Action: func(ctx *cli.Context) error {
					calls := newCalls()
					if err := walkSource(ctx.String("source-root"), calls); err != nil {
						return err
					}
					return calls.writeGraph(&NeoGraphOutput{})
				},
			},
			{
				Name:  "calls-dot",
				Usage: "Produce a .dot file of calls",
				Action: func(ctx *cli.Context) error {
					calls := newCalls()
					if err := walkSource(ctx.String("source-root"), calls); err != nil {
						return err
					}
					return calls.writeGraph(&DotGraphOutput{})
				},
			},
			{
				Name:  "called-missing-methods",
				Usage: "List any methods that are called but do not exist",
				Action: func(ctx *cli.Context) error {
					calls := newCalls()
					if err := walkSource(ctx.String("source-root"), calls); err != nil {
						return err
					}

					for fromModule, toModules := range calls.calls {
						for _, toModule := range toModules {
							if !slices.Contains(calls.methods, *toModule) {
								fmt.Printf("%s calls missing method %s\n", fromModule, toModule)
							}
						}
					}

					return nil
				},
			},
			{
				Name:  "lexer-check",
				Usage: "Ensure the lexer can correctly scan all source. This is mostly for debugging the lexer",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &lexCheck{})
				},
			},
			{
				Name:  "identifiers",
				Usage: "List identifier tokens, one per line.  This is mostly for debugging the lexer",
				Action: func(ctx *cli.Context) error {
					return doProcessAll(ctx.String("source-root"), &identifiers{})
				},
			},
			{
				Name:  "calls-stats-table",
				Usage: "Produce a table of module call counts",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "dsn",
						Usage: "bill pg mirror data source name",
						Value: "postgres://root:xxxxxxxx@hlsv0pgrs01.tp.private:5432/bill?sslmode=disable",
					},
				},

				Action: func(ctx *cli.Context) error {
					return callStatsTable(ctx.String("source-root"), ctx.String("dsn"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func doProcessAll(sourceRoot string, proc processor) error {
	err := proc.processAll(sourceRoot)
	if err != nil {
		return err
	}

	if err := proc.end(); err != nil {
		return err
	}
	return nil
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
