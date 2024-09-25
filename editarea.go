package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
	M_SELECT int = iota
	M_PENCIL
	M_RECTANGLE
	M_ELLIPSE
	M_FILL
)

type UndoMode int

const (
	NONE UndoMode = iota
	PENCIL
	RECTANGLE
	ELLIPSE
	FILL
	FLIP_HORIZONTALY
	FLIP_VERTICALY
	SWING_RIGHT
	SWING_LEFT
)

type ColorsPickCallback func(foreColor uint32, backColor uint32)
type PixbufModifyCallback func()

type mouseClickedCallback func(int, int)

type ButtonPressProc func(ea *editArea, ev *gdk.Event) bool
type ButtonReleaseProc func(ea *editArea, ev *gdk.Event) bool
type DrawProc func(ea *editArea, cr *cairo.Context)
type MotionNotifyProc func(ea *editArea, ev *gdk.Event) bool

type editArea struct {
	*gtk.DrawingArea
	imgBuf         *gdk.Pixbuf
	imgBufBak      *gdk.Pixbuf
	imgBufCopy     *gdk.Pixbuf
	nbPixelsH      int
	nbPixelsW      int
	cellSize       float64
	lastPx         int
	lastPy         int
	lastDrawPx     int
	lastDrawPy     int
	foreColor      uint32
	backColor      uint32
	fColorPick     bool
	mode           int
	scale          float64
	origin_x       float64
	origin_y       float64
	start_x        float64
	start_y        float64
	start_origin_x float64
	start_origin_y float64
	undo_mode      UndoMode
	selectRect     *SelectRect
	copyRect       *SelectRect
	clipboard      *gtk.Clipboard
	mouseClicked   mouseClickedCallback
	colorPick      ColorsPickCallback
	pixbufModify   PixbufModifyCallback
	buttonPress    ButtonPressProc
	buttonRelease  ButtonReleaseProc
	draw           DrawProc
	motionNotify   MotionNotifyProc
}

func (ea *editArea) SetLeftMouseClickedCallback(cb mouseClickedCallback) {
	ea.mouseClicked = cb
}

func (ea *editArea) SetColorsPickCallback(colorpickCb ColorsPickCallback) {
	ea.colorPick = colorpickCb
}

func (ea *editArea) SetPixBufCallback(pixbufModifyCb PixbufModifyCallback) {
	ea.pixbufModify = pixbufModifyCb
}

func EditAreaNew() *editArea {
	da, _ := gtk.DrawingAreaNew()
	pixbuf, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, 32, 32)
	//pixbuf.Fill(0xFFFFFFFF)
	pixbufbak, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, 32, 32)
	pixbufcopy, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, 32, 32)
	//pixbufbak.Fill(0xFFFFFFFF)
	atom := gdk.GdkAtomIntern("CLIPBOARD", false)
	clip, _ := gtk.ClipboardGet(atom)
	ea := &editArea{da, pixbuf, pixbufbak, pixbufcopy,
		32, 32, 14, -1, -1, -1, -1, 0, 0, false, 0, 1.0, 0, 0, 0, 0, 0, 0, NONE,
		SelectRectNew(0, 0, 0, 0), SelectRectNew(0, 0, 0, 0), clip,
		nil, nil, nil, nil, nil, nil, nil}

	ea.init()
	return ea

}

func (ea *editArea) PtInEditArea(px, py int) bool {
	if (px >= 0 && px < ea.nbPixelsW) && (py >= 0 && py < ea.nbPixelsH) {
		return true
	}
	return false
}

