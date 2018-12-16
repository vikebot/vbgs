package main

import (
	"time"

	"github.com/vikebot/vbgs/vbge"
)

type rotateObj struct {
	Angle *string `json:"angle"`
}
type rotatePacket struct {
	Type string    `json:"type"`
	Obj  rotateObj `json:"obj"`
}

func opRotate(c *ntcpclient, packet rotatePacket) {
	// c.Player.Rl.Rotate.Take()
	time.Sleep(500 * time.Millisecond)

	if packet.Obj.Angle == nil {
		c.Respond("Invalid packet. '.obj.angle' missing")
		return
	}

	angle := *packet.Obj.Angle
	if !vbge.IsAngle(angle) {
		c.RespondFmt("Invalid packet. '%s' is not a valid value for '.obj.angle'", angle)
		return
	}

	ngl := c.Player.Rotate(angle)
	c.RespondNil()

	for _, entity := range ngl {
		dist.GetClient(entity.Player.UserID).Push("game",
			struct {
				GRID  string           `json:"grid"`
				Type  string           `json:"type"`
				Angle string           `json:"angle"`
				Loc   *vbge.ARLocation `json:"loc"`
			}{
				c.Player.GRenderID,
				"rotate",
				angle,
				entity.ARLoc,
			},
			c.Log)
	}
}
