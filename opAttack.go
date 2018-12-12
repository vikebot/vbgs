package main

import (
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
	health, ng, relPos, err := c.Player.Attack(
		// func onHit
		func(e *vbge.Player, health int, ng vbge.NotifyGroup) {
			dist.PushGroup("game", ng.UserIDs(), struct {
				GRID  string `json:"grid"`
				Type  string `json:"type"`
				Value int    `json:"health"`
			}{
				e.GRenderID,
				"health",
				health,
			}, c.Log)
		},
		// func beforeRespawn
		func(e *vbge.Player, ng vbge.NotifyGroup) {
			dist.PushGroup("game", ng.UserIDs(), struct {
				GRID string `json:"grid"`
				Type string `json:"type"`
			}{
				e.GRenderID,
				"death",
			}, c.Log)
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

					dist.GetClient(ng[i].UserID).Push("game", struct {
						GRID       string          `json:"grid"`
						Type       string          `json:"type"`
						PlayerInfo vbge.PlayerResp `json:"playerinfo"`
					}{
						e.GRenderID,
						"spawn",
						playerResp,
					}, c.Log)
				} else {
					playerMapentity, err := vbge.GetViewableMapentity(vbge.RenderWidth, vbge.RenderHeight, e.UserID, battle, false)
					if err != nil {
						return
					}

					dist.GetClient(ng[i].UserID).Push("game", struct {
						GRID            string                  `json:"grid"`
						Type            string                  `json:"type"`
						Loc             *vbge.ARLocation        `json:"loc"`
						PlayerMapEntity *vbge.ViewableMapentity `json:"playermapentity"`
					}{
						e.GRenderID,
						"selfspawn",
						l.ToARLocation(),
						playerMapentity,
					}, c.Log)
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

			dist.PushGroup("game", ng.UserIDs(), struct {
				Stats []playerStats
			}{
				ps,
			}, c.Log)
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
		dist.GetClient(ng[i].UserID).Push("game", struct {
			GRID string           `json:"grid"`
			Type string           `json:"type"`
			Loc  *vbge.ARLocation `json:"loc"`
		}{
			c.Player.GRenderID,
			"attack",
			relPos[i].ToARLocation(),
		}, c.Log)
	}
}
