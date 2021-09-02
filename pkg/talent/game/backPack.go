package game

import (
	"encoding/json"
	"math/rand"
	"time"
)

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

func (hf *BackPack) GameEnded(g *GameClient, msg *JoinedMsg) bool {
	return false
}

//UserJoined aa
func (hf *BackPack) UserJoined(g *GameClient, msg *JoinedMsg) {

}

//SessionEnded ss
func (hf *BackPack) SessionEnded(g *GameClient, msg *SessionEndedMsg) {

}

//GameStated ss
func (hf *BackPack) GameStated(g *GameClient, mgs *GameStartedMsg) {

}

//GameRoundStarted ss
func (hf *BackPack) GameRoundStarted(g *GameClient, mgs *GameRoundMsg) {

}

//GameRoundEnded ss
func (hf *BackPack) GameRoundEnded(g *GameClient, mgs *GameRoundMsg) {
}

//PlayerUpdated ss PLAYER_UPDATED
func (hf *BackPack) PlayerUpdated(g *GameClient, msg []byte) {
	playerUpdated := &PlayerUpdated{}
	err := json.Unmarshal(msg, playerUpdated)
	if err != nil {
		g.stopWatch.Log("json Unmarshal err", err.Error())
		return
	}
	for _, move := range playerUpdated.Moves {
		switch move {
		case QUIT:
			if hf.currentRound != g.Round {
				hf.currentRound = g.Round
				g.stopWatch.Get(GAME_ROUND_STARTED, QUIT)
			} else {
				return
			}
			g.stopWatch.End(QUIT, "")
			action := Action{Action: QUIT, Player: playerUpdated.PlayerNumber}
			rand.Seed(time.Now().UnixNano())
			action.Data = []int{g.Round, 0}
			g.SendActionDelay(action, g.Round, rand.Intn(5)+5)
			return
		}
	}
}
