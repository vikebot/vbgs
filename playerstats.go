package main

import (
	"errors"

	"github.com/vikebot/vbdb"
)

type playerStats struct {
	Username string `json:"username"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
}

type playersStats []playerStats

// getPlayersStats returns the type playersStats which
// is a slice of playerStats, it's used for getting
// information of all players in the game
func getPlayersStats() (ps playersStats, err error) {
	for _, p := range battle.Players {
		user, success := vbdb.UserFromID(p.UserID)
		if !success {
			return ps, errors.New("unable to load User from db")
		}

		ps = append(ps, playerStats{
			Username: user.Username,
			Kills:    p.Kills,
			Deaths:   p.Deaths,
		})
	}

	return
}
