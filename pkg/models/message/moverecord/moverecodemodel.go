package moverecord

import "github.com/zeromicro/go-zero/core/stores/mon"

// const MoveRecodeCollectionName = "move_recode"

var _ MoveRecodeModel = (*customMoveRecodeModel)(nil)

type (
	// MoveRecodeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMoveRecodeModel.
	MoveRecodeModel interface {
		moveRecodeModel
	}

	customMoveRecodeModel struct {
		*defaultMoveRecodeModel
	}
)

// NewMoveRecodeModel returns a model for the mongo.
func NewMoveRecodeModel(url, db, MoveRecodeCollectionName string) MoveRecodeModel {
	conn := mon.MustNewModel(url, db, MoveRecodeCollectionName)
	return &customMoveRecodeModel{
		defaultMoveRecodeModel: newDefaultMoveRecodeModel(conn),
	}
}
