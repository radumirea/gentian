package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/outlay"
	"github.com/radumirea/gentian/painter"
	"github.com/radumirea/gentian/romutil"
)

const (
	TILES_SCALE = 2.0
	LEVEL_SCALE = 2.0
)

var clickableTiles = make([]widget.Clickable, 240)
var clickableLevelTiles = make([]widget.Clickable, 240)
var highlightBorder *image.RGBA

var loloData []byte
var levelsUncompressed = make([][]byte, 60)
var levelAmount int
var levelIndex int
var selectedTile int

var tiles = make([]image.Image, 10)

func main() {
	highlightBorder = image.NewRGBA(image.Rect(0, 0, 16, 16))
	painter.PaintBorder(highlightBorder, color.RGBA{255, 0, 0, 255})

	go func() {
		gameType, loloData := romutil.LoadROM("../data/lolo2.nes")
		_, levelAmount, levelsUncompressed = romutil.LoadAllLevels(gameType, loloData)
		tiles = romutil.ExtractTileTextures(true, true, false, gameType, loloData)
		w := app.NewWindow()
		err := run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	var ops op.Ops
	counter := 0
	th := material.NewTheme(gofont.Collection())
	var prevButton widget.Clickable
	var nextButton widget.Clickable
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			fmt.Println(counter)
			counter++
			gtx := layout.NewContext(&ops, e)
			if prevButton.Clicked() && levelIndex > 0 {
				levelIndex--
			}
			if nextButton.Clicked() && levelIndex < levelAmount-1 {
				levelIndex++
			}

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(th, &prevButton, "<<").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(th, &nextButton, ">>").Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(
							func(gtx layout.Context) layout.Dimensions {
								return (&outlay.Grid{}).Layout(gtx, 15, 16, cellDimensionerLevel, cellFuncLevel)
							},
						),
						layout.Rigid(
							func(gtx layout.Context) layout.Dimensions {
								return (&outlay.Grid{}).Layout(gtx, 15, 16, cellDimensionerTiles, cellFuncTiles)
							},
						),
					)
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}

func cellDimensionerTiles(axis layout.Axis, index, constraint int) int {
	return 16 * TILES_SCALE
}

func cellDimensionerLevel(axis layout.Axis, index, constraint int) int {
	return 16 * LEVEL_SCALE
}

func cellFuncLevel(gtx layout.Context, row, col int) layout.Dimensions {
	return buildLevelTile(gtx, &levelsUncompressed[levelIndex][row*16+col], row*16+col)
}

func cellFuncTiles(gtx layout.Context, row, col int) layout.Dimensions {
	return buildTile(gtx, tiles[row*16+col], row*16+col)
}

func buildLevelTile(gtx layout.Context, tag event.Tag, idx int) layout.Dimensions {
	for _, ev := range gtx.Queue.Events(tag) {
		if x, ok := ev.(pointer.Event); ok {
			switch x.Type {
			case pointer.Press:
				switch x.Buttons{
				case pointer.ButtonPrimary:
					levelsUncompressed[levelIndex][idx] = byte(selectedTile)
				case pointer.ButtonSecondary:
					selectedTile = int(levelsUncompressed[levelIndex][idx])
				}
			}
		}
	}

	defer clip.Rect{Max: image.Pt(16*LEVEL_SCALE, 16*LEVEL_SCALE)}.Push(gtx.Ops).Pop()

	pointer.InputOp{
		Tag:   tag,
		Types: pointer.Press,
	}.Add(gtx.Ops)
	texture := paint.NewImageOp(tiles[levelsUncompressed[levelIndex][idx]])
	tileImg := widget.Image{Src: texture, Scale: LEVEL_SCALE}
	return layout.Center.Layout(gtx, tileImg.Layout)
}

func buildTile(gtx layout.Context, tag event.Tag, idx int) layout.Dimensions {
	for _, ev := range gtx.Queue.Events(tag) {
		if x, ok := ev.(pointer.Event); ok {
			switch x.Type {
			case pointer.Press:
				selectedTile = idx
			}
		}
	}

	defer clip.Rect{Max: image.Pt(16*TILES_SCALE, 16*TILES_SCALE)}.Push(gtx.Ops).Pop()

	pointer.InputOp{
		Tag:   tag,
		Types: pointer.Press,
	}.Add(gtx.Ops)
	tileImg := widget.Image{Src: paint.NewImageOp(tiles[idx]), Scale: TILES_SCALE}
	if idx == selectedTile {
		borderImg := widget.Image{Src: paint.NewImageOp(highlightBorder), Scale: TILES_SCALE}
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, tileImg.Layout)
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, borderImg.Layout)
			}),
		)
	} else {
		return layout.Center.Layout(gtx, tileImg.Layout)
	}
}

func printLevelData(data []byte) {
	for y := 0; y <= 14; y++ {
		for x := 0; x <= 15; x++ {
			fmt.Printf("%2x ", data[y*16+x])
		}
		fmt.Println()
	}
	fmt.Println()
}
