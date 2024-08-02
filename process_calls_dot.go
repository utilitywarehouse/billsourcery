package main

import (
	"encoding/hex"
	"fmt"
	"log"

	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
)

func newGVCalls() *callsDot {
	return &callsDot{
		e: newExecutions(),
	}
}

type callsDot struct {
	f forms
	m methods
	p pubProcs
	e *executions
}

func (c *callsDot) upsertMethod(m *module) error {
	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"lightblue\"]\n", encodeIDForDotfile(m), m)
	return nil
}

func (c *callsDot) upsertForm(f *module) error {
	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"lightgreen\"]\n", encodeIDForDotfile(f), f)
	return nil
}

func (c *callsDot) upsertPublicProcedure(mod *module) error {
	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"yellow\"]\n", encodeIDForDotfile(mod), mod)
	return nil
}

func (c *callsDot) upsertCall(from *module, to *module) error {
	fmt.Printf("\t%s -> %s\n", encodeIDForDotfile(from), encodeIDForDotfile(to))
	return nil
}

func (c *callsDot) end() error {
	fmt.Println("digraph calls {")
	for _, m := range c.m.methods {
		if err := c.upsertMethod(&m); err != nil {
			return err
		}
	}
	for _, f := range c.f.forms {
		if err := c.upsertForm(&f); err != nil {
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
			if err := c.upsertPublicProcedure(&mod); err != nil {
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
					if err := c.upsertCall(&fromModule, &module{to, mtMethod}); err != nil {
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
	fmt.Println("}")
	return nil
}

func (c *callsDot) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, c)
}

func (c *callsDot) process(path string) error {
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
