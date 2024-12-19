package template

import (
	_ "embed"
	html "html/template"
	"io"
	"os"
	"path"
	text "text/template"

	"golang.org/x/text/language"
)

type IndexData struct {
	Scrabble string
	User     string
}

type AutoplayData struct {
	Scrabble string
	User     string
	Autoplay string
}

type ErrorData struct {
	Error string
}

type Templates interface {
	WriteTemplateScript(string) error
	WriteTemplateStyles(string) error
	WriteTemplate(w io.Writer, name string, data any)
	WriteError(w io.Writer, error string)
}

type _Templates struct {
	directory string
	text      *text.Template
	html      *html.Template
}

func CreateTemplates(language language.Tag) Templates {
	templates := &_Templates{
		directory: "templates",
	}
	templates.html = html.Must(html.ParseGlob(path.Join(templates.directory, "*.html")))
	//templates.text = text.Must(text.ParseGlob(path.Join(templates.directory, "*.text")))
	return templates
}

func (templates *_Templates) WriteTemplateStyles(dirName string) error {
	return templates.CopyTemplateFile(dirName, "styles.css")
}

func (templates *_Templates) WriteTemplateScript(dirName string) error {
	return templates.CopyTemplateFile(dirName, "script.js")
}

func (templates *_Templates) WriteTemplate(w io.Writer, name string, data any) {
	t := templates.html.Lookup(name)
	if t == nil {
		templates.WriteError(w, `Cannot locate template "error.html"`)
		return
	}

	if err := t.Execute(w, data); err != nil {
		templates.WriteError(w, err.Error())
		return
	}
}

func (templates *_Templates) WriteError(w io.Writer, error string) {
	t := templates.html.Lookup("error.html")
	if t == nil {
		panic(`Cannot locate template "error.html"` + "\n" + error)
	}
	data := ErrorData{error}
	if err := t.Execute(w, data); err != nil {
		panic(`Error executing template "error.html"` + "\n" + err.Error())
	}
}

func (templates *_Templates) CopyTemplateFile(dirName string, fileName string) error {
	var err error
	var content []byte

	sourceFilePath := path.Join(templates.directory, fileName)
	destinationFilePath := path.Join(dirName, fileName)
	destinationTmpFilePath := destinationFilePath + "~"
	defer func() {
		os.Remove(destinationTmpFilePath)
	}()
	if content, err = os.ReadFile(sourceFilePath); err != nil {
		return err
	}
	if err = os.WriteFile(destinationTmpFilePath, content, 0644); err != nil {
		return err
	}
	if err = os.Rename(destinationTmpFilePath, destinationFilePath); err != nil {
		return err
	}
	return nil
}
