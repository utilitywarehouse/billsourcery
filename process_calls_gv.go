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

func newGVCalls() *gvcalls {
	return &gvcalls{
		e: newExecutions(),
	}
}

type gvcalls struct {
	f forms
	m methods
	p pubProcs
	e *executions
}

func (c *gvcalls) end() error {
	fmt.Println("digraph calls {")
	for _, m := range c.f.forms {
		fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"lightgreen\"]\n", encodeIDForDotfile(m), m)
	}
	for _, m := range c.m.methods {
		//		log.Printf("method is %v \n", m)
		fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"lightblue\"]\n", encodeIDForDotfile(m), m)
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
					//fmt.Printf("%c  %c", to[0], to[len(to)-1])
					to = strings.ToLower(to)
					to = to[1 : len(to)-1]
					to = strings.TrimSuffix(to, ".jcl")

					//	log.Printf("from and to are %v %v\n", from, to)

					from_enc := encodeIDForDotfile(fromModule)
					to_mod := module{to, mtMethod}

					if !slices.Contains(c.m.methods, to_mod) {
						log.Printf("method %v is called from %v but not found - skipping\n", to_mod, fromModule)
					} else {

						to_enc := encodeIDForDotfile(to_mod)

						fmt.Printf("\t%s -> %s\n", from_enc, to_enc)
					}
				} else {
					log.Printf("call from %s to variable method '%s' - skipping", fromModule, to)
				}

			default:
				for i, t := range toks {
					log.Printf("tok %d is %v\n", i, t.lit)
				}
				return fmt.Errorf("unhandled type : '%#v' for statement %v\n", (toks[2].lit), stmt)
			}
		}
	}
	fmt.Println("}")
	return nil
}

func (c *gvcalls) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, c)
}

func (c *gvcalls) process(path string) error {
	if err := c.f.process(path); err != nil {
		return err
	}
	if err := c.m.process(path); err != nil {
		return err
	}
	if err := c.e.process(path); err != nil {
		return err
	}
	return c.p.process(path)
}

func encodeIDForDotfile(mod module) string {
	in := mod.moduleName
	// TODO: encode type in encoded name
	return "a" + hex.EncodeToString([]byte(in))
}
