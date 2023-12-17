package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

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
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

const (
	screenWidth  = 640
	screenHeight = 680
)

type Game struct {
	spline         ts.BSpline
	scene          *ebiten.Image
	drgPoint       *draggedCtrlPoint
	vertices       []ebiten.Vertex
	indices        []uint16
	ctrlPoints     [][2]int
	pastCtrlPoints [][2]int
	vs             []ebiten.Vertex
	is             []uint16
	counter        int
	aa             bool
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
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.aa = !g.aa
	}

	return nil
}

var curveScene *ebiten.Image

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(curveScene, &ebiten.DrawImageOptions{})
	target := curveScene

	g.pastCtrlPoints = g.ctrlPoints
	x, y := ebiten.CursorPosition()
	isPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
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
		return
	}
	isReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
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
		AntiAlias: g.aa,
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
		AntiAlias: g.aa,
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
		AntiAlias: g.aa,
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
	g := &Game{
		scene: ebiten.NewImage(500, 500),
	}
	return g
}

func main() {
	var g Game
	g.drgPoint = nil
	curveScene = ebiten.NewImage(640, 680)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Lines (Ebitengine Demo)")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
