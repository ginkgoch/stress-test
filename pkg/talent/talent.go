package talent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ginkgoch/stress-test/pkg/talent/game"
	"github.com/ginkgoch/stress-test/pkg/templates"
)

const (
	DefaultServiceEndpoint = "https://talent.test.moblab-us.cn/api/1"
	// DefaultServiceEndPoint = "http://talent:3000/api/1"
	signInUrl      = "/zhilian/login"
	informationUrl = "/student/information?ignoreTrait=true"
	statusUrl      = "/status"
	startGameUrl   = "/startGame/%s/%s"
	finishGameUrl  = "/game/finish/%s"
	statusGameUrl  = "/game/status/%s"
	summaryUrl     = "/summary?ignoreTrait=true"
)

var (
	ServiceEndpoint string
)

func NewTalentObject() *TalentObject {
	return new(TalentObject)
}

func (talent *TalentObject) Status(httpClient *http.Client) error {
	request, err := http.NewRequest("GET", talent.formalizeUrl(statusUrl), nil)
	if err != nil {
		return err
	}

	err = templates.HttpGet(request, httpClient)
	return err
}

func (talent *TalentObject) Summary(httpClient *http.Client) error {
	request, err := http.NewRequest("GET", talent.formalizeUrl(summaryUrl), nil)
	request.AddCookie(talent.Cookie)
	if err != nil {
		return err
	}

	res, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	infoData, err := templates.ConsumeResponse(res)
	if err != nil {
		return err
	}

	info := new(Summary)
	if err = json.Unmarshal(infoData, &info); err != nil {
		return err
	}
	if info.Success {
		fmt.Println("get Summary success")
		return nil
	} else {
		return fmt.Errorf("Get Summary failed, Info.Success=%v", info.Success)
	}
}

func (talent *TalentObject) SignIn(httpClient *http.Client) error {
	// if talent.Cookie != nil {
	// 	return nil
	// }

	request, err := http.NewRequest("GET", talent.formalizeUrl(signInUrl), nil)
	if err != nil {
		return err
	}

	request.Header.Set("x-forwarded-proto", "https")

	query := request.URL.Query()

	user := talent.SignInConfig.AsMap()
	for key := range user {
		query.Add(key, user[key])
	}
	query.Add("accessId", "111111")

	request.URL.RawQuery = query.Encode()
	res, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if _, err = templates.ConsumeResponse(res); err != nil {
		return err
	}

	for _, cookie := range res.Cookies() {
		if cookie.Name == "this.sid" {
			talent.Cookie = cookie
			break
		}
	}

	return nil
}

func (talent *TalentObject) Information(httpClient *http.Client) error {
	// if talent.UserId != "" {
	// 	return nil
	// }

	request, err := http.NewRequest("GET", talent.formalizeUrl(informationUrl), nil)
	request.AddCookie(talent.Cookie)
	if err != nil {
		return err
	}

	res, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	infoData, err := templates.ConsumeResponse(res)
	if err != nil {
		return err
	}

	info := new(Information)
	if err = json.Unmarshal(infoData, &info); err != nil {
		return err
	}

	talent.UserId = info.User.ID
	return nil
}

func (talent *TalentObject) StartGame(gameId string, httpClient *http.Client) error {
	relPath := fmt.Sprintf(startGameUrl, talent.UserId, gameId)

	request, err := http.NewRequest("GET", talent.formalizeUrl(relPath), nil)
	request.Header.Set("Content-Type", "application/json")

	if talent.Cookie != nil {
		request.AddCookie(talent.Cookie)
	}
	if err != nil {
		return err
	}

	res, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := templates.ConsumeResponse(res)
	if err != nil {
		return err
	} else {
		startGameData := new(StartGameData)
		err = json.Unmarshal(data, startGameData)
		if err != nil {
			return err
		}

		gameConfig := new(game.GameConfig)
		gameConfig.Server = startGameData.Data.Server
		gameConfig.ID = startGameData.Data.ID
		gameConfig.PlayerID, err = strconv.Atoi(startGameData.Data.PlayerID)
		gameConfig.GameURL = startGameData.Data.Gameurl
		gameConfig.RoomID = startGameData.Data.RoomID
		talent.GameConfig = gameConfig
		fmt.Println("get startGame success")
	}

	return err
}

func (talent *TalentObject) PlayGameOnly(gameId string, serverInfo string, playerId string, roomId string) (err error) {
	gameConfig := new(game.GameConfig)
	//if choose the playGame only, should set the prameter from cmd
	gameConfig.Server = serverInfo //startGameData.Data.Server
	gameConfig.ID = gameId         //startGameData.Data.ID
	gameConfig.PlayerID, err = strconv.Atoi(playerId)
	gameConfig.GameURL = "default"
	gameConfig.RoomID = roomId
	talent.GameConfig = gameConfig
	err = game.RunGame(talent.GameConfig)
	return
}

func (talent *TalentObject) StopGame(gameId string, httpClient *http.Client) (err error) {
	relPath := fmt.Sprintf(finishGameUrl, gameId)

	request, err := http.NewRequest("GET", talent.formalizeUrl(relPath), nil)
	request.Header.Set("Content-Type", "application/json")

	if talent.Cookie != nil {
		request.AddCookie(talent.Cookie)
	}
	if err != nil {
		return err
	}

	err = templates.HttpGet(request, httpClient)
	if err == nil {
		fmt.Println("get stopGame success")
	}
	return
}

func (talent *TalentObject) GameStatus(gameId string, httpClient *http.Client) (err error) {
	relPath := fmt.Sprintf(statusGameUrl, gameId)

	request, err := http.NewRequest("GET", talent.formalizeUrl(relPath), nil)
	request.Header.Set("Content-Type", "application/json")

	if talent.Cookie != nil {
		request.AddCookie(talent.Cookie)
	}
	if err != nil {
		return err
	}

	err = templates.HttpGet(request, httpClient)
	if err == nil {
		fmt.Println("get gameStatus success")
	}
	return
}

func (talent *TalentObject) PlayGame(gameId string) (err error) {
	// lib.InitWorkPools(10)
	// lib.RunDelayWorkPool()
	err = game.RunGame(talent.GameConfig)
	return
}

func (talent *TalentObject) formalizeUrl(url string) string {
	return fmt.Sprintf("%s%s", ServiceEndpoint, url)
}