func (ea *editArea) init() {

	//_, _ = glib.SignalNew("test-signal")
	ea.SetPencilMode()

	Width := ea.nbPixelsH*int(ea.cellSize) + 128
	//Height := ea.nbPixelsW*int(ea.cellSize) + 4
	// Setting parameter for drawing area
	// ea.da.SetHAlign(gtk.ALIGN_FILL)
	// ea.da.SetVAlign(gtk.ALIGN_FILL)
	// ea.da.SetHExpand(true)
	// ea.da.SetVExpand(true)
	ea.SetSizeRequest(Width, -1)
	// Event handlers
	ie := int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.BUTTON_MOTION_MASK | gdk.SCROLL_MASK)
	ea.AddEvents(ie)

	ea.Connect("motion-notify-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		eventMotion := gdk.EventMotionNewFromEvent(ev)

		if eventMotion.State()&gdk.ModifierType(gdk.BUTTON2_MASK) != 0 {
			//-- Translate canvas
			x, y := eventMotion.MotionVal()
			dx := x - ea.start_x
			dy := y - ea.start_y
			ea.origin_x = ea.start_origin_x + dx
			ea.origin_y = ea.start_origin_y + dy
			ea.QueueDraw()
			return false
		} else {
			return ea.motionNotify(ea, ev)
		}
	})

	ea.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {

		buttonEvent := gdk.EventButtonNewFromEvent(ev)
		if buttonEvent.Button() == gdk.BUTTON_MIDDLE {
			//-- Start translate canvas
			ea.start_x = buttonEvent.X()
			ea.start_y = buttonEvent.Y()
			ea.start_origin_x = ea.origin_x
			ea.start_origin_y = ea.origin_y
		}
		return ea.buttonPress(ea, ev)

	})

	ea.Connect("button-release-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
		buttonEvent := gdk.EventButtonNewFromEvent(ev)
		if buttonEvent.Button() == gdk.BUTTON_MIDDLE {
			ea.start_x = 0
			ea.start_y = 0
			ea.start_origin_x = ea.origin_x
			ea.start_origin_y = ea.origin_y

		}
		return ea.buttonRelease(ea, ev)

	})

	ea.Connect("scroll-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {

		ev1 := gdk.EventScrollNewFromEvent(ev)
		if ev1.Direction() == gdk.SCROLL_UP {
			if ea.scale < 5.0 {
				ea.scale += 0.05
			}
		} else if ev1.Direction() == gdk.SCROLL_DOWN {
			if ea.scale > 0.5 {
				ea.scale -= 0.05
			}
		}
		ea.QueueDraw()
		return true

	})

	ea.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		//--
		w := ea.GetAllocatedWidth() - 64
		h := ea.GetAllocatedHeight()
		//-- Calculer la taille du pixel affiché
		o1 := float64((w - 6)) / float64(ea.nbPixelsW)
		o2 := float64((h - 6)) / float64(ea.nbPixelsH)
		if o1 < o2 {
			ea.cellSize = o1 * ea.scale
		} else {
			ea.cellSize = o2 * ea.scale
		}

		cr.Translate(float64(ea.origin_x), float64(ea.origin_y))

		var x, y float64 = 0.0, 0.0
		cr.SetSourceRGBA(0.0, 0.0, 0.0, 1.0)
		cr.SetLineWidth(0.4)
		for j := 0; j <= ea.nbPixelsH; j++ {
			y = float64(j)*ea.cellSize + 4
			for i := 0; i <= ea.nbPixelsW; i++ {
				x = float64(i)*ea.cellSize + 4
				cr.MoveTo(x, y-1)
				cr.LineTo(x, y+1)
				cr.MoveTo(x-1, y)
				cr.LineTo(x+1, y)
			}
		}
		cr.Stroke()

		if ea.imgBuf != nil {
			var r, g, b, a byte
			pixs := ea.imgBuf.GetPixels()
			nChannels := ea.imgBuf.GetNChannels()
			rowStride := ea.imgBuf.GetRowstride()
			for j := 0; j < ea.nbPixelsH; j++ {
				y = float64(j)*ea.cellSize + 4 + 1
				for i := 0; i < ea.nbPixelsW; i++ {
					iPix := j*rowStride + i*nChannels
					r = pixs[iPix]
					g = pixs[iPix+1]
					b = pixs[iPix+2]
					a = pixs[iPix+3]
					if a != 0 {
						x = float64(i)*ea.cellSize + 4 + 1
						cr.SetSourceRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
						cr.Rectangle(x, y, ea.cellSize-2, ea.cellSize-2)
						cr.Fill()
					}
				}
			}
			gtk.GdkCairoSetSourcePixBuf(cr, ea.imgBuf, float64(ea.nbPixelsW+1)*ea.cellSize, 0)
			cr.Paint()
			gtk.GdkCairoSetSourcePixBuf(cr, ea.imgBufBak, float64(ea.nbPixelsW+1)*ea.cellSize, 64)
			cr.Paint()
			gtk.GdkCairoSetSourcePixBuf(cr, ea.imgBufCopy, float64(ea.nbPixelsW+1)*ea.cellSize, 128)
			cr.Paint()
		}

		//-- Current tool specificaly draw
		ea.draw(ea, cr)

	})

}

