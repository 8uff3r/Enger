package main

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
)

const (
	fontFaceRegular = "assets/fonts/VisbyRoundCF-Regular.otf"
	fontFaceBold    = "assets/fonts/VisbyRoundCF-Bold.otf"
)

type fonts struct {
	face         font.Face
	titleFace    font.Face
	bigTitleFace font.Face
	toolTipFace  font.Face
}

func loadButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.RGBA{R: 0, G: 0, B: 20, A: 255})

	hover := image.NewNineSliceColor(color.RGBA{R: 0, G: 0, B: 25, A: 255})

	pressed := image.NewNineSliceColor(color.RGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func loadFonts() (*fonts, error) {
	fontFace, err := loadFont(fontFaceRegular, 20)
	if err != nil {
		return nil, err
	}

	titleFontFace, err := loadFont(fontFaceBold, 24)
	if err != nil {
		return nil, err
	}

	bigTitleFontFace, err := loadFont(fontFaceBold, 28)
	if err != nil {
		return nil, err
	}

	toolTipFace, err := loadFont(fontFaceRegular, 15)
	if err != nil {
		return nil, err
	}

	return &fonts{
		face:         fontFace,
		titleFace:    titleFontFace,
		bigTitleFace: bigTitleFontFace,
		toolTipFace:  toolTipFace,
	}, nil
}
func loadFont(path string, size float64) (font.Face, error) {
	fontData, err := embeddedAssets.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(ttfFont, &truetype.Options{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	}), nil
}
func newImageFromFile(path string) (*ebiten.Image, error) {
	f, err := embeddedAssets.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	i, _, err := ebitenutil.NewImageFromReader(f)
	return i, err
}
func loadGraphicImages(idle string, disabled string) (*widget.ButtonImageImage, error) {
	idleImage, err := newImageFromFile(idle)
	if err != nil {
		return nil, err
	}

	var disabledImage *ebiten.Image
	if disabled != "" {
		disabledImage, err = newImageFromFile(disabled)
		if err != nil {
			return nil, err
		}
	}

	return &widget.ButtonImageImage{
		Idle:     idleImage,
		Disabled: disabledImage,
	}, nil
}

func loadImageNineSlice(path string, centerWidth int, centerHeight int) (*image.NineSlice, error) {
	i, err := newImageFromFile(path)
	if err != nil {
		return nil, err
	}
	w := i.Bounds().Dx()
	h := i.Bounds().Dy()
	return image.NewNineSlice(i,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight}),
		nil
}

func hexToColor(h string) color.Color {
	u, err := strconv.ParseUint(h, 16, 0)
	if err != nil {
		panic(err)
	}

	return color.NRGBA{
		R: uint8(u & 0xff0000 >> 16),
		G: uint8(u & 0xff00 >> 8),
		B: uint8(u & 0xff),
		A: 255,
	}
}
