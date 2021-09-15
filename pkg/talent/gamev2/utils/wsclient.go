package utils

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	Conn      *websocket.Conn
	ServerUrl string
	MessageId int
}

func NewWebsocketClient(serverUrl string) *WSClient {
	client := WSClient{
		ServerUrl: serverUrl,
	}

	return &client
}

func (wsc *WSClient) Close() error {
	if wsc.Conn != nil {
		return wsc.Conn.Close()
	}
	return nil
}

// step 1, connect
func (wsc *WSClient) Connect() (err error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsc.ServerUrl, nil)
	if err != nil {
		log.Println("connection failed")
		return
	}

	wsc.Conn = conn
	return
}

func (wsc *WSClient) ReadMessage(action string) (originMsg []byte, msg ReceivedMsg, err error) {
	_, originMsg, err = wsc.Conn.ReadMessage()

	if err != nil {
		log.Println(action + " reading failed")
		return
	}

	log.Println("read socket: ", string(originMsg))
	var receivedMsg []ReceivedMsg
	err = json.Unmarshal(originMsg, &receivedMsg)

	if err != nil {
		log.Println(action + " msg parsing failed")
		return
	}

	if len(receivedMsg) < 1 {
		errMsg := action + " msg format error"
		log.Println(errMsg)
		err = errors.New(errMsg)
		return
	}

	msg = receivedMsg[0]
	if msg.Error != "" {
		log.Println(action+" failed", msg.Error)
		err = errors.New(msg.Error)
		return
	}

	return
}

func (wsc *WSClient) WriteTextMessage(text string) error {
	return wsc.Conn.WriteMessage(websocket.TextMessage, []byte(text))
}

func (wsc *WSClient) WriteJSON(content interface{}) error {
	bytes, _ := json.Marshal([]interface{}{content})
	log.Println("write socket: ", string(bytes))
	return wsc.Conn.WriteJSON([]interface{}{content})
}

func (wsc *WSClient) IncrementMessageId() {
	wsc.MessageId++
}

func ValidateChannel(msg ReceivedMsg, action string, channel string) (err error) {
	if msg.Channel != channel {
		errMsg := action + " receive channel incorrect: " + msg.Channel
		log.Println(errMsg)
		err = errors.New(errMsg)
	}

	return
}
