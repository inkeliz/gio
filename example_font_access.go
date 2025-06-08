// Example demonstrating Local Font Access API for JS/WASM
// Build with: GOOS=js GOARCH=wasm go build -o example.wasm example.go
// Serve with an HTTP server and load in a Chrome browser

//go:build js

package main

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		// Create window
		w := new(app.Window)
		
		// Create text shaper with system fonts enabled
		shaper := text.NewShaper(text.WithCollection(gofont.Collection()))
		
		// Create material theme
		th := material.NewTheme()
		th.Shaper = shaper
		
		var ops op.Ops
		
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				return
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				
				// Create text content
				content := "Hello, World! This text should use system fonts if Local Font Access API is available."
				
				// Layout the text
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H1(th, "System Font Test").Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, content).Layout(gtx)
					}),
				)
				
				e.Frame(gtx.Ops)
			}
		}
	}()
	
	app.Main()
}