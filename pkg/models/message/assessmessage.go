package message

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"github.com/bytedance/sonic"
)

type AssessMessageKey struct {
	GameUid
	Step int
}

func (a AssessMessageKey) String() string {
	str, _ := sonic.MarshalString(a)
	return str
}

type AssessMessageValue struct {
	Edge  chess.Edge
	Score float64
}

func NewAssessMessageValue(s string) (newAssessMessageValue AssessMessageValue) {
	_ = sonic.UnmarshalString(s, &newAssessMessageValue)
	return newAssessMessageValue
}

func (a AssessMessageValue) String() string {
	str, _ := sonic.MarshalString(a)
	return str
}

type MovingHasBeenAssessedKey struct {
	GameUid
	Step int
	chess.Edge
}

func (m *MovingHasBeenAssessedKey) String() string {
	s, _ := sonic.MarshalString(m)
	return s
}
