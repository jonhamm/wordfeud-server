package main

import (
	"net/http"
	. "wordfeud/game"
	. "wordfeud/localize"
)

type autoplayData struct {
	Scrabble     string
	Autoplay     string
	MainMenu     string
	AutoplayGame string
}

func autoplayWWW(server *Server, w http.ResponseWriter, req *http.Request) {
	scrabble := getScrabble(server)
	lang := scrabble.options.Language
	data := autoplayData{
		Scrabble:     Localized(lang, "Scrabble"),
		Autoplay:     Localized(lang, "Two robot player game"),
		MainMenu:     Localized(lang, "Top level menu"),
		AutoplayGame: Localized(lang, "Play game"),
	}

	scrabble.writeTemplate(w, "autoplay.html", data)
}

func autoplayGameWWW(server *Server, w http.ResponseWriter, req *http.Request) {
	scrabble := getScrabble(server)
	game, err := NewGame(scrabble.options, scrabble.seqno, Players{BotPlayer(1), BotPlayer(2)})
	if err != nil {
		scrabble.writeError(w, err.Error())
		return
	}

	for n := 0; n < 1000; n++ {
		if !game.Play() {
			break
		}
	}
	scrabble.seqno++
	http.Redirect(w, req, "scrabble/autoplay", http.StatusFound)
}