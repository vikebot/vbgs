package vbge

import "errors"

// Battle represents a logical game instance without runtime or network infos
type Battle struct {
	Map     *MapEntity
	Players map[int]*Player
}

// GetGRIDFromPlayerID returns the GRID
func (b *Battle) GetGRIDFromPlayerID(id int) (GRID string, err error) {
	for _, p := range b.Players {
		if p.UserID == id {
			return p.GRenderID, nil
		}
	}
	return "", errors.New("No User found with the given ID")
}