func (ea *editArea) mouse_to_pixel(mx, my float64) (float64, float64) {
	px := int((mx - 4) / ea.cellSize)
	py := int((my - 4) / ea.cellSize)
	return float64(px), float64(py)
}

func (ea *editArea) pixel_to_mouse(x, y int) (float64, float64) {
	mx := float64(x)*ea.cellSize + 4.0
	my := float64(y)*ea.cellSize + 4.0
	return mx, my
}

func (ea *editArea) SetForegroundColor(col uint32) {
	ea.foreColor = col
}

func (ea *editArea) SetBackgroundColor(col uint32) {
	ea.backColor = col
}

func (ea *editArea) SetPixbuf(imgBuf *gdk.Pixbuf) {
	ea.nbPixelsH = imgBuf.GetHeight()
	ea.nbPixelsW = imgBuf.GetWidth()
	ea.imgBuf = imgBuf
	ea.imgBufBak, _ = gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, ea.nbPixelsW, ea.nbPixelsH)
	ea.imgBufCopy, _ = gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, ea.nbPixelsW, ea.nbPixelsH)
	ea.selectRect.Init()
	ea.QueueDraw()
}

func (ea *editArea) BackupSprite() {
	if ea.imgBuf != nil && ea.imgBufBak != nil {
		h := ea.imgBufBak.GetHeight()
		w := ea.imgBufBak.GetWidth()
		CopyArea(ea.imgBuf, 0, 0, w, h, ea.imgBufBak, 0, 0)
	}
}

func (ea *editArea) RestoreSprite() {
	if ea.imgBuf != nil && ea.imgBufBak != nil {
		h := ea.imgBuf.GetHeight()
		w := ea.imgBuf.GetWidth()
		CopyArea(ea.imgBufBak, 0, 0, w, h, ea.imgBuf, 0, 0)
	}
}

func CopyArea(srcPix *gdk.Pixbuf, xSrc, ySrc, wSrc, hSrc int, desPix *gdk.Pixbuf, xDes, yDes int) {

	if srcPix != nil && desPix != nil {

		srcNChannels := srcPix.GetNChannels()
		srcRowStride := srcPix.GetRowstride()
		srcPixs := srcPix.GetPixels()
		srcHeight := srcPix.GetHeight()
		srcWidth := srcPix.GetWidth()

		desNChannels := desPix.GetNChannels()
		desRowStride := desPix.GetRowstride()
		desPixs := desPix.GetPixels()
		desHeight := desPix.GetHeight()
		desWidth := desPix.GetWidth()

		var srcId, desId int
		var xs, ys, xd, yd int
		var srcIdY, desIdY int

		for j := 0; j < hSrc; j++ {

			ys = ySrc + j
			if ys >= srcHeight {
				break
			}
			yd = yDes + j
			srcIdY = ys * srcRowStride
			desIdY = yd * desRowStride
			if yd < desHeight {

				for i := 0; i < wSrc; i++ {

					xs = xSrc + i
					if xs >= srcWidth {
						break
					}
					xd = xDes + i

					if xd < desWidth {

						srcId = xs*srcNChannels + srcIdY
						desId = xd*desNChannels + desIdY

						desPixs[desId] = srcPixs[srcId]
						desPixs[desId+1] = srcPixs[srcId+1]
						desPixs[desId+2] = srcPixs[srcId+2]
						desPixs[desId+3] = srcPixs[srcId+3]

					}

				}

			}

		}

	}

}

