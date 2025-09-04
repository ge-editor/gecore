package screen

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	UPDATE_CONTENTS = 1 << iota
	MOVE_CURSOR
)

// 描画エリアを表す構造体
type DrawArea struct {
	ID       int                       // エリアの識別子
	dirty    uint32                    // フラグ管理
	cancel   context.CancelFunc        // 描画キャンセル用
	mu       sync.Mutex                // 排他制御
	drawChan chan struct{}             // 描画トリガー
	render   func(ctx context.Context) // 描画処理
}

// 描画リクエスト
func (da *DrawArea) RequestDraw(flag uint32) {
	da.mu.Lock()
	da.dirty |= flag
	da.mu.Unlock()

	select {
	case da.drawChan <- struct{}{}: // 描画トリガー送信
	default: // すでにトリガー済みの場合はスキップ
	}
}

// 描画ループ
func (da *DrawArea) startDrawLoop() {
	for range da.drawChan {
		da.mu.Lock()
		if da.cancel != nil {
			da.cancel() // 古い描画をキャンセル
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		da.cancel = cancel
		dirty := da.dirty
		da.dirty = 0 // フラグをリセット
		da.mu.Unlock()

		// 描画処理の実行
		if dirty != 0 {
			da.render(ctx)
		}
	}
}

// Renderer 全体を管理する構造体
type Renderer struct {
	areas  []*DrawArea  // 描画エリア
	screen tcell.Screen // tcell のスクリーン
}

// 新しい描画エリアを追加
func (r *Renderer) AddArea(id int, render func(ctx context.Context)) {
	area := &DrawArea{
		ID:       id,
		drawChan: make(chan struct{}, 1),
		render:   render,
	}
	go area.startDrawLoop() // 描画ループを開始
	r.areas = append(r.areas, area)
}

// 描画エリアを取得
func (r *Renderer) GetArea(id int) *DrawArea {
	for _, area := range r.areas {
		if area.ID == id {
			return area
		}
	}
	return nil
}

// キー入力の処理
func (r *Renderer) HandleInput(ev *tcell.EventKey) {
	switch ev.Rune() {
	case '1':
		area := r.GetArea(1)
		if area != nil {
			area.RequestDraw(UPDATE_CONTENTS)
		}
	case '2':
		area := r.GetArea(2)
		if area != nil {
			area.RequestDraw(UPDATE_CONTENTS)
		}
	case 'q':
		r.screen.Fini()
		os.Exit(0)
	}
}

func main() {
	// tcell のスクリーンを初期化
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create screen: %v\n", err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize screen: %v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	renderer := &Renderer{screen: screen}

	// 描画エリアを追加
	renderer.AddArea(1, func(ctx context.Context) {
		// エリア 1 の描画処理
		screen.Clear()
		st := tcell.StyleDefault.Foreground(tcell.ColorGreen)
		screen.SetContent(2, 2, '1', nil, st)
		screen.Show()
		fmt.Println("Area 1 Drawn")
	})

	renderer.AddArea(2, func(ctx context.Context) {
		// エリア 2 の描画処理
		screen.Clear()
		st := tcell.StyleDefault.Foreground(tcell.ColorBlue)
		screen.SetContent(10, 10, '2', nil, st)
		screen.Show()
		fmt.Println("Area 2 Drawn")
	})

	// イベントループ
	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				renderer.HandleInput(ev)
			}
		}
	}()

	// メインスレッドを維持
	select {}
}
