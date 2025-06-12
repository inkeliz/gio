# Local Font Access API for Gio WASM

This document explains how to use the Local Font Access API implementation in Gio for WASM/JS builds to access system fonts.

## Overview

The Local Font Access API is a browser feature (currently supported in Chrome/Chromium) that allows web applications to access locally installed fonts. This implementation enables Gio applications running in WASM to use system fonts instead of being limited to web fonts or embedded fonts.

## Browser Support

- ✅ Chrome 103+ (stable support)
- ✅ Chromium-based browsers (Edge, Opera, etc.)
- ❌ Firefox (not supported yet)
- ❌ Safari (not supported yet)

## How It Works

1. When creating a text shaper with system fonts enabled on JS/WASM, Gio will:
   - First attempt to use the Local Font Access API if available
   - Fall back to traditional system font loading (which typically fails on WASM)
   - Continue with any explicitly provided font collections

2. The implementation:
   - Calls `navigator.fonts.query()` to enumerate available fonts
   - Validates font formats (OpenType, TrueType, TrueType Collection)
   - Uses localStorage caching to avoid reloading unchanged fonts
   - Loads each supported font as a blob using `font.blob()`
   - Converts the blob to an ArrayBuffer and copies it to Go
   - Parses the font data using the existing OpenType parser
   - Registers the fonts with the text shaper
   - Handles permission errors gracefully with retry capability

3. Caching and Performance:
   - Fonts are cached in localStorage with SHA256 hashes for validation
   - Cache entries expire after 7 days
   - Only supported font formats are processed to avoid parser errors
   - Font loading is limited to 50 fonts to prevent browser performance issues

## Usage

```go
// Enable system fonts (same API as other platforms)
shaper := text.NewShaper(text.WithCollection(gofont.Collection()))

// Or explicitly enable system fonts
shaper := text.NewShaper()  // System fonts enabled by default
```

The API is identical to other platforms - no special WASM-specific code is required.

## Building and Testing

1. Build your Gio application for WASM:
   ```bash
   GOOS=js GOARCH=wasm go build -o app.wasm your-app.go
   ```

2. Serve the WASM file with an HTTP server (required for security):
   ```html
   <!DOCTYPE html>
   <html>
   <head>
       <meta charset="utf-8">
       <title>Gio WASM App</title>
   </head>
   <body>
       <script src="wasm_exec.js"></script>
       <script>
           const go = new Go();
           WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject)
               .then((result) => {
                   go.run(result.instance);
               });
       </script>
   </body>
   </html>
   ```

3. Open in Chrome and grant font access permission when prompted.

## Permissions

The Local Font Access API requires user permission. The browser will show a permission prompt the first time your application tries to access local fonts. Users can:

- Allow: Grant access to local fonts
- Block: Deny access (fallback to embedded fonts)
- Allow this time: Grant temporary access

### Late Authorization

If the user initially denies font access but later wants to grant it, your application can retry font loading without restarting:

```go
// When user grants permission (e.g., through a UI button)
if shaper, ok := myShaper.(*text.shaperImpl); ok {
    if err := shaper.TryReloadSystemFonts(); err != nil {
        log.Printf("Failed to reload system fonts: %v", err)
    }
}
```

Note: The `TryReloadSystemFonts()` method is only available on JS/WASM builds and will return an error on other platforms.

## Limitations

- **Performance**: Loading many fonts can be slow. The implementation limits to 50 fonts by default.
- **Browser support**: Only works in Chromium-based browsers.
- **Permissions**: Requires user consent.
- **Security**: Only works over HTTPS or localhost.

## Error Handling

The implementation handles various error conditions gracefully:

- API not available: Falls back to embedded fonts
- Permission denied: Falls back to embedded fonts  
- Network errors: Skips problematic fonts, continues with others
- Parse errors: Skips invalid fonts, continues with others

All errors are logged using the existing Gio logging system.

## Example

See `example_font_access.go` for a complete working example.