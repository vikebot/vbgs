package main

import (
	"encoding/json"
	"strconv"

	"go.uber.org/zap"
)

const (
	statusInternalServerError = "Internal Server Errror"
	statusInvalidJSON         = "Invalid JSON format"
)

func dispatch(c *ntcpclient, data []byte, packet typePacket) {
	var err error

	// Switch between packet types and dispatch them between the different op funcs
	switch *packet.Type {
	case "login":
		var login loginPacket
		err = json.Unmarshal(data, &login)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opLogin(c, login)
		return
	case "clienthello":
		var clienthello xhelloPacket
		err = json.Unmarshal(data, &clienthello)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opClienthello(c, clienthello)
		return
	case "agreeconn":
		// Check if this client has already a agreed connection
		if err = ntcpRegistry.Put(c); err != nil {
			log.Warn("multiple connections for same user", zap.Error(err))
			c.Respond("Connection already open - Please close any previous connections before initializing a new one.")
			return
		}

		c.AgreeconnDone = true
		c.Authenticated = true
		c.Player = battle.Players[c.UserID]

		c.RespondNil()
		dist.GetClient(strconv.Itoa(c.UserID)).PushInfo(true, c.IP, c.SDK, c.SDKLink, c.OS, c.Log)
		return
	case "rotate":
		var rotate rotatePacket
		err = json.Unmarshal(data, &rotate)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opRotate(c, rotate)
		return
	case "move":
		var move movePacket
		err = json.Unmarshal(data, &move)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opMove(c, move)
		return
	case "radar":
		var radar radarPacket
		err = json.Unmarshal(data, &radar)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opRadar(c, radar)
		return
	case "scout":
		var scout scoutPacket
		err = json.Unmarshal(data, &scout)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opScout(c, scout)
		return
	case "environment":
		var environment environmentPacket
		err = json.Unmarshal(data, &environment)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
	case "watch":
		var watch watchPacket
		err = json.Unmarshal(data, &watch)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opWatch(c, watch)
		return
	case "attack":
		var attack attackPacket
		err = json.Unmarshal(data, &attack)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opAttack(c, attack)
		return
	case "defend":
		var defend defendPacket
		err = json.Unmarshal(data, &defend)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opDefend(c, defend)
		return
	case "undefend":
		var undefend undefendPacket
		err = json.Unmarshal(data, &undefend)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opUndefend(c, undefend)
		return
	case "health":
		var health healthPacket
		err = json.Unmarshal(data, &health)
		if err != nil {
			c.Respond(statusInvalidJSON)
			return
		}
		opHealth(c, health)
		return
	default:
		c.CurType = "forbidden"
		c.Respond("Invalid packet. '.type' unknown")
		return
	}
}
