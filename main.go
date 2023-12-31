package main

import (
	"fmt"
	img "image"
	"image/color"
	"log"
	"math"
	"strconv"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
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
	drgPoint      *draggedCtrlPoint
	ui            *ebitenui.UI
	extraPts      []float64
	spline        Spline
	counter       int
	forceRerender bool
}
type Spline struct {
	curve      ts.BSpline
	uPoint     *Point
	ctrlPoints [][2]int
	degree     int
	u          float64
	knotIndex  int
	re         bool
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

func FlatFloatTo2dInt(s []float64) [][2]int {
	var result [][2]int
	for k := range s {
		if k%2 == 1 {
			continue
		}
		result = append(result, [2]int{int(s[k]), int(s[k+1])})
	}
	return result
}

func (g *Game) Update() error {
	// g.counter++
	g.ui.Update()

	return nil
}

var curveScene *ebiten.Image

func (g *Game) Draw(screen *ebiten.Image) {
	defer g.ui.Draw(screen)
	target := curveScene

	isPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	isReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	x, y := ebiten.CursorPosition()
	if x > .75*screenWidth {
		isReleased = false
		isPressed = false
	}

	if g.spline.curve != nil {
		g.spline.ctrlPoints = FlatFloatTo2dInt(g.spline.curve.GetControlPoints())
	}

	g.spline.re = false
	if isPressed {
		var drgPoint *draggedCtrlPoint
		for k, v := range g.spline.ctrlPoints {
			if In(v[0], v[1], x, y, 6) {
				drgPoint = &draggedCtrlPoint{
					index:      k,
					x:          v[0],
					y:          v[1],
					r:          6,
					isReleased: false,
				}
				break
			}
		}
		// FIXME:
		if g.drgPoint == drgPoint {
			g.drgPoint = nil
		} else {
			g.drgPoint = drgPoint
		}
		return
	}
	if isReleased {
		if g.drgPoint != nil {
			g.drgPoint = nil
			return
		}
	}
	reRender := g.forceRerender
	g.forceRerender = false
	if g.drgPoint != nil {
		g.spline.MoveCtrlPoint(g.drgPoint, x, y)
		reRender = true
	} else if x != 0 && y != 0 && isReleased {
		g.spline.ctrlPoints = append(g.spline.ctrlPoints, [2]int{x, y})
		g.NewSpline(target, g.spline.ctrlPoints)
		reRender = true
	}
	if !reRender {
		return
	}
	curveScene.Clear()
	g.AddPointByFlatList(target, g.extraPts, 3, color.RGBA{R: 255})

	for _, v := range g.spline.ctrlPoints {
		g.AddPointAt(target, v[0], v[1], 6, color.White)
	}
	g.drawLineByPoints(target, g.spline.ctrlPoints)
	if len(g.spline.ctrlPoints) < g.spline.degree+1 {
		return
	}
	ptsNum := 100 * (len(g.spline.ctrlPoints) / 5)
	g.drawSpline(target, g.spline.curve.Sample(ptsNum))

	if g.spline.curve != nil {
		knts := g.spline.curve.GetKnots()
		var res []int
		for k, v := range knts {
			knt, _ := GetPointAt(g.spline.curve, v)
			g.AddPointAt(target, int(knt[0]), int(knt[1]), 3, color.RGBA{R: 255})
			if k < g.spline.curve.GetDegree() || k > len(knts)-g.spline.curve.GetDegree()-1 {
				continue
			}
			res = append(res, int(v*float64((len(knts)-2*g.spline.curve.GetDegree())-1)))
		}
		if g.spline.knotIndex+g.spline.degree+1 > len(knts)-1 {
			return
		}
		// u := g.spline.u*knts[3+g.spline.knotIndex] + knts[g.spline.knotIndex+2]
		u := g.spline.u*(knts[g.spline.degree+g.spline.knotIndex]-knts[g.spline.knotIndex+g.spline.degree-1]) + knts[g.spline.knotIndex+g.spline.degree-1]
		if u > 1 {
			u = 1
		}
		pts, _ := GetPointAt(g.spline.curve, u)
		SetPoint(&g.spline.uPoint, pts[0], pts[1])
		g.AddPointAt(target, int(pts[0]), int(pts[1]), 12, color.RGBA{G: 255})
		fmt.Printf("%v\n", g.spline.uPoint.GetPoint())
	}
	if g.spline.uPoint != nil {
		g.spline.uPoint.ch = false
		g.AddPointAt(target, int(g.spline.uPoint.x), int(g.spline.uPoint.y), 10, color.RGBA{B: 255})
	}
	//	msg := fmt.Sprintf(`Press A to switch anti-aliasing.
	//
	// Press C to switch to draw the center lines
	// X: %d, Y: %d
	// %v
	//
	//	%v
	//
	// `, x, y, (g.spline.ctrlPoints), g.drgPoint)
	//
	//	ebitenutil.DebugPrint(target, msg)
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

func (s *Spline) MoveCtrlPoint(p *draggedCtrlPoint, x, y int) {
	if x == 0 && y == 0 || s.curve == nil {
		return
	}
	s.ctrlPoints[p.index][0] = x
	s.ctrlPoints[p.index][1] = y
	if len(s.ctrlPoints) < s.degree+1 {
		return
	}
	s.curve.SetControlPointVec2At(p.index, ts.NewVec2(float64(x), float64(y)))
}

func (g *Game) NewSpline(target *ebiten.Image, inpts [][2]int) {
	var flatInpts []float64
	for _, a := range inpts {
		flatInpts = append(flatInpts, []float64{float64(a[0]), float64(a[1])}...)
	}
	if len(inpts) < g.spline.degree+1 {
		return
	}
	g.spline.curve = ts.NewBSpline(len(inpts), 2, g.spline.degree)
	g.spline.curve.SetControlPoints(flatInpts)
}

func (g *Game) drawSpline(target *ebiten.Image, pts []float64) {
	var path vector.Path
	n := make([]struct{}, int(len(pts)/g.spline.curve.GetDimension()-1))
	for k := range n {
		p0x := pts[k*g.spline.curve.GetDimension()]
		p0y := pts[k*g.spline.curve.GetDimension()+1]
		p1x := pts[(k+1)*g.spline.curve.GetDimension()]
		p1y := pts[(k+1)*g.spline.curve.GetDimension()+1]
		if k == 0 {
			path.MoveTo(float32(p0x), float32(p0y))
		}
		path.LineTo(float32(p1x), float32(p1y))
	}
	op := &vector.StrokeOptions{}
	op.Width = float32(3)

	var vertices []ebiten.Vertex
	var indices []uint16
	vs, is := path.AppendVerticesAndIndicesForStroke(vertices[:0], indices[:0], op)
	for i := range vs {
		vs[i].SrcX = 0
		vs[i].SrcY = 0
		vs[i].ColorR = .5
		vs[i].ColorG = .5
		vs[i].ColorB = 1
		vs[i].ColorA = 1
	}
	target.DrawTriangles(vs, is, whiteSubImage, &ebiten.DrawTrianglesOptions{
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

	var vertices []ebiten.Vertex
	var indices []uint16
	vs, is := path.AppendVerticesAndIndicesForStroke(vertices[:0], indices[:0], op)
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

		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)

	ui := ebitenui.UI{
		Container: rootContainer,
	}
	g := &Game{
		ui:            &ui,
		drgPoint:      nil,
		forceRerender: false,
	}
	g.spline.degree = 2
	conToBezierBut := NewButton("Convert to bezier curves", func() {
		if g.spline.curve != nil {
			g.spline.curve = g.spline.curve.ToBeziers()
			g.forceRerender = true
		}
	})

	var uText *widget.Label
	uSlider := NewSlider(0, 100, 1, func(args *widget.SliderChangedEventArgs) {
		uText.Label = fmt.Sprintf("%0.2f", float64(args.Current)/100)
		g.forceRerender = true
		g.spline.u = float64(args.Current) / 100
	})
	uInput := NewInput("Knot index", func(args *widget.TextInputChangedEventArgs) {
		parsed, err := strconv.Atoi(args.InputText)
		if err != nil {
			return
		}
		g.spline.knotIndex = parsed
		g.spline.u = 0
	})
	uText = NewLabel(fmt.Sprintf("%d", uSlider.Current))

	degreeInput := NewInput("Degree", func(args *widget.TextInputChangedEventArgs) {
		parsed, err := strconv.Atoi(args.InputText)
		if err != nil {
			return
		}
		g.spline.degree = parsed
	})

	clearButton := NewButton("Clear scene", func() {
		g.spline.curve = nil
		g.spline.uPoint = nil
		g.spline.ctrlPoints = nil
		g.spline.re = true
		g.forceRerender = true
	})
	rightContainer.AddChild(clearButton)
	rightContainer.AddChild(conToBezierBut)
	rightContainer.AddChild(degreeInput)
	rightContainer.AddChild(uInput)
	rightContainer.AddChild(uSlider)
	rightContainer.AddChild(uText)
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
