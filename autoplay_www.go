package main

/*
import (
	"html/template"
	"net/http"
	. "wordfeud/context"
	. "wordfeud/game"
)

func autoplayWWW(server *Server, w http.ResponseWriter, req *http.Request) {
	server.serviceOptions.FileFormat = FILE_FORMAT_WWW
	server.serviceOptions.Directory = "www"
	game, err := NewGame(server.serviceOptions, server.seqno, Players{BotPlayer(1), BotPlayer(2)})
	p := game.Fmt()
	if err != nil {
		p.Println(w, err.Error())
		return
	}
	result.Width = int(game.Dimensions().Width)
	result.Height = int(game.Dimensions().Height)
	result.LetterScores = game.LetterScores()
	result.Board = game.Board()

	for n := 0; n < 1000; n++ {
		if !game.Play() {
			break
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
*/
