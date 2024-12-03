package main

import (
	_ "embed"
	"html/template"
	"net/http"
	"os"
	"path"
	. "wordfeud/context"
	. "wordfeud/localize"
)

type Scrabble struct {
	options     *GameOptions
	callCount   int
	seqno       int
	templateDir string
	wwwDir      string
	templates   *template.Template
}

type indexData struct {
	Scrabble string
	Autoplay string
}
type errorData struct {
	Error string
}

var scrabble = Scrabble{
	options:     nil,
	callCount:   0,
	seqno:       1,
	templateDir: "templates",
	wwwDir:      "www",
	templates:   nil,
}

func scrabbleWWW(server *Server, w http.ResponseWriter, req *http.Request) {
	if scrabble.callCount == 0 {
		scrabble.init(server)
	}
	scrabble.callCount++
	lang := scrabble.options.Language
	data := indexData{
		Scrabble: Localized(lang, "Scrabble"),
		Autoplay: Localized(lang, "Two robot player game"),
	}
	scrabble.writeTemplate(w, "index.html", data)
}

func (scrabble *Scrabble) init(server *Server) {
	scrabble.options = server.serviceOptions
	scrabble.templates = template.Must(template.ParseGlob(path.Join(scrabble.templateDir, "*.html")))
	scrabble.options.FileFormat = FILE_FORMAT_WWW
	scrabble.options.Directory = "www"
	writeTemplateScript(scrabble.options.Directory)
	writeTemplateStyles(scrabble.options.Directory)
}

func (scrabble *Scrabble) writeTemplate(w http.ResponseWriter, name string, data any) {
	t := scrabble.templates.Lookup(name)
	if t == nil {
		scrabble.writeError(w, `Cannot locate template "error.html"`)
		return
	}

	if err := t.Execute(w, data); err != nil {
		scrabble.writeError(w, err.Error())
		return
	}
}

func (scrabble *Scrabble) writeError(w http.ResponseWriter, error string) {
	t := scrabble.templates.Lookup("error.html")
	if t == nil {
		panic(`Cannot locate template "error.html"` + "\n" + error)
	}
	data := errorData{error}
	if err := t.Execute(w, data); err != nil {
		panic(`Error executing template "error.html"` + "\n" + err.Error())
	}
}

//go:embed templates/styles.css
var cssStyles string

func writeTemplateStyles(dirName string) error {
	var err error
	var f *os.File

	fileName := "styles.css"
	tmpFileName := fileName + "~"
	filePath := path.Join(dirName, fileName)
	tmpFilePath := path.Join(dirName, tmpFileName)
	if f, err = os.Create(tmpFilePath); err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name()) // clean up
		}
	}()

	if _, err = f.WriteString(cssStyles); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}
	f = nil
	return nil
}

//go:embed templates/script.js
var script string

func writeTemplateScript(dirName string) error {
	var err error
	var f *os.File

	fileName := "script.js"
	tmpFileName := fileName + "~"
	filePath := path.Join(dirName, fileName)
	tmpFilePath := path.Join(dirName, tmpFileName)
	if f, err = os.Create(tmpFilePath); err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name()) // clean up
		}
	}()

	if _, err = f.WriteString(script); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}
	f = nil
	return nil
}