func FillArea(desPix *gdk.Pixbuf, xSrc, ySrc, wSrc, hSrc int, col uint32) {

	desNChannels := desPix.GetNChannels()
	desRowStride := desPix.GetRowstride()
	desPixs := desPix.GetPixels()
	desHeight := desPix.GetHeight()
	desWidth := desPix.GetWidth()

	r, g, b, a := getRGBA(col)

	var (
		xd, yd        int
		desId, desIdY int
	)
	for j := 0; j < hSrc; j++ {
		yd = j + ySrc
		if yd >= desHeight {
			break
		}
		desIdY = yd * desRowStride
		for i := 0; i < wSrc; i++ {
			xd = xSrc + i
			if xd >= desWidth {
				break
			}
			desId = xd*desNChannels + desIdY
			desPixs[desId] = r
			desPixs[desId+1] = g
			desPixs[desId+2] = b
			desPixs[desId+3] = a

		}
	}

}

func GetPixel(pixBuf *gdk.Pixbuf, x, y int) uint32 {
	nChannels := pixBuf.GetNChannels()
	rowStride := pixBuf.GetRowstride()

	pixs := pixBuf.GetPixels()
	iPix := y*rowStride + x*nChannels
	r := uint32(pixs[iPix])   // Red
	g := uint32(pixs[iPix+1]) // Green
	b := uint32(pixs[iPix+2]) // Blue
	a := uint32(pixs[iPix+3]) // Alpha
	return RGBA(r, g, b, a)
}

func SetPixel(pixBuf *gdk.Pixbuf, x, y int, col uint32) {

	if pixBuf != nil {

		nChannels := pixBuf.GetNChannels()
		rowStride := pixBuf.GetRowstride()

		r, g, b, a := getRGBA(col)

		pixs := pixBuf.GetPixels()
		iPix := y*rowStride + x*nChannels
		pixs[iPix] = byte(r)   // Red
		pixs[iPix+1] = byte(g) // Green
		pixs[iPix+2] = byte(b) // Blue
		pixs[iPix+3] = byte(a) // Alpha

	}
}

func IntAbs(val int) int {
	if val < 0 {
		return -val
	} else {
		return val
	}
}

