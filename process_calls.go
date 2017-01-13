package main

import (
	"fmt"
	"log"
	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
)

func newCalls() *calls {
	return &calls{
		e: newExecutions(),
	}
}

type calls struct {
	m methods
	p pubProcs
	e *executions
	f forms
}

func (c *calls) end() error {
	for _, m := range c.m.methods {
		fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", encodeIDForDotfile(m), encodeIDForDotfile(m), m)
		fmt.Printf("SET %s :Method ;\n", encodeIDForDotfile(m))
	}
	for _, f := range c.f.forms {
		fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", encodeIDForDotfile(f), encodeIDForDotfile(f), f)
		fmt.Printf("SET %s :Form ;\n", encodeIDForDotfile(f))
	}
	for _, s := range c.p.stmts {
		switch {
		case s.tokens[0].tok == equilex.Public && s.tokens[1].tok == equilex.WS && s.tokens[2].tok == equilex.Procedure && s.tokens[3].tok == equilex.WS:
			value := s.tokens[4]
			if value.tok != equilex.Identifier {
				log.Panicf("bug : %v %v", value.tok, value.lit)
			}
			m := value.lit
			fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", encodeIDForDotfile(m), encodeIDForDotfile(m), m)
			fmt.Printf("SET %s :PublicProcedure ;\n", encodeIDForDotfile(m))
		default:
			log.Printf("skipping procedure %v\n", s)
		}
	}
	for from, e := range c.e.stmts {
		from = filenameToIdentifier(from)
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
			case equilex.OptimiseDatabase:
			case equilex.Method:
				to := toks[4].lit

				to = strings.ToLower(to)
				to = to[1 : len(to)-1]
				if strings.HasSuffix(to, ".jcl") {
					to = to[0 : len(to)-4]
				}

				//log.Printf("from and to are %v %v\n", from, to)
				fmt.Printf("MERGE (f:Node {id: \"%s\"}) MERGE (t:Node {id: \"%s\"}) MERGE (f)-[:calls]->(t);\n", encodeIDForDotfile(from), encodeIDForDotfile(to))
			default:
				for i, t := range toks {
					log.Printf("tok %d is %v\n", i, t.lit)
				}
				return fmt.Errorf("unhandled type : '%#v' for statement %v\n", (toks[2].lit), stmt)
			}
		}
	}
	return nil
}

func (c *calls) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, c)
}

func (c *calls) process(path string) error {
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

func encodeIDForDotfile(in string) string {
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
