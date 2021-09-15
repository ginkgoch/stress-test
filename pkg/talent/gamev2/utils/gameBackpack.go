package utils

//a="minimum_effort_airport"
//minimum_effort_airport_target
//https://talent.default.serverless-test.moblab-us.cn/api/1/startGame/603f59df15a8cc001645a1bd/backpack

//BackPack pushpll

const (
	ADD  = "ADD"
	QUIT = "QUIT"
)

type BackPack struct {
	Delay        int
	currentRound int
}

func NewBackPack() *BackPack {
	return &BackPack{Delay: 8, currentRound: -1}
}

func (hf *BackPack) GameEnded(g *GameClient, msg *GameJoinedMsg) bool {
	return false
}

//UserJoined aa
func (hf *BackPack) UserJoined(g *GameClient, msg *GameJoinedMsg) {

}

//SessionEnded ss
func (hf *BackPack) SessionEnded(g *GameClient, msg *GameSessionEndedMsg) {

}

//GameStated ss
func (hf *BackPack) GameStarted(g *GameClient, mgs *GameStartedMsg) {

}

//GameRoundStarted ss
func (hf *BackPack) GameRoundStarted(g *GameClient, mgs *GameRoundMsg) {

}

//GameRoundEnded ss
func (hf *BackPack) GameRoundEnded(g *GameClient, mgs *GameRoundMsg) {
}

//PlayerUpdated ss PLAYER_UPDATED
func (hf *BackPack) PlayerUpdated(g *GameClient, msg *GamePlayerUpdatedMsg) {
	playerUpdated := msg
	for _, move := range playerUpdated.Moves {
		switch move {
		case QUIT:
			if hf.currentRound != g.Round {
				hf.currentRound = g.Round
			} else {
				return
			}
			action := GamePlayerAction{Action: QUIT, Player: playerUpdated.PlayerNumber}
			action.Data = []interface{}{g.Round, 0}
			g.SendAction(action, "/service/game/"+g.SessionId)
			return
		}
	}
}
