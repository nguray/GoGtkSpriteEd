package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gio"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appId = "com.github.gotk3.gtkspriteed"

var (
	imgSelectBox       *gtk.Image
	imgPencil          *gtk.Image
	imgRectangle       *gtk.Image
	imgEllipse         *gtk.Image
	imgFill            *gtk.Image
	imgPlay            *gtk.Image
	imgPause           *gtk.Image
	toolBtnPlayMode    *gtk.ToolButton
	toolBtnCurrentMode *gtk.ToolButton
	editarea1          *editArea
	colorsbar1         *colorsBar
	spritesbar1        *spritesBar
	application        *gtk.Application
	appWin             *gtk.ApplicationWindow
	aboutDialog1       *gtk.AboutDialog
	err                error
	fPlay              bool
)

const fileScheme = "file"

type URI string

func isWindowsDrivePath(path string) bool {
	if len(path) < 3 {
		return false
	}
	return unicode.IsLetter(rune(path[0])) && path[1] == ':'
}

func isWindowsDriveURIPath(uri string) bool {
	if len(uri) < 4 {
		return false
	}
	return uri[0] == '/' && unicode.IsLetter(rune(uri[1])) && uri[2] == ':'
}

func filename(uri URI) (string, error) {
	if uri == "" {
		return "", nil
	}
	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}
	if u.Scheme != fileScheme {
		return "", fmt.Errorf("only file URIs are supported, got %q from %q", u.Scheme, uri)
	}
	// If the URI is a Windows URI, we trim the leading "/" and uppercase
	// the drive letter, which will never be case sensitive.
	if isWindowsDriveURIPath(u.Path) {
		u.Path = strings.ToUpper(string(u.Path[1])) + u.Path[2:]
	}
	return u.Path, nil
}

