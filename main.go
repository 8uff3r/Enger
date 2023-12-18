package main

import (
	"fmt"
	img "image"
	"image/color"
	"log"
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	ts "github.com/tinyspline/go"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(img.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	whiteImage.Fill(color.White)
}

const (
	screenWidth  = 960
	screenHeight = 680
)

type Game struct {
	spline         ts.BSpline
	drgPoint       *draggedCtrlPoint
	vertices       []ebiten.Vertex
	indices        []uint16
	ctrlPoints     [][2]int
	pastCtrlPoints [][2]int
	ui             *ebitenui.UI
	vs             []ebiten.Vertex
	is             []uint16
	counter        int
}
type draggedCtrlPoint struct {
	index      int
	x          int
	y          int
	r          int
	isReleased bool
}

func IntPow(n, m int) int {
	if m == 0 {
		return 1
	}
	result := n
	for i := 2; i <= m; i++ {
		result *= n
	}
	return result
}
func In(a, b, x, y, r int) bool {
	return math.Sqrt(float64(IntPow(x-a, 2)+IntPow((y-b), 2)))-float64(r)-2 <= 0
}

func (g *Game) Update() error {
	// g.counter++
	g.ui.Update()

	return nil
}

var curveScene *ebiten.Image

func (g *Game) Draw(screen *ebiten.Image) {
	// g.ui.Draw(curveScene)
	// screen.DrawImage(curveScene, &ebiten.DrawImageOptions{})
	target := curveScene
	// g.ui.Draw(target)

	g.pastCtrlPoints = g.ctrlPoints
	x, y := ebiten.CursorPosition()
	isPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if x > .75*screenWidth {
		isPressed = false
	}
	if isPressed {
		for k, v := range g.ctrlPoints {
			if In(v[0], v[1], x, y, 6) {
				g.drgPoint = &draggedCtrlPoint{
					index:      k,
					x:          v[0],
					y:          v[1],
					r:          6,
					isReleased: false,
				}
				break
			}
		}
	}
	g.ui.Draw(screen)
	if isPressed {
		return
	}
	isReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	if x > .75*screenWidth {
		isReleased = false
	}
	if isReleased {
		if g.drgPoint != nil {
			g.drgPoint = nil
			return
		}
	}
	reRender := 0
	if g.drgPoint != nil {
		g.MoveCtrlPoint(target, g.drgPoint, x, y)
		reRender = 1
	} else if x != 0 && y != 0 && isReleased {
		g.ctrlPoints = append(g.ctrlPoints, [2]int{x, y})
		g.DrawNewSpline(target, g.ctrlPoints)
		reRender = 1
	}
	if reRender == 0 {
		return
	}

	curveScene.Clear()

	for _, v := range g.ctrlPoints {
		g.AddPointAt(target, v[0], v[1], 6)
	}
	g.drawLineByPoints(target, g.ctrlPoints)
	if len(g.ctrlPoints) < 4 {
		return
	}
	g.drawSpline(target, g.spline.Sample())
	g.ui.Draw(screen)

	msg := fmt.Sprintf(`Press A to switch anti-aliasing.
Press C to switch to draw the center lines
X: %d, Y: %d
%v
		%v
`, x, y, (g.ctrlPoints), g.drgPoint)
	ebitenutil.DebugPrint(target, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) MoveCtrlPoint(target *ebiten.Image, p *draggedCtrlPoint, x, y int) {
	if x == 0 && y == 0 {
		return
	}
	g.ctrlPoints[p.index][0] = x
	g.ctrlPoints[p.index][1] = y
	// fmt.Printf("%v %v %v", x, y, p.index)
	if len(g.ctrlPoints) < 4 {
		return
	}
	g.spline.SetControlPointVec2At(p.index, ts.NewVec2(float64(x), float64(y)))
}

func (g *Game) DrawNewSpline(target *ebiten.Image, inpts [][2]int) {
	var flatInpts []float64
	for _, a := range inpts {
		flatInpts = append(flatInpts, []float64{float64(a[0]), float64(a[1])}...)
	}
	target.DrawTriangles(g.vs, g.is, whiteSubImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	})
	if len(inpts) < 4 {
		return
	}
	g.spline = ts.NewBSpline(len(inpts), 2, 3)

	g.spline.SetControlPoints(flatInpts)

	// pts := g.spline.Sample(100 * (len(inpts) / 5))
	// pts := spline.Sample()
	// g.drawLineByPoints(target, flatInpts)
	// g.drawSpline(target, pts)
}

func (g *Game) drawSpline(target *ebiten.Image, pts []float64) {
	var path vector.Path
	n := make([]struct{}, int(len(pts)/g.spline.GetDimension()-1))
	for k := range n {
		p0x := pts[k*g.spline.GetDimension()]
		p0y := pts[k*g.spline.GetDimension()+1]
		p1x := pts[(k+1)*g.spline.GetDimension()]
		p1y := pts[(k+1)*g.spline.GetDimension()+1]
		if k == 0 {
			path.MoveTo(float32(p0x), float32(p0y))
		}
		path.LineTo(float32(p1x), float32(p1y))
	}
	op := &vector.StrokeOptions{}
	op.Width = float32(3)
	g.vs, g.is = path.AppendVerticesAndIndicesForStroke(g.vertices[:0], g.indices[:0], op)
	for i := range g.vs {
		g.vs[i].SrcX = 0
		g.vs[i].SrcY = 0
		g.vs[i].ColorR = .5
		g.vs[i].ColorG = .5
		g.vs[i].ColorB = 1
		g.vs[i].ColorA = 1
	}
	target.DrawTriangles(g.vs, g.is, whiteSubImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	})
}

