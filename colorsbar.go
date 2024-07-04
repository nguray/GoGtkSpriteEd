package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const RGB_R_MASK uint32 = 0xFF << 24
const RGB_G_MASK uint32 = 0xFF << 16
const RGB_B_MASK uint32 = 0xFF << 8
const RGB_A_MASK uint32 = 0xFF

type ColorsChangeCallback func(uint32, uint32)

type colorsBar struct {
	da                   *gtk.DrawingArea
	nbLines              int
	nbColumns            int
	cellSize             int
	colors               []uint32
	foreColor            uint32
	backColor            uint32
	colorChangedCallback ColorsChangeCallback
}

func (cb *colorsBar) SetColorsChangeCallback(colorsChgCb ColorsChangeCallback) {
	cb.colorChangedCallback = colorsChgCb
}

func getRGBA(col uint32) (byte, byte, byte, byte) {
	r := (RGB_R_MASK & col) >> 24
	g := (RGB_G_MASK & col) >> 16
	b := (RGB_B_MASK & col) >> 8
	a := RGB_A_MASK & col
	return byte(r), byte(g), byte(b), byte(a)
}

func RGBA(red, green, blue, alpha uint32) uint32 {
	b := blue << 8
	g := green << 16
	r := red << 24
	a := alpha
	return r | g | b | a

}

func ColorsBarNew() *colorsBar {

	cb := new(colorsBar)
	cb.nbLines = 2
	cb.nbColumns = 32
	cb.cellSize = 18

	for l := 0; l < cb.nbLines; l++ {
		for c := 0; c < cb.nbColumns; c++ {
			cb.colors = append(cb.colors, 0x00FF00FF)
		}
	}

	cb.readPalette("palette.cfg")

	cb.da, _ = gtk.DrawingAreaNew()
	//cb.da.SetHAlign(gtk.ALIGN_FILL)
	//cb.da.SetVAlign(gtk.ALIGN_FILL)
	//cb.da.SetHExpand(true)
	//cb.da.SetVExpand(false)
	cb.da.SetSizeRequest(cb.nbColumns*cb.cellSize+2, cb.nbLines*cb.cellSize+2)
	// Event handlers
	ie := int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.BUTTON_MOTION_MASK)
	cb.da.AddEvents(ie)

	cb.da.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		// w := da.GetAllocatedWidth()
		// h := da.GetAllocatedHeight()

		var (
			x, y float64
			col  uint32
		)

		cb.drawCell(cr, 2, 2, float64(2*cb.cellSize-2), float64(2*cb.cellSize-2), cb.backColor)
		cb.drawCell(cr, 2, 2, float64(cb.cellSize-2), float64(cb.cellSize-2), cb.foreColor)

		for l := 0; l < cb.nbLines; l++ {
			for c := 0; c < cb.nbColumns; c++ {
				col = cb.colors[l*cb.nbColumns+c]
				x = float64((c+2)*cb.cellSize + 1)
				y = float64(l*cb.cellSize + 1)
				cb.drawCell(cr, x, y, float64(cb.cellSize-2), float64(cb.cellSize-2), col)

			}
		}

	})

	cb.da.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		buttonEvent := gdk.EventButtonNewFromEvent(ev)
		if buttonEvent.Button() == gdk.BUTTON_PRIMARY {

			if buttonEvent.Type() == gdk.EVENT_2BUTTON_PRESS {
				//fmt.Println("Double click Left Mouse button")
				id := cb.mouse2IdColor(buttonEvent.X(), buttonEvent.Y())
				if id >= 0 && id < (cb.nbColumns*cb.nbLines) {
					colorDlg, _ := gtk.ColorChooserDialogNew("Select Color", appWin)
					r, g, b, a := getRGBA(cb.foreColor)
					col := gdk.NewRGBA(float64(r)/255, float64(g)/255, float64(b)/255, float64(a)/255)
					colorDlg.SetRGBA(col)
					colorDlg.Connect("response", func(c *gtk.ColorChooserDialog, res int) {
						if res == int(gtk.RESPONSE_OK) {
							rgba := c.GetRGBA()
							fmt.Printf("%f %f %f %f\n", rgba.GetRed(), rgba.GetGreen(), rgba.GetBlue(), rgba.GetAlpha())
							cb.foreColor = RGBA(uint32(math.Round(rgba.GetRed()*255)),
								uint32(math.Round(rgba.GetGreen()*255)),
								uint32(math.Round(rgba.GetBlue()*255)),
								uint32(math.Round(rgba.GetAlpha()*255)))
							cb.colors[id] = cb.foreColor
							cb.da.QueueDraw()
							if cb.colorChangedCallback != nil {
								cb.colorChangedCallback(cb.foreColor, cb.backColor)
							}
						}
						colorDlg.Destroy()
					})
					//colorDlg.SetModal(true)
					colorDlg.Run()
				}
			} else {
				//fmt.Println("Press Left Mouse button")
				if buttonEvent.X() < float64(2*cb.cellSize) && buttonEvent.Y() < float64(2*cb.cellSize) {
					dum := cb.backColor
					cb.backColor = cb.foreColor
					cb.foreColor = dum
					cb.da.QueueDraw()
					cb.colorChangedCallback(cb.foreColor, cb.backColor)
				} else {
					id := cb.mouse2IdColor(buttonEvent.X(), buttonEvent.Y())
					if id >= 0 && id < (cb.nbColumns*cb.nbLines) {
						cb.foreColor = cb.colors[id]
						da.QueueDraw()
						if cb.colorChangedCallback != nil {
							cb.colorChangedCallback(cb.foreColor, cb.backColor)
						}
					}
				}

			}
			return true

		}
		return false
	})

	return cb
}

