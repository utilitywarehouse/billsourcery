package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"log"
	"strconv"
	"time"

	"cloud.google.com/go/bigquery"
)

func newTimeStatsUnaggBQProcessor(cacheDB string, earliest string, branches []string) *timeStatsUnaggBQProcessor {
	return &timeStatsUnaggBQProcessor{rawTimeStatsProcessor{AllStats: make(map[string][]*rawTimeStatsSummary), cacheDB: cacheDB, earliestRevision: earliest, branches: branches}}
}

type timeStatsUnaggBQProcessor struct {
	rawTimeStatsProcessor
}

type bqRawStatsRow struct {
	Branch       string
	Date         time.Time
	ModuleName   string
	ModuleType   string
	CommentCount int
	OtherCount   int
}

func (lp *timeStatsUnaggBQProcessor) end() error {
	// set GOOGLE_APPLICATION_CREDENTIALS to json containing service account key.

	ctx := context.Background()

	var rows []bqRawStatsRow

	for branch, ts := range lp.AllStats {
		for _, thisResult := range ts {
			for _, thisEntry := range thisResult.Entries {
				row := bqRawStatsRow{
					Branch:       branch,
					Date:         thisResult.Time,
					ModuleName:   thisEntry.FileName,
					ModuleType:   thisEntry.FileType,
					CommentCount: thisEntry.CommentCount,
					OtherCount:   thisEntry.OtherCount,
				}
				rows = append(rows, row)
			}
		}
	}

	if len(rows) == 0 {
		log.Printf("no rows. not uploading to bigquery")
		return nil
	}

	log.Printf("starting bigquery stuff\n")
	client, err := bigquery.NewClient(ctx, "uw-net" /*, option.WithHTTPClient(hc)*/)
	if err != nil {
		return err
	}

	tableName := "bill_source_stats_raw"

	if err := deleteAndRecreateBQ(ctx, client, "tmp", tableName, rows[0]); err != nil {
		return err
	}

	tab := client.Dataset("tmp").Table(tableName)

	format := "2006-01-02 15:04:05" // "YYYY-MM-DD HH:MM[:SS"

	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	for _, r := range rows {
		if err := cw.Write([]string{r.Branch, r.Date.Format(format), r.ModuleName, r.ModuleType, strconv.Itoa(r.CommentCount), strconv.Itoa(r.OtherCount)}); err != nil {
			return err
		}
	}
	cw.Flush()

	// ioutil.WriteFile("/tmp/bill_source_stats_raw.csv", buf.Bytes(), 0644)
	//	log.Printf("%s\n", buf.Bytes())

	rs := bigquery.NewReaderSource(&buf)

	j, err := tab.LoaderFrom(rs).Run(ctx)
	if err != nil {
		return err
	}
	status, err := j.Wait(ctx)
	if err != nil {
		return err
	}
	if status.State != bigquery.Done {
		panic("bug?")
	}
	return nil
}
