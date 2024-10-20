package main

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type PlayerNo uint

type Player struct {
	id   PlayerNo
	name string
}

type Players []*Player

var SystemPlayer = Player{id: 0, name: "__SYSTEM__"}

const MaxBotPlayers PlayerNo = 10

var botPlayers Players = make(Players, MaxBotPlayers)

func BotPlayer(no PlayerNo) *Player {
	if no < 1 || no >= MaxBotPlayers {
		return nil
	}
	if botPlayers[no] == nil {
		p := message.NewPrinter(language.Danish)

		botPlayers[no] = &Player{id: no, name: p.Sprintf("__BOT:%u__", no)}
	}
	return botPlayers[no]
}
