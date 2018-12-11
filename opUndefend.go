package main

type undefendObj struct {
}

type undefendPacket struct {
	Type string      `json:"type"`
	Obj  undefendObj `json:"obj"`
}

func opUndefend(c *ntcpclient, packtet undefendPacket) {
	c.Player.Rl.Defend.Take()

	ng, err := c.Player.Defend()
	if err != nil {
		c.Respond(err.Error())
		return
	}
	c.RespondNil()

	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"undefend"}`)), notifyChannelGroup, ng, c.log)
}
