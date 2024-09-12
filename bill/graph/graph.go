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
	calls := newGraph()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}

	var allNodes []*node
	for _, node := range calls.nodes {
		allNodes = append(allNodes, node)
	}
	sort.Slice(allNodes, func(i, j int) bool { return allNodes[i].Label < allNodes[j].Label })

	for _, node := range allNodes {
		if node.Type == nodeType {
			fmt.Println(node)
		}
	}
	return nil
}

func Graph(sourceRoot string, output string) error {

	var graphOutput graphOutput
	switch output {
	case "neo":
		graphOutput = &NeoGraphOutput{}
	case "dot":
		graphOutput = &DotGraphOutput{}
	default:
		return fmt.Errorf("unknown graph output : '%s'", output)
	}

	calls := newGraph()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}
	return calls.writeGraph(graphOutput)

}

func CalledMissingMethods(sourceRoot string) error {
	calls := newGraph()
	if err := walkSource(sourceRoot, calls); err != nil {
		return err
	}

	for _, fromModule := range calls.nodesSorted() {
		for _, toModule := range fromModule.refsSorted() {
			if toModule.Type == ntMethod {
				_, ok := calls.nodes[toModule]
				if !ok {
					fmt.Printf("%s calls missing method %s\n", fromModule.Name, toModule.Name)
				}
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
