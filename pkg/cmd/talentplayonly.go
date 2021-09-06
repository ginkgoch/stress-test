package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ginkgoch/stress-test/pkg/talent"
	"github.com/spf13/cobra"
)

var (
	gamesIds   []string
	serverInfo string
	playerId   string
	roomId     string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	toPlayCmd.PersistentFlags().StringArrayVarP(&gamesIds, "gamesIds", "g", []string{"competitive_math"}, `-g, game names, gamesIds are: competitive_math (default), ravens_matrices, push_pull, minimum_effort_airport, minimum_effort_airport_target, bomb_risk, backpack`)
	toPlayCmd.PersistentFlags().StringVarP(&serverInfo, "serverInfo", "", "", "--serverInfo")
	toPlayCmd.PersistentFlags().StringVarP(&playerId, "playerId", "", "", "--playerId")
	toPlayCmd.PersistentFlags().StringVarP(&roomId, "roomId", "", "", "--roomId, default competitive_math")

	toPlayCmd.Example = "stress-test talentplayonly -g competitive_math --serverInfo \"game-test.moblab-us.cn/gameserver-0/game.com-server\"  --playerId 789032002 --roomId baap_789032002"

	// gameID = "competitive_math"
	// gameID = "ravens_matrices"

	rootCmd.AddCommand(toPlayCmd)
}

var toPlayCmd = &cobra.Command{
	Use:   "talentplayonly",
	Short: "talentplayonly optimization test",
	Long:  `talentplayonly optimization test`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now().Unix()
		err := executeSingleTaskPlayGame()
		endTime := time.Now().Unix()
		if err != nil {
			fmt.Println("play game failed")
			return
		} else {
			durTime := endTime - startTime
			fmt.Printf("play game success, Total Time=%dsec\n", durTime)
		}
	},
}

func executeSingleTaskPlayGame() (err error) {
	talentObj := talent.NewTalentObject()
	for _, gameID := range gamesIds {
		err = talentObj.PlayGameOnly(gameID, serverInfo, playerId, roomId)
	}
	if err != nil {
		return err
	} else if debug {
		fmt.Println("debug - play game success")
	}
	return
}
