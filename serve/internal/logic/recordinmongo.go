package logic

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"serve/serve"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/message/moverecord"
)

func (l *PostGameInformationLogic) RecordInMongo(in *serve.GameInformationRequest) (err error) {
	ctx := l.ctx
	MongoUrl := l.svcCtx.Config.MongoConf.Url
	MongoDataBaseName := l.svcCtx.Config.MongoConf.DataBaseName

	if in.GameOver {
		var Winner string

		if in.Team1Score > in.Team2Score {
			Winner = chess.Player1.String()
		} else if in.Team1Score < in.Team2Score {
			Winner = chess.Player2.String()
		} else {
			Winner = "Draw"
		}

		recode := &moverecord.GameEndRecode{
			Winner: Winner,
		}

		mongoConn := moverecord.NewGameEndRecodeModel(MongoUrl, MongoDataBaseName, in.GameUid)
		if err = mongoConn.Insert(ctx, recode); err != nil {
			return err
		}

		return nil
	}

	if len(in.Edges) == 0 {
		recode := &moverecord.GameStartRecode{
			GameUid:   message.GameUid(in.GameUid),
			BoardSize: int(in.BoardSize),
			AI1:       in.AI1,
			AI2:       in.AI2,
		}

		mongoConn := moverecord.NewGameStartRecodeModel(MongoUrl, MongoDataBaseName, in.GameUid)
		if err = mongoConn.Insert(ctx, recode); err != nil {
			return err
		}

		return nil
	}

	recode := &moverecord.MoveRecode{
		StepCount:    int(in.StepCount),
		Player1Score: int(in.Team1Score),
		Player2Score: int(in.Team2Score),
		NowPlayer:    in.NowTurn,
		MoveEdge:     chess.Edge(in.MoveEdge).String(),
	}

	mongoConn := moverecord.NewMoveRecodeModel(MongoUrl, MongoDataBaseName, in.GameUid)
	if err = mongoConn.Insert(ctx, recode); err != nil {
		return err
	}

	return nil
}
