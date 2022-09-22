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
var convert bool
var in string
var deckName string

func main() {
	flag.StringVar(&query, "q", "", "query to show and record")
	flag.StringVar(&outputDir, "o", "", "output directory")
	flag.BoolVar(&convert, "c", false, "convert to mp3")
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

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("cancel")
		cancel()
		os.Exit(1)
	}()

	fmt.Printf("r = record / p = play recording / d = delete / n = next \n")

	for _, sentence := range ankiSentences {
		fmt.Println(sentence.Sentence.English)
		fmt.Println(sentence.Sentence.Pinyin)
		fmt.Println(sentence.Sentence.Chinese)
		loop(ctx, sentence.Sentence.Chinese)

		for _, hanzi := range sentence.Decomposition {
			fmt.Println(hanzi.Definitions)
			fmt.Println(hanzi.Readings)
			fmt.Println(hanzi.Ideograph)
			loop(ctx, hanzi.Ideograph)
		}
	}
}

func loop(ctx context.Context, query string) {
	queryHash := hash(query)
	path := filepath.Join(outputDir, queryHash)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "n" {
			break
		}
		switch text {
		case "r":
			err := record(ctx, path, convert)
			if err != nil {
				fmt.Printf("could not record: %v\n", err)
			}
		case "p":
			err := play(ctx, path, convert)
			if err != nil {
				fmt.Printf("could not play: %v\n", err)
			}
		case "d":
			err := deleteFile(ctx, path, convert)
			if err != nil {
				fmt.Printf("could not delete: %v\n", err)
			}
		default:

		}
		fmt.Printf("r = record / p = play recording / d = delete / n = next \n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("could not read input: %v\n", err)
		os.Exit(1)
	}
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