func (g *Game) drawLineByPoints(target *ebiten.Image, input [][2]int) {

	var pts []float64
	for _, a := range input {
		pts = append(pts, []float64{float64(a[0]), float64(a[1])}...)
	}
	var path vector.Path
	for k := range pts {
		if k%2 == 1 {
			continue
		}
		if k == 0 {
			path.MoveTo(float32(pts[k]), float32(pts[k+1]))
		} else {
			path.LineTo(float32(pts[k]), float32(pts[k+1]))
		}
	}
	op := &vector.StrokeOptions{}
	op.Width = float32(6)
	vs, is := path.AppendVerticesAndIndicesForStroke(g.vertices[:0], g.indices[:0], op)
	for i := range vs {
		vs[i].SrcX = 0
		vs[i].SrcY = 0
		vs[i].ColorR = .2
		vs[i].ColorG = 1
		vs[i].ColorB = .2
		vs[i].ColorA = .3
	}
	target.DrawTriangles(vs, is, whiteSubImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	})

}

func (g *Game) AddPointAt(target *ebiten.Image, x, y, r int) {
	vector.DrawFilledCircle(target, float32(x), float32(y), float32(r), color.White, true)
}
func (g *Game) clickPos() (int, int) {
	isReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	if isReleased {
		x, y := ebiten.CursorPosition()
		return x, y
	}
	return 0, 0
}
func NewGame() *Game {
	rootContainer := widget.NewContainer(
		// the container will use a plain color as its background
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(0)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true, true}),
		)),
	)
	curveScene = ebiten.NewImage(screenWidth*.75, screenHeight)
	curvesContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSlice(curveScene, [3]int{screenWidth * .75, 0, 0}, [3]int{screenHeight, 0, 0})),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(100, 600),
		),
	)
	rootContainer.AddChild(curvesContainer)

	rightContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(100, 600),
		),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{200, 0, 0, 0})),
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	idle, err := loadImageNineSlice("assets/button_idle.png", 12, 0)
	if err != nil {
		println(err)
	}
	hover, err := loadImageNineSlice("assets/button_hover.png", 12, 0)
	buttonImage := &widget.ButtonImage{
		Idle:  idle,
		Hover: hover,
		// Pressed: pressed,
	}
	face, _ := loadFont(12)
	conToBezierBut := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),

		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text("Convert to bezier curves", face, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),
		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			println("button clicked")
		}),
	)
	rightContainer.AddChild(conToBezierBut)
	rootContainer.AddChild(rightContainer)

	ui := ebitenui.UI{
		Container: rootContainer,
	}
	// rootContainer.AddChild(image.NewNineSlice(curveScene,)
	g := &Game{
		ui:       &ui,
		drgPoint: nil,
	}
	return g
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Lines (Ebitengine Demo)")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
