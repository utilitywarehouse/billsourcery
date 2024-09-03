package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"slices"

	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
)

type graphOutput interface {
	Start() error
	End() error
	UpsertMethod(m *module, missing bool) error
	UpsertForm(m *module) error
	UpsertPublicProcedure(m *module) error
	UpsertCall(from *module, to *module) error
}

func newGVCalls() *calls {
	return &calls{
		e:              newExecutions(),
		output:         &DotGraphOutput{},
		missingMethods: make(map[module]struct{}),
	}
}

func newCalls() *calls {
	return &calls{
		e:              newExecutions(),
		output:         &NeoGraphOutput{},
		missingMethods: make(map[module]struct{}),
	}
}

type calls struct {
	f forms
	m methods
	p pubProcs
	e *executions

	missingMethods map[module]struct{}

	output graphOutput
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

func (o *DotGraphOutput) UpsertMethod(m *module, missing bool) error {
	if missing {
		fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"red\"]\n", encodeIDForDotfile(m), m)
	} else {
		fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"lightblue\"]\n", encodeIDForDotfile(m), m)
	}
	return nil
}

func (o *DotGraphOutput) UpsertForm(f *module) error {
	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"lightgreen\"]\n", encodeIDForDotfile(f), f)
	return nil
}

func (o *DotGraphOutput) UpsertPublicProcedure(mod *module) error {
	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"yellow\"]\n", encodeIDForDotfile(mod), mod)
	return nil
}

func (o *DotGraphOutput) UpsertCall(from *module, to *module) error {
	fmt.Printf("\t%s -> %s\n", encodeIDForDotfile(from), encodeIDForDotfile(to))
	return nil
}

type NeoGraphOutput struct{}

func (o *NeoGraphOutput) Start() error {
	return nil
}

func (o *NeoGraphOutput) End() error {
	return nil
}

func (o *NeoGraphOutput) UpsertMethod(m *module, missing bool) error {
	id := encodeIDForNeo(m)

	fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", id, id, m)

	if missing {
		fmt.Printf("SET %s :Method \nSET %s : Missing;\n", id, id)
	} else {
		fmt.Printf("SET %s :Method ;\n", id)
	}
	return nil
}

func (o *NeoGraphOutput) UpsertForm(f *module) error {
	id := encodeIDForNeo(f)

	fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", id, id, f)
	fmt.Printf("SET %s :Form ;\n", id)
	return nil
}

func (o *NeoGraphOutput) UpsertPublicProcedure(mod *module) error {
	id := encodeIDForNeo(mod)

	fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", id, id, mod)
	fmt.Printf("SET %s :PublicProcedure ;\n", id)
	return nil
}

func (o *NeoGraphOutput) UpsertCall(from *module, to *module) error {
	fmt.Printf("MERGE (f:Node {id: \"%s\"}) MERGE (t:Node {id: \"%s\"}) MERGE (f)-[:calls]->(t);\n", encodeIDForNeo(from), encodeIDForNeo(to))
	return nil
}

func (c *calls) end() error {
	if err := c.output.Start(); err != nil {
		return err
	}
	for _, m := range c.m.methods {
		if err := c.output.UpsertMethod(&m, false); err != nil {
			return err
		}
	}
	for _, f := range c.f.forms {
		if err := c.output.UpsertForm(&f); err != nil {
			return err
		}
	}
	for _, s := range c.p.stmts {
		switch {
		case s.tokens[0].tok == equilex.Public && s.tokens[1].tok == equilex.WS && s.tokens[2].tok == equilex.Procedure && s.tokens[3].tok == equilex.WS:
			value := s.tokens[4]
			if value.tok != equilex.Identifier {
				log.Panicf("bug : %v %v", value.tok, value.lit)
			}
			m := value.lit
			mod := module{m, mtProcedure}
			if err := c.output.UpsertPublicProcedure(&mod); err != nil {
				return err
			}
		default:
			log.Printf("skipping procedure %v\n", s)
		}
	}
	for from, e := range c.e.stmts {
		fromModule := moduleFromFullFilename(from)
		for _, stmt := range e {
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

					if !slices.Contains(c.m.methods, to_mod) {
						c.missingMethods[to_mod] = struct{}{}
					}

					if err := c.output.UpsertCall(&fromModule, &to_mod); err != nil {
						return err
					}
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
	}

	for m := range c.missingMethods {
		if err := c.output.UpsertMethod(&m, true); err != nil {
			return err
		}
	}

	if err := c.output.End(); err != nil {
		return err
	}
	return nil
}

func (c *calls) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, c)
}

func (c *calls) process(path string) error {
	if err := c.f.process(path); err != nil {
		return err
	}
	if err := c.m.process(path); err != nil {
		return err
	}
	if err := c.e.process(path); err != nil {
		return err
	}
	if err := c.p.process(path); err != nil {
		return err
	}
	return nil
}

func encodeIDForDotfile(mod *module) string {
	in := mod.moduleName
	// TODO: encode type in encoded name
	return "a" + hex.EncodeToString([]byte(in))
}

func encodeIDForNeo(mod *module) string {
	in := mod.moduleName
	// TODO: add type into encoded name
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
	in = strings.Map(f, in)
	return "a_" + in
}
