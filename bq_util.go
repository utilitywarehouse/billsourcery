package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/bigquery"
)

func deleteAndRecreateBQ(ctx context.Context, client *bigquery.Client, dsName string, tableName string, example interface{}) error {
	tab := client.Dataset(dsName).Table(tableName)

	_, err := tab.Metadata(ctx)
	if err != nil {
		if !strings.Contains(err.Error(), "notFound") {
			return err
		}
		log.Printf("about to create table %s\n", tableName)
		s, err := bigquery.InferSchema(example)
		if err != nil {
			return err
		}
		if err := tab.Create(ctx, &bigquery.TableMetadata{
			Schema: s,
		}); err != nil {
			return err
		}
	} else {
		log.Printf("about to clear table %s\n", tableName)
		q := client.Query(fmt.Sprintf("DELETE FROM %s.%s where 1=1", dsName, tableName))
		q.UseLegacySQL = false
		j, err := q.Run(ctx)
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
		log.Printf("cleared table %s\n", tableName)
	}

	return nil
}
