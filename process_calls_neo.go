package main

import (
	"fmt"
	"log"

	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
)

func newCalls() *callsNeo {
	return &callsNeo{
		e: newExecutions(),
	}
}

type callsNeo struct {
	m methods
	p pubProcs
	e *executions
	f forms
}

func (c *callsNeo) end() error {
	for _, m := range c.m.methods {
		fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", encodeIDForNeo(m), encodeIDForNeo(m), m)
		fmt.Printf("SET %s :Method ;\n", encodeIDForNeo(m))
	}
	for _, f := range c.f.forms {
		fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", encodeIDForNeo(f), encodeIDForNeo(f), f)
		fmt.Printf("SET %s :Form ;\n", encodeIDForNeo(f))
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
			fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", encodeIDForNeo(mod), encodeIDForNeo(mod), mod)
			fmt.Printf("SET %s :PublicProcedure ;\n", encodeIDForNeo(mod))
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

					fmt.Printf("MERGE (f:Node {id: \"%s\"}) MERGE (t:Node {id: \"%s\"}) MERGE (f)-[:calls]->(t);\n", encodeIDForNeo(fromModule), encodeIDForNeo(module{to, mtMethod}))
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
	return nil
}

func (c *callsNeo) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, c)
}

func (c *callsNeo) process(path string) error {
	if err := c.m.process(path); err != nil {
		return err
	}
	if err := c.p.process(path); err != nil {
		return err
	}
	if err := c.e.process(path); err != nil {
		return err
	}
	if err := c.f.process(path); err != nil {
		return err
	}
	return nil
}

func encodeIDForNeo(mod module) string {
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
