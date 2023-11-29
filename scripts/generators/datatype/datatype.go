package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Datatype name not specified, you can use multiple words divided by space ie: 'make datatype shelf item'")
	}

	words := append([]string{}, os.Args[1:]...)
	for i, w := range words {
		words[i] = strings.ToLower(w)
		fmt.Printf("%d: %s\n", i+1, words[i])
	}

	joinedSpace := strings.Join(words, " ")
	joinedUnderscore := strings.Join(words, "_")
	data := struct {
		Upper          string
		Lower          string
		Store          string
		Multiple       string
		Letter         string
		MultiplePascal string
		Collection     string
	}{
		Upper:          Title(joinedSpace),
		Lower:          Lower(joinedSpace),
		Store:          AddStore(joinedSpace),
		Multiple:       AddS(joinedSpace),
		Letter:         FirstLetter(joinedSpace),
		MultiplePascal: Title(AddS(joinedSpace)),
		Collection:     AddS(joinedUnderscore),
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("getting working directory: %v", err)
	}

	templateContent, err := os.ReadFile(filepath.Join(currentDir, "/scripts/generators/datatype/", "datatype.tmpl"))
	if err != nil {
		log.Fatalf("reading template file: %v", err)
	}

	tmpl, err := template.New("datatype").Parse(string(templateContent))
	if err != nil {
		log.Fatalf("parsing template: %v", err)
	}

	file, err := os.Create(filepath.Join(currentDir, "data/", joinedUnderscore+".go"))
	if err != nil {
		log.Fatalf("creating datatyoe .go file: %v", err)
	}
	defer file.Close()

	if err = tmpl.Execute(file, data); err != nil {
		log.Fatalf("passing template to datatype .go file: %v", err)
	}

	fmt.Printf("file '%s' created successfully.\n", file.Name())
}

func Title(s string) string {
	return strings.ReplaceAll(cases.Title(language.English).String(s), " ", "")
}

func Lower(s string) string {
	s = Title(s)
	return strings.ToLower(s[:1]) + s[1:]
}

func AddS(s string) string {
	return s + "s"
}

func AddStore(s string) string {
	return Title(s) + "Store"
}

func FirstLetter(s string) string {
	return string(s[0])
}
