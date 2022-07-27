package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

var query string
var outputDir string
var convert bool

func main() {
	flag.StringVar(&query, "q", "", "query to show and record")
	flag.StringVar(&outputDir, "o", "./media", "output directory")
	flag.BoolVar(&convert, "c", false, "convert to mp3")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("cancel")
		cancel()
		os.Exit(1)
	}()

	fmt.Println(query)
	fmt.Printf("r = record / p = play recording / c = cancel / s = save \n")

	err := os.Mkdir(outputDir, os.ModeDir)
	if !errors.Is(err, os.ErrExist) {
		fmt.Printf("could not create output dir: %v\n", err)
		os.Exit(1)
	}
	path := filepath.Join(outputDir, query)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "c" {
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
		}
		fmt.Printf("r = record / p = play recording / d = delete / c = cancel \n")
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
			if err.Error() != "signal: killed" {
				fmt.Println(err)
			}
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		cancel()
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
