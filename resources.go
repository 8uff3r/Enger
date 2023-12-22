package main

import (
	"github.com/ebitenui/ebitenui/widget"
	"golang.org/x/image/font"
)

type buttonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    font.Face
	padding widget.Insets
}

const (
	textIdleColor       = "dff4ff"
	textDisabledColor   = "5a7a91"
	labelDisabledColor  = textDisabledColor
	buttonIdleColor     = textIdleColor
	buttonDisabledColor = labelDisabledColor
)

func newButtonResources() (*buttonResources, error) {
	idle, err := loadImageNineSlice("assets/graphics/button-idle.png", 12, 0)

	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("assets/graphics/button-hover.png", 12, 0)
	if err != nil {
		return nil, err
	}
	pressed_hover, err := loadImageNineSlice("assets/graphics/button-selected-hover.png", 12, 0)
	if err != nil {
		return nil, err
	}
	pressed, err := loadImageNineSlice("assets/graphics/button-pressed.png", 12, 0)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("assets/graphics/button-disabled.png", 12, 0)
	if err != nil {
		return nil, err
	}

	i := &widget.ButtonImage{
		Idle:         idle,
		Hover:        hover,
		Pressed:      pressed,
		PressedHover: pressed_hover,
		Disabled:     disabled,
	}

	return &buttonResources{
		image: i,

		text: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		},

		padding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}
