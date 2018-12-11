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

	matrix, ng, err := c.Player.Watch()
	if err != nil {
		c.RespondFmt(err.Error())
		return
	}

	c.RespondObj(&watchResponse{
		HealthMatrix: matrix,
	})
	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"watch"}`)), notifyChannelGroup, ng, c.log)
}
