package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
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
	health, ng, relPos, err := c.Player.Attack(
		// func onHit
		func(e *vbge.Player, health int, ng vbge.NotifyGroup) {
			updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"health","value":`+strconv.Itoa(health)+`}`)), notifyChannelGroup, ng, c.LogCtx)
		},
		// func beforeRespawn
		func(e *vbge.Player, ng vbge.NotifyGroup) {
			updateDist.Push(e, newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"death"}`)), notifyChannelGroup, ng, c.LogCtx)
		},
		// func afterRespawn
		func(e *vbge.Player, ng vbge.NotifyGroup) {
			// create generic player response packet
			playerResp := vbge.PlayerResp{
				GRID:          e.GRenderID,
				Health:        e.Health.HealthSynced(),
				CharacterType: e.CharacterType,
				WatchDir:      e.WatchDir,
			}

			for i := range ng {
				l := e.Location.RelativeFrom(ng[i].Location)
				if ng[i].UserID != e.UserID {
					playerResp.Location = *l

					// marshal response
					pr, err := json.Marshal(playerResp)
					if err != nil {
						c.LogCtx.Error("unable to marshal vbge.PlayerResp", zap.Error(err))
						return
					}

					updateDist.Push(ng[i], newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"spawn","playerinfo":`+string(pr)+`}`)), notifyChannelPrivate, nil, c.LogCtx)
				} else {
					playerMapentity, err := vbge.GetViewableMapentity(vbge.RenderWidth, vbge.RenderHeight, e.UserID, battle, false)
					if err != nil {
						return
					}

					pme, err := json.Marshal(playerMapentity)
					if err != nil {
						return
					}

					updateDist.Push(ng[i], newUpdate("game", []byte(`{"grid":"`+e.GRenderID+`","type":"selfspawn", "loc":{"isabs":false,"x":`+strconv.Itoa(l.X)+`,"y":`+strconv.Itoa(l.Y)+`},"playermapentity":`+string(pme)+`}`)), notifyChannelPrivate, nil, c.LogCtx)
				}
			}
		},
		// func ChangedStats
		func(p []vbge.Player, ng vbge.NotifyGroup) {
			var ps playersStats

			for i := range p {
				ps = append(ps, playerStats{
					GRID:   p[i].GRenderID,
					Kills:  p[i].Kills,
					Deaths: p[i].Deaths,
				})
			}

			statsObj, err := json.Marshal(ps)
			if err != nil {
				c.LogCtx.Error("unable to marshal playerStats", zap.Error(err))
				return
			}

			updateDist.Push(nil, newUpdate("stats", statsObj), notifyChannelGroup, ng, c.LogCtx)
		})
	if err != nil {
		c.Respond(err.Error())
		// updateDist.Push(c.Player, newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"attack"}`)), notifyChannelGroup, ng, c.LogCtx)
		return
	}

	c.RespondObj(&attackResponse{
		Health: health,
	})

	for i := range ng {
		updateDist.Push(ng[i], newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+
			`","type":"attack","loc":{"isabs":false,"x":`+strconv.Itoa(relPos[i].X)+`,"y":`+strconv.Itoa(relPos[i].Y)+`}}`)),
			notifyChannelPrivate, nil, c.LogCtx)
	}
}
