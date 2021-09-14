package utils

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	Conn      *websocket.Conn
	ServerUrl string
	ClientId  string
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

func (wsc *WSClient) readMessage(action string) (originMsg []byte, msg ReceivedMsg, err error) {
	_, originMsg, err = wsc.Conn.ReadMessage()

	if err != nil {
		log.Println(action + " reading failed")
		return
	}

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

func validateChannel(msg ReceivedMsg, action string, channel string) (err error) {
	if msg.Channel != channel {
		errMsg := action + " receive channel incorrect: " + msg.Channel
		log.Println(errMsg)
		err = errors.New(errMsg)
	}

	return
}

// step 2, handshake
func (wsc *WSClient) Handshake() (err error) {
	var handshakeBody = `[
		{
			"id":"1",
			"version":"1.0",
			"minimumVersion":"1.0",
			"channel":"/meta/handshake",
			"supportedConnectionTypes":[
				"websocket",
				"long-polling",
				"callback-polling"
			],
			"advice":{
				"timeout":60000,
				"interval":0
			},
			"ext":{
				"ack":true
			}
		}
	]`

	err = wsc.Conn.WriteMessage(websocket.TextMessage, []byte(handshakeBody))
	if err != nil {
		log.Println("handshake sending failed")
		return
	}

	originMsg, handshakeMsg, err := wsc.readMessage("handshake")
	if err != nil {
		return
	}

	err = validateChannel(handshakeMsg, "handshake", "/meta/handshake")
	if err != nil {
		return
	}

	var handshakeMsgs []HandshakeMsg
	err = json.Unmarshal(originMsg, &handshakeMsgs)
	if err != nil {
		log.Println("handshakeMsg json err:", err, handshakeMsgs)
		return
	}

	if len(handshakeMsgs) < 1 {
		errMsg := "handshakeMsg json item count err"
		log.Println(errMsg)
		err = errors.New(errMsg)
		return
	}

	wsc.ClientId = handshakeMsgs[0].ClientID
	return
}

// step 3, heartbeat
func (wsc *WSClient) Heartbeat() (err error) {
	connectBody := &Heartbeat{
		ConnectionType: "websocket",
		Ext:            map[string]interface{}{"ack": 0},
		Channel:        "/meta/connect",
	}

	wsc.MessageId++
	connectBody.ID = strconv.Itoa(wsc.MessageId)
	connectBody.ClientID = wsc.ClientId

	err = wsc.Conn.WriteJSON([]interface{}{connectBody})
	if err != nil {
		log.Println("heartbeat sending failed")
		return
	}

	originMsg, msg, err := wsc.readMessage("connect")
	if err != nil {
		return
	}

	err = validateChannel(msg, "connect", "/meta/connect")
	if err != nil {
		return
	}

	var heartbeats []Heartbeat
	err = json.Unmarshal(originMsg, &heartbeats)
	if err != nil {
		log.Println("connect json err:", err, heartbeats)
		return
	}

	heartbeatBody := heartbeats[0]
	heartbeatBody.ConnectionType = "websocket" //connectionType: 'websocket'
	err = wsc.Conn.WriteJSON([]interface{}{heartbeatBody})
	if err != nil {
		log.Println("send heartbeat failed", err)
		return
	}

	return
}
