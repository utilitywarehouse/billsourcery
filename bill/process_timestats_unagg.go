package bill

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/utilitywarehouse/equilex"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type rawTimeStatsSummary struct {
	Time    time.Time            `json:"timestamp"`
	Entries []*rawTimeStatsEntry `json:"entries"`
}

type rawTimeStatsEntry struct {
	FileName     string `json:"filename"`
	FileType     string `json:"filetype"`
	CommentCount int    `json:"comment-chars"`
	OtherCount   int    `json:"other-chars"`
}

type rawCacheEntry struct {
	TimeStats *rawTimeStatsSummary `json:"stats,omitempty"`
	Err       string               `json:"error,omitempty"`
}

type rawTimeStatsProcessor struct {
	AllStats map[string][]*rawTimeStatsSummary

	cacheDB          string
	earliestRevision string
	branches         []string
}

func (lp *rawTimeStatsProcessor) processRepository(sourceRoot string) error {
	for _, b := range lp.branches {
		if err := lp.processBranch(b, sourceRoot); err != nil {
			return err
		}
	}
	return nil
}

func (lp *rawTimeStatsProcessor) processBranch(branch string, sourceRoot string) error {
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

	cs := &rawstatsCache{lp.cacheDB}

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
				// log.Printf("appending for branch %s and count is %d\n", branch, len(c.TimeStats.Entries))
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

			blp := &rawTimeStatsBranchProcessor{branch: branch}
			err = walkSource(sourceRoot, blp)
			var tse *rawTimeStatsSummary
			if err != nil {
				log.Printf("revision %s was in error (will still cache): %s\n", revision, err.Error())
				if err := cs.put(revision, &rawCacheEntry{Err: err.Error()}); err != nil {
					return err
				}
			} else {
				tse = &rawTimeStatsSummary{date, blp.stats}
				lp.AllStats[branch] = append(lp.AllStats[branch], tse)
				if err := cs.put(revision, &rawCacheEntry{TimeStats: tse}); err != nil {
					return err
				}
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

	//	lp.AllStats[branch] = blp.stats

	return nil
}

type rawTimeStatsBranchProcessor struct {
	branch string
	stats  []*rawTimeStatsEntry
}

func (lp *rawTimeStatsBranchProcessor) process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var CommentCount int
	var OtherCount int

	l := equilex.NewLexer(transform.NewReader(f, charmap.Windows1252.NewDecoder()))

loop:
	for {
		tok, lit, err := l.Scan()
		if err != nil {
			return err
		}

		switch tok {
		case equilex.EOF:
			break loop
		case equilex.Comment:
			CommentCount += len(lit)
		default:
			OtherCount += len(lit)
		}

		switch tok {
		case equilex.EOF:
			break loop
		}
	}

	mod := moduleFromFullFilename(path)
	lp.stats = append(lp.stats, &rawTimeStatsEntry{
		FileName:     filepath.Base(mod.moduleName),
		FileType:     mod.moduleType.String(),
		CommentCount: CommentCount,
		OtherCount:   OtherCount,
	})

	return nil
}

type rawstatsCache struct {
	filename string
}

func (tsc *rawstatsCache) put(revision string, c *rawCacheEntry) error {
	db, err := bolt.Open(tsc.filename, 0o600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		value, err := json.Marshal(c)
		if err != nil {
			return err
		}

		b, err := tx.CreateBucketIfNotExists([]byte("timestats-raw"))
		if err != nil {
			return err
		}

		return b.Put([]byte(revision), value)
	}); err != nil {
		return err
	}
	return nil
}

func (tsc rawstatsCache) get(rev string) (*rawCacheEntry, error) {
	db, err := bolt.Open(tsc.filename, 0o600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var s *rawCacheEntry
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("timestats-raw"))
		if b != nil {
			data := b.Get([]byte(rev))
			if data != nil {
				s = &rawCacheEntry{}
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
