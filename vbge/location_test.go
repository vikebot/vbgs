package vbge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocation(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		Name string
		Loc  *Location
	}{
		{"Test01: create new Location", &Location{X: 10, Y: 10}},
		{"Test02: use negative Location", &Location{X: -10, Y: -10}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			newLoc := newLocation(tt.Loc.X, tt.Loc.Y)

			assert.Equal(tt.Loc, newLoc)
		})
	}
}

func TestNewARLocation(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		Name  string
		Loc   *Location
		ARLoc *ARLocation
	}{
		{"Test01: loc to ARloc", &Location{X: 10, Y: 10}, &ARLocation{IsAbs: false, Location: Location{X: 10, Y: 10}}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			arLoc := tt.Loc.ToARLocation()

			assert.Equal(tt.ARLoc, arLoc)
		})
	}
}

func TestAddDirection(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		Name      string
		Loc       *Location
		Direction string
		WantedLoc *Location
	}{
		{"Test01: Add direciton north", &Location{X: 10, Y: 10}, dirNorth, &Location{X: 10, Y: 9}},
		{"Test01: Add direciton east", &Location{X: 10, Y: 10}, dirEast, &Location{X: 11, Y: 10}},
		{"Test01: Add direciton south", &Location{X: 10, Y: 10}, dirSouth, &Location{X: 10, Y: 11}},
		{"Test01: Add direciton west", &Location{X: 10, Y: 10}, dirWest, &Location{X: 9, Y: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Loc.AddDirection(tt.Direction)

			assert.Equal(tt.WantedLoc, tt.Loc)
		})
	}
}

func TestDeepCopy(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		Name string
		Loc  *Location
	}{
		{"Test01: Test basic deep copy", &Location{X: 10, Y: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			newLoc := tt.Loc.DeepCopy()

			assert.Equal(tt.Loc, newLoc)
		})
	}
}
