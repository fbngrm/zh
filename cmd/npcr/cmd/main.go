package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/translate"
	"github.com/fgrimme/zh/internal/cedict"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

const idsSrc = "../../lib/cjkvi/ids.txt"
const cedictSrc = "../../lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const wordFrequencySrc = "../../lib/word_frequencies/global_wordfreq.release_UTF-8.txt"
const deckName = "new-practical-chinese-reader-01"

var tags = []string{"hsk1", deckName}

var lesson string

type Ignored map[string]struct{}

func (i Ignored) update(s string) {
	i[s] = struct{}{}
}

type Char struct {
	Chinese      string `yaml:"chinese"`
	English      string `yaml:"english"`
	Audio        string `yaml:"audio"`
	IsSingleRune bool   `yaml:"isSingleRune"`
}

type Word struct {
	Chinese      string `yaml:"chinese"`
	Pinyin       string `yaml:"pinyin"`
	English      string `yaml:"english"`
	Audio        string `yaml:"audio"`
	NewChars     []Char `yaml:"newChars"`
	AllChars     []Char `yaml:"allChars"`
	IsSingleRune bool   `yaml:"isSingleRune"`
}

type Sentence struct {
	Chinese      string `yaml:"chinese"`
	Pinyin       string `yaml:"pinyin"`
	English      string `yaml:"english"`
	Audio        string `yaml:"audio"`
	NewWords     []Word `yaml:"newWords"`
	AllWords     []Word `yaml:"allWords"`
	IsSingleRune bool   `yaml:"isSingleRune"`
}

type Dialog struct {
	Deck      string     `yaml:"deck"`
	Tags      []string   `yaml:"tags"`
	Chinese   string     `yaml:"chinese"`
	Pinyin    string     `yaml:"pinyin"`
	English   string     `yaml:"english"`
	Audio     string     `yaml:"audio"`
	Sentences []Sentence `yaml:"sentences"`
}

func main() {
	flag.StringVar(&lesson, "l", "", "lesson name")
	flag.Parse()

	tags = append(tags, lesson)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ignorePath := filepath.Join(cwd, "data", "ignore")
	ignored := loadIgnored(ignorePath)

	translationsPath := filepath.Join(cwd, "data", "translations")
	translations := loadTranslations(translationsPath)

	audioDirPath := filepath.Join(cwd, "data", lesson, "audio")

	// load dialogues from file
	dialogLines := loadDialogs(filepath.Join(cwd, "data", lesson, "input", "dialogues"))
	dialogues := getDialogues(dialogLines, ignored)
	for i, dialog := range dialogues {
		addPinyin(dialog)
		translateDialogNew(translations, dialog)
		translateDialog(translations, dialog)
		getAudio(audioDirPath, dialog)
		dialogPath := filepath.Join(cwd, "data", lesson, "output", fmt.Sprintf("dialog_%02d.yaml", i+1))
		writeToFile(dialog, dialogPath)
	}

	// load sentences from file
	sentenceLines := loadFile(filepath.Join(cwd, "data", lesson, "input", "sentences"))
	sentences := getSentences(ignored, sentenceLines)
	sentences = translateNewWordsInSentences(translations, sentences)
	sentences = translateAllWordsInSentences(translations, sentences)
	sentences = getSentenceAudio(context.Background(), audioDirPath, sentences)
	sentencesPath := filepath.Join(cwd, "data", lesson, "output", "sentences.yaml")
	writeToFile(sentences, sentencesPath)

	// load words from file
	var newWords []Word
	wordLines := loadFile(filepath.Join(cwd, "data", lesson, "input", "words"))
	for _, word := range wordLines {
		_, newWord := getWords(word, ignored)
		newWords = append(newWords, newWord...)
	}
	newWords = translateAllWords(translations, newWords)

	// write cards
	templatePath := filepath.Join(cwd, "tmpl")
	outPath := filepath.Join(cwd, "data", lesson, "output", "cards.md")
	_ = os.Remove(outPath)
	// add dialogs
	for _, dialog := range dialogues {
		writeDialogCards(dialog, deckName, templatePath, outPath)
	}
	// add from sentence list
	writeSentenceCards(sentences, deckName, templatePath, outPath)
	// add from vocab list
	writeWordsCards(newWords, deckName, templatePath, outPath)

	// write newly ignored words
	writeIgnored(ignored, ignorePath)
	writeTranslations(translations, translationsPath)
}

