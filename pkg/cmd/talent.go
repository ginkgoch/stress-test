package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ginkgoch/stress-test/pkg/client"
	"github.com/ginkgoch/stress-test/pkg/client/runner"
	"github.com/ginkgoch/stress-test/pkg/talent"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

var (
	filepath           string
	stage              int
	useQps             bool
	playGame           bool
	storeTalentObjects bool
	delay              int
	endless            bool
	//gameID             string
	games []string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	toCmd.PersistentFlags().IntVarP(&delay, "delay", "", 0, "--delay <ms>, default 0")
	toCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "", "<signing in user list file>[.talent].json")
	toCmd.PersistentFlags().StringVarP(&talent.ServiceEndpoint, "serverEndpoint", "u", talent.DefaultServiceEndpoint, "https://talent.test.moblab-us.cn/api/1")
	toCmd.PersistentFlags().IntVarP(&stage, "stage", "t", 0, "-t <stage>, default 0")
	toCmd.PersistentFlags().BoolVarP(&useQps, "qps", "q", false, "-q, default false")
	toCmd.PersistentFlags().BoolVarP(&playGame, "playGame", "p", false, "--playGame, default false")
	toCmd.PersistentFlags().BoolVarP(&endless, "endless", "", false, "--endless, default false")
	toCmd.PersistentFlags().StringArrayVarP(&games, "games", "g", []string{"competitive_math"}, `-g, game names, games are: competitive_math (default), ravens_matrices, push_pull, minimum_effort_airport, minimum_effort_airport_target`)
	toCmd.PersistentFlags().BoolVarP(&storeTalentObjects, "storeTalentObjects", "s", false, "-s, default false")
	toCmd.MarkFlagRequired("filepath")

	toCmd.Example = "stress-test talent -f ~/Downloads/2W-user.json"

	// gameID = "competitive_math"
	// gameID = "ravens_matrices"

	rootCmd.AddCommand(toCmd)
}

var toCmd = &cobra.Command{
	Use:   "talent",
	Short: "Talent optimization test",
	Long:  `Talent optimization test`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if limit < 1 {
			log.Fatalf("limit <%v> must greater than 0\n", limit)
		}

		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			log.Fatalf("file not exits <%s>", filepath)
		}

		fileBuffer, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatalf("open file failed - %v\n", err)
		}

		var userList []*talent.TalentObject
		if strings.HasSuffix(filepath, ".talent.json") {
			json.Unmarshal(fileBuffer, &userList)
		} else if strings.HasSuffix(filepath, ".json") {
			var signInConfigs []talent.SignInConfig
			json.Unmarshal(fileBuffer, &signInConfigs)

			for i := 0; i < len(signInConfigs); i++ {
				userList = append(userList, &talent.TalentObject{SignInConfig: &signInConfigs[i]})
			}
		}

		userLength := len(userList)
		if userLength == 0 {
			log.Fatal("no user loaded")
		}

		fmt.Printf("loaded %v users \n", userLength)

		var httpClient *http.Client
		if ParseBool(keepAlive) {
			httpClient = NewHttpClientWithoutRedirect(true)
		}

		if debug {
			debugErr := executeSingleTask(userList[0], httpClient, nil)
			if debugErr != nil {
				log.Fatal(debugErr)
			}
		} else {
			if endless {
				for {
					executeStressTest(userList, httpClient)
				}
			} else {
				executeStressTest(userList, httpClient)
			}
		}

		if storeTalentObjects {
			userJsonData, err := json.Marshal(userList)
			if err != nil {
				log.Fatal(err)
			}

			talentObjFilepath := strings.Replace(filepath, ".json", ".talent.json", -1)
			_, err = os.Stat(talentObjFilepath)

			if err == nil {
				os.Remove(talentObjFilepath)
			} else if err != nil && !os.IsNotExist(err) {
				log.Fatal(err)
			}

			ioutil.WriteFile(talentObjFilepath, userJsonData, 0777)
		}
	},
}

