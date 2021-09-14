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
	ID             string      `json:"id"`
	Channel        string      `json:"channel"`
	ClientID       string      `json:"clientId"`
	ConnectionType string      `json:"connectionType,omitempty"`
	Ext            interface{} `json:"ext"`
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
