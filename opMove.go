package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
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

	ng, relPos, err := c.Player.Move(dir)
	if err != nil {
		c.Respond(err.Error())
		return
	}

	// Move is successfully finished for client -> return nil
	c.RespondNil()

	// get new line for player
	newLine := vbge.GetNewLineMapentity(vbge.RenderWidth, c.Player.UserID, battle, dir)
	line, err := json.Marshal(&newLine)
	if err != nil {
		c.LogCtx.Error("unable to parse json", zap.Error(err))
		return
	}

	// create generic player response packet
	playerResp := vbge.PlayerResp{
		GRID:          c.Player.GRenderID,
		Health:        c.Player.Health.Value,
		CharacterType: c.Player.CharacterType,
		WatchDir:      c.Player.WatchDir,
	}

	// loop over all player's in the notifygroup and send an update
	for i := range ng {
		// set the relative posititon for the current opponent
		playerResp.Location = *relPos[i]

		// marshal response
		pr, err := json.Marshal(playerResp)
		if err != nil {
			c.LogCtx.Error("unable to marshal vbge.PlayerResp", zap.Error(err))
			return
		}

		updateDist.Push(ng[i],
			newUpdate("game",
				[]byte(`{"grid":"`+c.Player.GRenderID+
					`","type":"move","direction":"`+dir+`","playerinfo":`+string(pr)+
					`,"loc":{"isabs":false,"x":`+strconv.Itoa(relPos[i].X)+`,
					"y":`+strconv.Itoa(relPos[i].Y)+`},"newline":`+string(line)+`}`)),
			notifyChannelPrivate,
			nil,
			c.LogCtx)
	}
}