func getDialogues(sentences [][]string, ignored Ignored) []*Dialog {
	var dialogues []*Dialog
	for _, dialogSentences := range sentences {
		dialogues = append(dialogues, &Dialog{
			Chinese:   strings.Join(dialogSentences, " "),
			Sentences: getSentences(ignored, dialogSentences),
			Deck:      deckName,
		})
	}
	return dialogues
}

func addPinyin(d *Dialog) {
	dict, err := cedict.BySimplifiedHanzi(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}
	for i, sentence := range d.Sentences {
		for y, word := range sentence.AllWords {
			entries, _ := dict[word.Chinese]
			readings := make([]string, 0)
			for _, entry := range entries {
				readings = append(readings, entry.Readings...)
			}
			pinyin := "xxx"
			if len(readings) == 0 {
				fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
				fmt.Printf("no readings found for word \"%s\", please enter pinyin or skip to use %s\n", word.Chinese, pinyin)
				pinyin = getPinyinFromUser(sentence.Chinese, nil)
			}
			if len(readings) > 1 {
				fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
				fmt.Printf("more than 1 readings found for word \"%s\" please choose\n", word.Chinese)
				pinyin = getPinyinFromUser(sentence.Chinese, readings)
			}
			if len(readings) == 1 {
				pinyin = readings[0]
			}
			d.Sentences[i].AllWords[y].Pinyin = pinyin
		}
	}

	// dict, err := cedict.NewDict(cedictSrc)

	// idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	// if err != nil {
	// 	fmt.Printf("could not initialize ids decompose: %v\n", err)
	// 	os.Exit(1)
	// }

	// // we provide a word frequency index which needs to be initialized before first use.
	// frequencyIndex := frequency.NewWordIndex(wordFrequencySrc)

	// decomposer := hanzi.NewDecomposer(
	// 	dict,
	// 	kangxi.NewDict(),
	// 	search.NewSearcher(finder.NewFinder(dict)),
	// 	idsDecomposer,
	// 	nil,
	// 	frequencyIndex,
	// )

	// for _, sentence := range d.Sentences {
	// 	for _, word := range sentence.AllWords {
	// 		decomposition, err := decomposer.Decompose(word.Chinese, 3, 0)
	// 		if err != nil {
	// 			os.Stderr.WriteString(fmt.Sprintf("error decomposing %s: %v\n", word.Chinese, err))
	// 		}
	// 		if len(decomposition.Errs) != 0 {
	// 			for _, e := range decomposition.Errs {
	// 				os.Stderr.WriteString(fmt.Sprintf("errors decomposing %s: %v\n", word.Chinese, e))
	// 			}
	// 		}
	// 		if len(decomposition.Hanzi) != 1 {
	// 			os.Stderr.WriteString(fmt.Sprintf("expect exactly 1 decomposition for word: %s", word.Chinese))
	// 			os.Exit(1)
	// 		}
	// 		spew.Dump(decomposition.Hanzi[0].Readings)
	// 		// spew.Dump(decomposition)
	// 	}
	// }
}

func getPinyinFromUser(sentence string, options []string) string {
	if len(options) > 1 {
		fmt.Printf("s = skip / e = enter pinyin / o = choose option / c = cancel \n")
	} else {
		fmt.Printf("s = skip / e = enter pinyin / c = cancel \n")
	}
	fmt.Printf("sentence: %s\n", sentence)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "o":
			for i, o := range options {
				fmt.Printf("option %d: %s\n", i+1, o)
			}
			scanner.Scan()
			i, err := strconv.Atoi(scanner.Text())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return options[i-1]
		case "e":
			scanner.Scan()
			return scanner.Text()
		case "s":
			return "xxx"
		case "c":
			os.Exit(1)
		default:
		}
		fmt.Printf("s = skip / e = enter pinyin / c = cancel \n")
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("could not read input: %v\n", err)
		os.Exit(1)
	}
	return ""
}

