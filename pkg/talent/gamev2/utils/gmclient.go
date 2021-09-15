package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
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
	ClientId   string
	Round      int
	Player     GamePlayer
	SessionId  string
}

func NewGameClient(config *GameConfig) *GameClient {
	wsUrl := fmt.Sprintf("wss://%s/game-server/cometd", config.Server)

	player := NewGamePlayer(config.ID)
	if player == nil {
		log.Println("unknown game", config.ID)
	}

	return &GameClient{
		WSClient:   NewWebsocketClient(wsUrl),
		GameConfig: config,
		Player:     player,
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
	client.WSClient.MessageId++

	log.Println("heatbeat")
	if err = client.HeartbeatOnce(false); err != nil {
		return
	}

	log.Println("gamejoin")
	if err = client.JoinGame(); err != nil {
		return
	}

	log.Println("heatbeat")
	if err = client.HeartbeatOnce(true); err != nil {
		return
	}

	log.Println("playgame")
	if err = client.AutoPlay(); err != nil {
		return
	}

	log.Println("leavegame")
	if err = client.LeaveGame(); err != nil {
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

	client.ClientId = handshakeMsgs[0].ClientID
	return
}

func (client *GameClient) HeartbeatOnce(ignoreAck bool) (err error) {
	connectMsg := &HeartbeatMsg{
		ConnectionType: "websocket",
		// Ext:            map[string]interface{}{"ack": 0},
		Channel: "/meta/connect",
	}

	client.WSClient.MessageId++
	connectMsg.ID = strconv.Itoa(client.WSClient.MessageId)
	connectMsg.ClientID = client.ClientId
	err = client.WSClient.WriteJSON(connectMsg)
	if err != nil {
		err = fmt.Errorf("heartbeat sending failed: <%v>", err.Error())
		return
	}

	if ignoreAck {
		return
	}

	_, msg, err := client.WSClient.ReadMessage("connect")
	if err != nil {
		return
	}

	err = ValidateChannel(msg, "connect", "/meta/connect")
	if err != nil {
		return
	}

	return
}

// step 3, heartbeat
func (client *GameClient) Heartbeat() (err error) {
	connectMsg := &HeartbeatMsg{
		ConnectionType: "websocket",
		// Ext:            map[string]interface{}{"ack": 0},
		Channel: "/meta/connect",
	}

	client.WSClient.MessageId++
	connectMsg.ID = strconv.Itoa(client.WSClient.MessageId)
	connectMsg.ClientID = client.ClientId
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
	heartbeatMsg := heartbeats[0]
	heartbeatMsg.ConnectionType = "websocket" //connectionType: 'websocket'
	heartbeatMsg.ID = strconv.Itoa(client.WSClient.MessageId)
	err = client.WSClient.WriteJSON(heartbeatMsg)
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
func (client *GameClient) JoinGame() (err error) {
	err = client.processGameJoinOrLeave(client.GameConfig, "join")
	return
}

// step 5, leave game
func (client *GameClient) LeaveGame() (err error) {
	err = client.processGameJoinOrLeave(client.GameConfig, "leave")
	return
}

func (client *GameClient) SendAction(action interface{}, channel string) (err error) {
	actionMsg := ActionMsg{
		Data:    action,
		Channel: channel,
	}

	client.WSClient.IncrementMessageId()
	actionMsg.ID = strconv.Itoa(client.WSClient.MessageId)
	actionMsg.ClientID = client.ClientId
	err = client.WSClient.WriteJSON(actionMsg)
	return
}

func (client *GameClient) AutoPlay() error {
	for {
		time.Sleep(1 * time.Second)

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
			if err = json.Unmarshal(receiveMsg.Data, event); err != nil {
				log.Println("game room move json error", err)
				return err
			}

			exit, err := client.handleGameRoomEvent(event, &receiveMsg)
			if err != nil {
				log.Println(err.Error())
				return err
			}

			if exit {
				log.Print("game room event - session_ended, exit")
				return nil
			}
		case "/game":
			event := new(GameEvent)
			if err := json.Unmarshal(receiveMsg.Data, event); err != nil {
				log.Println("game room json error", err.Error())
			}

			exit, err := client.handleGameEvent(event, &receiveMsg)
			if err != nil {
				log.Println(err.Error())
				return err
			}

			if exit {
				log.Println("game event - game_ended, exit")
				return nil
			}
		}
	}
}

func (client *GameClient) handleGameRoomEvent(event *GameEvent, receivedMsg *GameReceiveMsg) (exit bool, err error) {
	log.Println("handle game room event -", event.Event)
	switch event.Event {
	case "UNAVAILABLE":
		err = errors.New("game UNAVAILABLE")
	case "USER_JOINED":
		err = client.handleGameJoinedEvent(receivedMsg)
	case "SESSION_ENDED":
		exit, err = client.handleSessionEndedEvent(receivedMsg)
	default:
		log.Println("unknown game event", event.Event)
	}

	return
}

func (client *GameClient) handleGameJoinedEvent(receivedMsg *GameReceiveMsg) error {
	joinedMsg := &GameJoinedMsg{}
	err := json.Unmarshal(receivedMsg.Data, joinedMsg)
	if err != nil {
		log.Println("game joined msg parse failed")
		return err
	}
	if !joinedMsg.Active {
		return errors.New("join game, not active")
	}

	client.Player.UserJoined(client, joinedMsg)
	return nil
}

func (client *GameClient) handleSessionEndedEvent(receivedMsg *GameReceiveMsg) (exit bool, err error) {
	joinedMsg := &GameSessionEndedMsg{}
	err = json.Unmarshal(receivedMsg.Data, joinedMsg)
	if err != nil {
		return
	}

	client.Player.SessionEnded(client, joinedMsg)
	exit = true
	return
}

func (client *GameClient) handleGameEvent(event *GameEvent, receivedMsg *GameReceiveMsg) (exit bool, err error) {
	log.Println("handle game event -", event.Event)

	eventData, err := json.Marshal(event.Data)
	if err != nil {
		log.Println("game event marshal error")
		return
	}

	switch event.Event {
	case "GAME_STARTED":
		if err = client.handleGameStartedEvent(eventData); err != nil {
			return
		}
	case "PLAYER_UPDATED":
		if err = client.handleGamePlayerUpdated(eventData); err != nil {
			return
		}
	case "GAME_ROUND_STARTED":
		if err = client.handleGameRoundEvent(eventData, event.Event, client.Player.GameRoundStarted); err != nil {
			return
		}
	case "GAME_ROUND_ENDED":
		if err = client.handleGameRoundEvent(eventData, event.Event, client.Player.GameRoundEnded); err != nil {
			return
		}
	case "GAME_ENDED":
		log.Println("game event - game_ended, exit")
		exit = true
	default:
		log.Println("unknown game event", event.Event)
	}

	return
}

func (client *GameClient) handleGameStartedEvent(eventData []byte) (err error) {
	msg := new(GameStartedMsg)
	if err = json.Unmarshal(eventData, msg); err != nil {
		log.Println("game started event json error", err.Error())
		return
	}

	if msg.Status != "RUNNING" {
		err = fmt.Errorf("game status error, expected RUNNING but got <%s>", msg.Status)
		log.Println(err.Error())
		return
	}

	log.Println("round -", msg.Round)
	client.Round = msg.Round
	client.SessionId = msg.GameID
	client.Player.GameStarted(client, msg)
	return
}

func (client *GameClient) handleGamePlayerUpdated(eventData []byte) (err error) {
	msg := new(GamePlayerUpdatedMsg)
	err = json.Unmarshal(eventData, msg)
	if err != nil {
		err = fmt.Errorf("game event - player_updated json unmarshal error <%v>", err.Error())
		return
	}

	client.HeartbeatOnce(true)
	time.Sleep(1 * time.Second)
	client.Player.PlayerUpdated(client, msg)
	return
}

func (client *GameClient) handleGameRoundEvent(eventData []byte, eventName string, playerCallback func(*GameClient, *GameRoundMsg)) (err error) {
	msg := new(GameRoundMsg)
	if err = json.Unmarshal(eventData, msg); err != nil {
		err = fmt.Errorf("game event - %s json unmarshal failed %v", eventName, err.Error())
		return
	}

	log.Println("round - ", msg.Round)
	client.Round = msg.Round
	playerCallback(client, msg)
	return
}