func (cb *colorsBar) drawCell(cr *cairo.Context, x, y, w, h float64, col uint32) {
	cr.SetLineWidth(0.5)
	r, g, b, a := getRGBA(col)
	if a == 0 {
		xLeft := x + 1
		yTop := y + 1
		xRight := xLeft + w - 2
		yBottom := yTop + h - 2
		cr.SetSourceRGBA(0, 0, 0, 1)
		cr.MoveTo(xLeft, yTop)
		cr.LineTo(xRight, yTop)
		cr.LineTo(xRight, yBottom)
		cr.LineTo(xLeft, yBottom)
		cr.LineTo(xLeft, yTop)
		cr.MoveTo(xLeft, yTop)
		cr.LineTo(xRight, yBottom)
		cr.MoveTo(xRight, yTop)
		cr.LineTo(xLeft, yBottom)
		cr.Stroke()
	} else {
		cr.SetSourceRGBA(float64(r)/255, float64(g)/255, float64(b)/255, float64(a)/255)
		cr.Rectangle(x+1, y+1, w, h)
		cr.Fill()
	}

}

func (cb *colorsBar) readPalette(fileName string) {

	f, err := os.Open(fileName)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var (
		id  int
		l   int = 0
		c   int = 0
		val uint64
		col uint32
		nbL int = 0
	)

	for scanner.Scan() {

		//--
		val, err = strconv.ParseUint(scanner.Text(), 10, 32)
		if err != nil {
			col = 0
		} else {
			col = uint32(val)
		}
		id = l*cb.nbColumns + c

		if nbL == 0 {
			cb.foreColor = col
		} else if nbL == 1 {
			cb.backColor = col
		} else {
			cb.colors[id] = col
			c++
			if c >= cb.nbColumns {
				c = 0
				l++
				if l >= cb.nbLines {
					break
				}
			}
		}
		nbL++

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

func (cb *colorsBar) savePalette(fileName string) {

	var (
		str1 string
	)

	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	str1 = fmt.Sprintf("%d\n", cb.foreColor)
	_, _ = f.WriteString(str1)

	str1 = fmt.Sprintf("%d\n", cb.backColor)
	_, _ = f.WriteString(str1)

	for _, col := range cb.colors {
		str1 := fmt.Sprintf("%d\n", col)
		_, _ = f.WriteString(str1)
	}

}

func (cb *colorsBar) mouse2IdColor(mx, my float64) int {
	c := int((mx - float64(2*cb.cellSize)) / float64(cb.cellSize))
	l := int(my / float64(cb.cellSize))
	if c >= 0 && c < cb.nbColumns && l >= 0 && l < cb.nbLines {
		return l*cb.nbColumns + c
	} else {
		return -1
	}

}

func (cb *colorsBar) GetForegroundColor() uint32 {
	return cb.foreColor
}

func (cb *colorsBar) GetBackgroundColor() uint32 {
	return cb.backColor
}

func (cb *colorsBar) SetBackgroundColor(col uint32) {
	cb.backColor = col
	cb.da.QueueDraw()
}

func (cb *colorsBar) SetForegroundColor(col uint32) {
	cb.foreColor = col
	cb.da.QueueDraw()
}