// parse

func getSentences(ignore Ignored, sentences []string) []Sentence {
	var results []Sentence
	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}
		allWords, newWords := getWords(sentence, ignore)
		results = append(results, Sentence{
			Chinese:      sentence,
			Audio:        hash(sentence),
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(sentence) == 1,
		})
	}
	return results
}

func getWords(sentence string, ignore Ignored) ([]Word, []Word) {
	sentence = strings.ReplaceAll(sentence, "。", " ")
	sentence = strings.ReplaceAll(sentence, ".", " ")
	sentence = strings.ReplaceAll(sentence, "，", " ")
	sentence = strings.ReplaceAll(sentence, ",", " ")
	sentence = strings.ReplaceAll(sentence, "?", " ")
	sentence = strings.ReplaceAll(sentence, "？", " ")
	sentence = strings.ReplaceAll(sentence, "！", " ")
	sentence = strings.ReplaceAll(sentence, "!", " ")

	var parts []string
	if strings.Contains(sentence, " ") {
		parts = strings.Split(sentence, " ")
	} else if strings.Contains(sentence, " ") {
		parts = strings.Split(sentence, " ")
	} else {
		parts = []string{sentence}
	}

	var allWords []Word
	for _, word := range parts {
		if word == "" {
			continue
		}
		allChars := getAllChars(word)
		allWords = append(allWords, Word{
			Chinese:      word,
			Audio:        hash(word),
			AllChars:     allChars,
			IsSingleRune: utf8.RuneCountInString(word) == 1,
		})
	}

	var newWords []Word
	for _, word := range allWords {
		if _, ok := ignore[word.Chinese]; ok {
			continue
		}
		ignore.update(word.Chinese)

		// set new chars after word has been added to ignore list,
		// we want to add words first, then chars
		word.NewChars = getNewChars(ignore, word.AllChars)
		newWords = append(newWords, word)
	}

	return allWords, newWords
}

func getAllChars(word string) []Char {
	allChars := make([]Char, 0)
	for _, char := range word {
		allChars = append(allChars, Char{
			Chinese:      string(char),
			IsSingleRune: true,
		})
	}
	return allChars
}

func getNewChars(ignore Ignored, allChars []Char) []Char {
	newChars := make([]Char, 0)
	for _, char := range allChars {
		if _, ok := ignore[char.Chinese]; ok {
			continue
		}
		newChars = append(newChars, char)
		ignore.update(char.Chinese)
	}
	return newChars
}

func loadDialogs(path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open dialogues file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	var dialogues [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			dialogues = append(dialogues, lines)
			lines = []string{}
			continue
		}
		lines = append(lines, line)
	}
	return dialogues
}

func loadFile(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open vocab file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func loadIgnored(path string) Ignored {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open ignore file: %v", err)
		os.Exit(1)
	}
	var i Ignored
	if err := yaml.Unmarshal(b, &i); err != nil {
		fmt.Printf("could not unmarshal ignore file: %v", err)
		os.Exit(1)
	}
	return i
}

func writeIgnored(i Ignored, path string) {
	data, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("could not marshal ignore file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write ignore file: %v", err)
		os.Exit(1)
	}
}

func writeToFile(i any, path string) {
	data, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("could not marshal dialog file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write dialog file: %v", err)
		os.Exit(1)
	}
}

// translate

type Translations map[string]string

func (t Translations) update(ch, en string) {
	t[ch] = en
}

func translateDialogNew(t Translations, d *Dialog) {
	translation, ok := t[d.Chinese]
	if !ok {
		var err error
		translation, err = translateText("en-US", d.Chinese)
		if err != nil {
			fmt.Println(err)
		}
	}
	d.English = translation
	t.update(d.Chinese, d.English)
	d.Sentences = translateNewWordsInSentences(t, d.Sentences)
}

