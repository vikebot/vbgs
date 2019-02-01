package vbge

import "errors"

var (
	// ErrNoEnemy is an error if there's no enemy
	ErrNoEnemy = errors.New("No Enemy")

	// ErrOutOfMap is an error when a position is out of the map
	ErrOutOfMap = errors.New("Position is out of Map")

	// ErrNoMoveOutOfMap comes when the player wants to move out of the map
	ErrNoMoveOutOfMap = errors.New("Cannot move outside the map")

	// ErrHasResident appears when a block already has a resident and an action
	// should happen
	ErrHasResident = errors.New("Block already has a resident")

	// ErrInaccessable appearse when a block is not accessable by a player
	ErrInaccessable = errors.New("Location is not accessable due to the block type")

	// ErrAlreadyDef describes that the player already is defending
	ErrAlreadyDef = errors.New("Player is already defending")

	// ErrAlreadyUndef describes that the player is already undefending
	ErrAlreadyUndef = errors.New("Player is already undefending")

	// ErrCantMoveOFDefending means thath the player cant move because of defending
	ErrCantMoveOFDefending = errors.New("Player is not able to move, because of defending")
)
