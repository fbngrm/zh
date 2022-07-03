## zh
`zh` is a command-line CEDICT dictionary, unihan search tool and word/hanzi decomposer.

### Features

- [x] translate words, hanzi or kangxi (chinese to english)
- [x] decompose words or hanzi into their components (hanzi, kangxi)
- [x] query unihan database by hanzi or unicode codepoint
- [x] lookup pinyin (reading/pronounciation) for words or hanzi
- [x] various output formats like yaml, json, go-templates
- [x] support scored search results for disambiguation
- [ ] HSK browser
- [ ] lookup hanzi containing radicals
- [ ] lookup words containing hanzi
- [ ] unihan browser
- [ ] translate english to chinese

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

### Todo

- [x] Omit empty yaml
- [x] Filter fields name mapping
- [x] Depth flag for decomposition
- [x] Force unihan search for hanzi
- [x] Make ids recursive
- [x] Do not decompose unihan hanzi by default
- [x] Decompose words
- [ ] Support En to Zh translation
- [ ] Batch decompose hanzi (from file)
- [x] HSK lists
- [ ] Batch decompose files
- [ ] HSK example sentences
- [ ] Filter * for slices
- [ ] Batch decompose sentences / split into words

