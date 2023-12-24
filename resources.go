package main

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
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
	labelIdleColor      = textIdleColor
	buttonIdleColor     = textIdleColor
	buttonDisabledColor = labelDisabledColor
)

func NewButton(txt string, handler func()) *widget.Button {
	idle, err := loadImageNineSlice("assets/button-idle.png", 12, 0)
	if err != nil {
		return nil
	}

	hover, err := loadImageNineSlice("assets/button-hover.png", 12, 0)
	if err != nil {
		return nil
	}
	pressed_hover, err := loadImageNineSlice("assets/button-selected-hover.png", 12, 0)
	if err != nil {
		return nil
	}
	pressed, err := loadImageNineSlice("assets/button-pressed.png", 12, 0)
	if err != nil {
		return nil
	}
	disabled, err := loadImageNineSlice("assets/button-disabled.png", 12, 0)
	if err != nil {
		return nil
	}
	buttonImage := &widget.ButtonImage{
		Idle:         idle,
		Hover:        hover,
		Pressed:      pressed,
		PressedHover: pressed_hover,
		Disabled:     disabled,
	}
	face, _ := loadFont(fontFaceRegular, 12)
	return widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text(txt, face, &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		}),
		widget.ButtonOpts.GraphicPadding(widget.Insets{
			Left:  30,
			Right: 30,
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
			handler()
		}),
	)
}

func NewSlider(min, max, step int, handler func(args *widget.SliderChangedEventArgs)) *widget.Slider {
	slider := widget.NewSlider(
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			handler(args)
		}),
		// Set the slider orientation - n/s vs e/w
		widget.SliderOpts.Direction(widget.DirectionHorizontal),
		// Set the minimum and maximum value for the slider
		widget.SliderOpts.MinMax(min, max),

		widget.SliderOpts.WidgetOpts(
			// Set the Widget to layout in the center on the screen
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			// Set the widget's dimensions
			widget.WidgetOpts.MinSize(200, 6),
		),
		widget.SliderOpts.Images(
			// Set the track images
			&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{200, 200, 200, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			},
			// Set the handle images
			&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Hover:   image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Pressed: image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
			},
		),
		// Set the size of the handle
		widget.SliderOpts.FixedHandleSize(6),
		// Set the offset to display the track
		widget.SliderOpts.TrackOffset(0),
		// Set the size to move the handle
		widget.SliderOpts.PageSizeFunc(func() int {
			return step
		}),
		// Set the callback to call when the slider value is changed
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			fmt.Println(args.Current)
		}),
	)
	// Set the current value of the slider
	slider.Current = 5
	return slider
}

func NewLabel(text string) *widget.Label {
	// fonts, err := loadFonts()
	// if err != nil {
	// 	return nil
	// }
	f, _ := loadFont(fontFaceBold, 12)
	t := widget.NewLabel(
		widget.LabelOpts.TextOpts(
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
				}),
			)),
		widget.LabelOpts.Text(text, f, &widget.LabelColor{
			Idle:     hexToColor(labelIdleColor),
			Disabled: hexToColor(labelDisabledColor),
		}),
	)
	return t
}
