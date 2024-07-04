package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------
//				Select Mode
//-----------------------------------------------------------------
func (ea *editArea) SetSelectMode() {
	ea.selectRect.Init()
	ea.buttonPress = SelectButtonPress
	ea.buttonRelease = SelectButtonRelease
	ea.motionNotify = SelectMotionNotify
	ea.draw = SelectDraw
	ea.mode = M_SELECT
}

func SelectButtonPress(ea *editArea, ev *gdk.Event) bool {

	buttonEvent := gdk.EventButtonNewFromEvent(ev)

	if buttonEvent.Button() == gdk.BUTTON_PRIMARY {
		//fmt.Println("Press Left Mouse button")
		px, py := ea.mouse_to_pixel(buttonEvent.X(), buttonEvent.Y())
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
				if ea.InSelectRect(buttonEvent.X(), buttonEvent.Y()) {
					idHandle := ea.HitHandle(buttonEvent.X(), buttonEvent.Y())
					if idHandle != -1 {
						//-- Start Move Handle
						ea.selectRect.selCorner = idHandle
					} else {
						// Start Move SelectRect
						ea.selectRect.mouseStartX = buttonEvent.X()
						ea.selectRect.mouseStartY = buttonEvent.Y()
						ea.selectRect.BackupPosition()
					}
				} else {
					ea.selectRect.mode = 0
					ea.selectRect.SetCorner(0, curPx, curPy)
					ea.selectRect.SetCorner(2, curPx, curPy)
					ea.selectRect.selCorner = -1
					ea.QueueDraw()
				}

			case 2:
				if ea.InSelectRect(buttonEvent.X(), buttonEvent.Y()) {
					//--
					ea.selectRect.mouseStartX = buttonEvent.X()
					ea.selectRect.mouseStartY = buttonEvent.Y()
					ea.selectRect.BackupPosition()
					return true
				} else {
					ea.selectRect.mode = 0
					ea.selectRect.SetCorner(0, curPx, curPy)
					ea.selectRect.SetCorner(2, curPx, curPy)
					ea.selectRect.selCorner = -1
					return true
				}

			}

		}
		return true
	}

	return false
}

func SelectButtonRelease(ea *editArea, ev *gdk.Event) bool {

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

func SelectMotionNotify(ea *editArea, ev *gdk.Event) bool {
	eventMotion := gdk.EventMotionNewFromEvent(ev)

	//s3 := eventMotion.State() & gdk.ModifierType(gdk.BUTTON1_MASK)
	//s4 := eventMotion.State() & gdk.ModifierType(gdk.BUTTON3_MASK)
	x, y := eventMotion.MotionVal()
	px, py := ea.mouse_to_pixel(x, y)
	curPx := int(px)
	curPy := int(py)

	//fmt.Println(s3, s4)
	if ea.PtInEditArea(curPx, curPy) {
		var left, top, right, bottom float64

		switch ea.selectRect.mode {
		case 0:
			ea.selectRect.SetCorner(2, curPx, curPy)
			ea.QueueDraw()

		case 1:
			if ea.selectRect.selCorner != -1 {
				ea.selectRect.SetCorner(ea.selectRect.selCorner, curPx, curPy)
				ea.QueueDraw()
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
					ea.QueueDraw()
				}
			}

		case 2:
			mdx := x - ea.selectRect.mouseStartX
			mdy := y - ea.selectRect.mouseStartY
			dx, dy := ea.mouse_to_pixel(mdx, mdy)
			if dx != 0.0 || dy != 0.0 {
				mdx := x - ea.selectRect.mouseStartX
				mdy := y - ea.selectRect.mouseStartY
				dx, dy := ea.mouse_to_pixel(mdx, mdy)
				if dx != 0.0 || dy != 0.0 {
					left = float64(ea.selectRect.savleft) + dx
					top = float64(ea.selectRect.savtop) + dy
					right = float64(ea.selectRect.savright) + dx
					bottom = float64(ea.selectRect.savbottom) + dy
					// Prevent the rectangle to go out limits
					if left < 0.0 {
						right -= left
						left = 0
					} else if right >= float64(ea.nbPixelsW) {
						left -= (right - float64(ea.nbPixelsW))
						right = float64(ea.nbPixelsW)
					}
					if top < 0.0 {
						bottom -= top
						top = 0
					} else if bottom >= float64(ea.nbPixelsH) {
						top -= (bottom - float64(ea.nbPixelsH))
						bottom = float64(ea.selectRect.bottom)
					}
					ea.selectRect.SetCorner(0, int(left), int(top))
					ea.selectRect.SetCorner(2, int(right), int(bottom))

					ea.RestoreSprite()
					CopyArea(ea.imgBufCopy, 0, 0, ea.selectRect.Width(), ea.selectRect.Height(), ea.imgBuf, ea.selectRect.left, ea.selectRect.top)
					ea.QueueDraw()
					if ea.pixbufModify != nil {
						ea.pixbufModify()
					}

				}
			}
			return !ea.selectRect.IsEmpty()

		}

	}

	return false
}

func SelectDraw(ea *editArea, cr *cairo.Context) {
	//-- Draw Select Frame
	ea.DrawSelectRect(cr)

}
