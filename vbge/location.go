package vbge

// Location represents a two-dimensional point in a matrix. It's values are
// zero-based. Therfore (0, 0) is the first point in the upper left.
type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func newLocation(y, x int) *Location {
	return &Location{
		X: x,
		Y: y,
	}
}

// DeepCopy allocates new memory and copies all the contents of the current
// location into the newly allocated one. Afterwards the pointer of the new
// struct is returned.
func (l *Location) DeepCopy() *Location {
	nl := Location{}
	nl.X = l.X
	nl.Y = l.Y
	return &nl
}

// AddDirection manipulates the current location by a factor of one into
// the cardinal direction specified by the `dir` parameter. `dir` must be a
// valid direction representation. Check with `vbge.IsDir`. This function
// doesn't validate the resulting location. Therefore it's maybe out-of-map.
// Check with `(*Location).IsInMap`
func (l *Location) AddDirection(dir string) {
	switch dir {
	case dirNorth:
		l.Y--
	case dirEast:
		l.X++
	case dirSouth:
		l.Y++
	case dirWest:
		l.X--
	}
}

// IsInMap checks whether the location is in the map or not. Determined with
// `(X && Y > 0) && (X < MapWidth) && (Y < MapHeight)`
func (l *Location) IsInMap() bool {
	return (l.X >= 0 &&
		l.X < MapWidth &&
		l.Y >= 0 &&
		l.Y < MapHeight)
}

// RelativeFrom returns the relative position from the given
// player location (pl) => the relative position is from the
// view of l
func (l *Location) RelativeFrom(pl *Location) *Location {
	nl := Location{
		X: l.X - pl.X,
		Y: l.Y - pl.Y,
	}
	return &nl
}

// ToARLocation converts a Location to a ARLocation with isAbs = faalse
func (l *Location) ToARLocation() *ARLocation {
	return &ARLocation{
		IsAbs:    false,
		Location: *l,
	}
}

// ARLocation is a type to define a location as absolut or relative
type ARLocation struct {
	Location
	IsAbs bool `json:"isabs"`
}
