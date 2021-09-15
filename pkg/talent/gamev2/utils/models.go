package utils

type ReceivedMsg struct {
	Channel string      `json:"channel"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Message struct {
	ID      string      `json:"id"`
	Channel string      `json:"channel"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HandshakeMsg struct {
	Message
	ClientID                 string                 `json:"clientId"`
	MinimumVersion           string                 `json:"minimumVersion"`
	SupportedConnectionTypes []string               `json:"supportedConnectionTypes"`
	Ext                      map[string]interface{} `json:"ext"`
	Version                  string                 `json:"version"`
	Successful               bool                   `json:"successful"`
}

type HeartbeatMsg struct {
	ID             string `json:"id"`
	Channel        string `json:"channel"`
	ClientID       string `json:"clientId"`
	ConnectionType string `json:"connectionType,omitempty"`
	// Ext            interface{} `json:"ext"`
}

type JoinGameMsg struct {
	Action string `json:"action"`
	Room   string `json:"room"`
	User   int    `json:"user"`
}

type ActionMsg struct {
	ID       string      `json:"id"`
	Channel  string      `json:"channel"`
	Data     interface{} `json:"data"`
	ClientID string      `json:"clientId"`
}

type GameReceiveMsg struct {
	Channel string
	Data    []byte
}

type GameEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type GameJoinedMsg struct {
	Active bool   `json:"active"`
	Event  string `json:"event"`
	Room   string `json:"room"`
}

type GameSessionEndedMsg struct {
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

type GamePlayerUpdatedMsg struct {
	Earning      int      `json:"earning"`
	Index        int      `json:"index"`
	Moves        []string `json:"moves"`
	Period       int      `json:"period"`
	PlayerNumber int      `json:"playerNumber"`
}

type GamePlayerAction struct {
	Action string      `json:"action"`
	Player int         `json:"player"`
	Data   interface{} `json:"data"`
}

type GamePlayer interface {
	UserJoined(*GameClient, *GameJoinedMsg)
	SessionEnded(*GameClient, *GameSessionEndedMsg)
	GameStarted(*GameClient, *GameStartedMsg)
	PlayerUpdated(*GameClient, *GamePlayerUpdatedMsg)
	GameRoundStarted(*GameClient, *GameRoundMsg)
	GameRoundEnded(*GameClient, *GameRoundMsg)
}
