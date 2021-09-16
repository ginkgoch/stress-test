package talent

import (
	"fmt"
	"net/http"

	"github.com/ginkgoch/stress-test/pkg/talent/game"
)

type TalentObject struct {
	UserId       string
	Cookie       *http.Cookie
	GameConfig   *game.GameConfig
	SignInConfig *SignInConfig
}

type SignInConfig struct {
	PhoneNumber   string `json:"phoneNumber"`
	Name          string `json:"name"`
	UserId        string `json:"userId"`
	OrgName       string `json:"orgName"`
	OrgCode       string `json:"orgCode"`
	MajorName     string `json:"majorName"`
	UserType      string `json:"userType"`
	SchoolName    string `json:"schoolName"`
	Authorization string `json:"authorization"`
}

func (c *SignInConfig) AsMap() map[string]string {
	confMap := make(map[string]string)
	confMap["phoneNumber"] = c.PhoneNumber
	confMap["name"] = c.Name
	confMap["userId"] = c.UserId
	confMap["orgName"] = c.OrgName
	confMap["orgCode"] = c.OrgCode
	confMap["majorName"] = c.MajorName
	confMap["userType"] = c.UserType
	confMap["schoolName"] = c.SchoolName
	confMap["authorization"] = c.Authorization

	return confMap
}

func (t *TalentObject) String() string {
	s := fmt.Sprintln()
	s += fmt.Sprintln("UserId:", t.UserId)
	s += fmt.Sprintln("Cookie:", t.Cookie)
	s += fmt.Sprintln("GameConfig:", t.GameConfig)
	return s
}

type Information struct {
	Success bool `json:"success"`
	User    struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phoneNumber"`
		ID          string `json:"id"`
	} `json:"user"`
}

type Summary struct {
	Data struct {
		Progress struct {
			Activities map[string]interface{} `json:"activities"`
			Profiles   map[string]interface{} `json:"profiles"`
			World      interface{}            `json:"world"`
		} `json:"progress"`
		Report struct {
			Category  map[string]interface{} `json:"category"`
			Report_Id string                 `json:"report_id"`
		} `json:"report"`
	} `json:"data"`
	Success bool `json:"success"`
}

// type GameConfig struct {
// 	ID       string `json:"id"`
// 	PlayerID int    `json:"playerId"`
// 	RoomID   string `json:"roomId"`
// 	Server   string `json:"server"`
// 	GameURL  string `json:"gameurl"`
// }

type StartGameData struct {
	Success bool `json:"success"`
	Data    struct {
		ID       string `json:"id"`
		PlayerID string `json:"playerId"`
		RoomID   string `json:"roomId"`
		Server   string `json:"serverAddress"`
		Gameurl  string `json:"gameurl"`
	} `json:"data"`
}
