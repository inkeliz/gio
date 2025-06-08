//go:build js

// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"fmt"
	"syscall/js"

	"gioui.org/font/opentype"
)

// loadSystemFontsJS attempts to load system fonts using the Local Font Access API
// available in Chrome and other Chromium-based browsers.
func loadSystemFontsJS(shaper *shaperImpl) error {
	// Check if the Local Font Access API is available
	navigator := js.Global().Get("navigator")
	if !navigator.Truthy() {
		return fmt.Errorf("navigator not available")
	}
	
	fonts := navigator.Get("fonts")
	if !fonts.Truthy() {
		return fmt.Errorf("Local Font Access API not available")
	}
	
	query := fonts.Get("query")
	if !query.Truthy() {
		return fmt.Errorf("fonts.query() not available")
	}
	
	// Create a promise to load fonts
	promise := query.Invoke()
	if !promise.Truthy() {
		return fmt.Errorf("failed to query fonts")
	}
	
	// Use a channel to handle the async operation
	done := make(chan error, 1)
	loaded := 0
	
	// Handle the promise resolution
	successCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic in font loading: %v", r)
			}
		}()
		
		if len(args) == 0 {
			done <- fmt.Errorf("no fonts returned")
			return nil
		}
		
		fonts := args[0]
		if !fonts.Truthy() {
			done <- fmt.Errorf("fonts array is empty")
			return nil
		}
		
		length := fonts.Get("length").Int()
		shaper.logger.Printf("found %d system fonts via Local Font Access API", length)
		
		// Load fonts sequentially to avoid overwhelming the browser
		for i := 0; i < length && i < 50; i++ { // Limit to 50 fonts to avoid performance issues
			font := fonts.Index(i)
			if err := loadSingleFontJS(shaper, font); err != nil {
				shaper.logger.Printf("failed to load font %d: %v", i, err)
			} else {
				loaded++
			}
		}
		
		shaper.logger.Printf("successfully loaded %d system fonts", loaded)
		done <- nil
		return nil
	})
	defer successCallback.Release()
	
	errorCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			done <- fmt.Errorf("error querying fonts: %s", args[0].Get("message").String())
		} else {
			done <- fmt.Errorf("unknown error querying fonts")
		}
		return nil
	})
	defer errorCallback.Release()
	
	promise.Call("then", successCallback).Call("catch", errorCallback)
	
	// Wait for the promise to resolve with a timeout
	select {
	case err := <-done:
		return err
	case <-timeoutChannel(10): // 10 second timeout
		return fmt.Errorf("timeout loading fonts via Local Font Access API")
	}
}

// timeoutChannel creates a channel that sends after the specified seconds
func timeoutChannel(seconds int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		// Use JavaScript's setTimeout to implement timeout
		timeout := js.Global().Get("setTimeout")
		if timeout.Truthy() {
			callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				close(ch)
				return nil
			})
			timeout.Invoke(callback, seconds*1000)
		} else {
			// Fallback: close immediately if setTimeout is not available
			close(ch)
		}
	}()
	return ch
}

// loadSingleFontJS loads a single font from the Local Font Access API
func loadSingleFontJS(shaper *shaperImpl, fontHandle js.Value) error {
	// Get font metadata
	family := fontHandle.Get("family").String()
	style := fontHandle.Get("style").String()
	
	if family == "" {
		return fmt.Errorf("font has no family name")
	}
	
	// Get the font blob
	blobMethod := fontHandle.Get("blob")
	if !blobMethod.Truthy() {
		return fmt.Errorf("font.blob() not available")
	}
	
	promise := blobMethod.Invoke()
	if !promise.Truthy() {
		return fmt.Errorf("failed to get font blob")
	}
	
	// Create channels to handle the async blob operation
	done := make(chan error, 1)
	var fontData []byte
	
	successCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic in blob loading: %v", r)
			}
		}()
		
		if len(args) == 0 {
			done <- fmt.Errorf("no blob returned")
			return nil
		}
		
		blob := args[0]
		if !blob.Truthy() {
			done <- fmt.Errorf("blob is null")
			return nil
		}
		
		// Convert blob to ArrayBuffer
		arrayBufferPromise := blob.Call("arrayBuffer")
		
		arrayBufferSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("panic in ArrayBuffer conversion: %v", r)
				}
			}()
			
			if len(args) == 0 {
				done <- fmt.Errorf("no ArrayBuffer returned")
				return nil
			}
			
			arrayBuffer := args[0]
			if !arrayBuffer.Truthy() {
				done <- fmt.Errorf("ArrayBuffer is null")
				return nil
			}
			
			// Copy the data to Go
			length := arrayBuffer.Get("byteLength").Int()
			fontData = make([]byte, length)
			js.CopyBytesToGo(fontData, js.Global().Get("Uint8Array").New(arrayBuffer))
			
			done <- nil
			return nil
		})
		defer arrayBufferSuccess.Release()
		
		arrayBufferError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				done <- fmt.Errorf("error converting ArrayBuffer: %s", args[0].Get("message").String())
			} else {
				done <- fmt.Errorf("unknown error converting ArrayBuffer")
			}
			return nil
		})
		defer arrayBufferError.Release()
		
		arrayBufferPromise.Call("then", arrayBufferSuccess).Call("catch", arrayBufferError)
		return nil
	})
	defer successCallback.Release()
	
	errorCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			done <- fmt.Errorf("error getting blob: %s", args[0].Get("message").String())
		} else {
			done <- fmt.Errorf("unknown error getting blob")
		}
		return nil
	})
	defer errorCallback.Release()
	
	promise.Call("then", successCallback).Call("catch", errorCallback)
	
	// Wait for the blob to be loaded with timeout
	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-timeoutChannel(5): // 5 second timeout per font
		return fmt.Errorf("timeout loading font blob for %s", family)
	}
	
	if len(fontData) == 0 {
		return fmt.Errorf("font data is empty for %s", family)
	}
	
	// Parse the font data
	face, err := opentype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("failed to parse font %s: %w", family, err)
	}
	
	// Load the font into the shaper
	fontFace := FontFace{
		Face: face,
		Font: face.Font(),
	}
	
	shaper.Load(fontFace)
	shaper.logger.Printf("loaded system font: %s (%s)", family, style)
	
	return nil
}