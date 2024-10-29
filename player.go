package main

import (
	"fmt"
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
		name := fmt.Sprintf("__BOT:%v__", no)
		botPlayers[no] = &Player{id: no, name: name}
	}
	return botPlayers[no]
}
