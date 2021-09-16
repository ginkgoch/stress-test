package game

const (
	USER_JOINED        = "USER_JOINED"
	UNAVAILABLE        = "UNAVAILABLE"
	SESSION_ENDED      = "SESSION_ENDED"
	GAME_STARTED       = "GAME_STARTED"
	GAME_ENDED         = "GAME_ENDED"
	GAME_ROUND_STARTED = "GAME_ROUND_STARTED"
	GAME_ROUND_ENDED   = "GAME_ROUND_ENDED"
	PLAYER_UPDATED     = "PLAYER_UPDATED"
)

//USER_JOINED ="USER_JOINED"
// /gameroom  USER_JOINED,UNAVAILABLE,SESSION_ENDED
// /game       GAME_STARTED,GAME_ENDED, GAME_ROUND_STARTED,GAME_ROUND_ENDED ,PLAYER_UPDATED

// Message represents general message
type Message struct {
	ID      string      `json:"id"`
	Channel string      `json:"channel"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GameMessage message
type GameMessage interface {
	SetMessage(messageid string, clientID string)
}

// Heartbeat represents Heartbeat message

//SetMessage set
func (hb *Heartbeat) SetMessage(messageid string, clientID string) {
	hb.ID = messageid
	hb.Channel = "/meta/connect"
	hb.ClientID = clientID
}

// HandshakeMsg represents handshake message
type HandshakeMsg struct {
	Message
	ClientID                 string                 `json:"clientId"`
	MinimumVersion           string                 `json:"minimumVersion"`
	SupportedConnectionTypes []string               `json:"supportedConnectionTypes"`
	Ext                      map[string]interface{} `json:"ext"`
	Version                  string                 `json:"version"`
	Successful               bool                   `json:"successful"`
}

//Player
// {
// 	"playerNumber": 0,
// 	"payoff": 115,
// 	"ranking": 1,
// 	"group": 0
//   }
type Player struct {
	PlayerNumber int    `json:"playerNumber"`
	UserID       int    `json:"userId"`
	Payoff       int    `json:"payoff"`
	Ranking      int    `json:"ranking"`
	Group        int    `json:"group"`
	PlayerName   string `json:"playerName"`
	Role         string `json:"role"`
}

// "player": {
// 	"playerNumber": 0,
// 	"userId": 274,
// 	"role": "player",
// 	"attributes": {
// 	  "strategy": "no_brain",
// 	  "earning": 0
// 	},
// 	"playerName": "22 ",
// 	"playPending": false,
// 	"recoverMode": false,
// 	"payoff": 0
//   }

//ReceivedMsg  event data
// type ReceivedMsg struct {
// 	Channel string `json:"channel"`
// 	Error   string `json:"error,omitempty"`
// 	Data    struct {
// 		Event string `json:"event"`
// 	} `json:"data,omitempty"`
// }
type ReceivedMsg struct {
	Channel    string      `json:"channel"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Id         *string     `json:"id"`
	Successful *bool       `json:"successful"`
}

//Event event type
type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

//JoginedMsg event USER_JOINED
// {
// 	"data": {
// 	  "active": true,
// 	  "event": "USER_JOINED",
// 	  "room": "b94695c8-d737-5592-b1b8-58760dbc45b3"
// 	},
// 	"channel": "/gameroom"
//   }
type JoinedMsg struct {
	Active bool   `json:"active"`
	Event  string `json:"event"`
	Room   string `json:"room"`
}

type LeaveMsg struct {
	Channel    string `json:"channel"`
	Id         string `json:"id"`
	Successful bool   `json:"successful"`
}

type UnsubscribeSend struct {
	Channel      string `json:"channel"`
	Id           string `json:"id"`
	ClientId     string `json:"clientId"`
	Subscription string `json:"subscription"`
}

//SessionEndedMsg   event SESSION_ENDED
// {
// 	"data": {
// 	  "game": "attention_control",
// 	  "sessionId": "5219a683-32f8-5b32-b68a-dd1254ccb6b8",
// 	  "event": "SESSION_ENDED",
// 	  "top_ranked": {
// 		"player": [
// 		  {
// 			"playerNumber": 0,
// 			"payoff": 115,
// 			"group": 0
// 		  }
// 		]
// 	  },
// 	  "player": {
// 		"playerNumber": 0,
// 		"payoff": 115,
// 		"ranking": 1,
// 		"group": 0
// 	  }
// 	},
// 	"channel": "/gameroom"
//   }
type SessionEndedMsg struct {
	Game      string `json:"game"`
	SessionID string `json:"sessionId"`
	Event     string `json:"event"`
	TopRanked struct {
		Challenger []struct {
			PlayerNumber int `json:"playerNumber"`
			Payoff       int `json:"payoff"`
			Group        int `json:"group"`
		} `json:"challenger"`
	} `json:"top_ranked"`
	Player struct {
		PlayerNumber int `json:"playerNumber"`
		Payoff       int `json:"payoff"`
		Ranking      int `json:"ranking"`
		Group        int `json:"group"`
	} `json:"player"`
}

