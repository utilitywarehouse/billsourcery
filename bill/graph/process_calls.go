package graph

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"

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

func (c *calls) nodesSorted() []*node {
	allNodes := make([]*node, 0, len(c.nodes))
	for _, node := range c.nodes {
		allNodes = append(allNodes, node)
	}
	sort.Slice(allNodes, func(i, j int) bool {
		return (allNodes[i].Label + "_" + allNodes[i].nodeId.id()) < (allNodes[j].Label + "_" + allNodes[j].nodeId.id())
	})
	return allNodes
}

func (c *calls) writeGraph(output graphOutput) error {
	if err := output.Start(); err != nil {
		return err
	}

	allNodes := c.nodesSorted()

	for _, n := range allNodes {
		id := sanitiseId(n.id())

		if err := output.AddNode(id, n.Label, []string{n.Type.String()}); err != nil {
			return err
		}
	}

	missingMethods := make(map[nodeId]struct{})

	for _, fromModule := range allNodes {
		for _, toModule := range fromModule.refsSorted() {
			if toModule.Type == ntMethod {
				_, ok := c.nodes[toModule]
				if !ok {
					missingMethods[toModule] = struct{}{}
				}

				if err := output.AddCall(sanitiseId(fromModule.id()), sanitiseId(toModule.id())); err != nil {
					return err
				}
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

func (cb *calls) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	br := bufio.NewReader(f)

	node := newNode()

	ppd := ""

	for {
		foo, err := br.Peek(4)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch string(foo) {
		case "FIL,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "FIL,130,")
			spl := strings.SplitN(s, ",", 2)

			fullName := spl[0]

			nameParts := strings.Split(fullName, ".")

			node.Name = strings.ToLower(nameParts[0])
			node.Label = nameParts[0]
			switch strings.ToLower(nameParts[1]) {
			case "jcl":
				node.Type = ntMethod
			case "imp":
				node.Type = ntImport
			case "exp":
				node.Type = ntExport
			case "frm":
				node.Type = ntForm
			case "qry":
				node.Type = ntQuery
			case "rep":
				node.Type = ntReport
			case "ppl":
				node.Type = ntPpl
			default:
				node.Type = "UNKNOWN"
			}

		case "GRP,":
			_, _ = br.ReadString('\n')
		case "FLD,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "FLD,12,")
			spl := strings.SplitN(s, ",", 2)

			node.addFieldRef(spl[0])
		case "IDX,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "IDX,04,")
			spl := strings.SplitN(s, ",", 2)

			node.addIndexRef(spl[0])
		case "WRK,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "WRK,10,")
			spl := strings.SplitN(s, ",", 2)

			node.addWrkRef(spl[0])
		case "TXT,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "TXT,132,")
			spl := strings.SplitN(s, ",", 2)

			c, err := strconv.Atoi(strings.TrimSpace(spl[0]))
			if err != nil {
				return err
			}
			buf := make([]byte, c)
			_, err = io.ReadAtLeast(br, buf, c)
			if err != nil {
				return err
			}
			text := string(buf)
			node.addText(text)

			// We should have a XTX next, which we can discard
			s, _ = br.ReadString('\n')
			if !strings.HasPrefix(s, "XTX,") {
				return fmt.Errorf("expected XTX prefix, but got %s in file %s", strings.TrimSpace(s), node.Name)
			}

			if node.nodeId.Type != "public_procedure_library" { // Hack. Skip ppl for now because we can't do it properly
				// Find method calls in text
				refs, err := findMethodRefs(node.nodeId, text)
				if err != nil {
					return err
				}
				for _, ref := range refs {
					node.addMethodRef(ref)
				}
			}
		case "SUB,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "SUB,27,")
			spl := strings.SplitN(s, ",", 2)

			node.addSubtableRef(spl[0])
		case "TBL,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "TBL,16,")
			spl := strings.SplitN(s, ",", 2)

			node.addSubtableRef(spl[0])
		case "PPC,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "PPC,18,")
			spl := strings.SplitN(s, ",", 2)

			node.addPublicProcedureRef(spl[0])
		case "PPD,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "PPD,17,")
			spl := strings.SplitN(s, ",", 2)
			name := spl[0]

			if ppd != "" {
				cb.nodes[node.nodeId] = node
			} else {
				if node.Type != "public_procedure_library" {
					log.Fatalf("found public procedure definitions outside of a public procedure library: %s, %s", name, node.Type)
				}

				if len(node.Refs) != 0 {
					log.Fatalf("found public procedure library with references outside of the procedure definitions: %s %#v", path, node)
				}
			}

			node = newNode()
			node.Label = name
			node.nodeId = newNodeId(name, ntPubProc)

			ppd = name

		case "BLK,", "KLB,":
			// Ignore "blocks"
			_, _ = br.ReadString('\n')
		case "VAD,", "VAR,":
			// Ignore local variables
			_, _ = br.ReadString('\n')
		case "LPD,", "LPC,":
			// Ignore local procedures
			_, _ = br.ReadString('\n')

		case "AUD,", "AUT,":
			// Ignore autovars
			_, _ = br.ReadString('\n')
		case "DBS,":
			// Ignore database reference
			_, _ = br.ReadString('\n')
		case "OBN,", "OBP,", "EQP,":
			// Ignore various "equinox" bits
			_, _ = br.ReadString('\n')
		case "DTW,", "DPW,", "DLW,", "DBP,", "DLD,", "OBD,":
			// Ignore various definitions
			_, _ = br.ReadString('\n')
		case "DLC,":
			//log.Println("implement dll calls")
			_, _ = br.ReadString('\n')
		case "OBC,":
			//log.Println("implement OBC")
			_, _ = br.ReadString('\n')
		default:
			fmt.Printf("foo: %s\n", foo)
		}
	}

	if node.Type != ntPpl {
		cb.nodes[node.nodeId] = node
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
	Txt   []string `json:"-"`
	Refs  map[nodeId]struct{}
}

func (m node) String() string {
	return m.Name
}

func (r *node) addMethodRef(name string) {
	r.Refs[newNodeId(name, ntMethod)] = struct{}{}
}

func newNode() *node {
	return &node{
		Txt:  make([]string, 0),
		Refs: make(map[nodeId]struct{}),
	}
}

func (r *node) addText(t string) {
	r.Txt = append(r.Txt, t)
}

func (r *node) addFieldRef(name string) {
	r.Refs[nodeId{Type: ntField, Name: name}] = struct{}{}
}

func (r *node) addIndexRef(name string) {
	r.Refs[nodeId{Type: ntIndex, Name: name}] = struct{}{}
}

func (r *node) addWrkRef(name string) {
	r.Refs[nodeId{Type: ntWorkArea, Name: name}] = struct{}{}
}

func (r *node) addSubtableRef(name string) {
	r.Refs[nodeId{Type: ntTable, Name: name}] = struct{}{}
}

func (r *node) addPublicProcedureRef(name string) {
	r.Refs[nodeId{Type: ntPubProc, Name: name}] = struct{}{}
}

func (r *node) refsSorted() []nodeId {
	nodes := make([]nodeId, 0, len(r.Refs))
	for k := range r.Refs {
		nodes = append(nodes, k)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].id() < nodes[j].id()
	})
	return nodes
}

type nodeType string

const (
	ntExport   nodeType = "export"
	ntField    nodeType = "field"
	ntForm     nodeType = "form"
	ntImport   nodeType = "import"
	ntIndex    nodeType = "index"
	ntMethod   nodeType = "method"
	ntPpl      nodeType = "public_procedure_library"
	ntProcess  nodeType = "process"
	ntPubProc  nodeType = "public_procedure"
	ntQuery    nodeType = "query"
	ntReport   nodeType = "report"
	ntTable    nodeType = "table"
	ntWorkArea nodeType = "work_area"
)

func (mt nodeType) String() string {
	return string(mt)
}
