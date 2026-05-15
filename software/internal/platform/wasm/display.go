// this is tinygo
package wasm

import "syscall/js"

type WasmDisplay struct {
	buf  [240 * 240]uint16
	cx   js.Value
	img  js.Value
	data js.Value
}

func NewDisplay(canvasID string) *WasmDisplay {
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", canvasID)
	ctx := canvas.Call("getContext", "2d")
	img := ctx.Call("createImageData", 240, 240)
	return &WasmDisplay{
		cx:   ctx,
		img:  img,
		data: img.Get("data"),
	}
}

func (d *WasmDisplay) Fill(color uint16) {
	for i := range d.buf {
		d.buf[i] = color
	}
}

func (d *WasmDisplay) Pixel(x, y int, c uint16) {
	if x < 0 || x >= 240 || y < 0 || y >= 240 {
		return
	}
	d.buf[y*240+x] = c
}

func (d *WasmDisplay) Present() error {
	rgba := make([]byte, 240*240*4)
	for i, px := range d.buf {
		r := byte((px >> 11) & 0x1f)
		g := byte((px >> 5) & 0x3f)
		b := byte(px & 0x1f)
		j := i * 4
		rgba[j+0] = (r << 3) | (r >> 2)
		rgba[j+1] = (g << 2) | (g >> 4)
		rgba[j+2] = (b << 3) | (b >> 2)
		rgba[j+3] = 0xff
	}
	js.CopyBytesToJS(d.data, rgba)
	d.cx.Call("putImageData", d.img, 0, 0)
	return nil
}
