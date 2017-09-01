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

func newTimeStatsImageProcessor(cacheDB string, earliest string, branches []string, outfile string) *timeStatsImageProcessor {
	return &timeStatsImageProcessor{AllStats: make(map[string][]*timeStatsEntry), cacheDB: cacheDB, earliestRevision: earliest, branches: branches, outfile: outfile}
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
	AllStats map[string][]*timeStatsEntry

	cacheDB          string
	earliestRevision string
	branches         []string
	outfile          string
}

func (lp *timeStatsImageProcessor) processAll(sourceRoot string) error {
	for _, b := range lp.branches {
		if err := lp.processBranch(b, sourceRoot); err != nil {
			return err
		}
	}
	return nil
}

func (lp *timeStatsImageProcessor) processBranch(branch string, sourceRoot string) error {
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

func (lp *timeStatsImageProcessor) end() error {

	// set up graph

	plotter.DefaultLineStyle.Width = vg.Points(1)
	plotter.DefaultGlyphStyle.Radius = vg.Points(1)

	p, err := plot.New()
	if err != nil {
		return err
	}
	p.Title.Text = "Bill code base"
	p.Legend.Top = true

	p.X.Label.Text = "Date"
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}
	p.Add(plotter.NewGrid())

	p.Y.Label.Text = "Code size\n(characters)"
	p.Y.Tick.Marker = &IntTicker{}

	for branch, ts := range lp.AllStats {

		// code data
		codeData := make(plotter.XYs, len(ts))
		for i := range codeData {
			res := ts[i].Results
			codeData[i].X = float64(ts[i].Time.Unix())
			codeData[i].Y = float64(res.OtherCount)
		}

		codeLine, codePoints, err := plotter.NewLinePoints(codeData)
		if err != nil {
			return err
		}
		codeLine.Color = color.RGBA{B: 255, A: 255}
		codePoints.Shape = draw.CircleGlyph{}
		codePoints.Color = color.RGBA{B: 255, A: 255}
		p.Add(codeLine, codePoints)
		p.Legend.Add(fmt.Sprintf("code (%s)", branch), codeLine)

		// comments data
		commentsData := make(plotter.XYs, len(ts))
		for i := range commentsData {
			res := ts[i].Results
			commentsData[i].X = float64(ts[i].Time.Unix())
			commentsData[i].Y = float64(res.CommentCount)
		}

		commentsLine, commentsPoints, err := plotter.NewLinePoints(commentsData)
		if err != nil {
			return err
		}
		commentsLine.Color = color.RGBA{R: 255, A: 255}
		commentsPoints.Shape = draw.CircleGlyph{}
		commentsPoints.Color = color.RGBA{B: 255, A: 255}
		p.Add(commentsLine, commentsPoints)
		p.Legend.Add(fmt.Sprintf("comments (%s)", branch), commentsLine)

		// totals data
		totalData := make(plotter.XYs, len(ts))
		for i := range totalData {
			res := ts[i].Results
			totalData[i].X = float64(ts[i].Time.Unix())
			totalData[i].Y = float64(res.CommentCount + res.OtherCount)
		}

		totalLine, totalPoints, err := plotter.NewLinePoints(totalData)
		if err != nil {
			return err
		}
		totalLine.Color = color.RGBA{G: 255, A: 255}
		totalPoints.Shape = draw.CircleGlyph{}
		totalPoints.Color = color.RGBA{B: 255, A: 255}
		p.Add(totalLine, totalPoints)
		p.Legend.Add(fmt.Sprintf("total (%s)", branch), totalLine)

		p.Y.Min = 0

	}

	// create output
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
