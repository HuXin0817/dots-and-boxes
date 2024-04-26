package message

import "github.com/google/uuid"

type GameUid string

func NewGameUid() GameUid {
	return GameUid(uuid.New().String())
}
