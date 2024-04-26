package moverecord

import "github.com/zeromicro/go-zero/core/stores/mon"

// const GameStartRecodeCollectionName = "game_start_recode"

var _ GameStartRecodeModel = (*customGameStartRecodeModel)(nil)

type (
	// GameStartRecodeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameStartRecodeModel.
	GameStartRecodeModel interface {
		gameStartRecodeModel
	}

	customGameStartRecodeModel struct {
		*defaultGameStartRecodeModel
	}
)

// NewGameStartRecodeModel returns a model for the mongo.
func NewGameStartRecodeModel(url, db, GameStartRecodeCollectionName string) GameStartRecodeModel {
	conn := mon.MustNewModel(url, db, GameStartRecodeCollectionName)
	return &customGameStartRecodeModel{
		defaultGameStartRecodeModel: newDefaultGameStartRecodeModel(conn),
	}
}
