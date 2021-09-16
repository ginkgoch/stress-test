package utils

import (
	"encoding/json"
	"errors"
	"log"
)

// joinOrLeave: join|leave
func (client *GameClient) processGameJoinOrLeave(config *GameConfig, joinOrLeave string, requireAck bool) (err error) {
	joinGame := JoinGameMsg{
		Action: joinOrLeave,
		Room:   config.RoomID,
		User:   config.PlayerID,
	}

	err = client.SendAction(joinGame, "/service/gameroom/"+config.RoomID)
	if err != nil {
		return
	}

	if !requireAck {
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
	err = client.processGameJoinOrLeave(client.GameConfig, "join", true)
	return
}

// step 5, leave game
func (client *GameClient) LeaveGame() (err error) {
	err = client.processGameJoinOrLeave(client.GameConfig, "leave", false)
	return
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
