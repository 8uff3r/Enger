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
	ui             *ebitenui.UI
	uPoint         *Point
	pastCtrlPoints [][2]int
	ctrlPoints     [][2]int
	indices        []uint16
	vs             []ebiten.Vertex
	is             []uint16
	extraPts       []float64
	vertices       []ebiten.Vertex
	counter        int
	forceRerender  bool
}
type Point struct {
	x  float64
	y  float64
	ch bool
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

	if g.spline != nil {
		g.ctrlPoints = [][2]int{}

		currCtrl := g.spline.GetControlPoints()
		for k := range currCtrl {
			if k%2 == 1 {
				continue
			}
			g.ctrlPoints = append(g.ctrlPoints, [2]int{int(currCtrl[k]), int(currCtrl[k+1])})
		}
	}
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
	if isPressed {
		g.ui.Draw(screen)
		return
	}
	isReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	if x > .75*screenWidth {
		isReleased = false
	}
	if isReleased {
		if g.drgPoint != nil {
			g.ui.Draw(screen)
			g.drgPoint = nil
			return
		}
	}
	reRender := g.forceRerender
	if g.drgPoint != nil {
		g.MoveCtrlPoint(target, g.drgPoint, x, y)
		reRender = true
	} else if x != 0 && y != 0 && isReleased {
		g.ctrlPoints = append(g.ctrlPoints, [2]int{x, y})
		if g.spline != nil {
			knts := g.spline.GetKnots()
			var res []int
			for k, v := range knts {
				if k < g.spline.GetDegree() || k > len(knts)-g.spline.GetDegree()-1 {
					continue
				}
				res = append(res, int(v*float64((len(knts)-2*g.spline.GetDegree())-1)))

			}

			// res := g.spline.ToBeziers()
			// res, _ := GetPointAt(g.spline, .1)
			// g.extraPts = append(g.extraPts, pts...)
			pts, _ := GetPointAt(g.spline, 0.2*knts[5])
			fmt.Printf("pts: %v\n", pts)
			SetPoint(&g.uPoint, pts[0], pts[1])

			// g.extraPts = res
			fmt.Printf("%v\n", g.uPoint.GetPoint())
		}
		g.DrawNewSpline(target, g.ctrlPoints)
		reRender = true
	}
	if !reRender {
		g.ui.Draw(screen)
		return
	}

	curveScene.Clear()
	g.AddPointByFlatList(target, g.extraPts, 9, color.RGBA{G: 255})

	for _, v := range g.ctrlPoints {
		g.AddPointAt(target, v[0], v[1], 6, color.White)
	}
	g.drawLineByPoints(target, g.ctrlPoints)
	if len(g.ctrlPoints) < 4 {
		return
	}
	ptsNum := 100 * (len(g.ctrlPoints) / 5)
	g.drawSpline(target, g.spline.Sample(ptsNum))
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
func (p *Point) GetPoint() *Point {
	p.ch = false
	return p
}
func SetPoint(p **Point, x, y float64) {
	*p = &Point{
		x:  x,
		y:  y,
		ch: true,
	}
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

func (g *Game) AddPointByFlatList(target *ebiten.Image, pts []float64, r int, color color.Color) {
	for k := range pts {
		if k%2 == 1 {
			continue
		}
		g.AddPointAt(target, int(pts[k]), int(pts[k+1]), r, color)

	}
}
func (g *Game) AddPointAt(target *ebiten.Image, x, y, r int, color color.Color) {
	vector.DrawFilledCircle(target, float32(x), float32(y), float32(r), color, true)
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
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x40, 0x1a, 0x22, 0xff})),
		// the container will use an anchor layout to layout its single child widget

		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.Insets{
				Left:   10,
				Right:  10,
				Top:    10,
				Bottom: 10,
			}),
		)),
	)

	ui := ebitenui.UI{
		Container: rootContainer,
	}
	// rootContainer.AddChild(image.NewNineSlice(curveScene,)
	g := &Game{
		ui:            &ui,
		drgPoint:      nil,
		forceRerender: false,
	}
	conToBezierBut := NewButton("Convert to bezier curves", func() {
		if g.spline != nil {
			g.spline = g.spline.ToBeziers()
			g.forceRerender = true
		}
	})
	rightContainer.AddChild(conToBezierBut)

	// var uText *widget.Label
	// uSlider := NewSlider(0, 100, 1, func(args *widget.SliderChangedEventArgs) {
	// 	uText.Label = fmt.Sprintf("%0.2f", float64(args.Current)/100)
	// })
	// uText = NewLabel(fmt.Sprintf("%d", uSlider.Current))
	// rightContainer.AddChild(uSlider)
	// rightContainer.AddChild(uText)
	rootContainer.AddChild(rightContainer)

	return g
}
func GetPointAt(spline ts.BSpline, u float64) ([]float64, []float64) {
	net := spline.Eval(u)
	pts := net.GetPoints()
	res := net.GetResult()
	// fmt.Printf("pts: %v, res: %v\n\n", pts, res)
	return res, pts
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Lines (Ebitengine Demo)")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
