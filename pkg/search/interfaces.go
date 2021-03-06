package search

import "github.com/fgrimme/zh/pkg/finder"

type Finder interface {
	SetSearchMode(query string)
	Find(query string, limit int) (finder.Matches, error)
	FindSorted(query string, limit int) (finder.Matches, error)
}
