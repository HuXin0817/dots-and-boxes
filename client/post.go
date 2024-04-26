package main

import (
	"context"
	"fmt"
	"log"
	"serve/serveclient"
	"time"

	"github.com/HuXin0817/dots-and-boxes/pkg/assess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/model"
)

func addEdgeLogic() {
	postGameInformationRequest := GetPostGameInformationRequest()
	postGameInformationResponse, err := ServeClient.PostGameInformation(context.Background(), postGameInformationRequest)
	if err != nil {
		log.Panicln(err)
	}

	log.Printf("=> {GameInformation %v, Step: %d}\n", postGameInformationRequest.GameUid, postGameInformationRequest.StepCount)

	if Window.GameInformation.Board.FreeEdgesCount() == 0 {
		return
	}

	if Window.GameInformation.NowPlayer == chess.Player1 && !AI1 {
		return
	}

	if Window.GameInformation.NowPlayer == chess.Player2 && !AI2 {
		return
	}

	time.Sleep(time.Second / 2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	totalCalNumber := int(postGameInformationResponse.TotalCalNumber)
	bar := model.NewBar(totalCalNumber, "Start Analyzing...")
	bar.Add(1)
	defer bar.Close()

	var BestEdge chess.Edge

FindingBestEdgeLogic:
	for waitingTime := 0; ; waitingTime++ {
		select {
		case <-ctx.Done():
			break FindingBestEdgeLogic
		default:
			timer := time.NewTimer(time.Second)

			edges := make(map[int64]bool)
			for e := range Window.GameInformation.Board.Edges {
				edges[int64(e)] = true
			}

			r := &serveclient.InquireBestEdgeRequest{
				GameUid:     string(GameUid),
				Step:        int64(len(Window.GameInformation.Board.Edges)),
				BoardSize:   int64(Window.GameInformation.BoardSize),
				Edges:       edges,
				StepCount:   int64(len(Window.GameInformation.Edges)),
				WaitingTime: int64(waitingTime),
			}

			resp, err := ServeClient.InquireBestEdge(ctx, r)
			if err != nil {
				log.Println(err)
			}

			bar.Goto(int(resp.CalculatedNumber) + 1)

			BestEdge = chess.Edge(resp.NowBestEdge)
			if resp.CalculatedNumber >= int64(totalCalNumber) {
				break FindingBestEdgeLogic
			}

			<-timer.C
		}
	}

	if BestEdge == 0 {
		BestEdge = assess.RandEdgeInBetterEdges(Window.GameInformation.Board)
	}

	fmt.Println()
	Window.AddEdge(BestEdge)
}
