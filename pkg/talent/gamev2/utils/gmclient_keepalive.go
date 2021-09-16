package utils

import (
	"fmt"
	"log"
	"strconv"
)

// step 3, heartbeat
func (client *GameClient) Heartbeat(ignoreAck bool) (err error) {
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

func isHeartbeatMsg(msg *ReceivedMsg) bool {
	return msg.Channel == "/meta/connect"
}

func (client *GameClient) KeepAlive() {
	for {
		log.Println("keepalive ack waiting")
		select {
		case msg := <-client.HeartbeatChan:
			log.Println("heartbeat triggered - dequeued")

			err := ValidateChannel(msg, "connect", "/meta/connect")
			if err != nil {
				log.Fatalln("heartbeat failed", err)
				return
			}

			client.Heartbeat(true)
		case exit := <-client.HeartbeatLeaveChan:
			if exit {
				log.Println("exit ensuring keep alive")
				return
			}
		}
	}
}
