package zh

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/sahilm/fuzzy"
)

var filenamesBytes []byte
var err error
var g *gocui.Gui
var f *Finder
var matchIndex int
var matches fuzzy.Matches

func InteractiveSearch(finder *Finder) {
	f = finder

	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = true

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("searchView", gocui.KeyArrowRight, gocui.ModNone, switchToDetailView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("searchView", gocui.KeyArrowDown, gocui.ModNone, switchToResultsView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("detailView", gocui.KeyArrowLeft, gocui.ModNone, switchToSearchView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("detailView", gocui.KeyEnter, gocui.ModNone, export); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("resultsView", gocui.KeyArrowRight, gocui.ModNone, switchToDetailView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("resultsView", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("resultsView", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func export(g *gocui.Gui, v *gocui.View) error {
	details, err := f.FormatDetails(matches[matchIndex].Index)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./output.json", []byte(details), os.ModePerm)
}

func printResults(i int, resultsView *gocui.View) {
	if len(matches) == 0 {
		return
	}
	resultsView.Clear()
	for ; i < len(matches); i++ {
		fmt.Fprintln(resultsView, f.FormatResult(matches[i].Index))
	}
}

func printDetails(i int) error {
	if len(matches) == 0 {
		return nil
	}
	detailView, err := g.View("detailView")
	if err != nil {
		return err
	}
	detailView.Clear()

	details, err := f.FormatDetails(matches[i].Index)
	if err != nil {
		return err
	}
	fmt.Fprintln(detailView, details)
	matchIndex = i

	return nil
}

var scrollOffset int

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return errors.New("view is nil")
	}

	detailView, err := g.View("detailView")
	if err != nil {
		return err
	}
	detailView.Clear()

	var i int
	cx, cy := v.Cursor()
	ox, oy := v.Origin()

	if err := v.SetCursor(cx, cy+1); err == nil {
		i = cy + 1
		if oy > 0 {
			i = i + oy
		}
		return printDetails(i)
	}

	i = cy
	err = v.SetOrigin(ox, oy+1)
	if err == nil {
		i = i + oy + 1
		return printDetails(i)
	}
	return err
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return errors.New("view is nil")
	}

	detailView, err := g.View("detailView")
	if err != nil {
		return err
	}
	detailView.Clear()

	var i int
	ox, oy := v.Origin()
	cx, cy := v.Cursor()

	if err := v.SetCursor(cx, cy-1); err == nil {
		i = cy - 1
		if oy > 0 {
			i = i + oy
		}
		return printDetails(i)
	}

	if oy == 0 {
		return switchToSearchView(g, v)
	}

	i = cy
	err = v.SetOrigin(ox, oy-1)
	if err == nil {
		i = i + oy - 1
		return printDetails(i)

	}
	return err
}

func switchToSearchView(g *gocui.Gui, view *gocui.View) error {
	if _, err := g.SetCurrentView("searchView"); err != nil {
		return err
	}
	return nil
}

func switchToDetailView(g *gocui.Gui, view *gocui.View) error {
	if _, err := g.SetCurrentView("detailView"); err != nil {
		return err
	}
	return nil
}

func switchToResultsView(g *gocui.Gui, view *gocui.View) error {
	if len(matches) == 0 {
		return nil
	}
	printDetails(0)

	if _, err := g.SetCurrentView("resultsView"); err != nil {
		return err
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("searchView", 0, 0, maxX/2-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Editable = true
		v.Frame = true
		v.Title = "Search"
		if _, err := g.SetCurrentView("searchView"); err != nil {
			return err
		}
		v.Editor = gocui.EditorFunc(finder)
	}

	if v, err := g.SetView("resultsView", 0, 3, maxX/2-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.Title = "Results"
	}

	if v, err := g.SetView("detailView", maxX/2, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.Title = "Data"
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func finder(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
		g.Update(func(gui *gocui.Gui) error {
			resultsView, err := g.View("resultsView")
			if err != nil {
				// handle error
			}
			resultsView.Clear()
			detailView, err := g.View("detailView")
			if err != nil {
				// handle error
			}
			detailView.Clear()

			// we only downgrade mode if this is a new search
			downgradeMode := len(v.ViewBuffer()) == 0
			f.SetMode(ch, downgradeMode)

			matches = fuzzy.FindFrom(strings.TrimSpace(v.ViewBuffer()), f)
			// elapsed := time.Since(time.Now())
			// fmt.Fprintf(resultsView, "found %v matches in %v\n\n", len(matches), elapsed)

			for _, match := range matches {
				// for i := 0; i < len(match.Str); i++ {
				// 	if contains(i, match.MatchedIndexes) {
				// 		fmt.Fprintf(resultsView, f.dict[i].Definition)
				// 	} else {
				// 		fmt.Fprintf(resultsView, f.dict[i].Definition)
				// 	}
				// }
				// fmt.Fprintln(resultsView, "")
				fmt.Fprintln(resultsView, f.FormatResult(match.Index))
			}
			return nil
		})
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
		g.Update(func(gui *gocui.Gui) error {
			resultsView, err := g.View("resultsView")
			if err != nil {
				// handle error
			}
			resultsView.Clear()
			detailView, err := g.View("detailView")
			if err != nil {
				// handle error
			}
			detailView.Clear()
			t := time.Now()
			matches := fuzzy.FindFrom(strings.TrimSpace(v.ViewBuffer()), f)
			elapsed := time.Since(t)
			fmt.Fprintf(resultsView, "found %v matches in %v\n", len(matches), elapsed)
			for _, match := range matches {
				for i := 0; i < len(match.Str); i++ {
					if contains(i, match.MatchedIndexes) {
						fmt.Fprintf(resultsView, f.dict[i].Definition)
					} else {
						fmt.Fprintf(resultsView, f.dict[i].Definition)
					}
				}
				fmt.Fprintln(resultsView, "")
			}
			return nil
		})
	case key == gocui.KeyDelete:
		v.EditDelete(false)
		g.Update(func(gui *gocui.Gui) error {
			resultsView, err := g.View("resultsView")
			if err != nil {
				// handle error
			}
			resultsView.Clear()
			detailView, err := g.View("detailView")
			if err != nil {
				// handle error
			}
			detailView.Clear()
			t := time.Now()
			matches := fuzzy.FindFrom(strings.TrimSpace(v.ViewBuffer()), f)
			elapsed := time.Since(t)
			fmt.Fprintf(resultsView, "found %v matches in %v\n", len(matches), elapsed)
			for _, match := range matches {
				for i := 0; i < len(match.Str); i++ {
					if contains(i, match.MatchedIndexes) {
						fmt.Fprintf(resultsView, f.dict[i].Definition)
					} else {
						fmt.Fprintf(resultsView, f.dict[i].Definition)
					}
				}
				fmt.Fprintln(resultsView, "")
			}
			return nil
		})
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	}
}

func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if needle == i {
			return true
		}
	}
	return false
}
