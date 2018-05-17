package main

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
)

func newTimeStatsBQProcessor(cacheDB string, earliest string, branches []string) *timeStatsBQProcessor {
	return &timeStatsBQProcessor{timeStatsProcessor{AllStats: make(map[string][]*timeStatsEntry), cacheDB: cacheDB, earliestRevision: earliest, branches: branches}}
}

type timeStatsBQProcessor struct {
	timeStatsProcessor
}

type bqStatsRow struct {
	Branch       string
	Date         time.Time
	FileCount    int
	CommentCount int
	OtherCount   int
}

func (lp *timeStatsBQProcessor) end() error {

	// set GOOGLE_APPLICATION_CREDENTIALS to json containing service account key.

	ctx := context.Background()

	var rows []bqStatsRow

	for branch, ts := range lp.AllStats {
		for _, thisStat := range ts {
			row := bqStatsRow{
				Branch:       branch,
				Date:         thisStat.Time,
				FileCount:    thisStat.Results.FileCount,
				CommentCount: thisStat.Results.CommentCount,
				OtherCount:   thisStat.Results.OtherCount,
			}
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		log.Printf("no rows. not uploading to bigquery")
		return nil
	}

	log.Printf("starting bigquery stuff\n")

	client, err := bigquery.NewClient(ctx, "uw-net")
	if err != nil {
		return err
	}
	deleteAndRecreateBQ(ctx, client, "tmp", "bill_source_stats", rows[0])

	tab := client.Dataset("tmp").Table("bill_source_stats")

	log.Printf("about to upload data\n")
	if err := tab.Uploader().Put(context.Background(), rows); err != nil {
		return err
	}

	return nil
}
