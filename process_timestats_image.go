package main

import (
	"fmt"
	"image/color"
	"log"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func newTimeStatsImageProcessor(cacheDB string, earliest string, branches []string, outfile string) *timeStatsImageProcessor {
	return &timeStatsImageProcessor{timeStatsProcessor{AllStats: make(map[string][]*timeStatsEntry), cacheDB: cacheDB, earliestRevision: earliest, branches: branches, outfile: outfile}}
}

type timeStatsImageProcessor struct {
	timeStatsProcessor
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
		totalLine.Width = 0.1 * vg.Millimeter
		totalPoints.Shape = draw.CircleGlyph{}
		totalPoints.Color = color.RGBA{B: 255, A: 255}
		totalPoints.Radius = 0.1 * vg.Millimeter
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
