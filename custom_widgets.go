package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type WidgetButton struct {
	widget.BaseWidget
	child     fyne.CanvasObject
	OnPressed func()
}

func (b *WidgetButton) CreateRenderer() fyne.WidgetRenderer {
	container := container.NewStack(widget.NewButton("", b.OnPressed), b.child)
	return widget.NewSimpleRenderer(container)
}

func NewWidgetButton(onPressed func(), child fyne.CanvasObject) *WidgetButton {
	b := &WidgetButton{OnPressed: onPressed, child: child}
	b.ExtendBaseWidget(b)
	return b
}

// func (b *WidgetButton) Tapped(_ *fyne.PointEvent) {
// 	if b.OnPressed != nil {
// 		b.OnPressed()
// 	}
// }
