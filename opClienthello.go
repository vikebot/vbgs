package main

import (
	"encoding/base64"
	"math/rand"
	"strings"

	"go.uber.org/zap"
)

type xhelloObj struct {
	Cipher *string `json:"cipher"`
}
type xhelloPacket struct {
	Type string    `json:"type"`
	Obj  xhelloObj `json:"obj"`
}

func opClienthello(c *ntcpclient, packet xhelloPacket) {
	if packet.Obj.Cipher == nil {
		c.Respond("Invalid packet. '.obj.cipher' missing")
		return
	}

	var plain string
	if envDisableCrypt {
		plain = *packet.Obj.Cipher
	} else {
		buf, err := base64.RawStdEncoding.DecodeString(*packet.Obj.Cipher)
		if err != nil {
			c.Respond("Invalid packet. '.obj.cipher' must be a base64 string")
			return
		}

		plainBuf, err := c.Crypt.Decrypt(buf)
		if err != nil {
			c.log.Warn("failed to decrypt", zap.Error(err))
			c.Respond("Invalid cipher text - unable to decrypt")
			return
		}

		plain = string(plainBuf)
	}

	if !strings.HasPrefix(plain, "clienthello:") || strings.Count(plain, ":") != 1 {
		c.Respond("Invalid plain text - expecting 'clienthello:YOURCHALLENGE'")
		return
	} else if plain == "clienthello:YOURCHALLENGE" {
		c.Respond("Invalid server challenge in '.obj.cipher' - 'YOURCHALLENGE' is not allowed")
		return
	}

	challenge := strings.Split(plain, ":")[1]
	if len(challenge) > 32 {
		c.Respond("Invalid server challenge in '.obj.cipher' - Maximum of 32 characters")
		return
	}

	// Define response builder for serverhello packet
	cipher := "serverhello:" + challenge
	if !envDisableCrypt {
		cipherBuf, err := c.Crypt.Encrypt([]byte(cipher))
		if err != nil {
			c.log.Error("failed to encrypt serverhello challenge response", zap.Error(err))
			c.Respond(statusInternalServerError)
			return
		}

		cipher = base64.RawStdEncoding.EncodeToString(cipherBuf)
	}
	c.MgmtWrite(xhelloPacket{
		Type: "serverhello",
		Obj: xhelloObj{
			Cipher: &cipher,
		},
	})
	c.ClienthelloDone = true

	// Connection verified -> enable complete encryption and send initial pc
	c.log = c.log.With(zap.Int("user_id", c.UserID))
	c.log.Info("authenticated")

	c.IsEncrypted = true
	c.Pc = rand.Uint32() / 2
	c.StartPc = c.Pc
	c.CurType = "initialpc"
	c.RespondNil()
}
