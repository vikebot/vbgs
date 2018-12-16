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

	matrix, ngl := c.Player.Environment()

	c.RespondObj(&environmentResponse{
		EnvironmentMatrix: matrix,
	})

	dist.PushGroup("game", ngl.UserIDs(), struct {
		GRID string `json:"grid"`
		Type string `json:"type"`
	}{
		c.Player.GRenderID,
		"environment",
	}, c.Log)
}
