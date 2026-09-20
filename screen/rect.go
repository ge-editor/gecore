package screen

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (r Rect) IsValid() bool {
	return r.Width >= 0 && r.Height >= 0
}

func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// FitsIn checks if rectangle r fits entirely inside rectangle b.
func (r Rect) FitsIn(b Rect) bool {
	return r == r.Intersection(b)
}

// Intersection returns the intersection rectangle of r and b.
// Intersection: the point or set of points where two lines or surfaces meet and cross each other.
func (r Rect) Intersection(b Rect) Rect {
	// No intersection cases
	// r and b do not overlap
	if r.X >= b.X+b.Width || r.Y >= b.Y+b.Height {
		return Rect{0, 0, 0, 0}
	}

	// r and b do not overlap
	if r.X+r.Width <= b.X || r.Y+r.Height <= b.Y {
		return Rect{0, 0, 0, 0}
	}

	// Adjust X or Width
	// Shrink the width of r to avoid overlapping with b
	if r.X+r.Width > b.X+b.Width {
		r.Width = (b.X + b.Width) - r.X
	}

	if r.X < b.X {
		r.Width -= b.X - r.X
		r.X = b.X
	}

	// Adjust Y or Height
	if r.Y+r.Height > b.Y+b.Height {
		r.Height = (b.Y + b.Height) - r.Y
	}

	if r.Y < b.Y {
		r.Height -= b.Y - r.Y
		r.Y = b.Y
	}

	return r
}
