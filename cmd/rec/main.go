package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/fbngrm/zh/internal/anki"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"gopkg.in/yaml.v2"
)

var query string
var outputDir string
var convertToMP3, force, autoDownload bool
var in string
var deckName string
var blacklistFile *os.File

func main() {
	flag.StringVar(&query, "q", "", "query to show and record")
	flag.StringVar(&outputDir, "o", "", "output directory")
	flag.BoolVar(&convertToMP3, "c", false, "convert to mp3")
	flag.BoolVar(&force, "f", false, "force re-recording for existing audio files")
	flag.BoolVar(&autoDownload, "a", false, "download all audios")
	flag.StringVar(&in, "i", "", "input file")
	flag.StringVar(&deckName, "d", "", "anki deck name")
	flag.Parse()

	ankiSentences, err := load(in)
	if err != nil {
		fmt.Printf("could not load input data: %v\n", err)
		os.Exit(1)
	}

	_, name := filepath.Split(in)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if outputDir == "" {
		outputDir = filepath.Join("data", "gen", deckName, "audio")
	}
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		fmt.Printf("could not create output dir: %v\n", err)
		os.Exit(1)
	}
	blacklistPath := filepath.Join("data", "lib", deckName, "blacklist")
	blacklistFile, err = os.OpenFile(blacklistPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		fmt.Printf("could not open blacklist: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("cancel")
		cancel()
		os.Exit(1)
	}()

	fmt.Printf("f = fetch / r = record / p = play recording / d = delete / n = next / b = blacklist / a = download all \n")

	for _, sentence := range ankiSentences {
		// we assume all entries have the same deckname
		if deckName == "" {
			deckName = sentence.DeckName
		}
		fmt.Println()
		fmt.Println(sentence.Sentence.English)
		fmt.Println(sentence.Sentence.Pinyin)
		fmt.Println(sentence.Sentence.Chinese)

		query := sentence.Sentence.Chinese
		queryHash := hash(query)
		topic := sentence.Tags
		dir := filepath.Join(outputDir, topic)
		filename := deckName + "-" + topic + "_" + queryHash
		if autoDownload {
			_, err := fetchAudio(ctx, query, dir, filename, true)
			if err != nil {
				fmt.Printf("could not download audio: %v\n", err)
			}
		} else {
			loop(ctx, query, filename)
		}

		for _, hanzi := range sentence.Decompositions {
			fmt.Println()
			fmt.Println(hanzi.Hanzi.Definitions)
			fmt.Println(hanzi.Hanzi.Readings)
			fmt.Println(hanzi.Hanzi.Ideograph)
			if len(hanzi.Hanzi.Definitions) == 0 || len(hanzi.Hanzi.Readings) == 0 {
				fmt.Println("WARNING: no translations|readings found, blacklist?")
			}

			query := hanzi.Hanzi.Ideograph
			queryHash := hash(query)
			dir := filepath.Join(outputDir, topic)
			filename := deckName + "-" + topic + "_" + queryHash
			if autoDownload {
				_, err := fetchAudio(ctx, query, dir, filename, hanzi.IsWord)
				if err != nil {
					fmt.Printf("could not download audio: %v\n", err)
				}
			} else {
				loop(ctx, query, filename)
			}
		}
	}
	fmt.Println("Done")
	blacklistFile.Close()
}

func loop(ctx context.Context, query, path string) {
	scanner := bufio.NewScanner(os.Stdin)
	if fileExists(ctx, path, convertToMP3) {
		fmt.Printf("file exists, skipping. use -f flag to overwrite: %s\n", path)
		return
	}
	filename := ""
	for scanner.Scan() {
		text := scanner.Text()
		if text == "n" {
			break
		}
		switch text {
		// case "f":
		// 	var err error
		// 	filename, err = fetchAudio(ctx, query, path)
		// 	if err != nil {
		// 		fmt.Printf("could not download audio: %v\n", err)
		// 	}
		case "r":
			var err error
			filename, err = record(ctx, path, convertToMP3)
			if err != nil {
				fmt.Printf("could not record: %v\n", err)
			}
		case "p":
			err := play(ctx, filename, convertToMP3)
			if err != nil {
				fmt.Printf("could not play: %v\n", err)
			}
		case "d":
			err := deleteFile(ctx, path, convertToMP3)
			if err != nil {
				fmt.Printf("could not delete: %v\n", err)
			}
		case "b":
			_, err := blacklistFile.WriteString(query + "\n")
			if err != nil {
				fmt.Printf("could not append to blacklist: %v\n", err)
			}
			return
		default:
		}
		fmt.Printf("f = fetch / r = record / p = play recording / d = delete / n = next / b = blacklist \n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("could not read input: %v\n", err)
		os.Exit(1)
	}
}

func fileExists(ctx context.Context, path string, convertToMP3 bool) bool {
	file := path + ".wav"
	if convertToMP3 {
		file = path + ".mp3"
	}
	if _, err := os.Stat(file); err == nil && !force {
		return true
	}
	return false
}

func deleteFile(ctx context.Context, path string, convertToMP3 bool) error {
	file := path + ".wav"
	if convertToMP3 {
		file = path + ".mp3"
	}
	return os.Remove(file)
}

func play(ctx context.Context, path string, convertToMP3 bool) error {
	fmt.Println(path)
	cmd := exec.CommandContext(ctx, "ffplay", "-autoexit", path)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func record(ctx context.Context, path string, convertToMP3 bool) (string, error) {
	fmt.Println("Recording, press Enter to stop")
	path = filepath.Join("data", "lib", path)
	wavPath := path + ".wav"
	ctxRec, cancel := context.WithCancel(ctx)
	go func() {
		cmd := exec.CommandContext(ctxRec, "ffmpeg", "-y", "-f", "alsa", "-i", "hw:0,0", wavPath)
		if err := cmd.Run(); err != nil {
			e := err.Error()
			if e != "signal: killed" && e != "exit status 1" {
				fmt.Println(err)
			}
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		cancel()
		time.Sleep(1 + time.Second) // why?
		return "", err
	}
	cancel()

	if !convertToMP3 {
		return wavPath, nil
	}
	mp3Path := path + ".mp3"
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", wavPath, "-vn", "-ar", "44100", "-ac", "2", "-b:a", "192k", mp3Path)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return mp3Path, os.Remove(wavPath)
}

func load(path string) ([]anki.Sentence, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var as []anki.Sentence
	err = yaml.Unmarshal(yamlFile, &as)
	if err != nil {
		return nil, err
	}
	return as, nil
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h.Sum(nil))
}

func fetchAudio(ctx context.Context, query, dir, filename string, isWord bool) (string, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	filename = filename + ".mp3"
	path := filepath.Join(dir, filename)

	if _, err := os.Stat(path); err == nil && !force {
		fmt.Printf("file exists, skipping. use -f flag to overwrite: %s\n", path)
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

	// if this is a word, we store it in a separate dir which we use to generate audio loops.
	// we don't want audio loops from single characters but we want audio for single characters
	// on the flashcards.
	if isWord {
		wordsOnlyDir := filepath.Join(dir, "words_only")
		if err := os.MkdirAll(wordsOnlyDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
		wordsOnlyPath := filepath.Join(wordsOnlyDir, filename)
		err = ioutil.WriteFile(wordsOnlyPath, resp.AudioContent, 0644)
		if err != nil {
			return "", err
		}
	}

	fmt.Printf("Audio content written to file: %v\n", path)
	return filename, nil
}