func Line(pixbuf *gdk.Pixbuf, x0, y0, x1, y1 int, col uint32) {
	/*----------------------------------------------------------------------------*\
		Description :



		Date de crÃ©ation : 16-02-2022                       Raymond NGUYEN THANH
	\*----------------------------------------------------------------------------*/

	var (
		width, height  int
		steep          bool
		deltax, deltay int
		error          int
		x, y, ystep    int
	)
	//-----------------------------------------------------------------------

	if pixbuf == nil {
		return
	}

	//n_channels = pixbuf.NChannels()
	//g_assert (pixbuf->get_colorspace() == Gdk::COLORSPACE_RGB); // gdk_pixbuf_get_colorspace (pixbuf)
	//g_assert (pixbuf->get_bits_per_sample()== 8); //  gdk_pixbuf_get_bits_per_sample (pixbuf)
	//g_assert (pixbuf->get_has_alpha()); //gdk_pixbuf_get_has_alpha (pixbuf)
	//g_assert (n_channels == 4);
	nChannels := pixbuf.GetNChannels()
	rowStride := pixbuf.GetRowstride()

	r, g, b, a := getRGBA(col)

	pixs := pixbuf.GetPixels()

	width = pixbuf.GetWidth()   // gdk_pixbuf_get_width (pixbuf);
	height = pixbuf.GetHeight() // gdk_pixbuf_get_height (pixbuf);

	putpixel := func(x, y int) {
		iPix := y*rowStride + x*nChannels
		pixs[iPix] = byte(r)   // Red
		pixs[iPix+1] = byte(g) // Green
		pixs[iPix+2] = byte(b) // Blue
		pixs[iPix+3] = byte(a) // Alpha

	}

	if ((x0 >= 0 && x0 < width) && (y0 >= 0 && y0 < height)) &&
		((x1 >= 0 && x1 < width) && (y1 >= 0 && y1 < height)) {

		steep = (IntAbs(y1-y0) > IntAbs(x1-x0))
		if steep {
			x0, y0 = y0, x0
			x1, y1 = y1, x1
		}
		if x0 > x1 {
			x0, x1 = x1, x0
			y0, y1 = y1, y0
		}

		deltax = x1 - x0
		deltay = IntAbs(y1 - y0)
		error = deltax / 2

		y = y0
		if y0 < y1 {
			ystep = 1
		} else {
			ystep = -1
		}

		for x = x0; x <= x1; x++ {
			if steep {
				putpixel(y, x)
			} else {
				putpixel(x, y)
			}
			error = error - deltay
			if error < 0 {
				y = y + ystep
				error = error + deltax
			}
		}
	}

}

func (ea *editArea) CutSelect() {
	if ea.mode == M_SELECT {
		if !ea.selectRect.IsEmpty() {
			CopyArea(ea.imgBuf, ea.selectRect.left, ea.selectRect.top,
				ea.selectRect.Width(), ea.selectRect.Height(), ea.imgBufCopy, 0, 0)
			ea.copyRect.left = ea.selectRect.left
			ea.copyRect.top = ea.selectRect.top
			ea.copyRect.right = ea.selectRect.right
			ea.copyRect.bottom = ea.selectRect.bottom

			tmpPixBuf, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8,
				ea.selectRect.Width(), ea.selectRect.Height())
			CopyArea(ea.imgBuf, ea.selectRect.left, ea.selectRect.top,
				ea.selectRect.Width(), ea.selectRect.Height(), tmpPixBuf, 0, 0)

			FillArea(ea.imgBuf, ea.selectRect.left, ea.selectRect.top,
				ea.selectRect.Width(), ea.selectRect.Height(), ea.backColor)

			ea.selectRect.Init()
			ea.QueueDraw()

			ea.clipboard.SetImage(tmpPixBuf)

		}
	}

}

func (ea *editArea) CopySelect() {
	if ea.mode == M_SELECT {
		if !ea.selectRect.IsEmpty() {
			CopyArea(ea.imgBuf, ea.selectRect.left, ea.selectRect.top,
				ea.selectRect.Width(), ea.selectRect.Height(), ea.imgBufCopy, 0, 0)
			ea.copyRect.left = ea.selectRect.left
			ea.copyRect.top = ea.selectRect.top
			ea.copyRect.right = ea.selectRect.right
			ea.copyRect.bottom = ea.selectRect.bottom

			tmpPixBuf, _ := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8,
				ea.selectRect.Width(), ea.selectRect.Height())
			CopyArea(ea.imgBuf, ea.selectRect.left, ea.selectRect.top,
				ea.selectRect.Width(), ea.selectRect.Height(), tmpPixBuf, 0, 0)

			ea.selectRect.Init()
			ea.QueueDraw()

			ea.clipboard.SetImage(tmpPixBuf)

		}
	}
}

