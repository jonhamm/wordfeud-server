package main

import (
	_ "embed"
	"net/http"
	. "wordfeud/context"
	. "wordfeud/localize"
	. "wordfeud/template"
)

type Scrabble struct {
	options     *GameOptions
	callCount   int
	seqno       int
	templateDir string
	wwwDir      string
	templates   Templates
}

var scrabbleData = Scrabble{
	options:   nil,
	callCount: 0,
	seqno:     1,
	templates: nil,
	wwwDir:    "www",
}

func getScrabble(server *Server) *Scrabble {
	if scrabbleData.options == nil {
		scrabbleData.init(server)
	}
	return &scrabbleData
}

func (scrabble *Scrabble) init(server *Server) {
	scrabble.options = server.serviceOptions
	scrabble.templates = CreateTemplates(scrabble.options.Language)
	scrabble.options.FileFormat = FILE_FORMAT_WWW
	scrabble.options.Directory = "www"
	scrabble.templates.WriteTemplateScript(scrabble.options.Directory)
	scrabble.templates.WriteTemplateStyles(scrabble.options.Directory)
}

func scrabbleWWW(server *Server, w http.ResponseWriter, req *http.Request) {
	userName := ""
	scrabble := getScrabble(server)
	scrabble.callCount++
	lang := scrabble.options.Language
	data := IndexData{
		Scrabble: Localized(lang, "Scrabble"),
		User:     userName,
	}
	scrabble.templates.WriteTemplate(w, "index.html", data)
}
