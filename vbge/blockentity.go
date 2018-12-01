package vbge

// BlockEntity represents a single point (block) in the map. It holds infos
// about it's environment and possible residents.
type BlockEntity struct {
	Resident  *Player
	Blocktype string
}

// HasResident reports if a player is currently in this block or not.
func (be *BlockEntity) HasResident() bool {
	return be.Resident != nil
}

// JoinArea marks the passed player as the current resident of this block.
func (be *BlockEntity) JoinArea(p *Player) {
	be.Resident = p
}

// LeaveArea dismarks the player which is currently the resident of this block
// to be a resident.
func (be *BlockEntity) LeaveArea() {
	be.Resident = nil
}
