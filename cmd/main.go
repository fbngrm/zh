package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type key string

const (
	kDefinition key = "kDefinition"
	kMandarin   key = "kMandarin"
)

func main() {
	if err := loadUnihanReadings(); err != nil {
		fmt.Printf("could not parse unihan data: %v", err)
		os.Exit(1)
	}

	//loadCJKDecompositionData()
}

func loadUnihanReadings() error {
	file, err := os.Open("./lib/Unihan_Readings.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	unihan := make(map[string]map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		words := strings.Fields(line)
		if len(line) < 3 {
			continue
		}
		uchar := words[0]
		char, err := toHanzi(uchar)
		if err != nil {
			return err
		}
		key := words[1]
		value := words[2]
		if _, ok := unihan[char]; !ok {
			unihan[char] = make(map[string]string)
		}
		unihan[char][key] = value
	}

	for k, v := range unihan {
		fmt.Println(k, v)
	}
	return scanner.Err()
}

func toHanzi(s string) (string, error) {
	if len(s) != 6 {
		return "", nil
	}
	r, err := strconv.ParseInt(s[2:], 16, 32)
	return string(rune(r)), err
}

func loadCJKDecompositionData() {
	file, err := os.Open("./lib//cjk-decomp-0.4.0.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cjkDecomposition := make(map[string]string)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		if len(line) < 2 {
			continue
		}
		key := line[0]
		value := line[1]
		cjkDecomposition[key] = match(value)
	}

	for k, v := range cjkDecomposition {
		fmt.Println(k, v)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func match(s string) string {
	i := strings.Index(s, "(")
	if i >= 0 {
		j := strings.Index(s, ")")
		if j >= 0 {
			return s[i+1 : j]
		}
	}
	return ""
}
