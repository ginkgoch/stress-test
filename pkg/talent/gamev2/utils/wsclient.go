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
	connectMsg := &HeartbeatMsg{
		ConnectionType: "websocket",
		Ext:            map[string]interface{}{"ack": 0},
		Channel:        "/meta/connect",
	}

	wsc.MessageId++
	connectMsg.ID = strconv.Itoa(wsc.MessageId)
	connectMsg.ClientID = wsc.ClientId

	err = wsc.Conn.WriteJSON([]interface{}{connectMsg})
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

	var heartbeats []HeartbeatMsg
	err = json.Unmarshal(originMsg, &heartbeats)
	if err != nil {
		log.Println("connect json err:", err, heartbeats)
		return
	}

	wsc.MessageId++
	heartbeatBody := heartbeats[0]
	heartbeatBody.ConnectionType = "websocket" //connectionType: 'websocket'
	heartbeatBody.ID = strconv.Itoa(wsc.MessageId)
	err = wsc.Conn.WriteJSON([]interface{}{heartbeatBody})
	if err != nil {
		log.Println("send heartbeat failed", err)
		return
	}

	return
}

// joinOrLeave: join|leave
func (wsc *WSClient) processGameJoinOrLeave(config *GameConfig, joinOrLeave string) (err error) {
	joinGame := JoinGameMsg{
		Action: joinOrLeave,
		Room:   config.RoomID,
		User:   config.PlayerID,
	}

	err = wsc.SendAction(joinGame, "/service/gameroom/"+config.RoomID)
	if err != nil {
		return
	}

	_, msg, err := wsc.readMessage(joinOrLeave + "-game")
	if err != nil {
		return
	}

	err = validateChannel(msg, joinOrLeave+"-game", "/service/gameroom/"+config.RoomID)
	if err != nil {
		return
	}

	return err
}

// step 4, join game
func (wsc *WSClient) JoinGame(config *GameConfig) (err error) {
	err = wsc.processGameJoinOrLeave(config, "join")
	return
}

// step 5, leave game
func (wsc *WSClient) LeaveGame(config *GameConfig) (err error) {
	err = wsc.processGameJoinOrLeave(config, "leave")
	return
}

func (wsc *WSClient) SendAction(action interface{}, channel string) (err error) {
	actionMsg := ActionMsg{
		Data:    action,
		Channel: channel,
	}

	wsc.MessageId++
	actionMsg.ID = strconv.Itoa(wsc.MessageId)
	err = wsc.Conn.WriteJSON([]interface{}{actionMsg})
	return
}

func (wsc *WSClient) AutoPlay() error {
	for {
		_, msg, err := wsc.readMessage("game move")
		if err != nil {
			return err
		}

		data, err := json.Marshal(msg.Data)
		if err != nil {
			log.Println("game move msg format error")
		}

		receiveMsg := GameReceiveMsg{
			Channel: msg.Channel,
			Data:    data,
		}

		switch msg.Channel {
		case "error":
			err = errors.New(string(data))
			return err
		case "/gameroom":
			event := new(GameEvent)
			err = json.Unmarshal(receiveMsg.Data, event)
			if err != nil {
				log.Println("game room move json error", err)
				return err
			}

			err = wsc.handleGameEvent(event, &receiveMsg)
			if err != nil {
				return err
			}
		}
	}
}

func (wsc *WSClient) handleGameEvent(event *GameEvent, receivedMsg *GameReceiveMsg) error {
	switch event.Event {
	case "UNAVAILABLE":
		return errors.New("game UNAVAILABLE")
	case "USER_JOINED":
		return wsc.handleGameJoinedEvent(receivedMsg)
	default:
		log.Println("unknown game event", event.Event)
		return nil
	}
}

func (wsc *WSClient) handleGameJoinedEvent(receivedMsg *GameReceiveMsg) error {
	joinedMsg := &GameJoinedMsg{}
	err := json.Unmarshal(receivedMsg.Data, joinedMsg)
	if err != nil {
		log.Println("game joined msg parse failed")
		return err
	}
	if !joinedMsg.Active {
		return errors.New("join game ,not active")
	}

	return nil
}
