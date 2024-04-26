package moverecord

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameEndRecode struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`

	Winner string
}
