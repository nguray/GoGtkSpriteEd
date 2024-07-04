package main

import (
	"fmt"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------
//				Rectangle Mode
//-----------------------------------------------------------------

func draw_rectangle(imgBuf *gdk.Pixbuf, x0, y0, x1, y1 int, c uint32) {

	var (
		startX, endX int
		startY, endY int
	)
	if x1 > x0 {
		startX, endX = x0, x1
	} else {
		startX, endX = x1, x0

	}

	if y1 > y0 {
		startY, endY = y0, y1
	} else {
		startY, endY = y1, y0
	}

	nChannels := imgBuf.GetNChannels()
	rowStride := imgBuf.GetRowstride()

	r, g, b, a := getRGBA(c)

	pixs := imgBuf.GetPixels()

	putpixel := func(x, y int) {
		iPix := y*rowStride + x*nChannels
		pixs[iPix] = byte(r)   // Red
		pixs[iPix+1] = byte(g) // Green
		pixs[iPix+2] = byte(b) // Blue
		pixs[iPix+3] = byte(a) // Alpha

	}

	if (endX != startX) || (endY != startY) {
		//-- Tracer le rectangle
		for x := startX; x <= endX; x++ {
			putpixel(x, startY)
			putpixel(x, endY)
		}
		y1 := startY + 1
		y2 := endY - 1
		for y := y1; y <= y2; y++ {
			putpixel(startX, y)
			putpixel(endX, y)
		}
	}

}

func fill_rectangle(imgBuf *gdk.Pixbuf, x0, y0, x1, y1 int, c uint32) {

	var (
		startX, endX int
		startY, endY int
	)
	if x1 > x0 {
		startX, endX = x0, x1
	} else {
		startX, endX = x1, x0

	}

	if y1 > y0 {
		startY, endY = y0, y1
	} else {
		startY, endY = y1, y0
	}

	nChannels := imgBuf.GetNChannels()
	rowStride := imgBuf.GetRowstride()

	r, g, b, a := getRGBA(c)

	pixs := imgBuf.GetPixels()

	putpixel := func(x, y int) {
		iPix := y*rowStride + x*nChannels
		pixs[iPix] = byte(r)   // Red
		pixs[iPix+1] = byte(g) // Green
		pixs[iPix+2] = byte(b) // Blue
		pixs[iPix+3] = byte(a) // Alpha

	}

	if (endX != startX) || (endY != startY) {
		//-- Tracer le rectangle
		for y := startY; y <= endY; y++ {
			for x := startX; x <= endX; x++ {
				putpixel(x, y)
			}
		}
	}

}

func (ea *editArea) InitRectangleMode() {
	ea.selectRect.Init()

}

func (ea *editArea) SetRectangleMode() {
	ea.selectRect.Init()
	ea.buttonPress = RectangleButtonPress
	ea.buttonRelease = RectangleButtonRelease
	ea.motionNotify = RectangleMotionNotify
	ea.draw = RectangleDraw
	ea.mode = M_RECTANGLE
}

func (ea *editArea) HitHandle(mx, my float64) int {
	for i := 0; i < 4; i++ {
		cx, cy := ea.selectRect.GetCorner(i)
		x, y := ea.pixel_to_mouse(cx, cy)
		if mx >= x && mx <= (x+ea.cellSize) &&
			my >= y && my <= (y+ea.cellSize) {
			return i
		}
	}
	return -1
}

func (ea *editArea) InSelectRect(mx, my float64) bool {
	px1, py1 := ea.selectRect.GetCorner(0)
	x1, y1 := ea.pixel_to_mouse(px1, py1)
	px2, py2 := ea.selectRect.GetCorner(2)
	x2, y2 := ea.pixel_to_mouse(px2, py2)

	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	y2 += ea.cellSize
	x2 += ea.cellSize
	return (mx >= x1) && (mx <= x2) && (my >= y1) && (my <= y2)
}

func RectangleButtonPress(ea *editArea, ev *gdk.Event) bool {

	buttonEvent := gdk.EventButtonNewFromEvent(ev)
	tmx := buttonEvent.X() - ea.origin_x
	tmy := buttonEvent.Y() - ea.origin_y

	if buttonEvent.Button() == gdk.BUTTON_PRIMARY {
		//fmt.Println("Press Left Mouse button")
		px, py := ea.mouse_to_pixel(tmx, tmy)
		//fmt.Printf("px=%f py=%f\n", px, py)
		curPx := int(px)
		curPy := int(py)
		if ea.PtInEditArea(curPx, curPy) {

			switch ea.selectRect.mode {
			case 0:
				ea.BackupSprite()
				ea.undo_mode = RECTANGLE
				ea.selectRect.SetCorner(0, curPx, curPy)
				ea.selectRect.SetCorner(2, curPx, curPy)
				ea.QueueDraw()

			case 1:
				if ea.InSelectRect(tmx, tmy) {
					idHandle := ea.HitHandle(tmx, tmy)
					if idHandle != -1 {
						//-- Start Move Handle
						ea.selectRect.selCorner = idHandle
					} else {
						// Start Move SelectRect
						ea.selectRect.mouseStartX = tmx
						ea.selectRect.mouseStartY = tmy
						ea.selectRect.BackupPosition()
					}
				} else {
					ea.selectRect.mode = 0
					ea.selectRect.SetCorner(0, curPx, curPy)
					ea.selectRect.SetCorner(2, curPx, curPy)
					ea.selectRect.selCorner = -1
					ea.QueueDraw()
				}

			}

		}
		return true
	}

	return false
}

func RectangleButtonRelease(ea *editArea, ev *gdk.Event) bool {

	if ea.selectRect.mode == 0 {
		x1, y1 := ea.selectRect.GetCorner(0)
		x2, y2 := ea.selectRect.GetCorner(2)
		if x1 != x2 && y1 != y2 {
			ea.selectRect.Normalize()
			ea.selectRect.mode = 1
		} else {
			ea.selectRect.Init()
		}
		ea.QueueDraw()
	}
	//--
	ea.selectRect.selCorner = -1

	// buttonEvent := gdk.EventButtonNewFromEvent(ev)
	// if buttonEvent.Button() == gdk.BUTTON_PRIMARY {
	// 	//fmt.Println("Release Left Mouse button")
	// 	px, py := ea.mouse_to_pixel(buttonEvent.X(), buttonEvent.Y())
	// 	curPx := int(px)
	// 	curPy := int(py)
	// 	fmt.Printf("px=%d py=%d\n", curPx, curPy)

	// 	// if ea.PtInEditArea(curPx, curPy) {
	// 	// 	if !ea.fColorPick {
	// 	// 		ea.lastDrawPx = curPx
	// 	// 		ea.lastDrawPy = curPy
	// 	// 	}
	// 	// }
	// 	return true
	// }
	return false
}

func RectangleMotionNotify(ea *editArea, ev *gdk.Event) bool {
	eventMotion := gdk.EventMotionNewFromEvent(ev)
	x, y := eventMotion.MotionVal()
	tmx := x - ea.origin_x
	tmy := y - ea.origin_y

	s3 := eventMotion.State() & gdk.ModifierType(gdk.BUTTON1_MASK)
	s4 := eventMotion.State() & gdk.ModifierType(gdk.BUTTON3_MASK)
	px, py := ea.mouse_to_pixel(tmx, tmy)
	curPx := int(px)
	curPy := int(py)

	fmt.Println(s3, s4)
	if ea.PtInEditArea(curPx, curPy) {
		var left, top, right, bottom float64

		switch ea.selectRect.mode {
		case 0:
			ea.selectRect.SetCorner(2, curPx, curPy)

		case 1:
			if ea.selectRect.selCorner != -1 {
				ea.selectRect.SetCorner(ea.selectRect.selCorner, curPx, curPy)
			} else {
				mdx := x - ea.selectRect.mouseStartX
				mdy := y - ea.selectRect.mouseStartY
				dx, dy := ea.mouse_to_pixel(mdx, mdy)
				if dx != 0.0 || dy != 0.0 {
					left = float64(ea.selectRect.savleft) + dx
					top = float64(ea.selectRect.savtop) + dy
					right = float64(ea.selectRect.savright) + dx
					bottom = float64(ea.selectRect.savbottom) + dy
					// Prevent the rectangle to go out limits
					if (left < 0.0) || (right >= float64(ea.nbPixelsW)) {
						left = float64(ea.selectRect.left)
						right = float64(ea.selectRect.right)
					}
					if (top < 0.0) || (bottom >= float64(ea.nbPixelsH)) {
						top = float64(ea.selectRect.top)
						bottom = float64(ea.selectRect.bottom)
					}
					ea.selectRect.SetCorner(0, int(left), int(top))
					ea.selectRect.SetCorner(2, int(right), int(bottom))

				}
			}

		}

		if !ea.selectRect.IsEmpty() {
			ea.RestoreSprite()

			var x0, y0, x1, y1 int

			if ea.selectRect.right > ea.selectRect.left {
				x0, x1 = ea.selectRect.left, ea.selectRect.right
			} else {
				x1, x0 = ea.selectRect.left, ea.selectRect.right
			}

			if ea.selectRect.top > ea.selectRect.bottom {
				y0, y1 = ea.selectRect.bottom, ea.selectRect.top
			} else {
				y1, y0 = ea.selectRect.bottom, ea.selectRect.top
			}

			if (eventMotion.State() & gdk.CONTROL_MASK) != 0 {
				// Fill Rectangle
				//fmt.Println("Draw Fill Rectangle")
				fill_rectangle(ea.imgBuf, x0, y0, x1, y1, ea.foreColor)
			} else {
				// Frame Rectangle
				//fmt.Println("Draw Frame Rectangle")
				draw_rectangle(ea.imgBuf, x0, y0, x1, y1, ea.foreColor)
			}
			ea.QueueDraw()
			if ea.pixbufModify != nil {
				ea.pixbufModify()
			}
		}

	}

	return false
}

func RectangleDraw(ea *editArea, cr *cairo.Context) {
	//-- Draw Select Frame
	ea.DrawSelectRect(cr)

}
