package cjkvi_test

import (
	"testing"

	"github.com/fbngrm/zh/internal/cjkvi"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const idsSrc = "../../lib/cjkvi/ids.txt"

func Test(t *testing.T) {
	d, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		t.Logf("could not initialize ids decompose: %v\n", err)
		t.FailNow()
	}

	testcases := []struct {
		name      string
		ideograph string
		depth     int
		hanzi     string
	}{
		{
			name:      "sucess hanzi, depth 1",
			ideograph: "好",
			depth:     1,
			hanzi: `mapping: U+597D
ideograph: 好
ids: ⿰女子
decomposition:
- mapping: U+5973
  ideograph: 女
  ids: 女
- mapping: U+5B50
  ideograph: 子
  ids: 子
`,
		},
		{
			name:      "sucess hanzi, depth 2",
			ideograph: "好",
			depth:     2,
			hanzi: `mapping: U+597D
ideograph: 好
ids: ⿰女子
decomposition:
- mapping: U+5973
  ideograph: 女
  ids: 女
- mapping: U+5B50
  ideograph: 子
  ids: 子
`,
		},
		{
			name:      "sucess hanzi 2, depth 1",
			ideograph: "子",
			depth:     1,
			hanzi: `mapping: U+5B50
ideograph: 子
ids: 子
`,
		},
		{
			name:      "sucess hanzi 3, depth 1",
			ideograph: "亮",
			depth:     1,
			hanzi: `mapping: U+4EAE
ideograph: 亮
ids: ⿱⿳亠口冖几[G]
decomposition:
- mapping: U+4EA0
  ideograph: 亠
  ids: ⿱丶一[GTK]
- mapping: U+53E3
  ideograph: 口
  ids: 口
- mapping: U+5196
  ideograph: 冖
  ids: 冖
- mapping: U+51E0
  ideograph: 几
  ids: 几
`,
		},
		{
			name:      "sucess hanzi 3, depth 2",
			ideograph: "亮",
			depth:     2,
			hanzi: `mapping: U+4EAE
ideograph: 亮
ids: ⿱⿳亠口冖几[G]
decomposition:
- mapping: U+4EA0
  ideograph: 亠
  ids: ⿱丶一[GTK]
  decomposition:
  - mapping: U+4E36
    ideograph: 丶
    ids: 丶
  - mapping: U+4E00
    ideograph: 一
    ids: 一
- mapping: U+53E3
  ideograph: 口
  ids: 口
- mapping: U+5196
  ideograph: 冖
  ids: 冖
- mapping: U+51E0
  ideograph: 几
  ids: 几
`,
		},
	}

	for _, tc := range testcases {
		h := d.Decompose(tc.ideograph, tc.depth)
		b, err := yaml.Marshal(h)
		if err != nil {
			t.Logf("could not format decomposition for query %s: %v\n", tc.ideograph, err)
			t.FailNow()
		}
		assert.Equal(t, tc.hanzi, string(b))
	}
}
