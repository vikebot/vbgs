package main

import (
	"encoding/json"

	"go.uber.org/zap"
)

func packetHandler(c *ntcpclient, data []byte) {
	c.CurType = "forbidden"

	// Decrypt packet
	if c.IsEncrypted && !envDisableCrypt {
		plainBuf, err := c.Crypt.DecryptBase64(data)
		if err != nil {
			c.LogCtx.Warn("failed to decrypt cipher", zap.Error(err))
			c.Respond("Invalid cipher text - unable to decrypt")
			return
		}
		data = plainBuf
	}

	// Log the incoming packet as debug message
	c.LogCtx.Debug("received",
		zap.String("packet", string(data)),
		zap.Uint32("seqnr", c.Pc-c.StartPc))

	// Check for basic packet structure
	var packet typePacket
	err := json.Unmarshal(data, &packet)
	if err != nil {
		c.Respond("Invalid JSON syntax")
		return
	}
	if packet.Type == nil {
		c.Respond("Invalid packet. '.type' missing")
		return
	}

	// Check for correct packet count
	if c.IsEncrypted {
		if packet.Pc == nil {
			c.Respond("Invalid packet. '.pc' missing")
			return
		}
		c.Pc++
		if *packet.Pc != c.Pc {
			c.Respond("Protocol mismatch. '.pc' value not increased")
			return
		}
	}

	// Set current packet type
	c.CurType = *packet.Type

	// Check if the login process isn't finished but the user tries to send another packet
	if (!c.LoginDone && *packet.Type != "login") &&
		(!c.ClienthelloDone && *packet.Type != "clienthello") &&
		(!c.AgreeconnDone && *packet.Type != "agreeconn") {
		c.Respond("You aren't allowed to send any packet type previous to a successful login")
		return
	}
	// Check if client has previously sent packets that are only allowed once
	if (c.LoginDone && *packet.Type == "login") ||
		(c.ClienthelloDone && *packet.Type == "clienthello") ||
		(c.AgreeconnDone && *packet.Type == "agreeconn") {
		c.CurType = *packet.Type
		c.RespondFmt("Protocol mismatch. '%s' already done", *packet.Type)
		return
	}

	// Dispatch the current event
	dispatch(c, data, packet)
}
