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
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fgrimme/zh/internal/anki"
	"gopkg.in/yaml.v2"
)

var query string
var outputDir string
var convertToMP3, force bool
var in string
var deckName string
var blacklistFile *os.File

func main() {
	flag.StringVar(&query, "q", "", "query to show and record")
	flag.StringVar(&outputDir, "o", "", "output directory")
	flag.BoolVar(&convertToMP3, "c", false, "convert to mp3")
	flag.BoolVar(&force, "f", false, "force re-recording for existing audio files")
	flag.StringVar(&in, "i", "", "input file")
	flag.StringVar(&deckName, "d", "", "anki deck name")
	flag.Parse()

	ankiSentences, err := load(in)
	if err != nil {
		fmt.Printf("could not load input data: %v\n", err)
		os.Exit(1)
	}

	dir, name := filepath.Split(in)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if outputDir == "" {
		outputDir = filepath.Join(dir, "audio")
	}
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		fmt.Printf("could not create output dir: %v\n", err)
		os.Exit(1)
	}
	blacklistPath := filepath.Join(dir, "blacklist")
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

	fmt.Printf("r = record / p = play recording / d = delete / n = next / b = blacklist \n")

	for _, sentence := range ankiSentences {
		// we assume all entries have the same deckname
		if deckName == "" {
			deckName = sentence.DeckName
		}
		fmt.Println()
		fmt.Println(sentence.Sentence.English)
		fmt.Println(sentence.Sentence.Pinyin)
		fmt.Println(sentence.Sentence.Chinese)
		loop(ctx, sentence.Sentence.Chinese)

		for _, hanzi := range sentence.Decomposition {
			fmt.Println()
			fmt.Println(hanzi.Definitions)
			fmt.Println(hanzi.Readings)
			fmt.Println(hanzi.Ideograph)
			if len(hanzi.Definitions) == 0 || len(hanzi.Readings) == 0 {
				fmt.Println("WARNING: no translations|readings found, blacklist?")
			}
			loop(ctx, hanzi.Ideograph)
		}
	}
	fmt.Println("Done")
	blacklistFile.Close()
}

func loop(ctx context.Context, query string) {
	queryHash := hash(query)
	path := filepath.Join(outputDir, deckName+"_"+queryHash)
	scanner := bufio.NewScanner(os.Stdin)
	if fileExists(ctx, path, convertToMP3) {
		fmt.Printf("file exists, skipping. use -f flag to overwrite: %s\n", path)
		return
	}
	for scanner.Scan() {
		text := scanner.Text()
		if text == "n" {
			break
		}
		switch text {
		case "r":
			err := record(ctx, path, convertToMP3)
			if err != nil {
				fmt.Printf("could not record: %v\n", err)
			}
		case "p":
			err := play(ctx, path, convertToMP3)
			if err != nil {
				fmt.Printf("could not play: %v\n", err)
			}
		case "d":
			err := deleteFile(ctx, path, convertToMP3)
			if err != nil {
				fmt.Printf("could not delete: %v\n", err)
			}
		case "b":
			_, err := blacklistFile.WriteString(query)
			if err != nil {
				fmt.Printf("could not append to blacklist: %v\n", err)
			}
			return
		default:
		}
		fmt.Printf("r = record / p = play recording / d = delete / n = next / b = blacklist \n")
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
	file := path + ".wav"
	if convertToMP3 {
		file = path + ".mp3"
	}
	fmt.Println(file)
	cmd := exec.CommandContext(ctx, "ffplay", "-autoexit", file)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func record(ctx context.Context, path string, convertToMP3 bool) error {
	fmt.Println("Recording, press Enter to stop")
	ctxRec, cancel := context.WithCancel(ctx)
	go func() {
		cmd := exec.CommandContext(ctxRec, "ffmpeg", "-y", "-f", "alsa", "-i", "hw:0,0", path+".wav")
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
		time.Sleep(1 + time.Second)
		return err
	}
	cancel()

	if !convertToMP3 {
		return nil
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", path+".wav", "-vn", "-ar", "44100", "-ac", "2", "-b:a", "192k", path+".mp3")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	return os.Remove(path + ".wav")
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
