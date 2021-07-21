package talent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ginkgoch/stress-test/pkg/templates"
)

const (
	DefaultServiceEndPoint = "https://talent.test.moblab-us.cn/api/1"
	// DefaultServiceEndPoint = "http://talent:3000/api/1"
	signInUrl      = "/zhilian/login"
	informationUrl = "/student/information?ignoreTrait=true"
	statusUrl      = "/status"
	startGameUrl   = "/startGame/%s/%s"
	finishGameUrl  = "/game/finish/%s"
)

func NewTalentObject(serviceEndpoint string) *TalentObject {
	if serviceEndpoint == "" {
		serviceEndpoint = DefaultServiceEndPoint
	}

	return &TalentObject{ServiceEndpoint: serviceEndpoint}
}

func (talent *TalentObject) Status(httpClient *http.Client) error {
	request, err := http.NewRequest("GET", talent.formalizeUrl(statusUrl), nil)
	if err != nil {
		return err
	}

	err = templates.HttpGet(request, httpClient)
	return err
}

func (talent *TalentObject) SignIn(user map[string]string, httpClient *http.Client) error {
	request, err := http.NewRequest("GET", talent.formalizeUrl(signInUrl), nil)
	if err != nil {
		return err
	}

	request.Header.Set("x-forwarded-proto", "https")

	query := request.URL.Query()
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

		gameConfig := new(GameConfig)
		gameConfig.Server = startGameData.Data.ServerAddress
		gameConfig.ID = startGameData.Data.ID
		gameConfig.PlayerID, err = strconv.Atoi(startGameData.Data.PlayerID)
		gameConfig.GameURL = startGameData.Data.Gameurl
		gameConfig.RoomID = startGameData.Data.RoomID
		talent.GameConfig = gameConfig
	}

	return err
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
	return
}

func (talent *TalentObject) formalizeUrl(url string) string {
	return fmt.Sprintf("%s%s", talent.ServiceEndpoint, url)
}