func executeStressTest(userList []*talent.TalentObject, httpClient *http.Client) {
	s := client.NewStressClientWithConcurrentNumber(1, len(userList))

	var rateLimiter ratelimit.Limiter

	if limit > 0 {
		rateLimiter = ratelimit.New(limit)
	}

	var index uint32 = 0
	s.Header()
	if !useQps {
		s.RunSingleTaskWithRateLimiter("talent", rateLimiter, func() error {
			tmpIndex := atomic.AddUint32(&index, 1)
			user := userList[tmpIndex-1]

			debugErr := executeSingleTask(user, httpClient, nil)
			return debugErr
		})
	} else {
		s.RunMultiTasksWithRateLimiter("talent", rateLimiter, func(ch chan<- *runner.TaskResult) error {
			tmpIndex := atomic.AddUint32(&index, 1)
			user := userList[tmpIndex-1]

			debugErr := executeSingleTask(user, httpClient, ch)
			return debugErr
		})
	}
}

func executeSingleStep(i int, action string, talentObj *talent.TalentObject, ch chan<- *runner.TaskResult, handler func() error) (int, error) {
	if stage == 0 || stage > i {
		t1 := time.Now()
		err := handler()
		enqueueMetrics(action, &t1, err, ch)

		if err != nil {
			return i, err
		} else if debug {
			fmt.Printf("debug - %s success: %v\n", action, talentObj.String())
		}

		i++
	}

	return i, nil
}

func executeSingleTask(user *talent.TalentObject, httpClient *http.Client, ch chan<- *runner.TaskResult) (err error) {
	if httpClient == nil {
		httpClient = NewHttpClientWithoutRedirect(false)
	}

	talentObj := user //talent.NewTalentObject()

	if stage == -1 {
		t1 := time.Now()
		err = talentObj.Status(httpClient)
		enqueueMetrics("status", &t1, err, ch)

		if err != nil {
			return err
		} else if debug {
			fmt.Printf("debug - status success")
		}

		return
	}

	i := 0
	if talentObj.Cookie == nil {
		if i, err = executeSingleStep(i, "sign-in", talentObj, ch, func() error {
			return talentObj.SignIn(httpClient)
		}); err != nil {
			return
		}
	} else {
		i++
	}

	if talentObj.UserId == "" {
		if i, err = executeSingleStep(i, "information", talentObj, ch, func() error {
			return talentObj.Information(httpClient)
		}); err != nil {
			return
		}
	} else {
		i++
	}

	var currentIndex = i
	for _, game := range games {
		if i, err = executeSingleStep(i, "start-game", talentObj, ch, func() error {
			processDelay()
			return talentObj.StartGame(game, httpClient)
		}); err != nil {
			return
		}

		if playGame && (stage == 0 || stage > i) {
			err = talentObj.PlayGame(game)
			if err != nil {
				return err
			} else if debug {
				fmt.Println("debug - play game success")
			}

			i++
		}

		if _, err = executeSingleStep(i, "stop-game", talentObj, ch, func() error {
			processDelay()
			return talentObj.StopGame(game, httpClient)
		}); err != nil {
			return
		}

		if _, err = executeSingleStep(i, "get-game-status", talentObj, ch, func() error {
			return talentObj.GameStatus(game, httpClient)
		}); err != nil {
			return
		}

		if _, err = executeSingleStep(i, "summary", talentObj, ch, func() error {
			return talentObj.Summary(httpClient)
		}); err != nil {
			return
		}

		i = currentIndex
	}

	return
}

func processDelay() {
	if delay > 0 {
		sleeping := delay + rand.Intn(1000)
		time.Sleep(time.Duration(sleeping) * time.Millisecond)
	}
}

func enqueueMetrics(name string, startTime *time.Time, err error, ch chan<- *runner.TaskResult) {
	endTime := time.Now()
	d := endTime.Sub(*startTime).Nanoseconds()

	if ch != nil {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		ch <- &runner.TaskResult{
			Success:     err == nil,
			ProcessTime: uint64(d),
			StartTime:   uint64(startTime.UnixNano()),
			EndTime:     uint64(endTime.UnixNano()),
			Category:    name,
			Err:         errMsg,
		}
	}
}
