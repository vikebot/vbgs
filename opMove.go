package main

import (
	"time"

	"github.com/vikebot/vbgs/vbge"
)

type moveObj struct {
	Direction *string `json:"direction"`
}
type movePacket struct {
	Type string  `json:"type"`
	Obj  moveObj `json:"obj"`
}

func opMove(c *ntcpclient, packet movePacket) {
	//c.Player.Rl.Move.Take()
	time.Sleep(1000 * time.Millisecond)
	if packet.Obj.Direction == nil {
		c.Respond("Invalid packet. '.obj.direction' missing")
		return
	}

	dir := *packet.Obj.Direction
	if !vbge.IsDir(dir) {
		c.RespondFmt("Invalid packet. '%s' is not a valid value for '.obj.direction'", *packet.Obj.Direction)
		return
	}

	ngl, err := c.Player.Move(dir)
	if err != nil {
		c.Respond(err.Error())
		return
	}

	// Move is successfully finished for client -> return nil
	c.RespondNil()

	// get new line for player
	newLine := vbge.GetNewLineMapentity(vbge.RenderWidth, c.Player.UserID, battle, dir)

	// create generic player response packet
	playerResp := vbge.PlayerResp{
		GRID:          c.Player.GRenderID,
		Health:        c.Player.Health.HealthSynced(),
		CharacterType: c.Player.CharacterType,
		WatchDir:      c.Player.WatchDir,
	}

	// loop over all player's in the notifygroup and send an update
	for _, entity := range ngl {
		// set the relative posititon for the current opponent
		playerResp.Location = entity.ARLoc

		dist.GetClient(entity.Player.UserID).Push("game",
			struct {
				GRID       string                  `json:"grid"`
				Type       string                  `json:"type"`
				Direction  string                  `json:"direction"`
				PlayerInfo vbge.PlayerResp         `json:"playerinfo"`
				Loc        *vbge.ARLocation        `json:"loc"`
				NewLine    *vbge.ViewableMapentity `json:"newline"`
			}{
				c.Player.GRenderID,
				"move",
				dir,
				playerResp,
				entity.Player.Location.ToARLocation(),
				newLine},
			c.Log)
	}
}
