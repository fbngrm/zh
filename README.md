## zh
`zh` is a command-line CEDICT dictionary, unihan search tool and word/hanzi decomposer.

### Features

- [x] translate words, hanzi or kangxi (chinese to english)
- [x] decompose words or hanzi into their components (hanzi, kangxi)
- [x] query unihan database by hanzi or unicode codepoint
- [x] lookup pinyin (reading/pronounciation) for words or hanzi
- [x] various output formats like yaml, json, go-templates
- [x] support scored search results for disambiguation
- [x] from file
- [ ] example words for hanzi by frequency and HSK
- [ ] example sentences for words
- [ ] example hanzi for kangxi
- [ ] sentence decomposition
- [ ] word frequency in BCC
- [x] translate english to chinese
- [ ] stroke index/count
- [ ] batch decompose sentences / split into words

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
