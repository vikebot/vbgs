package main

type playerStats struct {
	Username string `json:"username"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
}

type playersStats []playerStats

// getPlayersStats returns the type playersStats which
// is a slice of playerStats, it's used for getting
// information of all players in the game
func getPlayersStats() (ps playerStats, err error) {
	return
}
