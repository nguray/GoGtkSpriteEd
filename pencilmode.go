package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------
//				Pencil Mode
//-----------------------------------------------------------------

func (ea *editArea) SetPencilMode() {
	ea.buttonPress = PencilButtonPress
	ea.buttonRelease = PencilButtonRelease
	ea.draw = PencilDraw
	ea.motionNotify = PencilMotionNotify
	ea.mode = M_PENCIL
}

func PencilButtonPress(ea *editArea, ev *gdk.Event) bool {

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
			ea.undo_mode = PENCIL
			if (int(buttonEvent.State()) & int(gdk.CONTROL_MASK)) != 0 {
				// Draw Line from prevDraw pixel
				if ea.lastDrawPx >= 0 && ea.lastDrawPy >= 0 {
					//fmt.Println("Line to")
					Line(ea.imgBuf, ea.lastDrawPx, ea.lastDrawPy, curPx, curPy, ea.foreColor)
					ea.QueueDraw()
					if ea.pixbufModify != nil {
						ea.pixbufModify()
					}
				}
			} else if (int(buttonEvent.State()) & int(gdk.SHIFT_MASK)) != 0 {
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
				SetPixel(ea.imgBuf, ea.lastPx, ea.lastPy, ea.foreColor)
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
			ea.BackupSprite()
			ea.undo_mode = PENCIL
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
				SetPixel(ea.imgBuf, ea.lastPx, ea.lastPy, ea.backColor)
				ea.QueueDraw()
				if ea.pixbufModify != nil {
					ea.pixbufModify()
				}
			}
		}

	}
	return false

}

func PencilButtonRelease(ea *editArea, ev *gdk.Event) bool {
	buttonEvent := gdk.EventButtonNewFromEvent(ev)
	tmx := buttonEvent.X() - ea.origin_x
	tmy := buttonEvent.Y() - ea.origin_y
	if buttonEvent.Button() == gdk.BUTTON_PRIMARY || buttonEvent.Button() == gdk.BUTTON_SECONDARY {
		//fmt.Println("Release Left Mouse button")
		px, py := ea.mouse_to_pixel(tmx, tmy)
		curPx := int(px)
		curPy := int(py)
		if ea.PtInEditArea(curPx, curPy) {
			if !ea.fColorPick {
				ea.lastDrawPx = curPx
				ea.lastDrawPy = curPy
			}
		}
		return true
	}
	return false
}

func PencilDraw(ea *editArea, cr *cairo.Context) {
	// Nothing to draw
}

func PencilMotionNotify(ea *editArea, ev *gdk.Event) bool {
	eventMotion := gdk.EventMotionNewFromEvent(ev)
	x, y := eventMotion.MotionVal()
	tmx := x - ea.origin_x
	tmy := y - ea.origin_y

	s3 := eventMotion.State() & gdk.ModifierType(gdk.BUTTON1_MASK)
	s4 := eventMotion.State() & gdk.ModifierType(gdk.BUTTON3_MASK)

	px, py := ea.mouse_to_pixel(tmx, tmy)
	curPx := int(px)
	curPy := int(py)

	//fmt.Println(s1, s2, s3)
	if ea.PtInEditArea(curPx, curPy) {
		if s3 != 0 {
			//fmt.Println("M1 x=", int(px), "y=", int(py))
			if (eventMotion.State() & gdk.CONTROL_MASK) != 0 {
				if ea.lastPx != curPx || ea.lastPy != curPy {
					ea.RestoreSprite()
					ea.lastPx = curPx
					ea.lastPy = curPy
					Line(ea.imgBuf, ea.lastDrawPx, ea.lastDrawPy, curPx, curPy, ea.foreColor)
					ea.QueueDraw()
					if ea.pixbufModify != nil {
						ea.pixbufModify()
					}

				}

			} else if ea.lastPx != curPx || ea.lastPy != curPy {
				SetPixel(ea.imgBuf, curPx, curPy, ea.foreColor)
				ea.lastDrawPx = curPx
				ea.lastDrawPy = curPy
				ea.QueueDraw()
				if ea.pixbufModify != nil {
					ea.pixbufModify()
				}
			}
		} else if s4 != 0 {
			//fmt.Println("M3 x=", int(px), "y=", int(py))
			if ea.lastPx != curPx || ea.lastPy != curPy {
				SetPixel(ea.imgBuf, curPx, curPy, ea.backColor)
				ea.lastPx = curPx
				ea.lastPy = curPy
				ea.QueueDraw()
				if ea.pixbufModify != nil {
					ea.pixbufModify()
				}
			}
		}

	}
	return false
}
