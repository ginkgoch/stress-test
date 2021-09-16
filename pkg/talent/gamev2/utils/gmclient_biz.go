package utils

import (
	"encoding/json"
	"errors"
	"log"
)

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

func (client *GameClient) ListenGameMessages() error {
	for {
		_, msg, err := client.WSClient.ReadMessage("read game msg")
		if err != nil {
			return err
		}

		if isHeartbeatMsg(msg) {
			go func() {
				log.Println("heartbeat triggered - enqueue")
				client.HeartbeatChan <- msg
			}()
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

func (client *GameClient) ListenGameMovesAsync() error {
	errChan := make(chan *error)
	for {
		_, msg, err := client.WSClient.ReadMessage("read game msg")
		if err != nil {
			return err
		}

		if isHeartbeatMsg(msg) {
			client.HeartbeatChan <- msg
		} else {
			client.GameMoveChan <- msg
		}

		select {
		case gameMsg := <-client.GameMoveChan:
			client.HandleGameMoves(gameMsg, errChan)
		case errMsg := <-errChan:
			return *errMsg
		}
	}
}

func (client *GameClient) HandleGameMoves(msg *ReceivedMsg, errChan chan<- *error) {
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
		errChan <- &err
	case "/gameroom":
		event := new(GameEvent)
		if err = json.Unmarshal(receiveMsg.Data, event); err != nil {
			log.Println("game room move json error", err)
			errChan <- &err
		}

		exit, err := client.handleGameRoomEvent(event, &receiveMsg)
		if err != nil {
			log.Println(err.Error())
			errChan <- &err
		}

		if exit {
			log.Print("game room event - session_ended, exit")
			errChan <- &err
		}
	case "/game":
		event := new(GameEvent)
		if err := json.Unmarshal(receiveMsg.Data, event); err != nil {
			log.Println("game room json error", err.Error())
		}

		exit, err := client.handleGameEvent(event, &receiveMsg)
		if err != nil {
			log.Println(err.Error())
			errChan <- &err
		}

		if exit {
			log.Println("game event - game_ended, exit")
			errChan <- nil
		}
	}
}
