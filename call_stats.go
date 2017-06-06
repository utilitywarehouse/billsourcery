package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
)

func callStatsTable(sourceRoot string, dsn string) {

	all := &allModules{}
	all.processAll(sourceRoot)

	// map of module name to call count
	counts := make(map[module]int)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	q := `select
		modudetlastparam caller, modname callto, count(*)
		from equinox.modules mods
		inner join equinox.modudet dets on mods.equinox_lrn = dets.equinox_prn and modudetsofttype='Client/Server'
		group by caller, callto
		order by count(*) desc
		;
	`

	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		var caller string
		var callto string
		var count int
		if err := rows.Scan(&caller, &callto, &count); err != nil {
			log.Fatal(err)
		}
		spl := strings.Split(strings.ToLower(callto), ".")
		mod := module{moduleName: spl[0], moduleType: mapModExt(spl[1])}
		counts[mod] = counts[mod] + count
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	type tableEntry struct {
		ModuleName string
		ModuleType string
		CallCount  int
	}

	var results []tableEntry

	for _, mod := range all.modules {
		results = append(results, tableEntry{mod.moduleName, mod.moduleType, counts[mod]})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].ModuleName < results[j].ModuleName })
	sort.SliceStable(results, func(i, j int) bool { return results[i].CallCount < results[j].CallCount })

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader([]string{"module", "type", "call count"})
	for _, res := range results {
		tw.Append([]string{res.ModuleName, res.ModuleType, strconv.Itoa(res.CallCount)})
	}
	tw.Render()
}

func mapModExt(in string) string {
	switch in {
	case "exp":
		return "Exports"
	case "frm":
		return "Forms"
	case "imp":
		return "Imports"
	case "jcl":
		return "Methods"
		//	return "Procedures"
		//	return "Processes"
	case "qry":
		return "Queries"
	case "rep":
		return "Reports"
	default:
		panic(fmt.Sprintf("can't map module extension to type : %s\n", in))
	}
}
