package main

type defendObj struct {
}

type defendPacket struct {
	Type string    `json:"type"`
	Obj  defendObj `json:"obj"`
}

func opDefend(c *ntcpclient, packet defendPacket) {
	c.Player.Rl.Defend.Take()

	ng, err := c.Player.Defend()
	if err != nil {
		c.Respond(err.Error())
		return
	}
	c.RespondNil()

	updateDist.Push(c.Player, newUpdate("defend", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"defend"}`)), notifyChannelGroup, ng, c.Log)
}
