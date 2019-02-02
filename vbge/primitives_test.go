package vbge

import (
	"testing"

	"github.com/vikebot/vbcore"
)

func TestIsAngle(t *testing.T) {
	tests := []struct {
		name           string
		angleCandidate string
		want           bool
	}{
		{"left", "left", true},
		{"right", "left", true},
		{"angleLeft", angleLeft, true},
		{"angleRight", angleRight, true},

		{"Empty", "", false},
		{"dirNorth", dirNorth, false},
		{"dirEast", dirEast, false},
		{"dirSouth", dirSouth, false},
		{"dirWest", dirWest, false},
		{"Random 1", vbcore.FastRandomString(4), false},
		{"Random 2", vbcore.FastRandomString(8), false},
		{"Random 3", vbcore.FastRandomString(16), false},
		{"Random 4", vbcore.FastRandomString(32), false},
		{"Random 5", vbcore.FastRandomString(64), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAngle(tt.angleCandidate); got != tt.want {
				t.Errorf("IsAngle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	tests := []struct {
		name         string
		dirCandidate string
		want         bool
	}{
		{"north", "north", true},
		{"east", "east", true},
		{"south", "south", true},
		{"west", "west", true},
		{"dirNorth", dirNorth, true},
		{"dirEast", dirEast, true},
		{"dirSouth", dirSouth, true},
		{"dirWest", dirWest, true},

		{"Empty", "", false},
		{"angleLeft", angleLeft, false},
		{"angleRight", angleRight, false},
		{"Random 1", vbcore.FastRandomString(4), false},
		{"Random 2", vbcore.FastRandomString(8), false},
		{"Random 3", vbcore.FastRandomString(16), false},
		{"Random 4", vbcore.FastRandomString(32), false},
		{"Random 5", vbcore.FastRandomString(64), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDir(tt.dirCandidate); got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBlocktype(t *testing.T) {
	tests := []struct {
		name               string
		blocktypeCandidate string
		want               bool
	}{
		{"swamp", "swamp", true},
		{"stonetile", "stonetile", true},
		{"dirt", "dirt", true},
		{"grass", "grass", true},
		{"lava", "lava", true},
		{"lavarock", "lavarock", true},
		{"water", "water", true},
		{"endofmap", "endofmap", true},
		{"fog", "fog", true},
		{"swamp", blockSwamp, true},
		{"stonetile", blockStonetile, true},
		{"dirt", blockDirt, true},
		{"grass", blockGrass, true},
		{"lava", blockLava, true},
		{"lavarock", blockLavarock, true},
		{"water", blockWater, true},
		{"endofmap", blockEndOfMap, true},
		{"fog", blockFog, true},

		{"Empty", "", false},
		{"Random 1", vbcore.FastRandomString(4), false},
		{"Random 2", vbcore.FastRandomString(8), false},
		{"Random 3", vbcore.FastRandomString(16), false},
		{"Random 4", vbcore.FastRandomString(32), false},
		{"Random 5", vbcore.FastRandomString(64), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlocktype(tt.blocktypeCandidate); got != tt.want {
				t.Errorf("IsBlocktype() = %v, want %v", got, tt.want)
			}
		})
	}
}
