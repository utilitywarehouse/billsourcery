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
	sort.Strings(calls.procs)
	for _, procedure := range calls.procs {
		fmt.Println(procedure)
	}
	return nil
}

func Methods(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	for _, method := range calls.methods {
		fmt.Println(method)
	}
	return nil
}

func Forms(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	for _, form := range calls.forms {
		fmt.Println(form)
	}
	return nil
}

func Reports(sourceRoot string) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	for _, report := range calls.reports {
		fmt.Println(report)
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
			if !slices.Contains(calls.methods, *toModule) {
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
