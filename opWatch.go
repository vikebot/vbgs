package main

type watchObj struct {
}

type watchPacket struct {
	Type string   `json:"type"`
	Obj  watchObj `json:"obj"`
}

type watchResponse struct {
	HealthMatrix [][]int `json:"health_matrix"`
}

func opWatch(c *ntcpclient, packet watchPacket) {
	c.Player.Rl.Watch.Take()

	matrix, _ := c.Player.Watch()

	c.RespondObj(&watchResponse{
		HealthMatrix: matrix,
	})
}
