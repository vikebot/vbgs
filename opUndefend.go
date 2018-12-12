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

	dist.PushGroup("game", ng.UserIDs(), struct {
		GRID string `json:"grid"`
		Type string `json:"undefend"`
	}{
		c.Player.GRenderID,
		"undefend",
	}, c.Log)
}
