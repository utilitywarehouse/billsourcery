package bill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func StripComments(sourceRoot string) error {
	return walkSource(sourceRoot, &commentStripper{})
}

func ExtractPlainSource(sourceRoot string, targetDir string) error {
	pse := &plainSourceExtractor{sourceRoot, targetDir, 0, 0}
	if err := walkSource(sourceRoot, pse); err != nil {
		return err
	}
	fmt.Printf("Wrote %d plain output files from %d input files\n", pse.outputs, pse.inputs)
	return nil
}

func StringConstants(sourceRoot string) error {
	return walkSource(sourceRoot, &stringConsts{})
}

func LexerCheck(sourceRoot string) error {
	return walkSource(sourceRoot, &lexCheck{})
}

func Identifiers(sourceRoot string) error {
	return walkSource(sourceRoot, &identifiers{})
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
