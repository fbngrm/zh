package cedict_test

import (
	"testing"

	"github.com/fbngrm/zh/lib/cedict"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const src = "./testdata/cedict.txt"

func TestNewCedict(t *testing.T) {
	expected :=
		`- ideograph: 㐄
  kangxi: false
  source: cedict
  simplified: 㐄
  traditional: 㐄
  definitions:
    - component in Chinese characters, mirror image of 夂
  readings:
    - kua4
- ideograph: 㐌
  kangxi: false
  source: cedict
  simplified: 㐌
  traditional: 㐌
  definitions:
    - variant of 它
  readings:
    - ta1
`
	d, err := cedict.NewDict(src)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	b, err := yaml.Marshal(d)
	if err != nil {
		t.Logf("could not format dict: %v\n", err)
		t.FailNow()
	}
	assert.Equal(t, expected, string(b))
}
