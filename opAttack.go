package main

import (
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

	health, ngl, err := c.Player.Attack(
		// func onHit
		func(e *vbge.Player, health int, ngl vbge.NotifyGroupLocated) {
			dist.PushGroup("game", ngl.UserStringIDs(), struct {
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
		func(e *vbge.Player, ngl vbge.NotifyGroupLocated) {
			dist.PushGroup("game", ngl.UserStringIDs(), struct {
				GRID string `json:"grid"`
				Type string `json:"type"`
			}{
				e.GRenderID,
				"death",
			}, c.Log)
		},
		// func afterRespawn
		func(enemy *vbge.Player, ngl vbge.NotifyGroupLocated) error {
			// create generic player response packet
			playerResp := vbge.PlayerResp{
				GRID:          enemy.GRenderID,
				Health:        enemy.Health.HealthSynced(),
				CharacterType: enemy.CharacterType,
				WatchDir:      enemy.WatchDir,
			}

			// Inform the people around the enemies new location, that he has just
			// spawned.
			for _, entity := range ngl {
				if entity.Player.UserID != enemy.UserID {
					// set current entities location for response packet
					playerResp.Location = entity.ARLoc

					// send notification
					dist.GetClient(strconv.Itoa(entity.Player.UserID)).Push("game", struct {
						GRID       string          `json:"grid"`
						Type       string          `json:"type"`
						PlayerInfo vbge.PlayerResp `json:"playerinfo"`
					}{
						enemy.GRenderID,
						"spawn",
						playerResp,
					}, c.Log)
				}
			}

			// Inform the enemy itself that he has respawned
			playerMapentity, err := vbge.GetViewableMapentity(vbge.RenderWidth, vbge.RenderHeight, enemy.UserID, battle, false)
			if err != nil {
				return err
			}
			dist.GetClient(strconv.Itoa(enemy.UserID)).Push("game", struct {
				GRID            string                  `json:"grid"`
				Type            string                  `json:"type"`
				Loc             *vbge.ARLocation        `json:"loc"`
				PlayerMapEntity *vbge.ViewableMapentity `json:"playermapentity"`
			}{
				enemy.GRenderID,
				"selfspawn",
				enemy.Location.ToARLocation(),
				playerMapentity,
			}, c.Log)

			return nil
		},
		// func ChangedStats
		func(p []vbge.Player) {
			var ps playersStats

			for i := range p {
				ps = append(ps, playerStats{
					GRID:   p[i].GRenderID,
					Kills:  p[i].Kills,
					Deaths: p[i].Deaths,
				})
			}

			dist.PushBroadcast("game", struct {
				Stats []playerStats
			}{
				ps,
			}, c.Log)
		})
	if err != nil {
		c.Respond(err.Error())
		return
	}

	c.RespondObj(&attackResponse{
		Health: health,
	})

	for _, entity := range ngl {
		dist.GetClient(strconv.Itoa(entity.Player.UserID)).Push("game", struct {
			GRID string           `json:"grid"`
			Type string           `json:"type"`
			Loc  *vbge.ARLocation `json:"loc"`
		}{
			c.Player.GRenderID,
			"attack",
			entity.ARLoc,
		}, c.Log)
	}
}