// {
// "data":{"event":"UNAVAILABLE"},
// "channel":"/gameroom"
// }

//GameStartedMsg GAME_STARTED event
// {
// 	"gameId": "5219a683-32f8-5b32-b68a-dd1254ccb6b8/0",
// 	"period_interval": 10,
// 	"data":{}
// 	"playerState" {}
// 	"event": "GAME_STARTED",
// 	"type": "attention_control",
// 	"player": {}
// "status": "RUNNING",
// "round" : 0
// }
type GameStartedMsg struct {
	NumberOfPeriod   int    `json:"number_of_period"`
	RoundType        string `json:"round_type"`
	RoundDuration    int    `json:"round_duration"`
	RoundInterval    int    `json:"round_interval"`
	SessionStartTime int64  `json:"session_start_time"`
	ShowFeedback     bool   `json:"show_feedback"`
	Questions        int    `json:"questions"`
	Type             string `json:"type"`
	RoundEnd         int    `json:"roundEnd"`
	Payoffs          []int  `json:"payoffs"`
	TimeLeft         int    `json:"time_left"`
	RoundStart       int64  `json:"roundStart"`
	GameID           string `json:"gameId"`
	ShowTimer        bool   `json:"show_timer"`
	Period           int    `json:"period"`
	NumberOfRound    int    `json:"number_of_round"`
	EndTime          int64  `json:"end_time"`
	PeriodInterval   int    `json:"period_interval"`
	ResponsePayoff   int    `json:"response_payoff"`
	StartTime        int64  `json:"start_time"`
	AllowChat        bool   `json:"allow_chat"`
	GroupSize        int    `json:"group_size"`
	Round            int    `json:"round"`
	RealTimeLeft     int    `json:"real_time_left"`
	SolvePayoff      int    `json:"solve_payoff"`
	ProblemType      string `json:"problem_type"`
	Status           string `json:"status"`
	Desc             string `json:"desc"`
	Problems         int    `json:"problems"`
}

//GameEndedMsg GAME_ENDED event
// {
// 	"event": "GAME_ENDED"
// 	"gameId": "5219a683-32f8-5b32-b68a-dd1254ccb6b8/0",
//   "data" {}
// }
type GameEndedMsg struct {
	GameID string `json:"gameId"`
	Event  string `json:"event"`
	Data   struct {
		Status string `json:"status"`
	} `json:"data"`
}

//GameRoundMsg GAME_ROUND_STARTED GAME_ROUND_ENDED event
//{
// 	"event": "GAME_ROUND_STARTED"
// 	"data" : {
// 		"round": 1
// 		"status": "RUNNING"
// 	}
// }
type GameRoundMsg struct {
	RoundEnd     int64  `json:"roundEnd"`
	Period       int    `json:"period"`
	TimeLeft     int    `json:"time_left"`
	Round        int    `json:"round"`
	RealTimeLeft int    `json:"real_time_left"`
	RoundType    string `json:"round_type"`
	RoundStart   int64  `json:"roundStart"`
	EndTime      int64  `json:"end_time"`
	Status       string `json:"status"`
}

// GAME_ROUND_STARTED,GAME_ROUND_ENDED ,PLAYER_UPDATED

//SetMessage SetMessage on dataSend
func (ds *DataSend) SetMessage(messageid string, clientID string) {
	ds.ID = messageid
	ds.ClientID = clientID
}

//JoinGameSend send join game
// {
// 	"id": "3",
// 	"channel": "/service/gameroom/b94695c8-d737-5592-b1b8-58760dbc45b3",
// 	"data": {
// 	  "action": "join",
// 	  "room": "b94695c8-d737-5592-b1b8-58760dbc45b3",
// 	  "user": 274
// 	},
// 	"clientId": "11dmyf5kqmol9cll9krsglkxydh",
// 	"ext": {}
//   }
type JoinGameSend struct {
	Action string `json:"action"`
	Room   string `json:"room"`
	User   int    `json:"user"`
}

//PlayerUpdated ss
type PlayerUpdated struct {
	Earning      int      `json:"earning"`
	Index        int      `json:"index"`
	Moves        []string `json:"moves"`
	Period       int      `json:"period"`
	PlayerNumber int      `json:"playerNumber"`
}

// Action type
// {
// 	"action": "SOLVE",
// 	"player": 0,
// 	"data": [
// 		0,
// 		2,
// 		4
// 	]
// }
type Action struct {
	Action string      `json:"action"`
	Player int         `json:"player"`
	Data   interface{} `json:"data"`
}

// GameConfig start game response body
type GameConfig struct {
	ID          string `json:"id"`
	PlayerID    int    `json:"playerId"`
	RoomID      string `json:"roomId"`
	Server      string `json:"server"`
	GameURL     string `json:"gameurl"`
	PhoneNumber string `json:"phoneNumber"`
}

func (conf *GameConfig) WebSocketHost() (wsHost string) {
	// segs := strings.Split(conf.Server, "/")
	// if len(segs) != 3 {
	// 	log.Fatal("invalid game server id", conf.Server)
	// }

	// wsHost = segs[1] + ":8080"
	wsHost = conf.Server
	return
}
