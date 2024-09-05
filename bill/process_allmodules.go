package bill

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type module struct {
	moduleName string
	moduleType moduleType
}

func (m module) String() string {
	return m.moduleName
}

func moduleFromFullFilename(filename string) module {
	filename = strings.ToLower(filename)
	var mt moduleType
	switch {
	case strings.HasSuffix(filename, ".ex@.txt"):
		mt = mtExport
	case strings.HasSuffix(filename, ".fr@.txt"):
		mt = mtForm
	case strings.HasSuffix(filename, ".im@.txt"):
		mt = mtImport
	case strings.HasSuffix(filename, ".jc@.txt"):
		mt = mtMethod
	case strings.HasSuffix(filename, ".pp@.txt"):
		mt = mtProcedure
	case strings.HasSuffix(filename, ".pr@.txt"):
		mt = mtProcess
	case strings.HasSuffix(filename, ".qr@.txt"):
		mt = mtQuery
	case strings.HasSuffix(filename, ".re@.txt"):
		mt = mtReport
	default:
		log.Panicf("bug: %s\n", filename)
	}
	name := filename[0 : len(filename)-8]
	return module{name, mt}
}

type moduleType string

const (
	mtExport    moduleType = "export"
	mtForm      moduleType = "form"
	mtImport    moduleType = "import"
	mtMethod    moduleType = "method"
	mtProcedure moduleType = "procedure"
	mtProcess   moduleType = "process"
	mtQuery     moduleType = "query"
	mtReport    moduleType = "report"
)

func (mt moduleType) String() string {
	return string(mt)
}

type allModules struct {
	modules []module
}

func (m *allModules) print() error {
	for _, method := range m.modules {
		fmt.Println(method)
	}
	return nil
}

func (m *allModules) process(path string) error {
	_, file := filepath.Split(path)
	m.modules = append(m.modules, moduleFromFullFilename(file))
	return nil
}

func mapModExt(in string) moduleType {
	switch in {
	case "exp":
		return mtExport
	case "frm":
		return mtForm
	case "fr":
		return mtForm
	case "f":
		return mtForm
	case "imp":
		return mtImport
	case "jcl":
		return mtMethod
	case "jc":
		return mtMethod
	case "j":
		return mtMethod
		//	return mtProcedure
		//	return mtProcess
	case "qry":
		return mtQuery
	case "rep":
		return mtReport
	default:
		panic(fmt.Sprintf("can't map module extension to type : %s\n", in))
	}
}
