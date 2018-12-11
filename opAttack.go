package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/vikebot/vbgs/vbge"
)

type attackObj struct {
}

type attackPacket struct {
	Type string    `json:"type"`
	Obj  attackObj `json:"obj"`
}

type attackResponse struct {
	Health int `json:"health"`
}

func opAttack(c *ntcpclient, packet attackPacket) {
	// c.Player.Rl.Attack.Take()
	time.Sleep(300 * time.Millisecond)
	health, ng, err := c.Player.Attack(
		// func onHit
		func(e *vbge.Player, health int, ng vbge.NotifyGroup) {
			updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"health","value":`+strconv.Itoa(health)+`}`)), notifyChannelGroup, ng, c.log)
		},
		// func beforeRespawn
		func(e *vbge.Player, ng vbge.NotifyGroup) {
			updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"death"}`)), notifyChannelGroup, ng, c.log)
		},
		// func afterRespawn
		func(e *vbge.Player, ng vbge.NotifyGroup) {
			playerMapentity, err := vbge.GetViewableMapentity(vbge.RenderWidth, vbge.RenderHeight, e.UserID, battle, false)
			if err != nil {
				return
			}

			pme, err := json.Marshal(playerMapentity)
			if err != nil {
				return
			}

			for i := range ng {
				l := ng[i].Location.RelativeFrom(e.Location)
				if ng[i].UserID == e.UserID {
					updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"spawn", "loc":{"isabs":false,"x":`+strconv.Itoa(l.X)+`,"y":`+strconv.Itoa(l.Y)+`},"playermapentity":`+string(pme)+`}`)), notifyChannelGroup, nil, c.log)
				} else {
					updateDist.Push(ng[i], newUpdate("game", []byte(`{"grid":"`+ng[i].GRenderID+`","type":"selfspawn", "loc":{"isabs":false,"x":`+strconv.Itoa(l.X)+`,"y":`+strconv.Itoa(l.Y)+`}}`)), notifyChannelGroup, nil, c.log)
				}
			}
		})
	if err != nil {
		c.Respond(err.Error())
		updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"attack"}`)), notifyChannelGroup, ng, c.log)
		return
	}

	c.RespondObj(&attackResponse{
		Health: health,
	})

	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"attack"}`)), notifyChannelGroup, ng, c.log)
}
