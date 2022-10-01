package sentences

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitWords(t *testing.T) {
	t.Run("with punctuation", func(t *testing.T) {
		var expected = parsedSentences{
			"你好！我叫赵阳。我是工程师。": {
				Source:  "testdata",
				Chinese: "你好！我叫赵阳。我是工程师。",
				ChineseWords: []string{
					"你",
					"好",
					"！",
					"我",
					"叫",
					"赵",
					"阳",
					"。",
					"我",
					"是",
					"工程师",
					"。",
				},
				Pinyin:         "ni3 hao3! wo3 jiao4 Zhao4 Yang2. Wo3 shi4 gong1cheng2shi1.",
				English:        "Hello! I am Zhao Yang. I am an engineer.",
				EnglishLiteral: "",
			},
		}

		sourceName := "testdata"
		sourcePath, err := filepath.Abs("./testdata/with_punctuation")
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		got, err := Parse(sourceName, sourcePath)
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		assert.Equal(t, expected, got)
	})

	t.Run("no number in pinyin word", func(t *testing.T) {
		var expected = parsedSentences{
			"我的名字叫王丽丽。": {
				Source:  "testdata",
				Chinese: "我的名字叫王丽丽。",
				ChineseWords: []string{
					"我",
					"的",
					"名字",
					"叫",
					"王",
					"丽",
					"丽",
					"。",
				},
				Pinyin:         "wo3 de ming2zi jiao4 Wang2 Li Li.",
				English:        "My name is Wang Li Li.",
				EnglishLiteral: "",
			},
		}

		sourceName := "testdata"
		sourcePath, err := filepath.Abs("./testdata/no_number_in_pinyin")
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		got, err := Parse(sourceName, sourcePath)
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		assert.Equal(t, expected, got)
	})

	t.Run("no number in pinyin word 2", func(t *testing.T) {
		var expected = parsedSentences{
			"你是老师吗？我不是老师，我是工程师。": {
				Source:  "testdata",
				Chinese: "你是老师吗？我不是老师，我是工程师。",
				ChineseWords: []string{
					"你",
					"是",
					"老师",
					"吗",
					"？",
					"我",
					"不",
					"是",
					"老师",
					"，",
					"我",
					"是",
					"工程师",
					"。",
				},
				Pinyin:         "ni3 shi4 lao3shi1 ma? wo3 bu2 shi4 lao3shi1, wo3 shi4 gong1cheng2shi1.",
				English:        "Are you a teacher? I am not a teacher, I am an Engineer.",
				EnglishLiteral: "",
			},
		}

		sourceName := "testdata"
		sourcePath, err := filepath.Abs("./testdata/no_number_in_pinyin_2")
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		got, err := Parse(sourceName, sourcePath)
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}
		assert.Equal(t, expected, got)
	})

	t.Run("use five tone notation", func(t *testing.T) {
		var expected = parsedSentences{
			"我的名字叫王丽丽。": {
				Source:  "testdata",
				Chinese: "我的名字叫王丽丽。",
				ChineseWords: []string{
					"我",
					"的",
					"名字",
					"叫",
					"王",
					"丽丽",
					"。",
				},
				Pinyin:         "wo3 de ming2zi jiao4 Wang2 Li5Li5.",
				English:        "My name is Wang Lili.",
				EnglishLiteral: "",
			},
		}

		sourceName := "testdata"
		sourcePath, err := filepath.Abs("./testdata/with_five_tone_pinyin")
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		got, err := Parse(sourceName, sourcePath)
		if err != nil {
			t.Logf("unexpected error: %v", err)
			t.Fail()
		}

		assert.Equal(t, expected, got)
	})
}