func translateNewWordsInSentences(t Translations, sentences []Sentence) []Sentence {
	for i, sentence := range sentences {
		translation, ok := t[sentence.Chinese]
		if !ok {
			var err error
			translation, err = translateText("en-US", sentence.Chinese)
			if err != nil {
				log.Fatalf("could not translate sentence \"%s\": %v", sentence.Chinese, err)
			}
		}
		sentence.English = translation
		t.update(sentence.Chinese, sentence.English)

		for y, word := range sentence.NewWords {
			translation, ok := t[word.Chinese]
			if !ok {
				var err error
				translation, err = translateText("en-US", word.Chinese)
				if err != nil {
					log.Fatalf("could not translate word \"%s\": %v", word.Chinese, err)
				}
			}
			word.English = translation
			t.update(word.Chinese, word.English)

			for z, char := range word.NewChars {
				translation, ok := t[char.Chinese]
				if !ok {
					var err error
					translation, err = translateText("en-US", char.Chinese)
					if err != nil {
						log.Fatalf("could not translate char \"%s\": %v", char.Chinese, err)
					}
				}
				char.English = translation
				t.update(char.Chinese, char.English)
				word.NewChars[z] = char
			}
			sentence.NewWords[y] = word
		}
		sentences[i] = sentence
	}
	return sentences
}

func translateDialog(t Translations, d *Dialog) {
	d.Sentences = translateAllWordsInSentences(t, d.Sentences)
}

func translateAllWordsInSentences(t Translations, sentences []Sentence) []Sentence {
	for i, sentence := range sentences {
		sentence.AllWords = translateAllWords(t, sentence.AllWords)
		sentences[i] = sentence
	}
	return sentences
}

func translateAllWords(t Translations, words []Word) []Word {
	var translated []Word
	for _, word := range words {
		translation, ok := t[word.Chinese]
		if !ok {
			var err error
			translation, err = translateText("en-US", word.Chinese)
			if err != nil {
				log.Fatalf("could not translate word \"%s\": %v", word.Chinese, err)
			}
		}
		word.English = translation
		t.update(word.Chinese, word.English)

		for z, char := range word.AllChars {
			translation, ok := t[char.Chinese]
			if !ok {
				var err error
				translation, err = translateText("en-US", char.Chinese)
				if err != nil {
					log.Fatalf("could not translate char \"%s\": %v", char.Chinese, err)
				}
			}
			char.English = translation
			t.update(char.Chinese, char.English)
			word.AllChars[z] = char
		}
		translated = append(translated, word)
	}
	return translated
}

func translateText(targetLanguage, text string) (string, error) {
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	fmt.Printf("translate: %s...\n", text)
	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("translate returned empty response to text: %s", text)
	}
	trans := resp[0].Text
	trans = strings.ReplaceAll(trans, "&#39;", "'")
	fmt.Println(trans)
	return trans, nil
}

func loadTranslations(path string) Translations {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open translations file: %v", err)
		os.Exit(1)
	}
	var t Translations
	if err := yaml.Unmarshal(b, &t); err != nil {
		fmt.Printf("could not unmarshal translations file: %v", err)
		os.Exit(1)
	}
	return t
}

func writeTranslations(t Translations, path string) {
	data, err := yaml.Marshal(t)
	if err != nil {
		fmt.Printf("could not marshal translations file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write translations file: %v", err)
		os.Exit(1)
	}
}

// audio

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h.Sum(nil))
}

func getAudio(audioDirPath string, d *Dialog) {
	ctx := context.Background()

	filename, err := fetchAudio(ctx, d.Chinese, audioDirPath, hash(d.Chinese))
	if err != nil {
		fmt.Println(err)
	}
	d.Audio = filename
	d.Sentences = getSentenceAudio(ctx, audioDirPath, d.Sentences)
}

