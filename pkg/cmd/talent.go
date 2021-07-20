package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ginkgoch/stress-test/pkg/talent"
	"github.com/spf13/cobra"
)

var (
	filepath string
)

func init() {
	toCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "", "<signing in user list file>.json")

	toCmd.MarkFlagRequired("filepath")

	toCmd.Example = "stress-test talent signin -f ~/Downloads/2W-user.json"

	rootCmd.AddCommand(toCmd)
}

var toCmd = &cobra.Command{
	Use:   "talent <action: signin>",
	Short: "Talent optimization test",
	Long:  `Talent optimization test`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			log.Fatalf("file not exits <%s>", filepath)
		}

		fileBuffer, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatalf("open file failed - %v\n", err)
		}

		var userList []map[string]string

		json.Unmarshal(fileBuffer, &userList)

		userLength := len(userList)
		if userLength == 0 {
			log.Fatal("no user loaded")
		}

		fmt.Printf("loaded %v users \n", userLength)

		talent := new(talent.TalentObject)

		httpClient := NewHttpClientWithoutRedirect(true)
		err = talent.SignIn(userList[0], httpClient)
		if err != nil {
			log.Fatal("sign-in failed", err)
		}

		fmt.Println("response cookie", talent.Cookie)
	},
}
