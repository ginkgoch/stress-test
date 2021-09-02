package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ginkgoch/stress-test/pkg/log"
	"github.com/ginkgoch/stress-test/pkg/talent/lib"
)

// The GameClient type represents a game WebSocket connection and game data.
type GameClient struct {
	userID     int
	roomID     string
	gamePlayer GamePlayer
	Round      int
	gameID     string
	wsClient   *WebsocketClient
	stopWatch  *lib.StopWatch
	Delay      int
	WorkPools  *lib.DelayWorkPools
}

////NewGameClient  new a GameClient
//func NewGameClient(userID int, roomID string, serverURL string, player GamePlayer, phoneNumber string) *GameClient {
//	stopwatch := lib.NewStopWatch(phoneNumber + ":" + strconv.Itoa(userID))
//	gc := GameClient{
//		userID:     userID,
//		roomID:     roomID,
//		gamePlayer: player,
//		wsClient:   NewWebsocketClient(serverURL, userID),
//		stopWatch:  &stopwatch,
//	}
//	return &gc
//}

//NewGameClient  new a GameClient
func NewGameClient(config *GameConfig, player GamePlayer) *GameClient {
	stopwatch := lib.NewStopWatch(config.PhoneNumber + ":" + strconv.Itoa(config.PlayerID))

	wsUrl := fmt.Sprintf("wss://%s/game-server/cometd", config.WebSocketHost())
	log.Println("ws server", wsUrl)

	gc := GameClient{
		userID:     config.PlayerID,
		roomID:     config.RoomID,
		gamePlayer: player,
		wsClient:   NewWebsocketClient(wsUrl, config.PlayerID),
		stopWatch:  &stopwatch,
		WorkPools:  &lib.DelayWorkPools{DelayWorkChan: []chan *lib.DelayWork{}},
	}

	gc.WorkPools.InitWorkPools(10)
	gc.WorkPools.RunDelayWorkPool()
	return &gc
}

//Run game
func (gameClient *GameClient) Run() (err error) {
	defer gameClient.close()
	gameClient.stopWatch.Start("Run", strconv.Itoa(gameClient.userID))
	defer func() {
		gameClient.stopWatch.End("Run", fmt.Sprintf("%d %v", gameClient.userID, err))
	}()
	gameClient.stopWatch.Start("connect_wsclient", "")
	err = gameClient.wsClient.Connect()
	gameClient.stopWatch.End("connect_wsclient", fmt.Sprintf("%v", err))
	if err != nil {
		return
	}
	gameClient.joinGame()
	gameClient.stopWatch.Start("startJoin", "")

	time.Sleep(10 * time.Second)
	err = gameClient.handleMessage()
	gameClient.stopWatch.Log("DelayTime:", strconv.Itoa(gameClient.Delay))
	if err != nil {
		return err
	} else {
		gameClient.leaveGame()
	}
	return nil

}

func (g *GameClient) handleMessage() error {
	for receiveMsg := range g.wsClient.ReceivedMsgChan {
		log.Println("channel received:", receiveMsg.Channel)
		switch receiveMsg.Channel {
		case "error":
			return errors.New(string(receiveMsg.Data))
		case "/gameroom": //1
			event := Event{}
			err := json.Unmarshal(receiveMsg.Data, &event)
			if err != nil {
				return err
			}
			switch event.Event {
			case UNAVAILABLE:
				return errors.New("game UNAVAILABLE")
			case USER_JOINED:
				g.stopWatch.End("startJoin", USER_JOINED)
				joinedMsg := &JoinedMsg{}
				err = json.Unmarshal(receiveMsg.Data, joinedMsg)
				if err != nil {
					return err
				}
				if !joinedMsg.Active {
					return errors.New("join game ,not active")
				}
				g.stopWatch.Start(USER_JOINED, "")
				g.gamePlayer.UserJoined(g, joinedMsg)
			case SESSION_ENDED:
				g.stopWatch.End(USER_JOINED, SESSION_ENDED)
				g.stopWatch.End(GAME_ENDED, SESSION_ENDED)
				joinedMsg := &SessionEndedMsg{}
				err = json.Unmarshal(receiveMsg.Data, joinedMsg)
				if err != nil {
					return err
				}
				g.gamePlayer.SessionEnded(g, joinedMsg)
				return nil
			default:
				g.stopWatch.Start("/gameroom unhandled event: ", event.Event)
			}
		case "/game": //2
			event := Event{}
			if err := json.Unmarshal(receiveMsg.Data, &event); err != nil {
				g.stopWatch.Log("json unmarshal error", err.Error())
			}

			eventData, err := json.Marshal(event.Data)
			if err != nil {
				g.stopWatch.Log("json marshal error", err.Error())
			}

			switch event.Event {
			case GAME_STARTED: // game start
				g.stopWatch.Start(GAME_STARTED, "")
				g.stopWatch.Start(GAME_ROUND_STARTED, "FAKE")
				msg := &GameStartedMsg{}
				if err = json.Unmarshal(eventData, msg); err != nil {
					g.stopWatch.Log("json Unmarshal error", err.Error())
					return err
				}

				if msg.Status != "RUNNING" {
					g.stopWatch.Log("game status", msg.Status)
					return nil
				}
				g.gameID = msg.GameID
				g.Round = msg.Round
				g.gamePlayer.GameStated(g, msg)
			case PLAYER_UPDATED: // client response
				g.playerUpdated(eventData)

			case GAME_ROUND_STARTED: // round period start
				g.stopWatch.Start(GAME_ROUND_STARTED, "")
				g.stopWatch.End(GAME_ROUND_ENDED, GAME_ROUND_STARTED+":"+strconv.Itoa(g.Round))
				msg := &GameRoundMsg{}
				err = json.Unmarshal(eventData, msg)
				if err != nil {
					g.stopWatch.Log("json Unmarshal error", err.Error())
					return err
				}
				g.Round = msg.Round
				g.gamePlayer.GameRoundStarted(g, msg)
			case GAME_ROUND_ENDED: // round period end
				g.stopWatch.End(GAME_ROUND_STARTED, GAME_ROUND_ENDED+":"+strconv.Itoa(g.Round))
				g.stopWatch.Start(GAME_ROUND_ENDED, "")
				msg := &GameRoundMsg{}
				err = json.Unmarshal(eventData, msg)
				if err != nil {
					g.stopWatch.Log("json Unmarshal error", err.Error())
					return err
				}
				g.Round = msg.Round
				g.gamePlayer.GameRoundEnded(g, msg)
			case GAME_ENDED: // game end
				g.stopWatch.End(GAME_STARTED, GAME_ENDED)
				g.stopWatch.Start(GAME_ENDED, "")
				// return nil
			default:
				g.stopWatch.Log("/game unhandled event: ", event.Event)
			}
		default:
			g.stopWatch.Log("unhandled_channel", receiveMsg.Channel)
		}
	}
	return nil
}

