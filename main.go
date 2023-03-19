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

type Selection struct {
	origin      image.Point
	destination image.Point
	active      bool
}

var clickableTiles = make([]widget.Clickable, 240)
var clickableLevelTiles = make([]widget.Clickable, 240)
var highlightBorder *image.RGBA

var loloData []byte
var levelsUncompressed = make([][]byte, 60)
var levelAmount int
var levelIndex int
var selectedTile int
var levelGridTag bool
var dragSelection Selection

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
			//fmt.Println(counter)
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
								return buildLevelGrid(gtx, &levelGridTag)
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

func applyOverlaySelection() {
	for i := 0; i < 15; i++ {
		for j := 0; j < 16; j++ {
			y := i * 16 * LEVEL_SCALE
			x := j * 16 * LEVEL_SCALE
			pSrc := image.Point{x, y}
			pDest := image.Point{x + 16*LEVEL_SCALE - 1, y + 16*LEVEL_SCALE - 1}
			oSrc, oDest := alignRectPoints(dragSelection.origin, dragSelection.destination)
			if dragSelection.active && checkCollision(pSrc, pDest, oSrc, oDest) {
				levelsUncompressed[levelIndex][i*16+j] = byte(selectedTile)
			}
		}
	}
}

func buildLevelGrid(gtx layout.Context, tag event.Tag) layout.Dimensions {
	for _, ev := range gtx.Queue.Events(tag) {
		if x, ok := ev.(pointer.Event); ok {
			switch x.Type {
			case pointer.Press:
				switch x.Buttons {
				case pointer.ButtonPrimary:
					dragSelection.origin = x.Position.Round()
					dragSelection.destination = x.Position.Round()
				}
			case pointer.Release:
				if dragSelection.active {
					applyOverlaySelection()
					dragSelection.active = false
				}
			case pointer.Drag:
				switch x.Buttons {
				case pointer.ButtonPrimary:
					dragSelection.destination = x.Position.Round()
					dragSelection.active = true
				}
			}
		}
	}

	defer clip.Rect{Max: image.Pt(16*LEVEL_SCALE*16, 16*LEVEL_SCALE*15)}.Push(gtx.Ops).Pop()

	pointer.InputOp{
		Tag:   tag,
		Types: pointer.Drag | pointer.Press | pointer.Release,
	}.Add(gtx.Ops)

	grid := (&outlay.Grid{}).Layout(gtx, 15, 16, cellDimensionerLevel, cellFuncLevel)
	if dragSelection.active {
		shape := clip.Rect{Max: image.Pt(dragSelection.destination.X-dragSelection.origin.X, dragSelection.destination.Y-dragSelection.origin.Y)}
		defer op.Offset(dragSelection.origin).Push(gtx.Ops).Pop()
		paint.FillShape(gtx.Ops, color.NRGBA{0, 255, 255, 255},
			clip.Stroke{
				Path:  shape.Path(),
				Width: 2,
			}.Op())
	}

	return grid
}

func buildLevelTile(gtx layout.Context, tag event.Tag, idx int) layout.Dimensions {
	for _, ev := range gtx.Queue.Events(tag) {
		if x, ok := ev.(pointer.Event); ok {
			switch x.Type {
			case pointer.Press:
				switch x.Buttons {
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
	x := idx % 16 * 16 * LEVEL_SCALE
	y := idx / 16 * 16 * LEVEL_SCALE
	var texture image.Image
	pSrc := image.Point{x, y}
	pDest := image.Point{x + 16*LEVEL_SCALE - 1, y + 16*LEVEL_SCALE - 1}
	oSrc, oDest := alignRectPoints(dragSelection.origin, dragSelection.destination)
	if dragSelection.active && checkCollision(pSrc, pDest, oSrc, oDest) {
		texture = tiles[selectedTile]
	} else {
		texture = tiles[levelsUncompressed[levelIndex][idx]]
	}
	tileImg := widget.Image{Src: paint.NewImageOp(texture), Scale: LEVEL_SCALE}
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

func checkCollision(r1s, r1d, r2s, r2d image.Point) bool {
	if r1s.X <= r2d.X &&
		r1d.X >= r2s.X &&
		r1s.Y <= r2d.Y &&
		r1d.Y >= r2s.Y {
		return true
	} else {
		return false
	}
}

func alignRectPoints(rs, rd image.Point) (image.Point, image.Point) {
	if rs.Y > rd.Y {
		rs.Y, rd.Y = rd.Y, rs.Y
	}
	if rs.X > rd.X {
		rs.X, rd.X = rd.X, rs.X
	}
	return rs, rd
}
