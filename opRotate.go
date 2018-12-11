package main

import (
	"strconv"
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

	ng, relPos := c.Player.Rotate(angle)
	c.RespondNil()

	for i := range ng {
		updateDist.Push(ng[i], newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+
			`","type":"rotate","angle": "`+angle+`"`+
			`,"loc":{"isabs":false,"x":`+strconv.Itoa(relPos[i].X)+`,
	"y":`+strconv.Itoa(relPos[i].Y)+`}}`)), notifyChannelPrivate, nil, c.Log)

	}

}
