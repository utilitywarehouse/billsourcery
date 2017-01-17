package main

import (
	"log"
	"strings"
)

func filenameToIdentifier(filename string) string {
	filename = strings.ToLower(filename)
	switch {
	case strings.HasSuffix(filename, ".ex@.txt"):
	case strings.HasSuffix(filename, ".fr@.txt"):
	case strings.HasSuffix(filename, ".im@.txt"):
	case strings.HasSuffix(filename, ".jc@.txt"):
	case strings.HasSuffix(filename, ".pp@.txt"):
	case strings.HasSuffix(filename, ".pr@.txt"):
	case strings.HasSuffix(filename, ".qr@.txt"):
	case strings.HasSuffix(filename, ".re@.txt"):
	default:
		log.Panicf("bug: %s\n", filename)
	}
	return filename[0 : len(filename)-8]
}
