package main

import (
	"context"
	"log"
	"strings"
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
	client, err := bigquery.NewClient(ctx, "uw-net")
	if err != nil {
		return err
	}

	ds := client.Dataset("tmp")

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
	tab := ds.Table("bill_source_stats")

	log.Println("about to delete table")
	if err := tab.Delete(context.Background()); err != nil {
		if !strings.Contains(err.Error(), "notFound") {
			return err
		}
	}
delLoop:
	for {
		_, err := tab.Metadata(context.Background())
		switch {
		case err == nil:
			// still exists.
			time.Sleep(2 * time.Second)
		case strings.Contains(err.Error(), "notFound"):
			break delLoop
		default:
			return err
		}
	}
	log.Println("about to create table")
	s, err := bigquery.InferSchema(rows[0])
	if err != nil {
		return err
	}
	if err := tab.Create(context.Background(), &bigquery.TableMetadata{
		//		Name:   "tempbillstuff",
		Schema: s,
	}); err != nil {
		return err
	}

existLoop:
	for {
		_, err := tab.Metadata(context.Background())
		switch {
		case err == nil:
			// exists.
			break existLoop
		case strings.Contains(err.Error(), "notFound"):
			time.Sleep(2 * time.Second)
		default:
			return err
		}
	}

	log.Printf("about to upload data\n")
	if err := tab.Uploader().Put(context.Background(), rows); err != nil {
		return err
	}

	return nil
}
