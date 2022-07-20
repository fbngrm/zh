package search

import "github.com/fgrimme/zh/internal/finder"

type Finder interface {
	SetSearchMode(query string)
	Find(query string, limit int) finder.Matches
	FindSorted(query string, limit int) finder.Matches
}
