package moverecord

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameStartRecode struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`

	GameUid   message.GameUid
	BoardSize int
	AI1       bool
	AI2       bool
}
