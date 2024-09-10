package graph

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func newCalls() *calls {
	return &calls{
		nodes: make(map[nodeId]*node),
	}
}

type calls struct {
	nodes map[nodeId]*node
}

func (c *calls) writeGraph(output graphOutput) error {
	if err := output.Start(); err != nil {
		return err
	}

	var allNodes []*node
	for _, node := range c.nodes {
		allNodes = append(allNodes, node)
	}
	sort.Slice(allNodes, func(i, j int) bool { return allNodes[i].Label < allNodes[j].Label })

	for _, n := range allNodes {
		id := sanitiseId(n.id())

		if err := output.AddNode(id, n.Label, []string{n.Type.String()}); err != nil {
			return err
		}
	}

	missingMethods := make(map[nodeId]struct{})

	for _, fromModule := range allNodes {
		toModules := fromModule.Refs

		for toModule := range toModules {
			_, ok := c.nodes[toModule]
			if !ok {
				missingMethods[toModule] = struct{}{}
			}

			if err := output.AddCall(sanitiseId(fromModule.id()), sanitiseId(toModule.id())); err != nil {
				return err
			}
		}
	}

	missingSorted := make([]nodeId, 0, len(missingMethods))
	for missing := range missingMethods {
		missingSorted = append(missingSorted, missing)
	}
	sort.Slice(missingSorted, func(i int, j int) bool { return missingSorted[i].Name < missingSorted[j].Name })

	for _, n := range missingSorted {
		if err := output.AddNode(sanitiseId(n.id()), n.Name, []string{"method", "missing"}); err != nil {
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
	if strings.HasSuffix(dir, "/Forms/") ||
		strings.HasSuffix(dir, "/Methods/") ||
		strings.HasSuffix(dir, "/Reports/") {
		name, type_ := nodeFromFullFilename(file)
		node := newNode(name, type_)
		c.nodes[node.nodeId] = &node
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

func findMethodRefs(fromNodeId nodeId, text string) ([]string, error) {

	l := equilex.NewLexer(transform.NewReader(strings.NewReader(text), charmap.Windows1252.NewDecoder()))

	stmts := make([]*statement, 0)

	var stmt *statement

loop:
	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return nil, err
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

	var methodsRefs []string

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

				methodsRefs = append(methodsRefs, to)
			} else {
				log.Printf("call from %s to variable method '%s' - skipping", fromNodeId.Name, to)
			}
		default:
			for i, t := range toks {
				log.Printf("tok %d is %v\n", i, t.lit)
			}
			return nil, fmt.Errorf("unhandled type : '%#v' for statement %v", (toks[2].lit), stmt)
		}
	}

	return methodsRefs, nil

}

// TODO: this doesn't currently handle calls from procedures to methods.
func (c *calls) processMethodCalls(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	text := string(data)

	name, type_ := nodeFromFullFilename(filename(path))
	fromNodeId := newNodeId(name, type_)
	fromNode := c.nodes[fromNodeId]

	refs, err := findMethodRefs(fromNodeId, text)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		fromNode.addMethodRef(ref)
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
			node := newNode(s.tokens[4].lit, ntPubProc)
			c.nodes[node.nodeId] = &node
		} else {
			log.Printf("skipping procedure %v\n", s)
		}
	}

	return nil
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

type nodeId struct {
	Name string
	Type nodeType
}

func newNodeId(name string, type_ nodeType) nodeId {
	return nodeId{
		Name: strings.ToLower(name),
		Type: type_,
	}
}

func (n *nodeId) id() string {
	return n.Name + "_" + n.Type.String()
}

type node struct {
	nodeId
	Label string
	Refs  map[nodeId]struct{}
}

func (m node) String() string {
	return m.Name
}

func (r *node) addMethodRef(name string) {
	r.Refs[nodeId{Type: ntMethod, Name: name}] = struct{}{}
}

func newNode(name string, type_ nodeType) node {
	return node{
		nodeId: newNodeId(name, type_),
		Label:  name,
		Refs:   make(map[nodeId]struct{}),
	}
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

func nodeFromFullFilename(filename string) (string, nodeType) {
	filenameLower := strings.ToLower(filename)
	var mt nodeType
	switch {
	case strings.HasSuffix(filenameLower, ".ex@.txt"):
		mt = ntExport
	case strings.HasSuffix(filenameLower, ".fr@.txt"):
		mt = ntForm
	case strings.HasSuffix(filenameLower, ".im@.txt"):
		mt = ntImport
	case strings.HasSuffix(filenameLower, ".jc@.txt"):
		mt = ntMethod
	case strings.HasSuffix(filenameLower, ".pr@.txt"):
		mt = ntProcess
	case strings.HasSuffix(filenameLower, ".qr@.txt"):
		mt = ntQuery
	case strings.HasSuffix(filenameLower, ".re@.txt"):
		mt = ntReport
	default:
		log.Panicf("can't create node for filename : %s\n", filenameLower)
	}
	name := filename[0 : len(filename)-8]
	return name, mt
}