func (ea *editArea) PasteSelect() {
	if !ea.copyRect.IsEmpty() {
		ea.SetSelectMode()
		ea.selectRect.left = 0
		ea.selectRect.top = 0
		ea.selectRect.right = ea.copyRect.Width() - 1
		ea.selectRect.bottom = ea.copyRect.Height() - 1
		ea.selectRect.mode = 2
		ea.BackupSprite()

		if ea.clipboard.WaitIsImageAvailable() {

			tmpPixBuf, _ := ea.clipboard.WaitForImage()

			CopyArea(tmpPixBuf, 0, 0,
				tmpPixBuf.GetWidth(), tmpPixBuf.GetHeight(), ea.imgBufCopy, 0, 0)

			CopyArea(ea.imgBufCopy, 0, 0,
				ea.selectRect.Width(), ea.selectRect.Height(), ea.imgBuf, ea.selectRect.left, ea.selectRect.top)

			ea.QueueDraw()
			if ea.pixbufModify != nil {
				ea.pixbufModify()
			}

		}

	}
}

func (ea *editArea) DrawSelectRect(cr *cairo.Context) {

	//-- Draw Select Frame
	if !ea.selectRect.IsEmpty() {
		//--
		px1, py1 := ea.selectRect.GetCorner(0)
		px2, py2 := ea.selectRect.GetCorner(2)
		x1, y1 := ea.pixel_to_mouse(px1, py1)
		x2, y2 := ea.pixel_to_mouse(px2, py2)

		if x1 > x2 {
			x1, x2 = x2, x1
		}
		if y1 > y2 {
			y1, y2 = y2, y1
		}

		//-- Draw Handles
		cr.SetSourceRGBA(0.0, 0.0, 1.0, 1.0)
		dx := float64(ea.cellSize - 5)
		x := x1 + 2.0
		y := y1 + 2.0
		cr.Rectangle(x, y, dx, dx)
		cr.Fill()
		x = x2 + 2.0
		y = y1 + 2.0
		cr.Rectangle(x, y, dx, dx)
		cr.Fill()
		x = x2 + 2.0
		y = y2 + 2.0
		cr.Rectangle(x, y, dx, dx)
		cr.Fill()
		x = x1 + 2.0
		y = y2 + 2.0
		cr.Rectangle(x, y, dx, dx)
		cr.Fill()

		//-- Draw Frame
		y2 += float64(ea.cellSize)
		x2 += float64(ea.cellSize)
		cr.SetSourceRGBA(0.0, 0.0, 1.0, 0.1)
		cr.Rectangle(x1, y1, x2-x1, y2-y1)
		cr.Fill()
	}

}

func (ea *editArea) FlipHorizontaly() {
	//---------------------------------------------------------

	row_stride := ea.imgBuf.GetRowstride()
	n_channels := ea.imgBuf.GetNChannels()
	pixelsDes := ea.imgBuf.GetPixels()
	pixelsSrc := ea.imgBufBak.GetPixels()
	var ipSrc, ipDes, yOffset int
	width := ea.imgBuf.GetWidth()
	height := ea.imgBuf.GetHeight()

	w := width - 1
	for y := 0; y < height; y++ {
		yOffset = y * row_stride
		for x := 0; x < width; x++ {
			ipSrc = (yOffset + x*n_channels)
			ipDes = (yOffset + (w-x)*n_channels)
			pixelsDes[ipDes] = pixelsSrc[ipSrc]
			pixelsDes[ipDes+1] = pixelsSrc[ipSrc+1]
			pixelsDes[ipDes+2] = pixelsSrc[ipSrc+2]
			pixelsDes[ipDes+3] = pixelsSrc[ipSrc+3]
		}
	}
	ea.QueueDraw()
	if ea.pixbufModify != nil {
		ea.pixbufModify()
	}

}

