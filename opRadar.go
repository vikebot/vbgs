package main

type radarObj struct {
}

type radarPacket struct {
	Type string   `json:"type"`
	Obj  radarObj `json:"obj"`
}

type radarResponse struct {
	Counter int `json:"counter"`
}

func opRadar(c *ntcpclient, packet radarPacket) {
	c.Player.Rl.Radar.Take()

	counter, _ := c.Player.Radar()

	c.RespondObj(&radarResponse{
		Counter: counter,
	})
}
