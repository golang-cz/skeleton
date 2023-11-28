package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type templateData struct {
	Pascal   string
	Lower    string
	Store    string
	Multiple string
	Letter   string
}

func AddStore(s string) string {
	return strings.Title(s) + "Store"
}

func AddS(value string) string {
	return value + "s"
}

func FirstLetter(s string) string {
	return string(s[0])
}

func main() {
	dataName := flag.String("d", "", "Datatype name ie. '-d user'")

	flag.Parse()

	if *dataName == "" {
		flag.PrintDefaults()
		log.Fatal("Datatype name not specified")
	}

	data := templateData{
		Pascal:   strings.Title(*dataName),
		Lower:    strings.ToLower(*dataName),
		Store:    AddStore(*dataName),
		Multiple: AddS(*dataName),
		Letter:   FirstLetter(*dataName),
	}

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl := template.Must(template.New("myTemplate").Funcs(funcMap).Parse(string(templateString)))

	targetDir := "../../../data/"
	fileName := data.Lower + ".go"
	file, err := os.Create(filepath.Join(targetDir, fileName))
	if err != nil {
		panic(err)
	}

	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("File '%s' created successfully.\n", fileName)
}

var templateString string = `package data

import (
	"github.com/upper/db/v4"
	"github.com/golang-cz/skeleton/pkg/utc"
	"github.com/golang-cz/skeleton/proto"
)

type {{.Pascal}} struct {
	*proto.{{.Pascal}}
}

type {{.Store}} struct {
	db.Collection
}

// Interface checks
var _ = interface {
	db.Record
	db.BeforeCreateHook
	db.BeforeUpdateHook
}(&{{.Pascal}}{})

var _ = interface {
	db.Store
}(&{{.Store}}{})

func {{.Multiple}}(sess db.Session) *{{.Store}} {
	return &{{.Store}}{sess.Collection("{{ToLower .Multiple}}")}
}

func (u *{{.Pascal}}) Store(sess db.Session) db.Store {
	return {{.Multiple}}(sess)
}

func ({{.Letter}} *{{.Pascal}}) BeforeCreate(sess db.Session) error {
	{{.Letter}}.CreatedAt = utc.Now()
	{{.Letter}}.UpdatedAt = {{.Letter}}.CreatedAt

	return nil
}

func ({{.Letter}} *{{.Pascal}}) BeforeUpdate(sess db.Session) error {
	{{.Letter}}.UpdatedAt = utc.Now()

	return nil
}

func (s {{.Store}}) Find(conds ...interface{}) db.Result {
	return s.Collection.Find(append([]interface{}{db.Cond{}}, conds...)...)
}

func (s {{.Store}}) FindOne(conds ...interface{}) (*{{.Pascal}}, error) {
	var {{.Lower}} *{{.Pascal}}
	res := s.Find(conds...)
	exists, err := res.Exists()
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	err = res.One(&{{.Lower}})
	if err != nil {
		return nil, err
	}

	return {{.Lower}}, nil
}`
