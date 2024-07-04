package main

import (
	"strings"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type spriteChangeCallback func(spr *gdk.Pixbuf)

type spritesBar struct {
	da             *gtk.DrawingArea
	sprites        [8]*gdk.Pixbuf
	fileNames      [8]string
	nbCells        int
	cellSize       int
	idSelect       int
	spriteChange   spriteChangeCallback
	fAnimate       bool
	iAnimateSprite int
}

func (sb *spritesBar) SetSpriteChangeCallback(cb spriteChangeCallback) {
	sb.spriteChange = cb
}

func SpritesBarNew() *spritesBar {

	sb := new(spritesBar)
	sb.nbCells = 8
	sb.cellSize = 64
	sb.fAnimate = false
	sb.iAnimateSprite = 0

	for i := 0; i < sb.nbCells; i++ {
		sb.sprites[i], _ = gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, 32, 32)
		sb.sprites[i].Fill(0x00000000)
	}

	sb.da, _ = gtk.DrawingAreaNew()
	//sb.da.SetHAlign(gtk.ALIGN_FILL)
	//sb.da.SetVAlign(gtk.ALIGN_FILL)
	//sb.da.SetHExpand(false)
	//sb.da.SetVExpand(true)
	sb.da.SetSizeRequest(sb.cellSize, sb.nbCells*sb.cellSize)

	// Event handlers
	ie := int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.BUTTON_MOTION_MASK)
	sb.da.AddEvents(ie)

	sb.da.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		buttonEvent := gdk.EventButtonNewFromEvent(ev)
		if buttonEvent.Button() == gdk.BUTTON_PRIMARY {

			id := int((buttonEvent.Y() - 2) / float64(sb.cellSize))
			if (id != sb.idSelect) && (id < sb.nbCells) {
				sb.idSelect = id
				if sb.spriteChange != nil {
					sb.spriteChange(sb.sprites[id])
				}
				sb.da.QueueDraw()
			}

			return true
		}
		return false
	})

	sb.da.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		w := da.GetAllocatedWidth()
		//h := da.GetAllocatedHeight()

		// Draw Frame
		sb.cellSize = w - 8
		cr.SetAntialias(cairo.ANTIALIAS_NONE)
		cr.SetLineWidth(1)
		cr.SetSourceRGBA(0.5, 0.5, 0.5, 1.0)
		x := (w - sb.cellSize) / 2
		for i := 0; i < sb.nbCells; i++ {
			y := i*sb.cellSize + 2
			cr.Rectangle(float64(x), float64(y), float64(sb.cellSize), float64(sb.cellSize))
		}
		cr.Stroke()

		// Draw Select Mark
		if sb.idSelect >= 0 {
			xLeft := float64((w-sb.cellSize)/2 + 1)
			yTop := float64(sb.idSelect*sb.cellSize + 2 + 1)
			yBottom := yTop + float64(sb.cellSize-3)
			xRight := xLeft + float64(sb.cellSize-3)
			cr.SetSourceRGBA(1, 0.5, 0.5, 1)
			cr.SetLineWidth(1.5)
			//--
			cr.MoveTo(xLeft, yTop+6)
			cr.LineTo(xLeft, yTop)
			cr.LineTo(xLeft+6, yTop)
			cr.MoveTo(xLeft, yBottom-6)
			cr.LineTo(xLeft, yBottom)
			cr.LineTo(xLeft+6, yBottom)
			// --
			cr.MoveTo(xRight, yTop+6)
			cr.LineTo(xRight, yTop)
			cr.LineTo(xRight-6, yTop)
			cr.MoveTo(xRight, yBottom-6)
			cr.LineTo(xRight, yBottom)
			cr.LineTo(xRight-6, yBottom)
			cr.Stroke()

		}

		// Draw Sprites
		for i := 0; i < sb.nbCells; i++ {
			sprite := sb.sprites[i]
			if sprite != nil {
				x = (w - sprite.GetWidth()) / 2
				y := i*sb.cellSize + (sb.cellSize-sprite.GetHeight())/2
				gtk.GdkCairoSetSourcePixBuf(cr, sprite, float64(x), float64(y))
				cr.Paint()
			}
		}

		if sb.fAnimate {
			sprite := sb.sprites[sb.iAnimateSprite]
			if sprite != nil {
				x = (w - sprite.GetWidth()) / 2
				y := sb.nbCells*sb.cellSize + (sb.cellSize-sprite.GetHeight())/2
				gtk.GdkCairoSetSourcePixBuf(cr, sprite, float64(x), float64(y))
				cr.Paint()
			}

		}

	})

	return sb
}

func (sb *spritesBar) StartAnimate() {
	sb.fAnimate = true
	sb.iAnimateSprite = 0
	sb.da.QueueDraw()

}

func (sb *spritesBar) StopAnimate() {
	sb.fAnimate = false
	sb.da.QueueDraw()

}

func (sb *spritesBar) NextFrame() {
	sb.iAnimateSprite += 1
	sb.iAnimateSprite %= sb.nbCells
	sb.da.QueueDraw()
}

func (sb *spritesBar) GetCurrentSprite() *gdk.Pixbuf {
	if sb.idSelect >= 0 && sb.idSelect < sb.nbCells {
		return sb.sprites[sb.idSelect]
	}
	return nil
}

func (sb *spritesBar) SetCurrentSprite(spr *gdk.Pixbuf) {

	sb.sprites[sb.idSelect] = spr
	sb.da.QueueDraw()
}

func (sb *spritesBar) LoadCurrentSprite(fileName string) {
	sprite1, _ := gdk.PixbufNewFromFile(fileName)
	if sprite1 != nil {
		sb.fileNames[sb.idSelect] = fileName
		sb.sprites[sb.idSelect] = sprite1
		sb.da.QueueDraw()
		if sb.spriteChange != nil {
			sb.spriteChange(sprite1)
		}
	}
}

func (sb *spritesBar) SaveAsCurrentSprite(fileName string) {
	if !strings.HasSuffix(fileName, ".png") && !strings.HasSuffix(fileName, ".PNG") {
		fileName += ".png"
	}
	sb.fileNames[sb.idSelect] = fileName
	sprite1 := sb.sprites[sb.idSelect]
	sprite1.SavePNG(fileName, 9)
}

func (sb *spritesBar) SaveCurrentSprite() {
	fileName := sb.fileNames[sb.idSelect]
	if fileName != "" {
		sprite1 := sb.sprites[sb.idSelect]
		sprite1.SavePNG(fileName, 9)
	}
}

func (sb *spritesBar) GetCurrentSpriteFileName() string {
	return sb.fileNames[sb.idSelect]
}

func (sb *spritesBar) NewCurrentSprite(w, h int) {
	sb.fileNames[sb.idSelect] = ""
	sprite1, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, w, h)
	sprite1.Fill(0x00000000)
	sb.sprites[sb.idSelect] = sprite1
	sb.da.QueueDraw()
	if sb.spriteChange != nil {
		sb.spriteChange(sprite1)
	}
}
