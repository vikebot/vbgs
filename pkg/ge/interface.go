package ge

import (
	"github.com/vikebot/vbgs/pkg/syncc"
)

type Angle string
type Direction string
type UserID int
type CharacterType string

type Location struct {
	X int64
	Y int64
}
type ARLocation struct {
	marshalToAbs bool
	Abs          Location
	Rel          Location
}

type NotifyGroup []UserID
type NotifyGroupLocation []struct {
	UserID UserID
	ARLoc  ARLocation
}

type Attacker interface {
	DamagePoints() int
}

type Health interface {
	Points() int
	TakeDamage(attacker Attacker, sr syncc.Request)
}

type Player interface {
	UserID() UserID
	CharacterType() CharacterType
	Location() Location
	WatchDir() Direction
	Health() Health
	CharacterRotater
	CharacterMover
	CharacterRadarer
	CharacterScouter
}

type CharacterRotater interface {
	Rotate(angle Angle, sr syncc.Request) (ng NotifyGroupLocation)
}
type CharacterMover interface {
	Move(dir Direction, sr syncc.Request) (ng NotifyGroupLocation, err error)
}
type CharacterRadarer interface {
	Radar(sr syncc.Request) (count int, ng NotifyGroupLocation, err error)
}
type CharacterScouter interface {
	Scout(distance int, sr syncc.Request) (count int, ng NotifyGroupLocation, err error)
}
type CharacterEnvironmenter interface {
	Environment()
}

type MapSegment struct {
}

type Chunk struct {
}

type Map [][]Chunk
