package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type timeStatsEntry struct {
	Time    time.Time       `json:"timestamp"`
	Results *statsProcessor `json:"results"`
}

type cacheEntry struct {
	TimeStats *timeStatsEntry `json:"stats,omitempty"`
	Err       string          `json:"error,omitempty"`
}

type timeStatsProcessor struct {
	AllStats map[string][]*timeStatsEntry

	cacheDB          string
	earliestRevision string
	branches         []string
}

func (lp *timeStatsProcessor) processAll(sourceRoot string) error {
	for _, b := range lp.branches {
		if err := lp.processBranch(b, sourceRoot); err != nil {
			return err
		}
	}
	return nil
}

func (lp *timeStatsProcessor) processBranch(branch string, sourceRoot string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = sourceRoot
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "log", `--format=%H %cI`)
	cmd.Dir = sourceRoot
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	cs := &timestatsCache{lp.cacheDB}

	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		spl := strings.Split(s.Text(), " ")
		revision := spl[0]
		date, err := time.Parse(time.RFC3339, spl[1])
		if err != nil {
			panic(err)
		}

		c, err := cs.get(revision)
		if err != nil {
			return err
		}
		if c != nil {
			log.Printf("already have stats for revision %s\n", revision)
			if c.Err != "" {
				log.Printf("cached stats for revision %s is an error\n", revision)
			} else {
				lp.AllStats[branch] = append(lp.AllStats[branch], c.TimeStats)
			}
		} else {

			log.Printf("creating stats for revision %s\n", revision)

			cmd = exec.Command("git", "checkout", revision)
			cmd.Dir = sourceRoot
			err = cmd.Run()
			if err != nil {
				return err
			}

			sp := &statsProcessor{}
			err = sp.processAll(sourceRoot)
			var tse *timeStatsEntry
			if err != nil {
				log.Printf("revision %s was in error (will still cache): %s\n", revision, err.Error())
				cs.put(revision, &cacheEntry{Err: err.Error()})
			} else {
				tse = &timeStatsEntry{date, sp}
				lp.AllStats[branch] = append(lp.AllStats[branch], tse)
				cs.put(revision, &cacheEntry{TimeStats: tse})
			}
		}
		if revision == lp.earliestRevision {
			log.Printf("Exiting because earliest revision '%s' reached\n", revision)
			break
		}
	}

	cmd = exec.Command("git", "checkout", "master")
	cmd.Dir = sourceRoot
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type timestatsCache struct {
	filename string
}

func (tsc *timestatsCache) put(revision string, c *cacheEntry) error {
	db, err := bolt.Open(tsc.filename, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		value, err := json.Marshal(c)
		if err != nil {
			return err
		}

		b, err := tx.CreateBucketIfNotExists([]byte("timestats"))
		if err != nil {
			return err
		}

		return b.Put([]byte(revision), value)
	}); err != nil {
		return err
	}
	return nil
}

func (tsc timestatsCache) get(rev string) (*cacheEntry, error) {
	db, err := bolt.Open(tsc.filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var s *cacheEntry
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("timestats"))
		if b != nil {
			data := b.Get([]byte(rev))
			if data != nil {
				s = &cacheEntry{}
				err := json.Unmarshal(data, s)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s, nil
}
