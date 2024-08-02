package main

import (
	"fmt"
	"log"
	"slices"

	//"net/url"
	"strings"

	"github.com/utilitywarehouse/equilex"
)

func newCalledMissingMethods() *callsMissingMethod {
	return &callsMissingMethod{
		e: newExecutions(),
	}
}

type callsMissingMethod struct {
	f forms
	m methods
	p pubProcs
	e *executions
}

func (c *callsMissingMethod) end() error {

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
						fmt.Printf("%s calls missing method %s\n", fromModule, to)
					}
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

func (c *callsMissingMethod) processAll(sourceRoot string) error {
	return walkSource(sourceRoot, c)
}

func (c *callsMissingMethod) process(path string) error {
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
