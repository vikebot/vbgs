package main

import "time"

type defendObj struct {
}

type defendPacket struct {
	Type string    `json:"type"`
	Obj  defendObj `json:"obj"`
}

func opDefend(c *ntcpclient, packet defendPacket) {
	// c.Player.Rl.Defend.Take()
	time.Sleep(1000 * time.Millisecond)

	ng, err := c.Player.Defend()
	if err != nil {
		c.Respond(err.Error())
		return
	}
	c.RespondNil()

	dist.PushGroup("game", ng.UserIDs(), struct {
		GRID string `json:"grid"`
		Type string `json:"type"`
	}{
		c.Player.GRenderID,
		"defend",
	}, c.Log)
}
