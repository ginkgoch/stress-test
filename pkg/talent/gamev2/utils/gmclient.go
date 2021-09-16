package utils

import (
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
	WSClient           *WSClient
	GameConfig         *GameConfig
	ClientId           string
	Round              int
	Player             GamePlayer
	SessionId          string
	HeartbeatChan      chan *ReceivedMsg
	HeartbeatLeaveChan chan bool
	GameMoveChan       chan *ReceivedMsg
}

func NewGameClient(config *GameConfig) *GameClient {
	wsUrl := fmt.Sprintf("wss://%s/game-server/cometd", config.Server)

	player := NewGamePlayer(config.ID)
	if player == nil {
		log.Println("unknown game", config.ID)
	}

	return &GameClient{
		WSClient:           NewWebsocketClient(wsUrl),
		GameConfig:         config,
		Player:             player,
		HeartbeatChan:      make(chan *ReceivedMsg),
		HeartbeatLeaveChan: make(chan bool),
		GameMoveChan:       make(chan *ReceivedMsg),
	}
}

func (g *GameClient) Close() error {
	g.HeartbeatLeaveChan <- true
	time.Sleep(2 * time.Second)
	return g.WSClient.Close()
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

func (client *GameClient) Run() (err error) {
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
	if err = client.Heartbeat(false); err != nil {
		return
	}

	log.Println("gamejoin")
	if err = client.JoinGame(); err != nil {
		return
	}

	log.Println("heatbeat")
	if err = client.Heartbeat(true); err != nil {
		return
	}

	go client.KeepAlive()

	log.Println("playgame")
	if err = client.ListenGameMessages(); err != nil {
		return
	}

	log.Println("leavegame")
	if err = client.LeaveGame(); err != nil {
		return
	}

	return
}
