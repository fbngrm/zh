package zh

import (
	"errors"
	"fmt"
	"io/ioutil"
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
var result *Decomposition
var matches fuzzy.Matches

// type ButtonWidget struct {
// 	name    string
// 	x, y    int
// 	w       int
// 	label   string
// 	handler func(g *gocui.Gui, v *gocui.View) error
// }

// func NewButtonWidget(name string, x, y int, label string, handler func(g *gocui.Gui, v *gocui.View) error) *ButtonWidget {
// 	return &ButtonWidget{
// 		name:    name,
// 		x:       x,
// 		y:       y,
// 		w:       len(label) + 1,
// 		label:   label,
// 		handler: handler,
// 	}
// }

// func (w *ButtonWidget) Layout(g *gocui.Gui) error {
// 	v, err := g.SetView(w.name, w.x, w.y, w.x+w.w, w.y+2)
// 	if err != nil {
// 		if err != gocui.ErrUnknownView {
// 			return err
// 		}
// 		if _, err := g.SetCurrentView(w.name); err != nil {
// 			return err
// 		}
// 		if err := g.SetKeybinding(w.name, gocui.KeyEnter, gocui.ModNone, w.handler); err != nil {
// 			return err
// 		}
// 		fmt.Fprint(v, w.label)
// 	}
// 	return nil
// }

func InteractiveSearch(finder *Finder) (*Decomposition, error) {
	f = finder

	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = true

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return nil, err
	}

	if err := g.SetKeybinding("searchView", gocui.KeyArrowRight, gocui.ModNone, switchToDetailView); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("searchView", gocui.KeyArrowDown, gocui.ModNone, switchToResultsView); err != nil {
		return nil, err
	}

	if err := g.SetKeybinding("detailView", gocui.KeyArrowLeft, gocui.ModNone, switchToSearchView); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("detailView", gocui.KeyEnter, gocui.ModNone, export); err != nil {
		return nil, err
	}

	if err := g.SetKeybinding("resultsView", gocui.KeyArrowRight, gocui.ModNone, switchToDetailView); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("resultsView", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("resultsView", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return nil, err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return nil, err
	}
	return result, nil
}

func export(g *gocui.Gui, v *gocui.View) error {
	details, err := f.FormatDetails(matches[matchIndex].Index)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./output.json", []byte(details), os.ModePerm)
}

func printResults() error {
	if len(matches) == 0 {
		return nil
	}
	resultsView, err := g.View("resultsView")
	if err != nil {
		return err
	}
	resultsView.Clear()
	resultsView.Clear()
	for i := 0; i < len(matches); i++ {
		fmt.Fprintln(resultsView, f.FormatResult(matches[i].Index))
	}
	return printDetails(0)
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
	result = f.dict[matches[i].Index]
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
	// butdown := NewButtonWidget("butdown", 52, 7, "DOWN", nil)
	// butup := NewButtonWidget("butup", 58, 7, "UP", nil)
	// g.SetManager(butdown, butup)

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
			return printResults()
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
