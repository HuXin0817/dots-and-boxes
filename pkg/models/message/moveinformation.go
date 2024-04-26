package message

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"github.com/bytedance/sonic"
)

type MovingInformationMessage struct {
	TimeStamp
	GameUid
	StepCount int
	chess.Board
	MoveEdge chess.Edge
}

func NewMovingInformationMessage(str string) (newMovingInformationMessage MovingInformationMessage, err error) {
	err = sonic.UnmarshalString(str, &newMovingInformationMessage)
	return
}

func (m MovingInformationMessage) String() string {
	str, _ := sonic.MarshalString(m)
	return str
}