func (ea *editArea) FlipVerticaly() {
	//---------------------------------------------------------

	row_stride := ea.imgBuf.GetRowstride()
	n_channels := ea.imgBuf.GetNChannels()
	pixelsDes := ea.imgBuf.GetPixels()
	pixelsSrc := ea.imgBufBak.GetPixels()
	var ipSrc, ipDes, xOffset int
	width := ea.imgBuf.GetWidth()
	height := ea.imgBuf.GetHeight()
	var ySrcOffset, yDesOffset int

	h := height - 1
	for y := 0; y < height; y++ {
		ySrcOffset = y * row_stride
		yDesOffset = (h - y) * row_stride
		for x := 0; x < width; x++ {
			xOffset = x * n_channels
			ipSrc = (ySrcOffset + xOffset)
			ipDes = (yDesOffset + xOffset)
			pixelsDes[ipDes] = pixelsSrc[ipSrc]
			pixelsDes[ipDes+1] = pixelsSrc[ipSrc+1]
			pixelsDes[ipDes+2] = pixelsSrc[ipSrc+2]
			pixelsDes[ipDes+3] = pixelsSrc[ipSrc+3]
		}
	}
	ea.QueueDraw()
	if ea.pixbufModify != nil {
		ea.pixbufModify()
	}

}

func (ea *editArea) SwingLeft() {
	//---------------------------------------------------------
	var ipSrc, ipDes int
	w := ea.imgBufBak.GetWidth()
	h := ea.imgBufBak.GetHeight()
	rowStrideSrc := ea.imgBufBak.GetRowstride()
	nChannelsSrc := ea.imgBufBak.GetNChannels()
	pixelsSrc := ea.imgBufBak.GetPixels()
	pixelsDes := ea.imgBuf.GetPixels()

	offSetH := w - 1

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			ipSrc = x*nChannelsSrc + y*rowStrideSrc
			ipDes = y*nChannelsSrc + (offSetH-x)*rowStrideSrc

			pixelsDes[ipDes] = pixelsSrc[ipSrc]
			pixelsDes[ipDes+1] = pixelsSrc[ipSrc+1]
			pixelsDes[ipDes+2] = pixelsSrc[ipSrc+2]
			pixelsDes[ipDes+3] = pixelsSrc[ipSrc+3]

		}
	}

	//--
	ea.QueueDraw()
	if ea.pixbufModify != nil {
		ea.pixbufModify()
	}

}

func (ea *editArea) SwingRight() {
	//---------------------------------------------------------
	var ipSrc, ipDes int
	w := ea.imgBufBak.GetWidth()
	h := ea.imgBufBak.GetHeight()
	rowStrideSrc := ea.imgBufBak.GetRowstride()
	nChannelsSrc := ea.imgBufBak.GetNChannels()
	pixelsSrc := ea.imgBufBak.GetPixels()
	pixelsDes := ea.imgBuf.GetPixels()

	offSetW := h - 1

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			ipSrc = x*nChannelsSrc + y*rowStrideSrc
			ipDes = (offSetW-y)*nChannelsSrc + x*rowStrideSrc

			pixelsDes[ipDes] = pixelsSrc[ipSrc]
			pixelsDes[ipDes+1] = pixelsSrc[ipSrc+1]
			pixelsDes[ipDes+2] = pixelsSrc[ipSrc+2]
			pixelsDes[ipDes+3] = pixelsSrc[ipSrc+3]

		}
	}

	//--
	ea.QueueDraw()
	if ea.pixbufModify != nil {
		ea.pixbufModify()
	}

}

func (ea *editArea) Undo() {
	//---------------------------------------------------------

	ea.InitRectangleMode()
	ea.InitEllipseMode()

	switch ea.undo_mode {
	case PENCIL, RECTANGLE, ELLIPSE:
		ea.RestoreSprite()
		ea.QueueDraw()
	case FLIP_HORIZONTALY:
		ea.RestoreSprite()
		ea.QueueDraw()
	case FLIP_VERTICALY:
		ea.RestoreSprite()
		ea.QueueDraw()
	case SWING_LEFT:
		ea.BackupSprite()
		ea.SwingRight()
		ea.undo_mode = SWING_RIGHT
		ea.QueueDraw()
	case SWING_RIGHT:
		ea.BackupSprite()
		ea.SwingLeft()
		ea.undo_mode = SWING_LEFT
		ea.QueueDraw()
	}
}
