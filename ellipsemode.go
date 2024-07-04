package main

import (
	"fmt"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------
//				Ellipse Mode
//-----------------------------------------------------------------
func (ea *editArea) InitEllipseMode() {
	ea.selectRect.Init()

}

func (ea *editArea) SetEllipseMode() {
	ea.selectRect.Init()
	ea.buttonPress = EllipseButtonPress
	ea.buttonRelease = EllipseButtonRelease
	ea.motionNotify = EllipseMotionNotify
	ea.draw = EllipseDraw
	ea.mode = M_ELLIPSE
}

func FillEllipse(imgBuf *gdk.Pixbuf, left, top, right, bottom int, c uint32) {

	//---------------------------------------------
	a := (right - left) / 2
	b := (bottom - top) / 2

	x := 0
	y := b
	a2 := a * a
	b2 := b * b
	a2b2 := a2 + b2
	a2sqr := a2 + a2
	b2sqr := b2 + b2
	a4sqr := a2sqr + a2sqr
	b4sqr := b2sqr + b2sqr
	a8sqr := a4sqr + a4sqr
	b8sqr := b4sqr + b4sqr
	a4sqr_b4sqr := a4sqr + b4sqr
	_fn := a8sqr + a4sqr
	_fnn := a8sqr
	_fnnw := a8sqr
	_fnw := a8sqr + a4sqr - b8sqr*a + b8sqr
	_fnwn := a8sqr
	_fnwnw := a8sqr + b8sqr
	_fnww := b8sqr
	_fwnw := b8sqr
	_fww := b8sqr
	d1 := b2 - b4sqr*a + a4sqr

	red, green, blue, alpha := getRGBA(c)

	nChannels := imgBuf.GetNChannels()
	rowStride := imgBuf.GetRowstride()
	pixs := imgBuf.GetPixels()

	draw_horizontal_line := func(xLeft, xRight, y int) {

		for x := xLeft; x <= xRight; x++ {

			iPix := y*rowStride + x*nChannels
			pixs[iPix] = byte(red)     // Red
			pixs[iPix+1] = byte(green) // Green
			pixs[iPix+2] = byte(blue)  // Blue
			pixs[iPix+3] = byte(alpha) // Alpha

		}

	}

	for (_fnw < a2b2) || (d1 < 0) || (((_fnw - _fn) > b2) && (y > 0)) {

		draw_horizontal_line(left+x, right-x, top+y)
		//DrawHorizontalLine( pixbuf, left + x, right - x, top + y, col);
		draw_horizontal_line(left+x, right-x, bottom-y)
		//DrawHorizontalLine( pixbuf, left + x, right - x, bottom - y, col);

		y -= 1
		if (d1 < 0) || ((_fnw - _fn) > b2) {
			d1 = d1 + _fn
			_fn = _fn + _fnn
			_fnw = _fnw + _fnwn
		} else {
			x += 1
			d1 = d1 + _fnw
			_fn = _fn + _fnnw
			_fnw = _fnw + _fnwnw
		}

	}

	_fw := _fnw - _fn + b4sqr
	d2 := d1 + (_fw+_fw-_fn-_fn+a4sqr_b4sqr+a8sqr)/4
	_fnw = _fnw + (b4sqr - a4sqr)

	old_y := y + 1

	for x <= a {
		if y != old_y {
			draw_horizontal_line(left+x, right-x, top+y)
			//DrawHorizontalLine( pixbuf, left + x, right - x, top + y, col);
			draw_horizontal_line(left+x, right-x, bottom-y)
			//DrawHorizontalLine( pixbuf, left + x, right - x, bottom - y, col);
		}
		old_y = y

		x += 1
		if d2 < 0 {
			y -= 1
			d2 = d2 + _fnw
			_fw = _fw + _fwnw
			_fnw = _fnw + _fnwnw
		} else {
			d2 = d2 + _fw
			_fw = _fw + _fww
			_fnw = _fnw + _fnww
		}

	}

}

func BorderEllipse(imgBuf *gdk.Pixbuf, left, top, right, bottom int, c uint32) {

	a := (right - left) / 2
	b := (bottom - top) / 2

	var x int = 0
	y := b

	a2 := a * a
	b2 := b * b
	a2b2 := a2 + b2
	a2sqr := a2 + a2
	b2sqr := b2 + b2
	a4sqr := a2sqr + a2sqr
	b4sqr := b2sqr + b2sqr
	a8sqr := a4sqr + a4sqr
	b8sqr := b4sqr + b4sqr
	a4sqr_b4sqr := a4sqr + b4sqr

	_fn := a8sqr + a4sqr
	_fnn := a8sqr
	_fnnw := a8sqr
	_fnw := a8sqr + a4sqr - b8sqr*a + b8sqr
	_fnwn := a8sqr
	_fnwnw := a8sqr + b8sqr
	_fnww := b8sqr
	_fwnw := b8sqr
	_fww := b8sqr
	d1 := b2 - b4sqr*a + a4sqr

	red, green, blue, alpha := getRGBA(c)

	nChannels := imgBuf.GetNChannels()
	rowStride := imgBuf.GetRowstride()
	pixs := imgBuf.GetPixels()

	putpixel := func(x, y int) {
		iPix := y*rowStride + x*nChannels
		pixs[iPix] = byte(red)     // Red
		pixs[iPix+1] = byte(green) // Green
		pixs[iPix+2] = byte(blue)  // Blue
		pixs[iPix+3] = byte(alpha) // Alpha
	}

	for (_fnw < a2b2) || (d1 < 0) || (((_fnw - _fn) > b2) && (y > 0)) {

		//put_pixel(pixbuf, left + x, top + y, red, green, blue, alpha);
		putpixel(left+x, top+y)
		//put_pixel(pixbuf, right - x, top + y, red, green, blue, alpha);
		putpixel(right-x, top+y)
		//put_pixel(pixbuf, left + x, bottom - y, red, green, blue, alpha);
		putpixel(left+x, bottom-y)
		//put_pixel(pixbuf, right - x, bottom - y, red, green, blue, alpha);
		putpixel(right-x, bottom-y)

		y -= 1
		if (d1 < 0) || ((_fnw - _fn) > b2) {
			d1 += _fn
			_fn += _fnn
			_fnw += _fnwn
		} else {
			x += 1
			d1 += _fnw
			_fn += _fnnw
			_fnw += _fnwnw
		}
	}

	_fw := _fnw - _fn + b4sqr
	d2 := d1 + (_fw+_fw-_fn-_fn+a4sqr_b4sqr+a8sqr)/4
	_fnw = _fnw + (b4sqr - a4sqr)

	for x <= a {

		//put_pixel(pixbuf, left + x, top + y, red, green, blue, alpha);
		putpixel(left+x, top+y)
		//put_pixel(pixbuf, right - x, top + y, red, green, blue, alpha);
		putpixel(right-x, top+y)
		//put_pixel(pixbuf, left + x, bottom - y, red, green, blue, alpha);
		putpixel(left+x, bottom-y)
		//put_pixel(pixbuf, right - x, bottom - y, red, green, blue, alpha);
		putpixel(right-x, bottom-y)

		x += 1
		if d2 < 0 {
			y -= 1
			d2 += _fnw
			_fw += _fwnw
			_fnw += _fnwnw
		} else {
			d2 += _fw
			_fw += _fww
			_fnw += _fnww
		}
	}

}

func EllipseButtonPress(ea *editArea, ev *gdk.Event) bool {

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

func EllipseButtonRelease(ea *editArea, ev *gdk.Event) bool {

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

func EllipseMotionNotify(ea *editArea, ev *gdk.Event) bool {
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
				// Fill Ellipse
				//fmt.Println("Draw Fill Rectangle")
				FillEllipse(ea.imgBuf, x0, y0, x1, y1, ea.foreColor)
			} else {
				// Frame Ellipse
				//fmt.Println("Draw Frame Rectangle")
				BorderEllipse(ea.imgBuf, x0, y0, x1, y1, ea.foreColor)
			}
			ea.QueueDraw()
			if ea.pixbufModify != nil {
				ea.pixbufModify()
			}
		}

	}

	return false
}

func EllipseDraw(ea *editArea, cr *cairo.Context) {
	//-- Draw Select Frame
	ea.DrawSelectRect(cr)

}
