//go:build !windows
// +build !windows

package main

import (
	"os"
)

func defaultCacheName() string {
	return os.Getenv("HOME") + "/.billsourcery_timestats_cache"
}

func defaultOutputName() string {
	return "/tmp/billtimestats.png"
}
