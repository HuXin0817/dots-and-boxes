package moverecord

import "github.com/zeromicro/go-zero/core/stores/mon"

// const GameEndRecodeCollectionName = "game_end_recode"

var _ GameEndRecodeModel = (*customGameEndRecodeModel)(nil)

type (
	// GameEndRecodeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameEndRecodeModel.
	GameEndRecodeModel interface {
		gameEndRecodeModel
	}

	customGameEndRecodeModel struct {
		*defaultGameEndRecodeModel
	}
)

// NewGameEndRecodeModel returns a model for the mongo.
func NewGameEndRecodeModel(url, db, GameEndRecodeCollectionName string) GameEndRecodeModel {
	conn := mon.MustNewModel(url, db, GameEndRecodeCollectionName)
	return &customGameEndRecodeModel{
		defaultGameEndRecodeModel: newDefaultGameEndRecodeModel(conn),
	}
}
