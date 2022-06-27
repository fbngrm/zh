package cedict_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/fgrimme/zh/internal/cedict"
	"github.com/fgrimme/zh/internal/cjkvi"
	"github.com/fgrimme/zh/internal/finder"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	cedictSrc = "../../lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
	idsSrc    = "../../lib/cjkvi/ids.txt"
)

var decomposer *cedict.Decomposer

func TestMain(m *testing.M) {
	dict, err := cedict.NewDict(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}

	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}

	decomposer = cedict.NewDecomposer(dict, finder.NewFinder(dict), idsDecomposer)
	os.Exit(m.Run())
}

func TestWordDecomposition(t *testing.T) {
	testcases := []struct {
		name    string
		query   string
		results int
		depth   int
		hanzi   string
		err     error
		errs    []error
	}{
		{
			name:    "expect successful word decomposition with depth 1",
			query:   "漂亮",
			results: 1,
			depth:   1,
			hanzi: `ideograph: 漂亮
source: cedict
simplified: 漂亮
traditional: 漂亮
definitions:
    - pretty
    - beautiful
readings:
    - piao4
    - liang5
decompositions:
    - ideograph: 漂
      mapping: U+6F02
      simplified: 漂
      traditional: 漂
      other_definitions:
        - to float, to drift
      other_readings:
        - piao1
      ids: ⿰氵票
    - ideograph: 亮
      mapping: U+4EAE
      simplified: 亮
      traditional: 亮
      other_definitions:
        - bright, clear, resonant, to shine, to show, to reveal
      other_readings:
        - liang4
      ids: ⿱⿳亠口冖几[G]
`,
			errs: []error{
				errors.New("no reading match found for hanzi decomposition 漂"),
				errors.New("no definition match found for hanzi decomposition 漂"),
				errors.New("no reading match found for hanzi decomposition 亮"),
				errors.New("no definition match found for hanzi decomposition 亮"),
			},
		},
	}

	for _, tc := range testcases {
		h, errs, err := decomposer.BuildWordDecomposition(tc.query, tc.results, tc.depth)
		if err != tc.err {
			t.Logf("expected error %v but got %v\n", tc.err, err)
			t.FailNow()
		}
		if len(errs) != len(tc.errs) {
			t.Logf("expected errors %d but got %d\n", len(tc.errs), len(errs))
			for _, e := range errs {
				t.Log(e)
			}
			t.FailNow()
		}

		b, err := yaml.Marshal(h)
		if err != nil {
			t.Logf("could not format hanzi decomposition for query %s: %v\n", tc.query, err)
			t.FailNow()
		}
		assert.Equal(t, tc.hanzi, string(b))
	}
}
