package logic

import (
	"context"
	"strconv"

	"github.com/HuXin0817/dots-and-boxes/pkg/assess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"

	"serve/internal/svc"
	"serve/serve"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/zeromicro/go-zero/core/logx"
)

type PostGameInformationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostGameInformationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostGameInformationLogic {
	return &PostGameInformationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PostGameInformationLogic) PostGameInformation(in *serve.GameInformationRequest) (*serve.GameInformationRequestResponse, error) {
	if in.BoardSize > 10 {
		return nil, BoardSizeOutOfRangeErr
	}

	if err := l.RecordInMongo(in); err != nil {
		return nil, err
	}

	if in.GameOver {
		if _, err := l.svcCtx.RedisClient.Del(in.GameUid); err != nil {
			return nil, err
		}
		return &serve.GameInformationRequestResponse{}, nil
	}

	if err := l.svcCtx.RedisClient.Setex(in.GameUid, strconv.FormatInt(in.StepCount, 10), 120); err != nil {
		return nil, err
	}

	if in.NowTurn == chess.Player1.String() && !in.AI1 {
		return &serve.GameInformationRequestResponse{}, nil
	}

	if in.NowTurn == chess.Player2.String() && !in.AI2 {
		return &serve.GameInformationRequestResponse{}, nil
	}

	var edges []chess.Edge
	for e := range in.Edges {
		edges = append(edges, chess.Edge(e))
	}

	b := chess.NewBoard(int(in.BoardSize), edges...)
	edgeMessages := assess.BetterEdges(b)
	if err := SendMessageToRedisLists(l.svcCtx.RedisClient, l.svcCtx.PartitionPusher, message.GameUid(in.GameUid), int(in.StepCount), b, edgeMessages...); err != nil {
		return nil, err
	}

	return &serve.GameInformationRequestResponse{
		TotalCalNumber: int64(len(edgeMessages)),
	}, nil
}
