//go:build !js

// SPDX-License-Identifier: Unlicense OR MIT

package text

import "fmt"

// loadSystemFontsJS is not available on non-JS platforms
func loadSystemFontsJS(shaper *shaperImpl) error {
	return fmt.Errorf("Local Font Access API only available on JS/WASM")
}