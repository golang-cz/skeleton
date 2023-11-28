package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"

	"github.com/BurntSushi/toml"
)

const (
	blackColor   = "\033[30m"
	redColor     = "\033[31m"
	greenColor   = "\033[32m"
	yellowColor  = "\033[33m"
	blueColor    = "\033[34m"
	magentaColor = "\033[35m"
	bgRedColor   = "\033[41m"
	bgGreenColor = "\033[42m"
	bgBlueColor  = "\033[44m"
	resetColor   = "\033[0m"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: toml-compare <file1.toml> <file2.toml>")
		return
	}

	file1Path := os.Args[1]
	file2Path := os.Args[2]

	file1Keys := getKeys(file1Path)
	file2Keys := getKeys(file2Path)

	missingInFile1 := findMissingKeys(file1Keys, file2Keys)
	missingInFile2 := findMissingKeys(file2Keys, file1Keys)

	fmt.Printf("%v\n1 - %v\n2 - %v\n",
		colorize(" CHECK DIFFERENCES IN TOML KEYS CONFIGS ", blackColor, bgBlueColor),
		colorize(file1Path, magentaColor),
		colorize(file2Path, magentaColor),
	)

	if len(missingInFile1) > 0 || len(missingInFile2) > 0 {
		if len(missingInFile1) > 0 {
			printMissingKeys(file1Path, missingInFile1)
		}

		if len(missingInFile2) > 0 {
			printMissingKeys(file2Path, missingInFile2)
		}

		fmt.Println()
		os.Exit(1)
	}

	fmt.Printf("%v\n\n", colorize(" PASS ", blackColor, bgGreenColor))
}

func getKeys(filePath string) []string {
	var data interface{}
	if _, err := toml.DecodeFile(filePath, &data); err != nil {
		fmt.Println("Error decoding TOML:", err)
		os.Exit(1)
	}

	keys := extractKeys(data, "")
	sort.Strings(keys)
	return keys
}

func extractKeys(data interface{}, prefix string) []string {
	keys := make([]string, 0)
	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			newKey := prefix + key.String()
			keys = append(keys, newKey)
			keys = append(keys, extractKeys(v.MapIndex(key).Interface(), newKey+".")...)
		}
	}

	return keys
}

func findMissingKeys(keysToCheck, referenceKeys []string) []string {
	missingKeys := []string{}

	referenceKeySet := make(map[string]bool)
	for _, key := range referenceKeys {
		referenceKeySet[key] = true
	}

	for _, key := range keysToCheck {
		if _, ok := referenceKeySet[key]; !ok {
			missingKeys = append(missingKeys, key)
		}
	}

	return missingKeys
}

func colorize(text string, colors ...string) string {
	var colorizedString string
	for _, c := range colors {
		colorizedString += c
	}

	colorizedString += text + resetColor

	return colorizedString
}

func printMissingKeys(filepath string, missingKeys []string) {
	fmt.Printf("%s %v in %s\n", colorize(" FAIL ", blackColor, bgRedColor), colorize("missing keys", redColor), colorize(filepath, blueColor))

	digits := countDigits(len(missingKeys))

	for i, e := range missingKeys {
		fmt.Printf("%-*d - %s\n", digits, i+1, colorize(e, yellowColor))
	}
}

func countDigits(number int) int {
	numberStr := strconv.Itoa(number)

	return len(numberStr)
}
