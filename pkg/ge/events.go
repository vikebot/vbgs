package ge

// AttackHitEvent is called when a player has been hit
type AttackHitEvent func(p *Player, healthAfterHit int, ng NotifyGroupLocation)

// AttackDeathEvent is called when a player has died
type AttackDeathEvent func(p *Player, ng NotifyGroupLocation)

// AttackSpawnEvent is called when a player has spawned
type AttackSpawnEvent func(p *Player, ng NotifyGroupLocation)

// StatsEvent is called when the stats of a player
// has changed
type StatsEvent func(p []Player, ng NotifyGroupLocation)

type AttackEvents struct {
	Hit   AttackHitEvent
	Death AttackDeathEvent
	Spawn AttackSpawnEvent
}
