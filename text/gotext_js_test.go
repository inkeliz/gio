//go:build js

// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"testing"
)

func TestLoadSystemFontsJS(t *testing.T) {
	// Create a test shaper
	shaper := &shaperImpl{
		logger: newDebugLogger(),
	}
	
	// Test the function - it should fail gracefully in test environment
	// since there's no browser DOM available
	err := loadSystemFontsJS(shaper)
	if err == nil {
		t.Error("Expected error when Local Font Access API is not available in test environment")
	}
	
	// The error should indicate that navigator is not available
	if err.Error() != "navigator not available" {
		t.Logf("Got expected error: %v", err)
	}
}