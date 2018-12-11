package main

import "strconv"

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

	health, ng := c.Player.GetHealth()

	c.RespondObj(&healthResponse{
		Health: health,
	})

	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"health","value":"`+strconv.Itoa(health)+`"}`)), notifyChannelGroup, ng, c.Log)
}
