package main

import (
	"encoding/json"
	"strconv"
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

	ng, err := c.Player.Move(dir)
	if err != nil {
		c.Respond(err.Error())
		return
	}

	newLine, err := vbge.GetNewLineMapentity(vbge.RenderWidth, c.Player.UserID, battle, dir)
	if err != nil {
		c.Respond(err.Error())
		return
	}

	line, err := json.Marshal(&newLine)
	if err != nil {
		c.Respond(err.Error())
		return
	}
	c.RespondNil()

	playerResp := vbge.PlayerResp{
		GRID:          c.Player.GRenderID,
		Health:        c.Player.Health.Value,
		CharacterType: c.Player.CharacterType,
		WatchDir:      c.Player.WatchDir,
	}

	for i := 0; i < len(ng); i++ {
		l := vbge.Location{
			X: ng[i].Location.X - c.Player.Location.X,
			Y: ng[i].Location.Y - c.Player.Location.Y,
		}

		playerResp.Location = *c.Player.Location.RelativeFrom(ng[i].Location)

		pr, err := json.Marshal(playerResp)
		if err != nil {
			c.Respond(err.Error())
			return
		}

		updateDist.Push(ng[i], newUpdate("game", []byte(`{"grid":"`+c.Player.GRenderID+`","type":"move","direction":"`+dir+`","playerinfo":`+string(pr)+`,"loc":{"isabs":false,"x":`+strconv.Itoa(l.X)+`,"y":`+strconv.Itoa(l.Y)+`},"newline":`+string(line)+`}`)), notifyChannelPrivate, nil, c.LogCtx)
	}
}
