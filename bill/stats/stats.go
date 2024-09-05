package stats

func Stats(sourceRoot string) error {
	stats := &statsProcessor{}
	if err := walkSource(sourceRoot, stats); err != nil {
		return err
	}
	return stats.print()
}

func TimestatsImage(sourceRoot string, cacheDb string, earliest string, branches []string, output string) error {
	processor := newTimeStatsImageProcessor(cacheDb, earliest, branches, output)
	if err := processor.scanRepository(sourceRoot); err != nil {
		return err
	}
	if err := processor.outputGraph(); err != nil {
		return err
	}
	return nil
}

func TimestatsBqRaw(sourceRoot string, cacheDb string, earliest string,
	branches []string) error {
	processor := newTimeStatsUnaggBQProcessor(
		cacheDb,
		earliest,
		branches,
	)
	err := processor.processRepository(sourceRoot)
	if err != nil {
		return err
	}

	if err := processor.end(); err != nil {
		return err
	}
	return nil
}
