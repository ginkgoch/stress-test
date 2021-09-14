package main

import (
	"log"

	"github.com/ginkgoch/stress-test/pkg/talent/gamev2/utils"
)

func main() {
	gameConfig := utils.GameConfig{
		ID:       "backpack",
		PlayerID: 14258464,
		RoomID:   "baap_14258464",
		Server:   "gameserver.moblab-us.cn/gameserver-0",
		GameURL:  "https://vgame.moblab-us.cn/backpack/dev/",
	}
	gameClient := utils.NewGameClient(&gameConfig)
	err := gameClient.PlayGame()
	if err != nil {
		log.Println(err)
	} else {
		log.Println("completed")
	}
}
