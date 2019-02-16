package main

import "time"

type undefendObj struct {
}

type undefendPacket struct {
	Type string      `json:"type"`
	Obj  undefendObj `json:"obj"`
}

func opUndefend(c *ntcpclient, packtet undefendPacket) {
	// c.Player.Rl.Defend.Take()
	time.Sleep(1000 * time.Millisecond)

	ng, err := c.Player.Undefend()
	if err != nil {
		c.Respond(err.Error())
		return
	}
	c.RespondNil()

	dist.PushGroup("game", ng.UserStringIDs(), struct {
		GRID string `json:"grid"`
		Type string `json:"undefend"`
	}{
		c.Player.GRenderID,
		"undefend",
	}, c.Log)
}
