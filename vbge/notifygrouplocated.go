package vbge

// NotifyGroupLocated specifies a NotifyGroup that also includes location
// information for the specific players.
type NotifyGroupLocated []*NotifyGroupLocatedEntity

// NotifyGroupLocatedEntity is a single Player that should be notified about
// an event, that occurred at ARLoc (a relative location to the Player).
type NotifyGroupLocatedEntity struct {
	Player *Player
	ARLoc  *ARLocation
}

// UserIDs returns all userIDs from a NotifyGroupLocated in an slice of int
func (ngl NotifyGroupLocated) UserIDs() (userIDs []int) {
	userIDs = make([]int, len(ngl))

	for i := range ngl {
		userIDs[i] = ngl[i].Player.UserID
	}
	return
}
