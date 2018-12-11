package main

import (
	"encoding/base64"

	"github.com/vikebot/vbdb"
	"go.uber.org/zap"
)

type loginObj struct {
	RoundTicket *string `json:"roundticket"`
}
type loginPacket struct {
	Type string   `json:"type"`
	Obj  loginObj `json:"obj"`
}

func opLogin(c *ntcpclient, packet loginPacket) {
	if packet.Obj.RoundTicket == nil {
		c.Respond("Invalid packet. '.obj.roundticket' missing")
		return
	}

	v, exists, success := vbdb.RoundentryFromRoundticketCtx(*packet.Obj.RoundTicket, c.log)
	if !success {
		c.Respond(statusInternalServerError)
		return
	}
	if !exists {
		c.Respond("Your rounticket doesn't reference any game.")
		return
	}

	if config.Battle.RoundID != v.RoundID {
		c.Respond("Your roundticket references an already finished game.")
		c.log.Warn("valid watchtoken references invalid round",
			zap.Int("config_round_id", config.Battle.RoundID),
			zap.Int("watchtoken_round_id", v.RoundID))
		return
	}

	c.UserID = v.UserID

	keybuf, err := base64.StdEncoding.DecodeString(*v.AESKey)
	if err != nil {
		c.log.Error("failed to decode base64 string", zap.String("aeskey", *v.AESKey))
		c.Respond(statusInternalServerError)
		return
	}
	err = c.InitAes(keybuf)
	if err != nil {
		c.log.Error("failed to init AES from key buffer", zap.Error(err))
		c.Respond(statusInternalServerError)
		return
	}

	c.LoginDone = true
	c.RespondNil()
}
