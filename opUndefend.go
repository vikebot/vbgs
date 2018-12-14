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

	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"undefend"}`)), notifyChannelGroup, ng, c.LogCtx)
}
