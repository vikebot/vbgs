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

	counter, ngl := c.Player.Scout(distance)

	c.RespondObj(&scoutResponse{
		Counter: counter,
	})

	for _, entity := range ngl {
		dist.GetClient(entity.Player.UserID).Push("game",
			struct {
				GRID     string           `json:"grid"`
				Type     string           `json:"type"`
				Distance int              `json:"distance"`
				Loc      *vbge.ARLocation `json:"loc"`
			}{
				c.Player.GRenderID,
				"scout",
				distance,
				entity.ARLoc,
			}, c.Log)
	}
}
