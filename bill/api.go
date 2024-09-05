package bill

import (
	"os"
	"path/filepath"
	"strings"
)

func StripComments(sourceRoot string) error {
	return walkSource(sourceRoot, &commentStripper{})
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
