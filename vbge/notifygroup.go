package vbge

// NotifyGroup specifies a set of `*Player` that should be informed about
// a specific event.
type NotifyGroup []*Player

// UserIDs returns all userIDs from a NotifyGroup in an slice of int
func (ng NotifyGroup) UserIDs() (userIDs []int) {
	userIDs = make([]int, len(ng))

	for i := range ng {
		userIDs[i] = ng[i].UserID
	}
	return
}
