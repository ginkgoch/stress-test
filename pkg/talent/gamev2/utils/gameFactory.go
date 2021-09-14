package utils

func NewGamePlayer(gameId string) GamePlayer {
	switch gameId {
	case "backpack":
		return NewBackPack()
	}

	return nil
}