func getSentenceAudio(ctx context.Context, audioDirPath string, sentences []Sentence) []Sentence {
	for x, sentence := range sentences {
		filename, err := fetchAudio(ctx, sentence.Chinese, audioDirPath, hash(sentence.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		sentence.Audio = filename

		for y, word := range sentence.NewWords {
			filename, err := fetchAudio(ctx, word.Chinese, audioDirPath, hash(word.Chinese))
			if err != nil {
				fmt.Println(err)
			}
			word.Audio = filename

			for z, ch := range word.NewChars {
				filename, err := fetchAudio(ctx, ch.Chinese, audioDirPath, hash(ch.Chinese))
				if err != nil {
					fmt.Println(err)
				}
				ch.Audio = filename
				word.NewChars[z] = ch
			}
			sentence.NewWords[y] = word
		}
		sentences[x] = sentence
	}
	return sentences
}

func fetchAudio(ctx context.Context, query, audioDir, filename string) (string, error) {
	if err := os.MkdirAll(audioDir, os.ModePerm); err != nil {
		return "", err
	}
	filename = filename + ".mp3"
	path := filepath.Join(audioDir, filename)
	globalPath := filepath.Join(audioDir, "..", "..", "audio", filename)

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("audio file exists: %s\n", path)
		return filename, nil
	}
	if _, err := os.Stat(globalPath); err == nil {
		fmt.Printf("audio file exists: %s\n", globalPath)
		return filename, nil
	}

	time.Sleep(1 * time.Second)
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: query},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "cmn-CN",
			Name:         "cmn-CN-Wavenet-C",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", err
	}

	// The resp's AudioContent is binary.
	err = ioutil.WriteFile(path, resp.AudioContent, 0644)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(globalPath, resp.AudioContent, 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("%v\n", query)
	fmt.Printf("audio content written to file: %v\n", path)
	return filename, nil
}

// export

func writeDialogCards(d *Dialog, deckName, templatePath, outPath string) {
	text, err := fillDialogTemplate(d, deckName, templatePath)
	if err != nil {
		fmt.Printf("could not fill template: %v\n", err)
		os.Exit(1)
	}
	appendToAnkiCards(text, outPath)
}

func writeSentenceCards(s []Sentence, deckName, templatePath, outPath string) {
	text, err := fillSentencesTemplate(s, deckName, templatePath)
	if err != nil {
		fmt.Printf("could not fill template: %v\n", err)
		os.Exit(1)
	}
	appendToAnkiCards(text, outPath)
}

func writeWordsCards(w []Word, deckName, templatePath, outPath string) {
	text, err := fillWordsTemplate(w, deckName, templatePath)
	if err != nil {
		fmt.Printf("could not fill template: %v\n", err)
		os.Exit(1)
	}
	appendToAnkiCards(text, outPath)
}
func appendToAnkiCards(text string, outPath string) {
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("could not open anki cards file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		fmt.Printf("could not append to anki cards file: %v\n", err)
	}
}

var tplFuncMap = template.FuncMap{
	"audio": func(query string) string {
		return "[sound:" + hash(query) + ".mp3]"
	},
	"removeSpaces": func(s string) string {
		return strings.ReplaceAll(s, " ", "")
	},
	"deckName": func() string {
		return deckName
	},
	"tags": func() string {
		return strings.Join(tags, ", ")
	},
}

func fillDialogTemplate(d *Dialog, deckName, tmplPath string) (string, error) {
	tmpl, err := template.New("dialog.tmpl").Funcs(tplFuncMap).ParseGlob(tmplPath + "/*.tmpl")
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, d)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func fillSentencesTemplate(s []Sentence, deckName, tmplPath string) (string, error) {
	tmpl, err := template.New("sentences.tmpl").Funcs(tplFuncMap).ParseGlob(tmplPath + "/*.tmpl")
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func fillWordsTemplate(s []Word, deckName, tmplPath string) (string, error) {
	tmpl, err := template.New("words.tmpl").Funcs(tplFuncMap).ParseGlob(tmplPath + "/*.tmpl")
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeCardsToFile(data, outPath string) {
	if err := os.WriteFile(outPath, []byte(data), 0644); err != nil {
		fmt.Printf("could not write anki cards: %v", err)
		os.Exit(1)
	}
}
