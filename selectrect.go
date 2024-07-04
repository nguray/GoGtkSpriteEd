package main

type SelectRect struct {
	left        int
	top         int
	right       int
	bottom      int
	mode        int
	selCorner   int
	mouseStartX float64
	mouseStartY float64
	savleft     int
	savtop      int
	savright    int
	savbottom   int
}

func SelectRectNew(l, t, r, b int) *SelectRect {
	sr := &SelectRect{l, t, r, b, 0, 0, 0, 0, 0, 0, 0, 0}
	return sr
}

func (sr *SelectRect) BackupPosition() {
	sr.savleft = sr.left
	sr.savtop = sr.top
	sr.savright = sr.right
	sr.savbottom = sr.bottom

}

func (sr *SelectRect) RestorePosition() {
	sr.left = sr.savleft
	sr.top = sr.savtop
	sr.right = sr.savright
	sr.bottom = sr.savbottom

}

func (sr *SelectRect) GetCorner(n int) (int, int) {
	switch n {
	case 0:
		return sr.left, sr.top
	case 1:
		return sr.right, sr.top
	case 2:
		return sr.right, sr.bottom
	case 3:
		return sr.left, sr.bottom
	default:
		return 0, 0
	}
}

func (sr *SelectRect) SetCorner(n int, x, y int) {
	switch n {
	case 0:
		sr.left = x
		sr.top = y
	case 1:
		sr.right = x
		sr.top = y
	case 2:
		sr.right = x
		sr.bottom = y
	case 3:
		sr.left = x
		sr.bottom = y
	}
}

func (sr *SelectRect) Normalize() {
	if sr.left > sr.right {
		sr.left, sr.right = sr.right, sr.left
	}
	if sr.top > sr.bottom {
		sr.top, sr.bottom = sr.bottom, sr.top
	}
}

func (sr *SelectRect) Empty() {
	sr.left = 0
	sr.top = 0
	sr.right = 0
	sr.bottom = 0
}

func (sr *SelectRect) IsEmpty() bool {
	return sr.left == sr.right || sr.top == sr.bottom
}

func (sr *SelectRect) Init() {
	sr.Empty()
	sr.mode = 0
	sr.mouseStartX = 0
	sr.mouseStartY = 0
}

func (sr *SelectRect) Width() int {
	return sr.right - sr.left + 1
}

func (sr *SelectRect) Height() int {
	return sr.bottom - sr.top + 1
}

func (sr *SelectRect) Offset(dx int, dy int) {
	sr.left += dx
	sr.right += dx
	sr.top += dy
	sr.bottom += dy
}
