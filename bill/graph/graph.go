package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func PublicProcs(sourceRoot string) error {
	return listNodeType(sourceRoot, ntPubProc)
}

func Methods(sourceRoot string) error {
	return listNodeType(sourceRoot, ntMethod)
}

func Forms(sourceRoot string) error {
	return listNodeType(sourceRoot, ntForm)
}

func Reports(sourceRoot string) error {
	return listNodeType(sourceRoot, ntReport)
}

func listNodeType(sourceRoot string, nodeType nodeType) error {
	calls := newCalls()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}

	var allNodes []*node
	for _, node := range calls.nodes {
		allNodes = append(allNodes, node)
	}
	sort.Slice(allNodes, func(i, j int) bool { return allNodes[i].Name < allNodes[j].Name })

	for _, node := range allNodes {
		if node.Type == nodeType {
			fmt.Println(node)
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
			_, ok := calls.nodes[toModule]
			if !ok {
				fmt.Printf("%s calls missing method %s\n", fromModule.Name, toModule.Name)
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
