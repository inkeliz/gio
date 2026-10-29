// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"slices"

	"gioui.org/f32"
	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/ops"
)

const rasterCacheBytes = 8 << 20

// Require sustained reuse before paying for a dedicated layer texture.
const minStableFrames = 8

type rasterPaint struct {
	clip     image.Rectangle
	material material
	path     opKey
	offset   f32.Point
	kind     uint8
}
type rasterEntry struct {
	texture driver.Texture
	size    image.Point
	paints  []rasterPaint
	used    uint64
	stable  uint8
	valid   bool
}
type rasterCache struct {
	entries map[ops.Key]*rasterEntry
	free    [64]*rasterEntry
	nfree   int
	scratch []rasterPaint
	frame   uint64
	bytes   int
}

func (c *rasterCache) release() {
	for _, e := range c.entries {
		if e.texture != nil {
			e.texture.Release()
		}
	}
	*c = rasterCache{}
}
func (c *rasterCache) remove(key ops.Key, e *rasterEntry) {
	if e.texture != nil {
		e.texture.Release()
		c.bytes -= e.size.X * e.size.Y * 4
		e.texture = nil
	}
	delete(c.entries, key)
	clear(e.paints)
	e.paints = e.paints[:0]
	e.valid, e.stable, e.used = false, 0, 0
	if c.nfree < len(c.free) {
		c.free[c.nfree] = e
		c.nfree++
	}
}
func (c *rasterCache) evict() bool {
	var key ops.Key
	var oldest *rasterEntry
	for k, e := range c.entries {
		if e.used != c.frame && (oldest == nil || e.used < oldest.used) {
			key, oldest = k, e
		}
	}
	if oldest == nil {
		return false
	}
	c.remove(key, oldest)
	return true
}

// Exact paint and clipping signatures validate hits; source operation keys
// provide lookup only. Eight matching observations on separate frames precede
// GPU allocation. Unstable groups keep the ordinary layer path, and metadata
// and texture storage are reused instead of recreated on every invalidation.
func (c *rasterCache) prepare(ctx driver.Device, layers []opacityLayer, paints []imageOp, disabled bool) {
	c.frame++
	for k, e := range c.entries {
		if e.used+120 < c.frame {
			c.remove(k, e)
		}
	}
	if disabled {
		return
	}
	for i := range layers {
		l := &layers[i]
		size := l.clip.Size()
		if size.X <= 0 || size.Y <= 0 || size.X > 1024 || size.Y > 1024 || l.opEnd-l.opStart < 3 {
			continue
		}
		// Collection records layers in preorder, so a first child is next.
		if i+1 < len(layers) && layers[i+1].parent == i {
			continue
		}
		clear(c.scratch)
		c.scratch = c.scratch[:0]
		origin := l.clip.Min
		for _, img := range paints[l.opStart:l.opEnd] {
			if img.culled {
				continue
			}
			c.scratch = append(c.scratch, rasterPaint{clip: img.clip.Sub(origin), material: img.material})
			for p := img.path; p != nil; p = p.parent {
				if p.path {
					c.scratch = append(c.scratch, rasterPaint{kind: 1, path: p.pathKey, offset: p.off.Sub(f32.Pt(float32(origin.X), float32(origin.Y)))})
				}
			}
		}
		if len(c.scratch) < 3 {
			continue
		}
		e := c.entries[l.key]
		match := e != nil && e.size == size && slices.Equal(e.paints, c.scratch)
		if e != nil && e.used == c.frame && !match {
			continue
		}
		if e == nil {
			if len(c.entries) >= len(c.free) && !c.evict() {
				continue
			}
			if c.entries == nil {
				c.entries = make(map[ops.Key]*rasterEntry, len(c.free))
			}
			if c.nfree != 0 {
				c.nfree--
				e = c.free[c.nfree]
				c.free[c.nfree] = nil
			} else {
				e = new(rasterEntry)
			}
			c.entries[l.key] = e
		} else if !match {
		}
		if !match {
			if e.texture != nil && e.size != size {
				e.texture.Release()
				c.bytes -= e.size.X * e.size.Y * 4
				e.texture = nil
			}
			e.size = size
			clear(e.paints)
			e.paints = append(e.paints[:0], c.scratch...)
			e.valid, e.stable = false, 1
		} else if e.used != c.frame && e.stable < minStableFrames {
			e.stable++
		}
		e.used = c.frame
		if e.stable < minStableFrames {
			continue
		}
		if e.texture == nil {
			need := size.X * size.Y * 4
			for c.bytes+need > rasterCacheBytes && c.evict() {
			}
			if c.bytes+need > rasterCacheBytes {
				continue
			}
			tex, err := ctx.NewTexture(driver.TextureFormatSRGBA, size.X, size.Y, driver.FilterNearest, driver.FilterNearest, driver.BufferBindingTexture|driver.BufferBindingFramebuffer)
			if err != nil {
				continue
			}
			e.texture = tex
			c.bytes += need
		}
		l.cache = e
		if e.valid {
			for j := l.opStart; j < l.opEnd; j++ {
				paints[j].culled = true
			}
		} else {
		}
	}
}
