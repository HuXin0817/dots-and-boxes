package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/HuXin0817/colog"
	"image/color"
	"os"
	"sync"
	"time"
)

// 定义棋盘UI的常量
var (
	DotDistance = float32(75) // 点之间的距离
	DotWidth    = float32(15) // 点的宽度
	DotMargin   = float32(50) // 棋盘边距

	EdgeFilledColor       = color.RGBA{R: 128, G: 128, B: 128, A: 128} // 边的颜色
	Player1HighLightColor = color.NRGBA{R: 30, G: 30, B: 255, A: 128}  // 玩家1的高亮颜色
	Player2HighLightColor = color.NRGBA{R: 255, G: 30, B: 30, A: 128}  // 玩家2的高亮颜色

	TipColor = color.NRGBA{R: 255, G: 255, B: 30, A: 50} // 提示颜色

	Player1FilledColor = color.NRGBA{R: 30, G: 30, B: 128, A: 128} // 玩家1填充颜色
	Player2FilledColor = color.NRGBA{R: 128, G: 30, B: 30, A: 128} // 玩家2填充颜色

	BoxSize = DotDistance - DotWidth // 方块的大小

	mainWindowSize = DotDistance*float32(BoardSize) + DotMargin // 主窗口大小
)

// 转换坐标
func transPosition(x int) float32 { return DotMargin + float32(x)*DotDistance }

// 获取点的坐标
func getDotPosition(d Dot) (float32, float32) { return transPosition(d.X()), transPosition(d.Y()) }

// 创建点的画布元素
func NewDotCanvas(d Dot) (newDotCanvas *canvas.Circle) {
	newDotCanvas = canvas.NewCircle(color.White)
	newDotCanvas.Resize(fyne.NewSize(DotWidth, DotWidth))
	newDotCanvas.Move(fyne.NewPos(getDotPosition(d)))
	return
}

// 创建边的画布元素
func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + DotWidth/2
	y1 := transPosition(e.Dot1().Y()) + DotWidth/2
	x2 := transPosition(e.Dot2().X()) + DotWidth/2
	y2 := transPosition(e.Dot2().Y()) + DotWidth/2

	newEdge := canvas.NewLine(EdgeFilledColor)
	newEdge.Position1 = fyne.NewPos(x1, y1)
	newEdge.Position2 = fyne.NewPos(x2, y2)
	newEdge.StrokeWidth = DotWidth
	return newEdge
}

// 创建方块的画布元素
func NewBox(s Box) *canvas.Rectangle {
	d := Dot(s)
	x := transPosition(d.X()) + DotWidth
	y := transPosition(d.Y()) + DotWidth

	r := canvas.NewRectangle(color.Black)
	r.Move(fyne.NewPos(x, y))
	r.Resize(fyne.NewSize(BoxSize, BoxSize))
	return r
}

// BoardUI 定义棋盘UI结构体
type BoardUI struct {
	board        Board
	aiPlayer1    bool
	aiPlayer2    bool
	player1Score int
	player2Score int
	nowTurn      Turn
	Container    *fyne.Container
	edges        map[Edge]*canvas.Line
	boxes        map[Box]*canvas.Rectangle
	buttons      map[Edge]*widget.Button
	signChan     chan struct{}
	mu           sync.Mutex
}

// 颜色插值函数
func interpolateColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8((float64(r1)*(1-t) + float64(r2)*t) / 256)
	g := uint8((float64(g1)*(1-t) + float64(g2)*t) / 256)
	b := uint8((float64(b1)*(1-t) + float64(b2)*t) / 256)
	a := uint8((float64(a1)*(1-t) + float64(a2)*t) / 256)

	return color.RGBA{R: r, G: g, B: b, A: a}
}

