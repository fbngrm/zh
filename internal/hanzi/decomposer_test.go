package hanzi_test

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

func TestWordWordDecomposition(t *testing.T) {
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
		{
			name:    "expect successful word decomposition with depth 2",
			query:   "漂亮",
			results: 1,
			depth:   2,
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
      decompositions:
        - ideograph: 氵
          mapping: U+6C35
          simplified: 氵
          traditional: 氵
          definitions:
            - '"water" radical in Chinese characters (Kangxi radical 85), occurring in 沒|没'
          readings:
            - shui3
          ids: 氵
        - ideograph: 票
          mapping: U+7968
          simplified: 票
          traditional: 票
          definitions:
            - ticket, ballot, banknote, CL:張|张
          readings:
            - piao4
          ids: ⿱覀示
    - ideograph: 亮
      mapping: U+4EAE
      simplified: 亮
      traditional: 亮
      other_definitions:
        - bright, clear, resonant, to shine, to show, to reveal
      other_readings:
        - liang4
      ids: ⿱⿳亠口冖几[G]
      decompositions:
        - ideograph: 亠
          mapping: U+4EA0
          simplified: 亠
          traditional: 亠
          definitions:
            - '"lid" radical in Chinese characters (Kangxi radical 8)'
          readings:
            - tou2
          ids: ⿱丶一[GTK]
        - ideograph: 口
          mapping: U+53E3
          simplified: 口
          traditional: 口
          definitions:
            - mouth, classifier for things with mouths (people, domestic animals, cannons, wells etc), classifier for bites or mouthfuls
          readings:
            - kou3
          ids: 口
        - ideograph: 冖
          mapping: U+5196
          simplified: 冖
          traditional: 冖
          definitions:
            - '"cover" radical in Chinese characters (Kangxi radical 14), occurring in 軍|军'
          readings:
            - mi4
          ids: 冖
        - ideograph: 几
          mapping: U+51E0
          simplified: 几
          traditional: 几
          definitions:
            - small table
          readings:
            - ji1
          ids: 几
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
		h, errs, err := decomposer.Decompose(tc.query, tc.results, tc.depth)
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

func TestDecomposition(t *testing.T) {
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
			name:    "expect successful decomposition with depth 1",
			query:   "漂",
			results: 1,
			depth:   1,
			hanzi: `ideograph: 漂
mapping: U+6F02
simplified: 漂
traditional: 漂
definitions:
    - to float, to drift
readings:
    - piao1
ids: ⿰氵票
decompositions:
    - ideograph: 氵
      mapping: U+6C35
      simplified: 氵
      traditional: 氵
      definitions:
        - '"water" radical in Chinese characters (Kangxi radical 85), occurring in 沒|没'
      readings:
        - shui3
      ids: 氵
    - ideograph: 票
      mapping: U+7968
      simplified: 票
      traditional: 票
      definitions:
        - ticket, ballot, banknote, CL:張|张
      readings:
        - piao4
      ids: ⿱覀示
`,
			errs: []error{},
		},
		{
			name:    "expect successful decomposition with depth 1",
			query:   "漂",
			results: 1,
			depth:   2,
			hanzi: `ideograph: 漂
mapping: U+6F02
simplified: 漂
traditional: 漂
definitions:
    - to float, to drift
readings:
    - piao1
ids: ⿰氵票
decompositions:
    - ideograph: 氵
      mapping: U+6C35
      simplified: 氵
      traditional: 氵
      definitions:
        - '"water" radical in Chinese characters (Kangxi radical 85), occurring in 沒|没'
      readings:
        - shui3
      ids: 氵
    - ideograph: 票
      mapping: U+7968
      simplified: 票
      traditional: 票
      definitions:
        - ticket, ballot, banknote, CL:張|张
      readings:
        - piao4
      ids: ⿱覀示
      decompositions:
        - ideograph: 覀
          mapping: U+8980
          ids: ⿱一⿻口⿰丨丨
        - ideograph: 示
          mapping: U+793A
          simplified: 示
          traditional: 示
          definitions:
            - to show, to reveal
          readings:
            - shi4
          ids: 示
`,
			errs: []error{},
		},
		{
			name:    "expect successful decomposition with depth 1",
			query:   "漂",
			results: 1,
			depth:   3,
			hanzi: `ideograph: 漂
mapping: U+6F02
simplified: 漂
traditional: 漂
definitions:
    - to float, to drift
readings:
    - piao1
ids: ⿰氵票
decompositions:
    - ideograph: 氵
      mapping: U+6C35
      simplified: 氵
      traditional: 氵
      definitions:
        - '"water" radical in Chinese characters (Kangxi radical 85), occurring in 沒|没'
      readings:
        - shui3
      ids: 氵
    - ideograph: 票
      mapping: U+7968
      simplified: 票
      traditional: 票
      definitions:
        - ticket, ballot, banknote, CL:張|张
      readings:
        - piao4
      ids: ⿱覀示
      decompositions:
        - ideograph: 覀
          mapping: U+8980
          ids: ⿱一⿻口⿰丨丨
          decompositions:
            - ideograph: 一
              mapping: U+4E00
              simplified: 一
              traditional: 一
              definitions:
                - one, 1, single, a (article), as soon as, entire, whole, all, throughout, "one" radical in Chinese characters (Kangxi radical 1), also pr.
              readings:
                - yi1
              ids: 一
            - ideograph: 口
              mapping: U+53E3
              simplified: 口
              traditional: 口
              definitions:
                - mouth, classifier for things with mouths (people, domestic animals, cannons, wells etc), classifier for bites or mouthfuls
              readings:
                - kou3
              ids: 口
            - ideograph: 丨
              mapping: U+4E28
              simplified: 丨
              traditional: 丨
              definitions:
                - vertical stroke (in Chinese characters), referred to as 豎筆|竖笔
              readings:
                - shu4
              ids: 丨
            - ideograph: 丨
              mapping: U+4E28
              simplified: 丨
              traditional: 丨
              definitions:
                - vertical stroke (in Chinese characters), referred to as 豎筆|竖笔
              readings:
                - shu4
              ids: 丨
        - ideograph: 示
          mapping: U+793A
          simplified: 示
          traditional: 示
          definitions:
            - to show, to reveal
          readings:
            - shi4
          ids: 示
`,
			errs: []error{},
		},
	}

	for _, tc := range testcases {
		h, errs, err := decomposer.Decompose(tc.query, tc.results, tc.depth)
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
