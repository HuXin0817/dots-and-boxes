package logic

import (
	"context"
	"math"
	"serve/internal/svc"
	"serve/serve"

	"github.com/HuXin0817/dots-and-boxes/pkg/assess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/zeromicro/go-zero/core/logx"
)

type InquireBestEdgeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInquireBestEdgeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InquireBestEdgeLogic {
	return &InquireBestEdgeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func GetBestEdge(in *serve.InquireBestEdgeRequest, members []string) (BestEdge chess.Edge) {
	var edges []chess.Edge
	for e := range in.Edges {
		edges = append(edges, chess.Edge(e))
	}

	board := chess.NewBoard(int(in.BoardSize), edges...)
	BestScore := -assess.INF
	MinScore := assess.INF
	for _, m := range members {
		v := message.NewAssessMessageValue(m)
		if v.Score > BestScore {
			BestScore = v.Score
			BestEdge = v.Edge
		}
		MinScore = math.Min(MinScore, BestScore)
	}

	if MinScore == BestScore {
		return assess.RandEdgeInBetterEdges(board)
	}

	return
}

func (l *InquireBestEdgeLogic) InquireBestEdge(in *serve.InquireBestEdgeRequest) (*serve.InquireBestEdgeResponse, error) {
	if in.BoardSize > 10 {
		return nil, BoardSizeOutOfRangeErr
	}

	key := message.AssessMessageKey{
		GameUid: message.GameUid(in.GameUid),
		Step:    int(in.Step),
	}

	members, err := l.svcCtx.RedisClient.Smembers(key.String())
	if err != nil {
		return nil, err
	}

	if in.WaitingTime != 0 && in.WaitingTime%60 == 0 {
		var edges []chess.Edge
		for e := range in.Edges {
			edges = append(edges, chess.Edge(e))
		}

		b := chess.NewBoard(int(in.BoardSize), edges...)

		nextEdges := assess.BetterEdges(b)
		edgeMap := make(map[chess.Edge]struct{})
		for _, e := range nextEdges {
			edgeMap[e] = struct{}{}
		}

		for _, m := range members {
			v := message.NewAssessMessageValue(m)
			delete(edgeMap, v.Edge)
		}

		var edgeMessages []chess.Edge
		for e := range edgeMap {
			edgeMessages = append(edgeMessages, e)
		}

		SendMessageToRedisLists(l.svcCtx.RedisClient, l.svcCtx.PartitionPusher, message.GameUid(in.GameUid), int(in.StepCount), b, edgeMessages...)
	}

	BestEdge := GetBestEdge(in, members)
	return &serve.InquireBestEdgeResponse{
		NowBestEdge:      int64(BestEdge),
		CalculatedNumber: int64(len(members)),
	}, nil
}
