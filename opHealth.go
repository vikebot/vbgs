package main

type healthObj struct {
}

type healthPacket struct {
	Type string    `json:"type"`
	Obj  healthObj `json:"obj"`
}

type healthResponse struct {
	Health int `json:"value"`
}

func opHealth(c *ntcpclient, packtet healthPacket) {
	c.Player.Rl.Health.Take()

	health := c.Player.GetHealth()

	c.RespondObj(&healthResponse{
		Health: health,
	})
}
