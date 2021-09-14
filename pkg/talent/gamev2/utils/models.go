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

type Heartbeat struct {
	ID             string      `json:"id"`
	Channel        string      `json:"channel"`
	ClientID       string      `json:"clientId"`
	ConnectionType string      `json:"connectionType,omitempty"`
	Ext            interface{} `json:"ext"`
}
