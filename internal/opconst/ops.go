// SPDX-License-Identifier: Unlicense OR MIT

package opconst

type OpType byte

// Start at a high number for easier debugging.
const firstOpIndex = 200

const (
	TypeMacro OpType = iota + firstOpIndex
	TypeCall
	TypeTransform
	TypeLayer
	TypeInvalidate
	TypeImage
	TypePaint
	TypeColor
	TypeLinearGradient
	TypeArea
	TypePointerInput
	TypePass
	TypeWriteClipboard
	TypeReadClipboard
	TypeKeyInput
	TypeKeyFocus
	TypeKeySoftKeyboard
	TypePush
	TypePop
	TypeAux
	TypeClip
	TypeProfile
)

const (
	TypeMacroLen           = 1 + 4 + 4
	TypeCallLen            = 1 + 4 + 4
	TypeTransformLen       = 1 + 4*6
	TypeLayerLen           = 1
	TypeRedrawLen          = 1 + 8
	TypeImageLen           = 1
	TypePaintLen           = 1
	TypeColorLen           = 1 + 4
	TypeLinearGradientLen  = 1 + 8*2 + 4*2
	TypeAreaLen            = 1 + 1 + 4*4
	TypePointerInputLen    = 1 + 1 + 1
	TypePassLen            = 1 + 1
	TypeWriteClipboardLen  = 1
	TypeReadClipboardLen   = 1
	TypeKeyInputLen        = 1
	TypeKeyFocusLen        = 1 + 1
	TypeKeySoftKeyboardLen = 1 + 1
	TypePushLen            = 1
	TypePopLen             = 1
	TypeAuxLen             = 1
	TypeClipLen            = 1 + 4*4 + 4 + 2 + 4
	TypeProfileLen         = 1
)

func (t OpType) Size() int {
	return [...]int{
		TypeMacroLen,
		TypeCallLen,
		TypeTransformLen,
		TypeLayerLen,
		TypeRedrawLen,
		TypeImageLen,
		TypePaintLen,
		TypeColorLen,
		TypeLinearGradientLen,
		TypeAreaLen,
		TypePointerInputLen,
		TypePassLen,
		TypeWriteClipboardLen,
		TypeReadClipboardLen,
		TypeKeyInputLen,
		TypeKeyFocusLen,
		TypeKeySoftKeyboardLen,
		TypePushLen,
		TypePopLen,
		TypeAuxLen,
		TypeClipLen,
		TypeProfileLen,
	}[t-firstOpIndex]
}

func (t OpType) NumRefs() int {
	switch t {
	case TypeKeyInput, TypePointerInput, TypeProfile, TypeCall, TypeReadClipboard, TypeWriteClipboard:
		return 1
	case TypeImage:
		return 2
	default:
		return 0
	}
}
