//go:build js

// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"fmt"
	"syscall/js"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

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
	
	// Try to get cached fonts first
	cache, err := getCachedFonts()
	if err != nil {
		shaper.logger.Printf("failed to get font cache: %v", err)
		cache = make(map[string]fontCacheEntry)
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
			if err := loadSingleFontJS(shaper, font, cache); err != nil {
				shaper.logger.Printf("failed to load font %d: %v", i, err)
			} else {
				loaded++
			}
		}
		
		// Update cache
		if err := setCachedFonts(cache); err != nil {
			shaper.logger.Printf("failed to update font cache: %v", err)
		}
		
		shaper.logger.Printf("successfully loaded %d system fonts", loaded)
		done <- nil
		return nil
	})
	defer successCallback.Release()
	
	errorCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 && args[0].Truthy() {
			msg := "unknown error"
			if msgVal := args[0].Get("message"); msgVal.Truthy() {
				msg = msgVal.String()
			}
			// Handle permission denied specifically for late-authorization
			if args[0].Get("name").String() == "NotAllowedError" {
				done <- fmt.Errorf("font access permission denied - user can grant permission later")
			} else {
				done <- fmt.Errorf("error querying fonts: %s", msg)
			}
		} else {
			done <- fmt.Errorf("unknown error querying fonts")
		}
		return nil
	})
	defer errorCallback.Release()
	
	promise.Call("then", successCallback).Call("catch", errorCallback)
	
	// Wait for the promise to resolve
	return <-done
}


// fontCacheEntry represents a cached font entry
type fontCacheEntry struct {
	Family    string    `json:"family"`
	Style     string    `json:"style"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}

// getCachedFonts retrieves cached font metadata from localStorage
func getCachedFonts() (map[string]fontCacheEntry, error) {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return nil, fmt.Errorf("localStorage not available")
	}
	
	cacheData := localStorage.Call("getItem", "gio_font_cache")
	if !cacheData.Truthy() {
		return make(map[string]fontCacheEntry), nil
	}
	
	var cache map[string]fontCacheEntry
	if err := json.Unmarshal([]byte(cacheData.String()), &cache); err != nil {
		// Clear invalid cache and start fresh
		localStorage.Call("removeItem", "gio_font_cache")
		return make(map[string]fontCacheEntry), nil
	}
	
	return cache, nil
}

// setCachedFonts stores font metadata in localStorage
func setCachedFonts(cache map[string]fontCacheEntry) error {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return fmt.Errorf("localStorage not available")
	}
	
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	
	localStorage.Call("setItem", "gio_font_cache", string(data))
	return nil
}

// getFontHash generates a hash for the font data
func getFontHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// isSupportedFontFormat checks if the font format is supported by the OpenType parser
func isSupportedFontFormat(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	
	// Check for common OpenType/TrueType signatures
	signature := string(data[:4])
	switch signature {
	case "OTTO": // OpenType with CFF outlines
		return true
	case "\x00\x01\x00\x00": // TrueType signature
		return true
	case "true": // TrueType on Mac
		return true
	case "typ1": // Type 1 font (not supported by go-text/typesetting)
		return false
	default:
		// Check for TrueType Collection signature
		if len(data) >= 4 && string(data[:4]) == "ttcf" {
			return true
		}
		return false
	}
}

// loadSingleFontJS loads a single font from the Local Font Access API
func loadSingleFontJS(shaper *shaperImpl, fontHandle js.Value, cache map[string]fontCacheEntry) error {
	// Get font metadata
	family := fontHandle.Get("family").String()
	style := fontHandle.Get("style").String()
	
	if family == "" {
		return fmt.Errorf("font has no family name")
	}
	
	cacheKey := family + ":" + style
	
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
			if len(args) > 0 && args[0].Truthy() {
				msg := "unknown error"
				if msgVal := args[0].Get("message"); msgVal.Truthy() {
					msg = msgVal.String()
				}
				done <- fmt.Errorf("error converting ArrayBuffer: %s", msg)
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
		if len(args) > 0 && args[0].Truthy() {
			msg := "unknown error"
			if msgVal := args[0].Get("message"); msgVal.Truthy() {
				msg = msgVal.String()
			}
			done <- fmt.Errorf("error getting blob: %s", msg)
		} else {
			done <- fmt.Errorf("unknown error getting blob")
		}
		return nil
	})
	defer errorCallback.Release()
	
	promise.Call("then", successCallback).Call("catch", errorCallback)
	
	// Wait for the blob to be loaded
	if err := <-done; err != nil {
		return err
	}
	
	if len(fontData) == 0 {
		return fmt.Errorf("font data is empty for %s", family)
	}
	
	// Check if the font format is supported
	if !isSupportedFontFormat(fontData) {
		return fmt.Errorf("unsupported font format for %s", family)
	}
	
	// Check cache to avoid reprocessing fonts
	fontHash := getFontHash(fontData)
	if cached, exists := cache[cacheKey]; exists {
		// Check if the font hasn't changed and isn't too old (cache for 7 days)
		if cached.Hash == fontHash && time.Since(cached.Timestamp) < 7*24*time.Hour {
			shaper.logger.Printf("using cached font: %s (%s)", family, style)
			// Font is cached and still valid, skip loading
			return nil
		}
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
	
	// Update cache with successful load
	cache[cacheKey] = fontCacheEntry{
		Family:    family,
		Style:     style,
		Hash:      fontHash,
		Timestamp: time.Now(),
	}
	
	return nil
}