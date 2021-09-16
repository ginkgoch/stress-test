package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

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
