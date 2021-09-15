package main

import (
	"log"

	"github.com/ginkgoch/stress-test/pkg/talent/gamev2/utils"
)

func main() {
	gameConfig := utils.GameConfig{
		ID:       "backpack",
		PlayerID: 789035009,
		RoomID:   "baap_789035009",
		Server:   "game-test.moblab-us.cn/gameserver-0/game.com-server",
		GameURL:  "https://game-dist-test.moblab-us.cn/backpack/dev",
	}
	gameClient := utils.NewGameClient(&gameConfig)
	err := gameClient.PlayGame()
	if err != nil {
		log.Println(err)
	} else {
		log.Println("completed")
	}
}
