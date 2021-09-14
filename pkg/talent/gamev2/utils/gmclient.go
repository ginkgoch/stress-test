package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
)

type GameConfig struct {
	ID          string `json:"id"`
	PlayerID    int    `json:"playerId"`
	RoomID      string `json:"roomId"`
	Server      string `json:"server"`
	GameURL     string `json:"gameurl"`
	PhoneNumber string `json:"phoneNumber"`
}

type GameClient struct {
	WSClient   *WSClient
	GameConfig *GameConfig
}

func NewGameClient(config *GameConfig) *GameClient {
	wsUrl := fmt.Sprintf("wss://%s/game-server/cometd", config.Server)
	return &GameClient{
		WSClient:   NewWebsocketClient(wsUrl),
		GameConfig: config,
	}
}

func (g *GameClient) Close() error {
	return g.WSClient.Close()
}

func (client *GameClient) PlayGame() (err error) {
	defer client.Close()

	log.Println("connecting")
	if err = client.WSClient.Connect(); err != nil {
		return
	}

	log.Println("handshake")
	if err = client.Handshake(); err != nil {
		return
	}

	log.Println("heatbeat")
	if err = client.Heartbeat(); err != nil {
		return
	}

	return
}

// step 2, handshake
func (client *GameClient) Handshake() (err error) {
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

	err = client.WSClient.WriteTextMessage(handshakeBody)
	if err != nil {
		log.Println("handshake sending failed")
		return
	}

	originMsg, handshakeMsg, err := client.WSClient.ReadMessage("handshake")
	if err != nil {
		return
	}

	err = ValidateChannel(handshakeMsg, "handshake", "/meta/handshake")
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

	client.WSClient.ClientId = handshakeMsgs[0].ClientID
	return
}

// step 3, heartbeat
func (client *GameClient) Heartbeat() (err error) {
	connectMsg := &HeartbeatMsg{
		ConnectionType: "websocket",
		Ext:            map[string]interface{}{"ack": 0},
		Channel:        "/meta/connect",
	}

	client.WSClient.MessageId++
	connectMsg.ID = strconv.Itoa(client.WSClient.MessageId)
	connectMsg.ClientID = client.WSClient.ClientId

	err = client.WSClient.WriteJSON(connectMsg)
	if err != nil {
		log.Println("heartbeat sending failed")
		return
	}

	originMsg, msg, err := client.WSClient.ReadMessage("connect")
	if err != nil {
		return
	}

	err = ValidateChannel(msg, "connect", "/meta/connect")
	if err != nil {
		return
	}

	var heartbeats []HeartbeatMsg
	err = json.Unmarshal(originMsg, &heartbeats)
	if err != nil {
		log.Println("connect json err:", err, heartbeats)
		return
	}

	client.WSClient.IncrementMessageId()
	heartbeatBody := heartbeats[0]
	heartbeatBody.ConnectionType = "websocket" //connectionType: 'websocket'
	heartbeatBody.ID = strconv.Itoa(client.WSClient.MessageId)
	err = client.WSClient.WriteJSON(heartbeatBody)
	if err != nil {
		log.Println("send heartbeat failed", err)
		return
	}

	return
}

// joinOrLeave: join|leave
func (client *GameClient) processGameJoinOrLeave(config *GameConfig, joinOrLeave string) (err error) {
	joinGame := JoinGameMsg{
		Action: joinOrLeave,
		Room:   config.RoomID,
		User:   config.PlayerID,
	}

	err = client.SendAction(joinGame, "/service/gameroom/"+config.RoomID)
	if err != nil {
		return
	}

	_, msg, err := client.WSClient.ReadMessage(joinOrLeave + "-game")
	if err != nil {
		return
	}

	err = ValidateChannel(msg, joinOrLeave+"-game", "/service/gameroom/"+config.RoomID)
	if err != nil {
		return
	}

	return err
}

// step 4, join game
func (client *GameClient) JoinGame(config *GameConfig) (err error) {
	err = client.processGameJoinOrLeave(config, "join")
	return
}

// step 5, leave game
func (client *GameClient) LeaveGame(config *GameConfig) (err error) {
	err = client.processGameJoinOrLeave(config, "leave")
	return
}

func (client *GameClient) SendAction(action interface{}, channel string) (err error) {
	actionMsg := ActionMsg{
		Data:    action,
		Channel: channel,
	}

	client.WSClient.IncrementMessageId()
	actionMsg.ID = strconv.Itoa(client.WSClient.MessageId)
	err = client.WSClient.WriteJSON(actionMsg)
	return
}

func (client *GameClient) AutoPlay() error {
	for {
		_, msg, err := client.WSClient.ReadMessage("game move")
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

			err = client.handleGameEvent(event, &receiveMsg)
			if err != nil {
				return err
			}
		}
	}
}

func (client *GameClient) handleGameEvent(event *GameEvent, receivedMsg *GameReceiveMsg) error {
	switch event.Event {
	case "UNAVAILABLE":
		return errors.New("game UNAVAILABLE")
	case "USER_JOINED":
		return client.handleGameJoinedEvent(receivedMsg)
	default:
		log.Println("unknown game event", event.Event)
		return nil
	}
}

func (client *GameClient) handleGameJoinedEvent(receivedMsg *GameReceiveMsg) error {
	joinedMsg := &GameJoinedMsg{}
	err := json.Unmarshal(receivedMsg.Data, joinedMsg)
	if err != nil {
		log.Println("game joined msg parse failed")
		return err
	}
	if !joinedMsg.Active {
		return errors.New("join game ,not active")
	}

	//TODO: user joined
	return nil
}
