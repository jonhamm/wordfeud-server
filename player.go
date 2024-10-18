package main

type Player struct {
	id   int
	name string
}

var SystemPlayer = Player{id: 0, name: "__SYSTEM__"}
