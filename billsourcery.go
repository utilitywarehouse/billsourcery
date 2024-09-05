package main

import (
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/billsourcery/bill"

	_ "net/http/pprof"
)

func init() {
	go func() {
		panic(http.ListenAndServe(":6060", nil))
	}()
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
				Action: func(ctx *cli.Context) error {
					return bill.Stats(ctx.String("source-root"))
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
					return bill.TimestatsImage(
						cCtx.String("source-root"),
						cCtx.String("cache-db"),
						cCtx.String("earliest"),
						cCtx.StringSlice("branches"),
						cCtx.String("output"))
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
					return bill.TimestatsBqRaw(
						ctx.String("source-root"),
						ctx.String("cache-db"),
						ctx.String("earliest"),
						ctx.StringSlice("branches"),
					)
				},
			},
			{
				Name:  "strip-comments",
				Usage: "Remove comments from the source files",
				Action: func(ctx *cli.Context) error {
					return bill.StripComments(ctx.String("source-root"))
				},
			},
			{
				Name:  "string-constants",
				Usage: "Dump all \" delimited string constants found in the source, one per line, to stdout (multi-line strings not included)",
				Action: func(ctx *cli.Context) error {
					return bill.StringConstants(ctx.String("source-root"))
				},
			},
			{
				Name:  "public-procs",
				Usage: "List public procedures",
				Action: func(ctx *cli.Context) error {
					return bill.PublicProcs(ctx.String("source-root"))
				},
			},
			{
				Name:  "methods",
				Usage: "List method names",
				Action: func(ctx *cli.Context) error {
					return bill.Methods(ctx.String("source-root"))
				},
			},
			{
				Name:  "forms",
				Usage: "List form names",
				Action: func(ctx *cli.Context) error {
					return bill.Forms(ctx.String("source-root"))
				},
			},
			{
				Name:  "reports",
				Usage: "List report names",
				Action: func(ctx *cli.Context) error {
					return bill.Reports(ctx.String("source-root"))
				},
			},
			{
				Name:  "all-modules",
				Usage: "List all modules (not procedures)",
				Action: func(ctx *cli.Context) error {
					return bill.AllModules(ctx.String("source-root"))
				},
			},
			{
				Name:  "calls-neo",
				Usage: "Produce neo4j cypher statements to create bill call graph. (Procedures not supported properly yet)",
				Action: func(ctx *cli.Context) error {
					return bill.CallsNeo(ctx.String("source-root"))
				},
			},
			{
				Name:  "calls-dot",
				Usage: "Produce a .dot file of calls",
				Action: func(ctx *cli.Context) error {
					return bill.CallsDot(ctx.String("source-root"))
				},
			},
			{
				Name:  "called-missing-methods",
				Usage: "List any methods that are called but do not exist",
				Action: func(ctx *cli.Context) error {
					return bill.CalledMissingMethods(ctx.String("source-root"))
				},
			},
			{
				Name:  "lexer-check",
				Usage: "Ensure the lexer can correctly scan all source. This is mostly for debugging the lexer",
				Action: func(ctx *cli.Context) error {
					return bill.LexerCheck(ctx.String("source-root"))
				},
			},
			{
				Name:  "identifiers",
				Usage: "List identifier tokens, one per line.  This is mostly for debugging the lexer",
				Action: func(ctx *cli.Context) error {
					return bill.Identifiers(ctx.String("source-root"))
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
					return bill.CallStatsTable(ctx.String("source-root"), ctx.String("dsn"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
