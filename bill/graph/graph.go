package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func PublicProcs(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	sort.Slice(calls.nodes, func(i, j int) bool { return calls.nodes[i].nodeName < calls.nodes[j].nodeName })
	for _, procedure := range calls.nodes {
		if procedure.nodeType == ntPubProc {
			fmt.Println(procedure)
		}
	}
	return nil
}

func Methods(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	for _, method := range calls.nodes {
		if method.nodeType == ntMethod {
			fmt.Println(method)
		}
	}
	return nil
}

func Forms(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	for _, form := range calls.nodes {
		if form.nodeType == ntForm {
			fmt.Println(form)
		}
	}
	return nil
}

func Reports(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	for _, report := range calls.nodes {
		if report.nodeType == ntReport {
			fmt.Println(report)
		}
	}
	return nil
}

func CallsNeo(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	return calls.writeGraph(&NeoGraphOutput{})
}

func CallsDot(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	return calls.writeGraph(&DotGraphOutput{})
}

func CalledMissingMethods(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}

	for fromModule, toModules := range calls.calls {
		for _, toModule := range toModules {
			if !slices.Contains(calls.nodes, *toModule) {
				fmt.Printf("%s calls missing method %s\n", fromModule, toModule)
			}
		}
	}

	return nil
}

func walkSource(sourceRoot string, proc fileProcessor) error {
	inSourceDir := func(root, path string) bool {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			panic(err)
		}
		switch filepath.Dir(relative) {
		case "Exports", "Forms", "Imports", "Methods", "Procedures", "Processes", "Queries", "Reports":
			return true
		default:
			return false
		}
	}

	return filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".txt") && inSourceDir(sourceRoot, path) {
			//	log.Println(path)
			err := proc.process(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type fileProcessor interface {
	process(path string) error
}
