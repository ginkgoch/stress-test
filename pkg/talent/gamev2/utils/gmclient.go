package utils

import (
	"fmt"
	"log"
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

func (g *GameClient) PlayGame() (err error) {
	defer g.Close()

	log.Println("connecting")
	if err = g.WSClient.Connect(); err != nil {
		return
	}

	log.Println("handshake")
	if err = g.WSClient.Handshake(); err != nil {
		return
	}

	log.Println("heatbeat")
	if err = g.WSClient.Heartbeat(); err != nil {
		return
	}

	return
}