func (g *GameClient) playerUpdated(eventData []byte) {
	playerUpdated := &PlayerUpdated{}
	err := json.Unmarshal(eventData, playerUpdated)
	if err != nil {
		g.stopWatch.Log("error", err.Error())
		return
	}
	g.stopWatch.Log("PLAYER_UPDATED_moves", strings.Join(playerUpdated.Moves, ","))

	g.stopWatch.Start(PLAYER_UPDATED, "round:"+strconv.Itoa(g.Round))
	g.gamePlayer.PlayerUpdated(g, eventData)
}

func (g *GameClient) close() {
	g.wsClient.close()
}

//SendActionDelay send action,  delay second
func (g *GameClient) SendActionDelay(action Action, round int, delay int) {
	g.Delay += delay
	g.WorkPools.SendWork(func() {
		g.SendAction(action, round)
	}, delay)
}

//SendAction send action, use go SendAction
func (g *GameClient) SendAction(action Action, round int) {
	if g.Round != round { // need lock?  g.Round
		return
	}
	channel := "/service/game/" + g.gameID
	g.stopWatch.Start(action.Action, "")
	g.wsClient.SendAction(action, channel)
}

func (g *GameClient) joinGame() {
	joinGame := JoinGameSend{
		Action: "join",
		Room:   g.roomID,
		User:   g.userID,
	}
	g.wsClient.SendAction(joinGame, "/service/gameroom/"+g.roomID)
}

func (g *GameClient) leaveGame() {
	log.Println("leaving game room")
	joinGame := JoinGameSend{
		Action: "leave",
		Room:   g.roomID,
		User:   g.userID,
	}
	g.wsClient.SendAction(joinGame, "/service/gameroom/"+g.roomID)
}

type GameID string

const (
	RM    GameID = "ravens_matrices"
	CM    GameID = "competitive_math"
	PP    GameID = "push_pull"
	AIR   GameID = "minimum_effort_airport"
	AIR_T GameID = "minimum_effort_airport_target"
	BR    GameID = "bomb_risk"
	BP    GameID = "backpack"
	//a="minimum_effort_airport"
	//minimum_effort_airport_target
	//HF              GameID = "hearts_flowers"
	//CB              GameID = "Corsiblocks"
	//AC              GameID = "attention_control"
	//NUMERACY        GameID = "numeracy"
	//GRIT_ASSESSMENT GameID = "grit_assessment"
)

func RunGame(gameConf *GameConfig) (err error) {
	var player GamePlayer
	switch GameID(gameConf.ID) {
	case RM:
		player = NewRevensMatrices(5)
	case CM:
		player = NewCompetitiveMath(5)
	case PP:
		player = NewPushPull()
	case AIR:
		player = NewAirport()
	case AIR_T:
		player = NewAirport()
	case BR:
		player = NewBombRisk()
	case BP:
		player = NewBackPack()
	default:
		err = fmt.Errorf("player %d, no such game:%s ", gameConf.PlayerID, gameConf.ID)
		return
	}
	gameClient := *NewGameClient(gameConf, player)
	err = gameClient.Run()
	return
}
