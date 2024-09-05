package graph

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

func newCalls() *calls {
	return &calls{
		calls: make(map[node]([]*node)),
	}
}

type calls struct {
	forms   []node
	methods []node
	reports []node
	procs   []node
	calls   map[node]([]*node)
}

func (c *calls) writeGraph(output graphOutput) error {
	if err := output.Start(); err != nil {
		return err
	}

	sort.Slice(c.procs, func(i, j int) bool { return c.procs[i].nodeName < c.procs[j].nodeName })

	for _, m := range c.methods {
		id := encodeID(&m)

		if err := output.AddNode(id, m.nodeName, []string{m.nodeType.String()}); err != nil {
			return err
		}
	}
	for _, f := range c.forms {
		id := encodeID(&f)
		if err := output.AddNode(id, f.nodeName, []string{f.nodeType.String()}); err != nil {
			return err
		}
	}
	for _, r := range c.reports {
		id := encodeID(&r)
		if err := output.AddNode(id, r.nodeName, []string{r.nodeType.String()}); err != nil {
			return err
		}
	}

	for _, s := range c.procs {
		id := encodeID(&s)
		if err := output.AddNode(id, s.nodeName, []string{s.nodeType.String()}); err != nil {
			return err
		}
	}

	fromModuleSorted := make([]node, 0, len(c.calls))
	for k := range c.calls {
		fromModuleSorted = append(fromModuleSorted, k)
	}
	sort.Slice(fromModuleSorted, func(i int, j int) bool { return fromModuleSorted[i].nodeName < fromModuleSorted[j].nodeName })

	missingMethods := make(map[node]struct{})

	for _, fromModule := range fromModuleSorted {
		toModules := c.calls[fromModule]

		for _, toModule := range toModules {
			if !slices.Contains(c.methods, *toModule) {
				missingMethods[*toModule] = struct{}{}
			}

			if err := output.AddCall(encodeID(&fromModule), encodeID(toModule)); err != nil {
				return err
			}
		}
	}

	missingSorted := make([]node, 0, len(missingMethods))
	for missing := range missingMethods {
		missingSorted = append(missingSorted, missing)
	}
	sort.Slice(missingSorted, func(i int, j int) bool { return missingSorted[i].nodeName < missingSorted[j].nodeName })

	for _, m := range missingSorted {
		id := encodeID(&m)

		if err := output.AddNode(id, m.nodeName, []string{"method", "missing"}); err != nil {
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
		c.forms = append(c.forms, nodeFromFullFilename(file))
	} else if strings.HasSuffix(dir, "/Methods/") {
		c.methods = append(c.methods, nodeFromFullFilename(file))
	} else if strings.HasSuffix(dir, "/Reports/") {
		c.reports = append(c.reports, nodeFromFullFilename(file))
	}

	if !strings.HasSuffix(dir, "/Procedures/") {
		if err := c.processMethodCalls(path); err != nil {
			return err
		}
	}

	if err := c.processPublicProcs(path); err != nil {
		return err
	}
	return nil
}

// TODO: this doesn't currently handle calls from procedures to methods.
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

	fromModule := nodeFromFullFilename(filename(path))
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

				to_mod := node{to, ntMethod}

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
			c.procs = append(c.procs, node{s.tokens[4].lit, ntPubProc})
		} else {
			log.Printf("skipping procedure %v\n", s)
		}
	}

	return nil
}

func encodeID(node *node) string {
	baseId := node.nodeName + "_" + node.nodeType.String()
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

type node struct {
	nodeName string
	nodeType nodeType
}

func (m node) String() string {
	return m.nodeName
}

type nodeType string

const (
	ntExport  nodeType = "export"
	ntForm    nodeType = "form"
	ntImport  nodeType = "import"
	ntMethod  nodeType = "method"
	ntPubProc nodeType = "public_procedure"
	ntProcess nodeType = "process"
	ntQuery   nodeType = "query"
	ntReport  nodeType = "report"
)

func (mt nodeType) String() string {
	return string(mt)
}

func nodeFromFullFilename(filename string) node {
	filename = strings.ToLower(filename)
	var mt nodeType
	switch {
	case strings.HasSuffix(filename, ".ex@.txt"):
		mt = ntExport
	case strings.HasSuffix(filename, ".fr@.txt"):
		mt = ntForm
	case strings.HasSuffix(filename, ".im@.txt"):
		mt = ntImport
	case strings.HasSuffix(filename, ".jc@.txt"):
		mt = ntMethod
	case strings.HasSuffix(filename, ".pr@.txt"):
		mt = ntProcess
	case strings.HasSuffix(filename, ".qr@.txt"):
		mt = ntQuery
	case strings.HasSuffix(filename, ".re@.txt"):
		mt = ntReport
	default:
		log.Panicf("can't create node for filename : %s\n", filename)
	}
	name := filename[0 : len(filename)-8]
	return node{name, mt}
}
