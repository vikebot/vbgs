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

	dist.PushGroup("game", ng.UserIDs(), struct {
		GRID string `json:"grid"`
		Type string `json:"type"`
	}{
		c.Player.GRenderID,
		"environment",
	}, c.Log)
}
