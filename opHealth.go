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

	c.RespondObj(&healthResponse{
		Health: c.Player.Health.HealthSynced(),
	})
}
