package main

import (
	"errors"

	"github.com/vikebot/vbdb"
)

type playerStats struct {
	GRID     string `json:"grid"`
	Username string `json:"username"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
}

type playersStats []playerStats

// getPlayersStats returns the type playersStats which
// is a slice of playerStats, it's used for getting
// information of all players in the game
func getPlayersStats() (ps playersStats, err error) {
	usernames, success := vbdb.UsernamesFromRoundID(config.Battle.RoundID)
	if !success {
		return ps, errors.New("unable to load usernames from db")
	}

	for _, p := range battle.Players {
		ps = append(ps, playerStats{
			GRID:     p.GRenderID,
			Username: usernames[p.UserID],
			Kills:    p.Kills,
			Deaths:   p.Deaths,
		})
	}

	return
}
