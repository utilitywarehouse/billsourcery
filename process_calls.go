package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"

	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type graphOutput interface {
	Start() error
	End() error
	AddNode(id string, name string, tags []string) error
	AddCall(from_id string, to_id string) error
}

func newCalls() *calls {
	return &calls{
		calls:          make(map[module]([]*module)),
		missingMethods: make(map[module]struct{}),
	}
}

type calls struct {
	forms   []module
	methods []module
	reports []module
	procs   []string
	calls   map[module]([]*module)

	missingMethods map[module]struct{}
}

type DotGraphOutput struct{}

func (o *DotGraphOutput) Start() error {
	fmt.Println("digraph calls {")
	return nil
}

func (o *DotGraphOutput) End() error {
	fmt.Println("}")
	return nil
}

func (o *DotGraphOutput) AddNode(id string, name string, tags []string) error {
	colour := ""

	if slices.Contains(tags, "form") {
		colour = "lightgreen"
	} else if slices.Contains(tags, "report") {
		colour = "orange"
	} else if slices.Contains(tags, "public_procedure") {
		colour = "yellow"
	} else if slices.Contains(tags, "method") {
		if slices.Contains(tags, "missing") {
			colour = "red"
		} else {
			colour = "lightblue"
		}
	}

	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"%s\"]\n", id, name, colour)

	return nil
}

func (o *DotGraphOutput) AddCall(from string, to string) error {
	fmt.Printf("\t%s -> %s\n", from, to)
	return nil
}

type NeoGraphOutput struct{}

func (o *NeoGraphOutput) Start() error {
	return nil
}

func (o *NeoGraphOutput) End() error {
	return nil
}

func (o NeoGraphOutput) AddNode(id string, name string, tags []string) error {

	fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", id, id, name)

	if slices.Contains(tags, "form") {
		fmt.Printf("SET %s :Form ;\n", id)
	} else if slices.Contains(tags, "report") {
		fmt.Printf("SET %s :Report ;\n", id)
	} else if slices.Contains(tags, "public_procedure") {
		fmt.Printf("SET %s :PublicProcedure ;\n", id)
	} else if slices.Contains(tags, "method") {
		if slices.Contains(tags, "missing") {
			fmt.Printf("SET %s :Method \nSET %s : Missing;\n", id, id)
		} else {
			fmt.Printf("SET %s :Method ;\n", id)
		}
	}

	return nil
}

func (o *NeoGraphOutput) AddCall(from string, to string) error {
	fmt.Printf("MERGE (f:Node {id: \"%s\"}) MERGE (t:Node {id: \"%s\"}) MERGE (f)-[:calls]->(t);\n", from, to)
	return nil
}

func (c *calls) writeGraph(output graphOutput) error {
	if err := output.Start(); err != nil {
		return err
	}
	for _, m := range c.methods {
		id := encodeID(&m)

		if err := output.AddNode(id, m.moduleName, []string{"method"}); err != nil {
			return err
		}
	}
	for _, f := range c.forms {
		id := encodeID(&f)
		if err := output.AddNode(id, f.moduleName, []string{"form"}); err != nil {
			return err
		}
	}
	for _, r := range c.reports {
		id := encodeID(&r)
		if err := output.AddNode(id, r.moduleName, []string{"report"}); err != nil {
			return err
		}
	}

	sort.Strings(c.procs)

	for _, s := range c.procs {
		mod := module{s, mtProcedure}
		id := encodeID(&mod)

		if err := output.AddNode(id, mod.moduleName, []string{"public_procedure"}); err != nil {
			return err
		}
	}

	fromModuleSorted := make([]module, 0, len(c.calls))
	for k := range c.calls {
		fromModuleSorted = append(fromModuleSorted, k)
	}
	sort.Slice(fromModuleSorted, func(i int, j int) bool { return fromModuleSorted[i].moduleName < fromModuleSorted[j].moduleName })

	for _, fromModule := range fromModuleSorted {
		toModules := c.calls[fromModule]

		for _, toModule := range toModules {
			if !slices.Contains(c.methods, *toModule) {
				c.missingMethods[*toModule] = struct{}{}
			}

			if err := output.AddCall(encodeID(&fromModule), encodeID(toModule)); err != nil {
				return err
			}
		}
	}

	missingSorted := make([]module, 0, len(c.missingMethods))
	for missing := range c.missingMethods {
		missingSorted = append(missingSorted, missing)
	}
	sort.Slice(missingSorted, func(i int, j int) bool { return missingSorted[i].moduleName < missingSorted[j].moduleName })

	for _, m := range missingSorted {
		id := encodeID(&m)

		if err := output.AddNode(id, m.moduleName, []string{"method", "missing"}); err != nil {
			return err
		}
	}

	if err := output.End(); err != nil {
		return err
	}
	return nil
}

