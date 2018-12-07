package vbge

type playerStats struct {
	Username string `json:"username"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
}

type playersStats []playerStats
