package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------
//				Fill Mode
//-----------------------------------------------------------------
func (ea *editArea) SetFillMode() {
	ea.selectRect.Init()
	ea.buttonPress = FillButtonPress
	ea.buttonRelease = FillButtonRelease
	ea.motionNotify = FillMotionNotify
	ea.draw = FillDraw
	ea.mode = M_FILL

}

type IntPt struct {
	x, y int
}

type StackIntPt []*IntPt

func (s *StackIntPt) IsEmpty() bool {
	return len(*s) == 0
}

func (s *StackIntPt) Len() int {
	return len(*s)
}

func (s *StackIntPt) Push(pt *IntPt) {
	*s = append(*s, pt)
}

func (s *StackIntPt) Pop() (*IntPt, bool) {
	if s.IsEmpty() {
		return nil, false
	} else {
		index := len(*s) - 1
		elem := (*s)[index]
		*s = (*s)[:index]
		return elem, true
	}
}

func FloodFill(imgBuf *gdk.Pixbuf, px, py int, c uint32) {

	var (
		x            int
		y            int
		fStartNord   bool = false
		fStartSud    bool = false
		xStartLine   int  = 0
		xEndLine     int  = 0
		fNord        bool = false
		fSud         bool = false
		pix_color    uint32
		f            bool
		target_color uint32
		start_x      int
		start_y      int
	)

	width := imgBuf.GetWidth()
	height := imgBuf.GetHeight()
	nChannels := imgBuf.GetNChannels()
	rowStride := imgBuf.GetRowstride()
	pixs := imgBuf.GetPixels()

	getpixel := func(gpx, gpy int) (uint32, bool) {
		if gpx >= 0 && gpx < width && gpy >= 0 && gpy < height {
			iPix := gpy*rowStride + gpx*nChannels
			r := uint32(pixs[iPix])   // Red
			g := uint32(pixs[iPix+1]) // Green
			b := uint32(pixs[iPix+2]) // Blue
			a := uint32(pixs[iPix+3]) // Alpha
			return RGBA(r, g, b, a), true

		}
		return 0, false
	}

	target_color, f = getpixel(px, py)
	if !f || target_color == c {
		return
	}

	red, green, blue, alpha := getRGBA(c)

	putpixel := func(ppx, ppy int) {
		iPix := ppy*rowStride + ppx*nChannels
		pixs[iPix] = byte(red)     // Red
		pixs[iPix+1] = byte(green) // Green
		pixs[iPix+2] = byte(blue)  // Blue
		pixs[iPix+3] = byte(alpha) // Alpha
	}

	//-- Créer la pile
	stk := new(StackIntPt)

	var pt *IntPt

	pt = &IntPt{px, py}
	stk.Push(pt)

	for stk.Len() != 0 {

		pt, f = stk.Pop()
		if !f {
			break
		}
		start_x = pt.x
		start_y = pt.y

		//-- Vérifier au Nord
		fStartNord = false
		if start_y > 0 {
			pix_color, f = getpixel(start_x, start_y-1)
			if f {
				if pix_color == target_color {
					pt = &IntPt{start_x, start_y - 1}
					stk.Push(pt)
					fStartNord = true
				}
			}
		}

		//-- Vérifier au sud
		fStartSud = false
		if start_y < (height - 1) {
			pix_color, f = getpixel(start_x, start_y+1)
			if f {
				if pix_color == target_color {
					pt = &IntPt{start_x, start_y + 1}
					stk.Push(pt)
					fStartSud = true
				}
			}
		}

		y = start_y

		//-- Aller vers l'est
		xEndLine = start_x
		fNord = fStartNord
		fSud = fStartSud
		if xEndLine < (width - 1) {

			x = xEndLine + 1
			for {
				pix_color, f = getpixel(x, y)
				if !f || pix_color != target_color {
					break
				}

				//-- Vérifier au Nord
				pix_color, f = getpixel(x, y-1)
				if f && (y > 0) {
					if target_color == pix_color {
						if !fNord {
							pt = &IntPt{x, y - 1}
							stk.Push(pt)
							fNord = true
						}
					} else {
						fNord = false
					}
				} else {
					fNord = false
				}

				//-- Vérifier au sud
				pix_color, f = getpixel(x, y+1)
				if f && (y < (height - 1)) {
					if target_color == pix_color {
						if !fSud {
							pt = &IntPt{x, y + 1}
							stk.Push(pt)
							fSud = true
						}
					} else {
						fSud = false
					}
				} else {
					fSud = false
				}

				xEndLine = x
				x += 1
				if x >= width {
					break
				}
			}

		} else {
			xEndLine = width - 1

		}

		//-- Aller vers l'ouest
		xStartLine = start_x
		fNord = fStartNord
		fSud = fStartSud
		if xStartLine > 0 {

			x = xStartLine - 1

			for {

				pix_color, f = getpixel(x, y)
				if !f || (pix_color != target_color) {
					break
				}

				//-- Vérifier au Nord
				pix_color, f = getpixel(x, y-1)
				if f && (y > 0) {
					if target_color == pix_color {
						if !fNord {
							pt = &IntPt{x, y - 1}
							stk.Push(pt)
							fNord = true
						}
					} else {
						fNord = false
					}
				} else {
					fNord = false
				}

				//-- Vérifier au sud
				pix_color, f = getpixel(x, y+1)
				if (y < (height - 1)) && f {
					if target_color == pix_color {
						if !fSud {
							pt = &IntPt{x, y + 1}
							stk.Push(pt)
							fSud = true
						}
					} else {
						fSud = false
					}
				} else {
					fSud = false
				}

				xStartLine = x
				x -= 1
				if x < 0 {
					break
				}
			}

		} else {
			xStartLine = 0
		}

		//-- Tracer la line
		for x := xStartLine; x <= xEndLine; x++ {
			putpixel(x, y)

		}

	}

}