// 新建棋盘UI
func NewBoardUI() (board *BoardUI) {
	background := canvas.NewRectangle(color.Black)
	background.Move(fyne.NewPos(0, 0))
	background.Resize(fyne.NewSize(1e10, 1e10))

	board = &BoardUI{
		board:     make(Board),
		nowTurn:   Player1Turn,
		Container: container.NewWithoutLayout(background),
		edges:     make(map[Edge]*canvas.Line),
		boxes:     make(map[Box]*canvas.Rectangle),
		buttons:   make(map[Edge]*widget.Button),
		signChan:  make(chan struct{}, 1),
	}

	// 初始化方块UI
	for _, b := range Boxes {
		boxUi := NewBox(b)
		board.boxes[b] = boxUi
		board.Container.Add(boxUi)
	}

	// 初始化边UI和按钮
	for _, e := range EdgesMap {
		edgeUi := NewEdgeCanvas(e)
		board.edges[e] = edgeUi
		board.Container.Add(edgeUi)
		board.buttons[e] = widget.NewButton("", func() {
			switch {
			case board.aiPlayer1 && board.nowTurn == Player1Turn:
				return
			case board.aiPlayer2 && board.nowTurn == Player2Turn:
				return
			}
			board.AddEdge(e)
		})

		var SizeX, SizeY float32
		if e.Dot1().X() == e.Dot2().X() {
			SizeX = DotWidth
			SizeY = DotDistance
		} else {
			SizeX = DotDistance
			SizeY = DotWidth
		}

		board.buttons[e].Resize(fyne.NewSize(SizeX, SizeY))
		PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - board.buttons[e].Size().Width/2 + DotWidth/2
		PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - board.buttons[e].Size().Height/2 + DotWidth/2
		board.buttons[e].Move(fyne.NewPos(PosX, PosY))
		board.Container.Add(board.buttons[e])
	}

	// 初始化点UI
	for _, d := range Dots {
		dotUi := NewDotCanvas(d)
		board.Container.Add(dotUi)
	}

	// 启动AI处理协程
	go func() {
		for range board.signChan {
			e := GenerateBestEdge(board.board)
			board.AddEdge(e)
		}
	}()

	return
}

// 添加边
func (b *BoardUI) AddEdge(e Edge) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查边是否已存在
	if _, c := b.board[e]; c {
		return
	}

	defer b.Container.Refresh()

	// 根据当前玩家设置边的颜色
	if b.nowTurn == Player1Turn {
		b.edges[e].StrokeColor = Player1HighLightColor
	} else {
		b.edges[e].StrokeColor = Player2HighLightColor
	}

	// 计算得分
	score := b.board.ObtainsScore(e)
	if b.nowTurn == Player1Turn {
		b.player1Score += score
	} else {
		b.player2Score += score
	}

	// 获取获得的方块并填充颜色
	boxes := b.board.ObtainsBoxes(e)
	for _, box := range boxes {
		if b.nowTurn == Player1Turn {
			b.boxes[box].FillColor = Player1FilledColor
		} else {
			b.boxes[box].FillColor = Player2FilledColor
		}
	}

	// 记录操作日志
	if b.nowTurn == Player1Turn {
		colog.Infof("Player1 Edge: %s Player1 Score: %d, Player2 Score: %d", e.ToString(), b.player1Score, b.player2Score)
	}
	if b.nowTurn == Player2Turn {
		colog.Infof("Player2 Edge: %s Player1 Score: %d, Player2 Score: %d", e.ToString(), b.player1Score, b.player2Score)
	}

	// 隐藏已点击的按钮
	if button, c := b.buttons[e]; c {
		button.Hide()
	}

	// 如果没有得分，则切换玩家
	if score == 0 {
		b.nowTurn.Change()
	}

	// 更新棋盘状态
	b.board[e] = struct{}{}

	// 检查是否有接近完成的方块并提示
	for _, box := range Boxes {
		edgesCountInBox := b.board.EdgesCountInBox(box)
		if edgesCountInBox == 3 {
			nowStep := len(b.board)
			go func() {
				startColor := TipColor
				endColor := color.Black
				step := 100
				d := time.Second / time.Duration(step)

				for {
					for i := 0; i <= step; i++ {
						time.Sleep(d)
						t := float64(i) / float64(step)
						if len(b.board) != nowStep {
							return
						}
						b.boxes[box].FillColor = interpolateColor(startColor, endColor, t)
						b.boxes[box].Refresh()
					}

					for i := 0; i <= step; i++ {
						time.Sleep(d)
						t := float64(i) / float64(step)
						if len(b.board) != nowStep {
							return
						}
						b.boxes[box].FillColor = interpolateColor(endColor, startColor, t)
						b.boxes[box].Refresh()
					}
				}
			}()
		}
	}

	// 检查游戏是否结束
	if len(EdgesMap) == len(b.board) {
		switch {
		case b.player1Score > b.player2Score:
			colog.Info("Player1 Win! Score:", b.player1Score)
		case b.player2Score > b.player1Score:
			colog.Info("Player2 Win! Score:", b.player2Score)
		case b.player1Score == b.player2Score:
			colog.Infof("Draw!")
		}

		time.Sleep(time.Second * 3)
		os.Exit(0)
	}

	// 启动AI
	if b.aiPlayer1 && b.nowTurn == Player1Turn {
		b.signChan <- struct{}{}
	} else if b.aiPlayer2 && b.nowTurn == Player2Turn {
		b.signChan <- struct{}{}
	}
}