func main() {

	// Create a new application.
	application, err = gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)
	errorCheck(err)

	// Connect function to application startup event, this is not required.
	application.Connect("startup", func() {
		log.Println("application startup")
	})

	// Connect function to application activate event
	application.Connect("activate", func() {
		log.Println("application activate")

		gresource1, err1 := gio.LoadGResource("myapp.gresource")
		errorCheck(err1)
		gio.RegisterGResource(gresource1)

		// Get the GtkBuilder UI definition in the glade file.
		builder, err := gtk.BuilderNewFromResource("/res/gtk_sprite_ed.glade")
		//builder, err := gtk.BuilderNewFromFile("res/gtk_sprite_ed.glade")
		errorCheck(err)

		// Map the handlers to callback functions, and connect the signals
		// to the Builder.
		signals := map[string]interface{}{
			"on_window1_destroy":             onMainWindowDestroy,
			"on_select_mode":                 onSelectMode,
			"on_pencil_mode":                 onPencilMode,
			"on_rectangle_mode":              onRectangleMode,
			"on_ellipse_mode":                onEllipseMode,
			"on_fill_mode":                   onFillMode,
			"on_play_mode":                   onPlayMode,
			"menu_item_quit_activate":        onMenuItemQuit,
			"menu_item_New_activate":         onMenuItemNew,
			"menu_item_open_activate":        onMenuItemOpen,
			"menu_item_save_activate":        onMenuItemSave,
			"menu_item_saveas_activate":      onMenuItemSaveAs,
			"item_undo_activate":             onMenuItemUndo,
			"item_cut_activate":              onMenuItemCut,
			"item_copy_activate":             onMenuItemCopy,
			"item_paste_activate":            onMenuItemPaste,
			"item_about_activate":            onMenuItemAbout,
			"item_flip_horizontaly_activate": onMenuItemFlipHorizontaly,
			"item_flip_verticaly_activate":   onMenuItemFlipVerticaly,
			"item_swing_left_activate":       onMenuItemSwingLeft,
			"item_swing_right_activate":      onMenuItemSwingRight,
		}
		builder.ConnectSignals(signals)

		// Get the object with the id of "main_window".
		obj, err := builder.GetObject("window1")
		errorCheck(err)
		// Verify that the object is a pointer to a gtk.ApplicationWindow.
		//win, err := isWindow(obj)
		//errorCheck(err)
		appWin, _ = obj.(*gtk.ApplicationWindow)
		appWin.SetTitle("Gotk3SpriteEd")

		colorsbar1 = ColorsBarNew()
		editarea1 = EditAreaNew()
		spritesbar1 = SpritesBarNew()

		//t_uri0, _ := gtk.TargetEntryNew("text/plain", gtk.TARGET_OTHER_APP, uint(gdk.TARGET_STRING))
		t_uri1, _ := gtk.TargetEntryNew("text/uri-list", gtk.TARGET_OTHER_APP, 0)
		//t_uri2, _ := gtk.TargetEntryNew("STRING", gtk.TARGET_OTHER_APP, uint(gdk.TARGET_STRING))
		appWin.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*t_uri1}, gdk.ACTION_COPY|gdk.ACTION_MOVE|gdk.ACTION_DEFAULT)
		appWin.Connect("drag-data-received", func(g *gtk.ApplicationWindow, ctx *gdk.DragContext, x int, y int, data *gtk.SelectionData, info uint, _ uint) {
			//fmt.Println("Drag-Data-Received")
			if data.GetLength() > 0 {
				uri := data.GetData()
				if strings.HasPrefix(string(uri), "file://") {
					// Remove \n and \r
					var t []byte
					s := string(uri)
					for i := 0; i < len(s); i++ {
						c := s[i]
						if c != '\n' && c != '\r' {
							t = append(t, c)
						}
					}
					s = string(t)
					fileName, _ := filename(URI(s))
					if strings.HasSuffix(fileName, ".png") ||
						strings.HasSuffix(fileName, ".PNG") {
						spritesbar1.LoadCurrentSprite(fileName)
					}

				}
			}
		})

		editarea1.SetForegroundColor(colorsbar1.GetForegroundColor())
		editarea1.SetBackgroundColor(colorsbar1.GetBackgroundColor())

		colorsbar1.SetColorsChangeCallback(colorsChangeCallback)

		if sprite1 := spritesbar1.GetCurrentSprite(); sprite1 != nil {
			editarea1.SetPixbuf(sprite1)
		}

		spritesbar1.SetSpriteChangeCallback(func(spr *gdk.Pixbuf) {
			if spr != nil {
				editarea1.SetPixbuf(spr)
			}
		})

		editarea1.SetColorsPickCallback(func(foreColor, backColor uint32) {
			colorsbar1.SetForegroundColor(foreColor)
			colorsbar1.SetBackgroundColor(backColor)
		})

		editarea1.SetPixBufCallback(func() {
			spritesbar1.da.QueueDraw()
		})

		hbox1, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
		hbox1.PackStart(editarea1, true, true, 0)
		hbox1.PackEnd(spritesbar1.da, false, true, 0)

		obj, _ = builder.GetObject("VBox2")
		vbox2 := obj.(*gtk.Box)
		vbox2.PackStart(hbox1, true, true, 2)
		vbox2.PackEnd(colorsbar1.da, false, false, 2)

		obj, _ = builder.GetObject("imageSelectBox1")
		imgSelectBox = obj.(*gtk.Image)

		obj, _ = builder.GetObject("imagePencil1")
		imgPencil = obj.(*gtk.Image)

		obj, _ = builder.GetObject("imageRectangle1")
		imgRectangle = obj.(*gtk.Image)

		obj, _ = builder.GetObject("imageEllipse1")
		imgEllipse = obj.(*gtk.Image)

		obj, _ = builder.GetObject("imageFill1")
		imgFill = obj.(*gtk.Image)

		obj, _ = builder.GetObject("imagePlay1")
		imgPlay = obj.(*gtk.Image)

		obj, _ = builder.GetObject("imagePause")
		imgPause = obj.(*gtk.Image)

		fPlay = false
		obj, _ = builder.GetObject("PlayMode")
		toolBtnPlayMode = obj.(*gtk.ToolButton)

		obj, _ = builder.GetObject("CurrentMode")
		toolBtnCurrentMode = obj.(*gtk.ToolButton)

		obj, _ = builder.GetObject("AboutDialog1")
		aboutDialog1 = obj.(*gtk.AboutDialog)

		// Show the Window and all of its components.
		appWin.ShowAll()
		application.AddWindow(appWin)

		//--Set Pencil Mode
		onPencilMode()

	})

	// Connect function to application shutdown event, this is not required.
	application.Connect("shutdown", func() {
		log.Println("application shutdown")
		colorsbar1.savePalette("palette.cfg")
	})

	// Launch the application
	os.Exit(application.Run(os.Args))

}

