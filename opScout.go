package main

import (
	"strconv"
	"time"

	"github.com/vikebot/vbgs/vbge"
)

type scoutObj struct {
	Distance *int `json:"distance"`
}

type scoutPacket struct {
	Type string   `json:"type"`
	Obj  scoutObj `json:"obj"`
}

type scoutResponse struct {
	Counter int `json:"counter"`
}

func opScout(c *ntcpclient, packet scoutPacket) {
	// c.Player.Rl.Scout.Take()
	time.Sleep(500 * time.Millisecond)
	if packet.Obj.Distance == nil {
		c.Respond("Invalid packet. `obj.distance' missing")
		return
	}

	distance := *packet.Obj.Distance
	if !vbge.IsDistance(distance) {
		c.RespondFmt("Invalid packet. '%s' is not a valid value for '.obj.distance'", strconv.Itoa(*packet.Obj.Distance))
		return
	}

	counter, ng, relPos, err := c.Player.Scout(distance)
	if err != nil {
		c.Respond(err.Error())
		return
	}

	c.RespondObj(&scoutResponse{
		Counter: counter,
	})

	for i := range ng {
		dist.GetClient(ng[i].UserID).Push("game",
			struct {
				GRID     string           `json:"grid"`
				Type     string           `json:"type"`
				Distance int              `json:"distance"`
				Loc      *vbge.ARLocation `json:"loc"`
			}{
				c.Player.GRenderID,
				"scout",
				distance,
				relPos[i].ToARLocation(),
			}, c.Log)
	}
}