func (c *calls) process(path string) error {
	dir, file := filepath.Split(path)
	if strings.HasSuffix(dir, "/Forms/") {
		c.forms = append(c.forms, moduleFromFullFilename(file))
	} else if strings.HasSuffix(dir, "/Methods/") {
		c.methods = append(c.methods, moduleFromFullFilename(file))
	} else if strings.HasSuffix(dir, "/Reports/") {
		c.reports = append(c.reports, moduleFromFullFilename(file))
	}

	if err := c.processMethodCalls(path); err != nil {
		return err
	}

	if err := c.processPublicProcs(path); err != nil {
		return err
	}
	return nil
}

func (c *calls) processMethodCalls(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	stmts := make([]*statement, 0)

	var stmt *statement

loop:
	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.EOF:
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
			break loop
		case equilex.Execute:
			stmt = &statement{}
			stmt.add(tok, lit)
		case equilex.NewLine:
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
			stmt = nil
		default:
			if stmt != nil {
				stmt.add(tok, lit)
			}
		}
	}

	fromModule := moduleFromFullFilename(filename(path))
	for _, stmt := range stmts {
		toks := stmt.tokens
		for toks[0].tok != equilex.Execute {
			toks = toks[1:]
		}
		switch toks[2].tok {
		case equilex.Export:
		case equilex.Task:
		case equilex.Form:
		case equilex.FormSwap:
		case equilex.Query:
		case equilex.Process:
		case equilex.System:
		case equilex.Report:
		case equilex.ReportPreview:
		case equilex.Shell:
		case equilex.Command:
		case equilex.Import:
		case equilex.EmptyDatabase:
		case equilex.MethodSwap:
		case equilex.MethodSetup:
		case equilex.OptimiseDatabase:
		case equilex.OptimiseTable:
		case equilex.OptimiseTableIndexes:
		case equilex.OptimiseDatabaseIndexes:
		case equilex.OptimiseAllDatabases:
		case equilex.OptimiseAllDatabasesIndexes:
		case equilex.OptimiseDatabaseHelper:
		case equilex.ConvertAllDatabases:
		case equilex.Method:
			to := toks[4].lit

			if to[0] == '"' && to[len(to)-1] == '"' {
				to = strings.ToLower(to)
				to = to[1 : len(to)-1]
				to = strings.TrimSuffix(to, ".jcl")

				to_mod := module{to, mtMethod}

				c.calls[fromModule] = append(c.calls[fromModule], &to_mod)

			} else {
				log.Printf("call from %s to variable method '%s' - skipping", fromModule, to)
			}
		default:
			for i, t := range toks {
				log.Printf("tok %d is %v\n", i, t.lit)
			}
			return fmt.Errorf("unhandled type : '%#v' for statement %v", (toks[2].lit), stmt)
		}
	}

	return nil
}

func (c *calls) processPublicProcs(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

	stmts := make([]*statement, 0)

	stmt := &statement{}

loop:
	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch {
		case tok == equilex.EOF:
			if !stmt.empty() {
				stmts = append(stmts, stmt)
			}
			break loop
		case stmt.empty() && tok == equilex.Public:
			stmt.add(tok, lit)
		case tok == equilex.NewLine && !stmt.empty():
			stmts = append(stmts, stmt)
			stmt = &statement{}
		case !stmt.empty():
			stmt.add(tok, lit)
		}
	}

	for _, s := range stmts {
		if s.tokens[0].tok == equilex.Public && s.tokens[1].tok == equilex.WS && s.tokens[2].tok == equilex.Procedure && s.tokens[3].tok == equilex.WS {
			c.procs = append(c.procs, s.tokens[4].lit)
		} else {
			log.Printf("skipping procedure %v\n", s)
		}
	}

	return nil
}

func encodeID(mod *module) string {
	baseId := mod.moduleName + "_" + mod.moduleType.String()
	return sanitiseId(baseId)
}

func sanitiseId(baseId string) string {
	f := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}
	return "a_" + strings.Map(f, baseId)
}

func filename(path string) string {
	_, file := filepath.Split(path)
	return file
}
