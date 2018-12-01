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
	health, ng, err := c.Player.Attack(func(e *vbge.Player, notifyGroup vbge.NotifyGroup) {
		updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"death"}`)), notifyChannelGroup, notifyGroup, c.LogCtx)
	}, func(e *vbge.Player, notifyGroup vbge.NotifyGroup) {
		playerMapentity, err := vbge.GetViewableMapentity(vbge.RenderWidth, vbge.RenderHeight, e.UserID, battle, false)
		if err != nil {
			return
		}

		pme, err := json.Marshal(playerMapentity)
		if err != nil {
			return
		}

		for i := 0; i < len(notifyGroup); i++ {
			l := notifyGroup[i].Location.RelativeFrom(e.Location)
			if notifyGroup[i].UserID == e.UserID {
				updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"spawn", "loc":{"isabs":false,"x":`+strconv.Itoa(l.X)+`,"y":`+strconv.Itoa(l.Y)+`},"playermapentity":`+string(pme)+`}`)), notifyChannelGroup, nil, c.LogCtx)
			} else {
				updateDist.Push(notifyGroup[i], newUpdate("game", []byte(`{"grid":"`+notifyGroup[i].GRenderID+`","type":"selfspawn", "loc":{"isabs":false,"x":`+strconv.Itoa(l.X)+`,"y":`+strconv.Itoa(l.Y)+`}}`)), notifyChannelGroup, nil, c.LogCtx)
			}
		}
	})
	if err != nil {
		c.Respond(err.Error())
		updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"attack"}`)), notifyChannelGroup, ng, c.LogCtx)
		return
	}

	c.RespondObj(&attackResponse{
		Health: health,
	})

	updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"attack"}`)), notifyChannelGroup, ng, c.LogCtx)
}
