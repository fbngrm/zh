package unihan

import "github.com/fgrimme/zh/pkg/reflection"

const (
	KDefinition  string = "kDefinition"
	KMandarin    string = "kMandarin"
	KCantonese   string = "kCantonese"
	KHanyuPinyin string = "kHanyuPinyin"
	KXHC1983     string = "kXHC1983"
	KHangul      string = "kHangul"
	KHanyuPinlu  string = "kHanyuPinlu"
	KVietnamese  string = "kVietnamese"
	KJapaneseOn  string = "kJapaneseOn"
	KJapaneseKun string = "kJapaneseKun"
	KTang        string = "kTang"
	KKorean      string = "kKorean"
)

type Hanzi struct {
	KDefinition  string
	KMandarin    string
	KCantonese   string
	KHanyuPinyin string
	KXHC1983     string
	KHangul      string
	KHanyuPinlu  string
	KVietnamese  string
	KJapaneseOn  string
	KJapaneseKun string
	KTang        string
	KKorean      string
}

func (h *Hanzi) SetFields(m map[string]string) error {
	for k, v := range m {
		err := reflection.SetField(h, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