func errorCheck(e error) {
	if e != nil {
		// panic for any errors.
		log.Panic(e)
	}
}

func onMainWindowDestroy() {
	//--
	log.Println("onMainWindowDestroy")

}

func onSelectMode() {
	//--
	log.Println("onSelectMode")
	toolBtnCurrentMode.SetIconWidget(imgSelectBox)
	editarea1.SetSelectMode()

}

func onPencilMode() {
	//--
	log.Println("onPencilMode")
	toolBtnCurrentMode.SetIconWidget(imgPencil)
	editarea1.SetPencilMode()

}

func onRectangleMode() {
	//--
	log.Println("onRectangleMode")
	toolBtnCurrentMode.SetIconWidget(imgRectangle)
	editarea1.SetRectangleMode()

}

func onEllipseMode() {
	//--
	log.Println("onEllipseMode")
	toolBtnCurrentMode.SetIconWidget(imgEllipse)
	editarea1.SetEllipseMode()

}

func onFillMode() {
	//--
	log.Println("onFillMode")
	toolBtnCurrentMode.SetIconWidget(imgFill)
	editarea1.SetFillMode()

}

func updateFrame() {
	//--
	if fPlay {
		//log.Println("UpdateTimer")
		spritesbar1.NextFrame()
		glib.TimeoutAdd(500, updateFrame)
	}
}

func onPlayMode() {
	//--
	if fPlay {
		log.Println("onPauseMode")
		fPlay = false
		toolBtnPlayMode.SetIconWidget(imgPlay)
		spritesbar1.StopAnimate()

	} else {
		log.Println("onPlayMode")
		fPlay = true
		toolBtnPlayMode.SetIconWidget(imgPause)
		spritesbar1.StartAnimate()
		glib.TimeoutAdd(500, updateFrame)

	}

}

func colorsChangeCallback(foreColor, backColor uint32) {
	//--
	editarea1.SetForegroundColor(foreColor)
	editarea1.SetBackgroundColor(backColor)
}

func onMenuItemQuit() {
	//--
	log.Println("onMenuItemQuit")
	application.Quit()
}

func onMenuItemOpen() {
	//--
	log.Println("onMenuItemOpen")

	filechooser, _ := gtk.FileChooserDialogNewWith2Buttons(
		"Open...",
		appWin,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"Cancel",
		gtk.RESPONSE_DELETE_EVENT,
		"Open",
		gtk.RESPONSE_ACCEPT)
	filter, _ := gtk.FileFilterNew()
	filter.AddPattern("*.png")
	filter.SetName(".png")
	filechooser.AddFilter(filter)

	switcher := filechooser.Run()
	log.Println(switcher, ": switcher")
	if switcher == gtk.RESPONSE_ACCEPT {
		fileName := filechooser.GetFilename()
		if strings.HasSuffix(fileName, ".png") ||
			strings.HasSuffix(fileName, ".PNG") {
			spritesbar1.LoadCurrentSprite(fileName)
			log.Println(fileName)
		}
	}
	filechooser.Destroy()

}

func onMenuItemSaveAs() {
	//--
	log.Println("onMenuItemSaveAs")

	filechooser, _ := gtk.FileChooserDialogNewWith2Buttons(
		"Save as...",
		appWin,
		gtk.FILE_CHOOSER_ACTION_SAVE,
		"Cancel",
		gtk.RESPONSE_DELETE_EVENT,
		"Save",
		gtk.RESPONSE_ACCEPT)
	filter, _ := gtk.FileFilterNew()
	filter.AddPattern("*.png")
	filter.SetName(".png")
	filechooser.AddFilter(filter)

	switcher := filechooser.Run()
	log.Println(switcher, ": switcher ")
	if switcher == gtk.RESPONSE_ACCEPT {
		fileName := filechooser.GetFilename()
		spritesbar1.SaveAsCurrentSprite(fileName)
		log.Println(fileName)
	}
	filechooser.Destroy()

}

