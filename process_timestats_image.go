package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
)

func newTimeStatsImageProcessor(cacheDB string, earliest string, outfile string) *timeStatsImageProcessor {
	return &timeStatsImageProcessor{cacheDB: cacheDB, earliestRevision: earliest, outfile: outfile}
}

type timeStatsEntry struct {
	Time    time.Time       `json:"timestamp"`
	Results *statsProcessor `json:"results"`
}

type cacheEntry struct {
	TimeStats *timeStatsEntry `json:"stats,omitempty"`
	Err       string          `json:"error,omitempty"`
}

type timeStatsImageProcessor struct {
	AllStats []*timeStatsEntry

	cacheDB          string
	earliestRevision string
	outfile          string
}

func (lp *timeStatsImageProcessor) processAll(sourceRoot string) error {
	cmd := exec.Command("git", "log", `--format=%H %cI`)
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
				lp.AllStats = append(lp.AllStats, c.TimeStats)
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
				lp.AllStats = append(lp.AllStats, tse)
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

func (lp *timeStatsImageProcessor) end() error {

	ts := lp.AllStats

	plotter.DefaultLineStyle.Width = vg.Points(1)
	plotter.DefaultGlyphStyle.Radius = vg.Points(3)

	data := make(plotter.XYs, len(ts))
	for i := range data {
		res := ts[i].Results
		data[i].X = float64(ts[i].Time.Unix())
		data[i].Y = float64(res.CommentCount + res.OtherCount)
	}

	p, err := plot.New()
	if err != nil {
		return err
	}
	p.Title.Text = "Bill code base"

	p.X.Label.Text = "Date"
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}
	p.Add(plotter.NewGrid())

	p.Y.Label.Text = "Code size\n(characters)"
	p.Y.Tick.Marker = &IntTicker{}

	line, points, err := plotter.NewLinePoints(data)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{G: 255, A: 255}
	points.Shape = draw.CircleGlyph{}
	points.Color = color.RGBA{B: 255, A: 255}

	p.Add(line, points)

	err = p.Save(30*vg.Centimeter, 15*vg.Centimeter, lp.outfile)
	if err != nil {
		return err
	}
	log.Printf("saved output to '%s'\n", lp.outfile)
	return nil
}

type IntTicker struct{}

func (it *IntTicker) Ticks(min float64, max float64) []plot.Tick {
	def := plot.DefaultTicks{}
	ticks := def.Ticks(min, max)
	for i, t := range ticks {
		t.Label = fmt.Sprintf("%d", uint64(t.Value))
		ticks[i] = t
	}
	return ticks
}
