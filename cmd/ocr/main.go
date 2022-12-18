package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	vision "cloud.google.com/go/vision/apiv1"
)

func main() {
	in := "/home/f/data/rslsync/ocr/17 texts"
	out := "/home/f/work/src/github.com/fbngrm/zh/data/lib/hsk1/ocr/17 texts"
	err := os.MkdirAll(out, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(in)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
		annotation, err := detectTextURI(filepath.Join(in, f.Name()))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(annotation)
		err = os.WriteFile(filepath.Join(out, f.Name()+".txt"), []byte(annotation), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// detectText gets text from the Vision API for an image at the given file path.
func detectTextURI(filepath string) (string, error) {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return "", err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %v", err)
	}
	defer file.Close()
	image, err := vision.NewImageFromReader(file)
	if err != nil {
		return "", fmt.Errorf("Failed to create image: %v", err)
	}

	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		return "", err
	}

	if len(annotations) == 0 {
		return "", errors.New("No text found.")
	}

	return annotations[0].Description, nil
}
