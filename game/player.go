package game

type PlayerNo uint8
type PlayerId uint

type Player struct {
	id   PlayerId
	name string
}

type Players []*Player

const MaxBotPlayers PlayerNo = PlayerNo(10)
const NoPlayer = PlayerNo(0)
const SystemPlayerId = PlayerId(0)

var BotPlayerNames = [MaxBotPlayers]string{
	"*Alice*",
	"*Bob*",
	"*John*",
	"*Emma*",
	"*Fred*",
	"*Lisa*",
	"*Paul*",
	"*Vera*",
	"*Bill*",
	"*Karen*",
}

var nextPlayerId = 1000

var SystemPlayer = &Player{id: SystemPlayerId, name: "__SYSTEM__"}

var botPlayers Players = make(Players, MaxBotPlayers)

func BotPlayer(no PlayerNo) *Player {
	if no == NoPlayer || no >= MaxBotPlayers {
		return nil
	}
	if botPlayers[no] == nil {
		name := BotPlayerNames[no-1]
		botPlayers[no] = &Player{id: PlayerId(no + 100), name: name}
	}
	return botPlayers[no]
}
