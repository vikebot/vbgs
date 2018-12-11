package main

type environmentObj struct {
}

type environmentPacket struct {
	Type string         `json:"type"`
	Obj  environmentObj `json:"obj"`
}

type environmentResponse struct {
	EnvironmentMatrix [][]string `json:"environment_matrix"`
}

func opEnvironment(c *ntcpclient, packet environmentPacket) {
	c.Player.Rl.Environment.Take()

	matrix, ng, err := c.Player.Environment()

	if err != nil {
		c.RespondFmt(err.Error())
		return
	}

	c.RespondObj(&environmentResponse{
		EnvironmentMatrix: matrix,
	})

	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"environment"}`)), notifyChannelGroup, ng, c.log)
}
