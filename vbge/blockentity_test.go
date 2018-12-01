package vbge

import "testing"

func newBlockEntity() BlockEntity {
	be := BlockEntity{
		Resident: &Player{
			UserID:   1,
			Map:      NewMapEntity(MapWidth, MapHeight),
			WatchDir: dirNorth,
			Location: &Location{
				X: HrHeight,
				Y: HrWidth,
			},
		},
		Blocktype: blockDirt,
	}
	return be
}

func TestBlockEntity_LeaveArea(t *testing.T) {
	tests := []struct {
		name string
		be   BlockEntity
	}{
		{"Test01", newBlockEntity()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.be.JoinArea(tt.be.Resident)
			tt.be.LeaveArea()
			if tt.be.HasResident() == true {
				t.Fail()
				t.Errorf(tt.name)
			}
			if tt.be.Blocktype != "dirt" {
				t.Fail()
				t.Error(tt.name + " " + tt.be.Blocktype)
			}
		})
	}
}
