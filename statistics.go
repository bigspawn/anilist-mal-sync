package main

import "log"

type Statistics struct {
	UpdatedCount int
	SkippedCount int
	TotalCount   int
}

func (s Statistics) Print(prefix string) {
	log.Printf("[%s] Updated %d out of %d\n", prefix, s.UpdatedCount, s.TotalCount)
	log.Printf("[%s] Skipped %d\n", prefix, s.SkippedCount)
}