func onMenuItemSave() {
	//--
	log.Println("onMenuItemSave")
	fileName := spritesbar1.GetCurrentSpriteFileName()
	if fileName == "" {
		onMenuItemSaveAs()
	} else {
		spritesbar1.SaveCurrentSprite()
	}
}

func onMenuItemNew() {
	//--

	dialog, _ := gtk.DialogNew()
	dialog.SetTitle("Enter Sizes")
	dialog.SetModal(true)
	dialog.SetDefaultSize(200, 120)
	dialog.SetTransientFor(appWin)

	hbox1, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
	width_entry, _ := gtk.EntryNew()
	width_entry.SetText("32")
	width_entry.SetWidthChars(6)
	width_entry.SetAlignment(0.5)
	width_label, _ := gtk.LabelNew("Width : ")
	hbox1.PackStart(width_label, true, false, 4)
	hbox1.PackEnd(width_entry, true, false, 4)

	hbox2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
	height_entry, _ := gtk.EntryNew()
	height_entry.SetText("32")
	height_entry.SetWidthChars(6)
	height_entry.SetAlignment(0.5)
	height_label, _ := gtk.LabelNew("Height :")
	hbox2.PackStart(height_label, true, false, 4)
	hbox2.PackEnd(height_entry, true, false, 4)

	vbox, _ := dialog.GetContentArea()
	vbox.PackStart(hbox1, false, true, 4)
	vbox.PackStart(hbox2, false, true, 4)

	dlgBtnCancel, _ := dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)
	dlgBtnCancel.Connect("clicked", func() {
		dialog.Destroy()
	})
	dlgBtnOk, _ := dialog.AddButton("Ok", gtk.RESPONSE_OK)
	dlgBtnOk.Connect("clicked", func() {
		wString, _ := width_entry.GetText()
		hString, _ := height_entry.GetText()
		w, _ := strconv.ParseInt(wString, 10, 32)
		h, _ := strconv.ParseInt(hString, 10, 32)
		spritesbar1.NewCurrentSprite(int(w), int(h))
		dialog.Destroy()
	})
	dialog.ShowAll()
}

func onMenuItemUndo() {
	//--
	log.Println("onMenuItemUndo")
	editarea1.Undo()
}

func onMenuItemCut() {
	//--
	log.Println("onMenuItemCut")
	editarea1.CutSelect()
}

func onMenuItemCopy() {
	//--
	log.Println("onMenuItemCopy")
	editarea1.CopySelect()
}

func onMenuItemPaste() {
	//--
	log.Println("onMenuItemPaste")
	editarea1.PasteSelect()
}

func onMenuItemAbout() {
	//--
	aboutDialog1.SetDestroyWithParent(true)
	aboutDialog1.SetModal(true)
	aboutDialog1.SetTransientFor(appWin)
	aboutDialog1.SetPosition(gtk.WIN_POS_CENTER_ON_PARENT)
	aboutDialog1.Run()
	aboutDialog1.Hide()

}

func onMenuItemFlipHorizontaly() {
	//--
	log.Println("onMenuItemFlipHorizontaly")
	editarea1.BackupSprite()
	editarea1.FlipHorizontaly()
	editarea1.undo_mode = FLIP_HORIZONTALY

}

func onMenuItemFlipVerticaly() {
	//--
	log.Println("onMenuItemFlipVerticaly")
	editarea1.BackupSprite()
	editarea1.FlipVerticaly()
	editarea1.undo_mode = FLIP_VERTICALY

}

func onMenuItemSwingLeft() {
	//--
	log.Println("onMenuItemSwingLeft")
	if editarea1.imgBufBak.GetWidth() == editarea1.imgBufBak.GetHeight() {
		editarea1.BackupSprite()
		editarea1.SwingLeft()
		editarea1.undo_mode = SWING_LEFT
	}

}

func onMenuItemSwingRight() {
	//--
	log.Println("onMenuItemSwingRight")
	if editarea1.imgBufBak.GetWidth() == editarea1.imgBufBak.GetHeight() {
		editarea1.BackupSprite()
		editarea1.SwingRight()
		editarea1.undo_mode = SWING_RIGHT
	}
}
