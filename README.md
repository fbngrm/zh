## zh
`zh` is a command-line dictionary, unihan search tool and word/hanzi/kangxi decomposer.

### Installation

Clone the repo and install via Go tool:
```
git clone git@github.com:fbngrm/zh.git
cd zh
go install cmd/zh/zh.go
whereis zh
```

### Usage

#### Translate English to Chinese

Result set limited to the single most frequently used translation (based on BCC):
```
zh -q teacher -r 1

老师	lao3, shi1	teacher
```
Include first level of decompositions:
```
zh -q teacher -r 1 -v

老师	lao3, shi1	teacher
老	lao3		old
师	shi1		teacher, master, expert, model, army division, (old) troops, to dispatch troops
```
Get all results, ordered by frequency of usage (based on BCC):
```
zh -q teacher

老师	lao3, shi1	teacher
先生	xian1, sheng5	teacher, husband, doctor (dialect)
教师	jiao4, shi1	teacher
师	shi1		teacher, master, expert, model, army division, (old) troops, to dispatch troops
教员	jiao4, yuan2	teacher, instructor
业师	ye4, shi1	teacher, one's teacher
师尊	shi1, zun1	teacher, master
```

#### Translate Chinese to English
Simple translation:
```
zh -q 苹果

苹果	ping2, guo3	apple
```
Include first level of decompositions:
```
zh -q 苹果 -v

苹果	ping2, guo3	apple
苹	ping2		(artemisia), duckweed, apple
果	guo3		fruit, result, resolute, indeed, if really, variant of 果
```

#### Decomposition
Decomposition is based on Ideographic Description Sequences from CHISE database.
The decomposition data will be printed in YAML format when supplying the flag `-fmt yaml`.
Hanzi are recursively decomposed down to Kangxi or component level.
For just gettting the first level of decomposition the use the `-v` flag.
```
zh -q engineer -fmt yaml

- ideograph: 工程师
  ...
  components:
    - 工
    - 程
    - 师
  components_decompositions:
    - ideograph: 工
      ...
    - ideograph: 程
      ...
      components:
        - 禾
        - 呈
      components_decompositions:
        - ideograph: 禾
          ...
        - ideograph: 呈
          ...
          components:
            - 口
            - 王
          components_decompositions:
            - ideograph: 口
              ...
            - ideograph: 王
              ...
    - ideograph: 师
      ...
      components:
        - 丨
        - 丿
        - 帀
      components_decompositions:
        ...
        - ideograph: 帀
          ...
          components:
            - 一
            - 巾
          components_decompositions:
            - ideograph: 一
            - ideograph: 巾
              ...
```

#### Support for tranditional script
By default, the Chinese translation in Mandarin uses simplified Hanzi (Chinese character/ideographs).
Traditional Hanzi equivalents can be obtained with the decomposition data.
Example:
```
zh -q 师 -fmt yaml -r 1

- ideograph: 师
  simplified:
    - 师
  traditional:
    - 師
  definitions:
    ...
```


### Features

- [x] Translate words, hanzi or kangxi (Chinese to English)
- [x] Translate English words to Chinese
- [x] Decompose words or Hanzi into their components (Hanzi, Kangxi, components)
- [x] Query unihan database by Hanzi or unicode codepoint
- [x] Lookup pinyin (reading/pronounciation) for words or Hanzi
- [x] Various output formats like yaml, json, go-templates
- [x] Support scored search results for disambiguation
- [x] Batch translate/decompose from file
- [x] Hsk lookup
- [x] Support Kangxi equivalents
- [x] Example sentences for words
- [x] Support frequency in BCC for translations from English to Chinese
- [ ] Part of speech
- [ ] Example Hanzi for Kangxi
- [ ] Sentence decomposition
- [ ] Stroke index/count
- [ ] Batch decompose sentences / split into words
- [ ] Example words for Hanzi
  - [ ] by frequency
  - [ ] by HSK

### Data Sources

#### CEDICT
- Community maintained free Chinese-English dictionary.
- Published by MDBG
- For documentation, see https://cc-cedict.org/wiki/

#### IDS Decompositions
- CJKVI Database
- Based on CHISE IDS Database
- For documentation, see https://github.com/cjkvi/cjkvi-ids

#### Unihan
- Unicode Character Database
- Unicode version: 13.0.0
- For documentation, see http://www.unicode.org/reports/tr38/

#### Word Frequencies
- BCC
- See https://www.github.com/fbngrm/zh/lib/word_frequencies/global_wordfreq.release_UTF-8.txt
