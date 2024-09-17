package graph

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"strings"

	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func newGraph() *graph {
	return &graph{
		nodes: make(map[nodeId]*node),
		used:  make(map[nodeId]struct{}),
	}
}

type graph struct {
	nodes map[nodeId]*node
	used  map[nodeId]struct{}
}

func (c *graph) addNode(node *node) {
	c.nodes[node.nodeId] = node

	// Some referenced node types implicity exist even though we haven't
	// "found" them anywhere, because they don't exist in the source code.
	for referenced := range node.Refs {
		if referenced.Type == ntTable || referenced.Type == ntField || referenced.Type == ntIndex || referenced.Type == ntWorkArea {
			// Create node if missing
			_, ok := c.nodes[referenced]
			if !ok {
				node := newNode()
				node.nodeId = referenced
				node.Label = referenced.Name
				c.nodes[referenced] = node
			}

			// Mark as used
			c.used[referenced] = struct{}{}
		}
	}

}

func (c *graph) nodesSorted() []*node {
	allNodes := make([]*node, 0, len(c.nodes))
	for _, node := range c.nodes {
		allNodes = append(allNodes, node)
	}
	sort.Slice(allNodes, func(i, j int) bool {
		return (allNodes[i].Label + "_" + allNodes[i].nodeId.id()) < (allNodes[j].Label + "_" + allNodes[j].nodeId.id())
	})
	return allNodes
}

func (c *graph) writeGraph(output graphOutput) error {
	if err := output.Start(); err != nil {
		return err
	}

	allNodes := c.nodesSorted()

	for _, n := range allNodes {
		id := sanitiseId(n.id())

		var labels []string
		_, used := c.used[n.nodeId]
		if used {
			labels = []string{n.Type.String(), "used"}
		} else {
			labels = []string{n.Type.String()}
		}

		if err := output.AddNode(id, n.Label, labels); err != nil {
			return err
		}
	}

	missingRefs := make(map[nodeId]struct{})

	for _, fromModule := range allNodes {
		for _, toModule := range fromModule.refsSorted() {
			_, ok := c.nodes[toModule]
			if !ok {
				missingRefs[toModule] = struct{}{}
			}

			if err := output.AddReference(sanitiseId(fromModule.id()), sanitiseId(toModule.id())); err != nil {
				return err
			}
		}
	}

	missingSorted := make([]nodeId, 0, len(missingRefs))
	for missing := range missingRefs {
		missingSorted = append(missingSorted, missing)
	}
	sort.Slice(missingSorted, func(i int, j int) bool { return missingSorted[i].Name < missingSorted[j].Name })

	for _, n := range missingSorted {
		if err := output.AddNode(sanitiseId(n.id()), n.Name, []string{n.Type.String(), "missing"}); err != nil {
			return err
		}
	}

	if err := output.End(); err != nil {
		return err
	}
	return nil
}

func (cb *graph) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	br := bufio.NewReader(f)

	node := newNode()

	ppd := ""

	for {
		prefix, err := br.Peek(4)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch string(prefix) {
		case "FIL,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "FIL,130,")
			spl := strings.SplitN(s, ",", 2)

			fullName := spl[0]

			node.nodeId, node.Label = idAndLabelFromFullName(fullName)
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

			if node.nodeId.Type != ntPpl { // Hack. Skip ppl for now because we can't do it properly
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

			// mark this procedure as used
			cb.markPublicProcedureUsed(spl[0])

		case "PPD,":
			s, _ := br.ReadString('\n')

			s = strings.TrimPrefix(s, "PPD,17,")
			spl := strings.SplitN(s, ",", 2)
			name := spl[0]

			if ppd != "" {
				cb.addNode(node)
			} else {
				if node.Type != ntPpl {
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
			fmt.Printf("foo: %s\n", prefix)
		}
	}

	if node.Type != ntPpl {
		cb.addNode(node)
	}
	return nil
}

func (g *graph) markPublicProcedureUsed(name string) {
	id := newNodeId(name, ntPubProc)
	g.used[id] = struct{}{}
}

func idAndLabelFromFullName(fullName string) (nodeId, string) {
	var id nodeId
	var label string

	nameParts := strings.Split(fullName, ".")

	id.Name = strings.ToLower(nameParts[0])
	label = nameParts[0]

	switch strings.ToLower(nameParts[1]) {
	case "jcl":
		id.Type = ntMethod
	case "imp":
		id.Type = ntImport
	case "exp":
		id.Type = ntExport
	case "frm":
		id.Type = ntForm
	case "qry":
		id.Type = ntQuery
	case "rep":
		id.Type = ntReport
	case "ppl":
		id.Type = ntPpl
	default:
		id.Type = "UNKNOWN"
	}

	return id, label
}

func (g *graph) applyModules(modulesCsv, modudetCsv string) error {
	switch {
	case modulesCsv == "" && modudetCsv == "":
		return nil
	case (modulesCsv == "" && modudetCsv != "") || (modudetCsv == "" && modulesCsv != ""):
		return errors.New("module CSV files must both be provided")
	}

	modules, err := os.Open(modulesCsv)
	if err != nil {
		return fmt.Errorf("failed to open module file %s : %e", modulesCsv, err)
	}
	modulesReader := csv.NewReader(bufio.NewReader(modules))
	modulesReader.ReuseRecord = true
	modulesReader.TrimLeadingSpace = true

	// Map of logic id to module name
	modNames := make(map[string]string)

	// skip header
	_, _ = modulesReader.Read()

	for {
		rec, err := modulesReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading modules csv : %e", err)
		}

		modNames[rec[6]] = rec[0]
	}

	modudet, err := os.Open(modudetCsv)
	if err != nil {
		return fmt.Errorf("failed to open module file %s : %e", modulesCsv, err)
	}
	modudetReader := csv.NewReader(bufio.NewReader(modudet))
	modudetReader.ReuseRecord = true
	modudetReader.TrimLeadingSpace = true

	since, err := time.Parse("2006-01-02", "2024-01-01")
	if err != nil {
		return err
	}

	usedModuleLogics := make(map[string]struct{})

	// skip header
	_, _ = modudetReader.Read()

	for {
		rec, err := modudetReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading modules csv : %e", err)
		}

		time, err := time.Parse("2006-01-02", rec[0])
		if err != nil {
			return fmt.Errorf("error parsing date '%s' from ModuDet CSV : %e", rec[0], err)
		}
		if !time.Before(since) {
			usedModuleLogics[rec[13]] = struct{}{}
		}
	}

	for logic := range usedModuleLogics {
		name, ok := modNames[logic]
		if !ok {
			return fmt.Errorf("unknown module with logic id %s", logic)
		}

		// Some are truncated. Skip those.
		if strings.Contains(name, ".") {
			id, _ := idAndLabelFromFullName(name)
			if id.Type != "UNKNOWN" {
				g.used[id] = struct{}{}
			}
		}
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
	r.Refs[newNodeId(name, ntField)] = struct{}{}
}

func (r *node) addIndexRef(name string) {
	r.Refs[newNodeId(name, ntIndex)] = struct{}{}
}

func (r *node) addWrkRef(name string) {
	r.Refs[newNodeId(name, ntWorkArea)] = struct{}{}
}

func (r *node) addSubtableRef(name string) {
	r.Refs[newNodeId(name, ntTable)] = struct{}{}
}

func (r *node) addPublicProcedureRef(name string) {
	r.Refs[newNodeId(name, ntPubProc)] = struct{}{}
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