func FillButtonPress(ea *editArea, ev *gdk.Event) bool {

	buttonEvent := gdk.EventButtonNewFromEvent(ev)
	tmx := buttonEvent.X() - ea.origin_x
	tmy := buttonEvent.Y() - ea.origin_y

	//ea.Emit("test-signal")
	if buttonEvent.Button() == gdk.BUTTON_PRIMARY {
		//fmt.Println("Press Left Mouse button")
		px, py := ea.mouse_to_pixel(tmx, tmy)
		//fmt.Printf("px=%f py=%f\n", px, py)
		curPx := int(px)
		curPy := int(py)

		ea.fColorPick = false
		if ea.PtInEditArea(curPx, curPy) {
			ea.BackupSprite()
			if (int(buttonEvent.State()) & int(gdk.SHIFT_MASK)) != 0 {
				// Pick Foreground Color
				ea.foreColor = GetPixel(ea.imgBuf, curPx, curPy)
				if ea.colorPick != nil {
					ea.colorPick(ea.foreColor, ea.backColor)
				}
				ea.fColorPick = true
			} else {
				ea.lastPx = curPx
				ea.lastPy = curPy
				ea.lastDrawPx = curPx
				ea.lastDrawPy = curPy
				FloodFill(ea.imgBuf, curPx, curPy, ea.foreColor)
				ea.QueueDraw()
				if ea.pixbufModify != nil {
					ea.pixbufModify()
				}
			}
		}
		//ea.mouseClicked(ea.lastPx, ea.lastPy)
		return true
	} else if buttonEvent.Button() == gdk.BUTTON_SECONDARY {
		px, py := ea.mouse_to_pixel(tmx, tmy)
		//fmt.Printf("px=%f py=%f\n", px, py)
		curPx := int(px)
		curPy := int(py)
		ea.fColorPick = false
		if ea.PtInEditArea(curPx, curPy) {
			if (int(buttonEvent.State()) & int(gdk.SHIFT_MASK)) != 0 {
				// Pick Background Color
				ea.backColor = GetPixel(ea.imgBuf, curPx, curPy)
				if ea.colorPick != nil {
					ea.colorPick(ea.foreColor, ea.backColor)
				}
				ea.fColorPick = true
			} else {
				ea.lastPx = curPx
				ea.lastPy = curPy
				FloodFill(ea.imgBuf, curPx, curPy, ea.backColor)
				ea.QueueDraw()
				if ea.pixbufModify != nil {
					ea.pixbufModify()
				}
			}
		}

	}
	return false

}

func FillButtonRelease(ea *editArea, ev *gdk.Event) bool {

	return false
}

func FillDraw(ea *editArea, cr *cairo.Context) {
	// Nothing to draw
}

func FillMotionNotify(ea *editArea, ev *gdk.Event) bool {

	return false
}
