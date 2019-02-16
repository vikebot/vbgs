package vbge

import (
	"errors"
	"math/rand"
	"strconv"

	"github.com/vikebot/vbcore"
)

// Player represents a single character in the game. It collects all infos
// needed.
type Player struct {
	UserID        int
	GRenderID     string
	PicLink       string
	Map           *MapEntity
	Location      *Location
	WatchDir      string
	Health        *Health
	IsDefending   bool
	Kills         int
	Deaths        int
	Rl            *OpLimitations
	CharacterType string
}

// NewPlayerWithSpawn creates a new player and spawn the player on the map
func NewPlayerWithSpawn(userID int, m *MapEntity) (p *Player, err error) {
	p = &Player{
		UserID:        userID,
		Map:           m,
		GRenderID:     strconv.Itoa(userID),
		WatchDir:      dirNorth,
		Health:        NewDefaultHealth(),
		Rl:            NewOpLimitations(),
		CharacterType: humanThugMale,
	}

	// Search random picture
	/* #nosec G404 */
	if rand.Int()%2 == 0 {
		p.PicLink = "male/avatar" + strconv.Itoa((rand.Int()%20)+1) + ".png" /* #nosec G404 */
	} else {
		p.PicLink = "female/avatar" + strconv.Itoa((rand.Int()%15)+1) + ".png" /* #nosec G404 */
	}

	// Spawn the player
	err = p.SpawnSynced()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// NewDebugPlayer returns a player that can be used for DEMO-purposes
func NewDebugPlayer(m *MapEntity) *Player {
	return &Player{
		UserID:    1,
		Map:       m,
		GRenderID: "testid",
		Location: &Location{
			X: 20,
			Y: 20,
		},
		WatchDir: dirNorth,
		Health:   NewDefaultHealth(),
		Rl:       NewOpLimitations(),
	}
}

// Rotate implements https://sdk-wiki.vikebot.com/#rotate
func (p *Player) Rotate(angle string) (ngl NotifyGroupLocated) {
	if angle == angleRight {
		switch p.WatchDir {
		case dirNorth:
			p.WatchDir = dirEast
		case dirEast:
			p.WatchDir = dirSouth
		case dirSouth:
			p.WatchDir = dirWest
		case dirWest:
			p.WatchDir = dirNorth
		}
	} else {
		switch p.WatchDir {
		case dirNorth:
			p.WatchDir = dirWest
		case dirEast:
			p.WatchDir = dirNorth
		case dirSouth:
			p.WatchDir = dirEast
		case dirWest:
			p.WatchDir = dirSouth
		}
	}

	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	// find out which players need to be informed about this action and their
	// relative positions to us
	return p.Map.PInRenderArea(p.Location)
}

// Move implements https://sdk-wiki.vikebot.com/#move
func (p *Player) Move(dir string) (ngl NotifyGroupLocated, err error) {
	if p.IsDefending {
		return nil, ErrCantMoveOFDefending
	}

	// make a real value-copy of the location, add the proposed user-direction
	// and see if it's inside the map
	locc := p.Location.DeepCopy()
	locc.AddDirection(dir)
	if !locc.IsInMap() {
		return nil, ErrNoMoveOutOfMap
	}

	if !locc.IsAccessable(p.Map) {
		return nil, ErrInaccessable
	}

	if !locc.IsAccessable(p.Map) {
		return nil, ErrInaccessable
	}

	// lock the map
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	// check whether the proposed field has a resident or not
	if p.Map.Matrix[locc.Y][locc.X].HasResident() {
		return nil, ErrHasResident
	}

	// proposed field has no resident -> leave old and join new
	p.Map.Matrix[p.Location.Y][p.Location.X].LeaveArea()
	p.Map.Matrix[locc.Y][locc.X].JoinArea(p)

	// set the new player's location ref to his new location
	oldL := p.Location
	p.Location = locc

	// find out which players need to be informed about this action and their
	// relative positions to us
	return p.Map.PInExtendedRenderArea(oldL, p.Location), nil
}

// Radar implements https://sdk-wiki.vikebot.com/#radar
func (p *Player) Radar() (playerCount int, ngl NotifyGroupLocated) {
	// calculate enclosing
	startX := vbcore.MaxInt(0, p.Location.X-radarRadius)
	endX := vbcore.MinInt(MapWidth, p.Location.X+radarRadius)
	startY := vbcore.MaxInt(0, p.Location.Y-radarRadius)
	endY := vbcore.MinInt(MapHeight, p.Location.Y+radarRadius)

	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	// calculate player count
	pCount := 0
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			if p.Map.Matrix[y][x].HasResident() {
				pCount++
			}
		}
	}

	// find out which players need to be informed about this action and their
	// relative positions to us
	return pCount, p.Map.PInRenderArea(p.Location)
}

// Scout implements https://sdk-wiki.vikebot.com/#scout
func (p *Player) Scout(distance int) (playerCount int, ngl NotifyGroupLocated) {
	pCount := 0

	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	switch p.WatchDir {
	case dirNorth:
		for i := 1; i < distance+1; i++ {
			var y = vbcore.MaxInt(p.Location.Y-i, 0)
			if y == 0 {
				break
			}
			if p.Map.Matrix[y][p.Location.X].HasResident() {
				pCount++
			}
		}
	case dirEast:
		for i := 1; i < distance+1; i++ {
			var x = vbcore.MinInt(p.Location.X+1, MapWidth-1)
			if x == MapWidth-1 {
				break
			}
			if p.Map.Matrix[p.Location.Y][vbcore.MinInt(p.Location.X+i, MapWidth-1)].HasResident() {
				pCount++
			}
		}
	case dirSouth:
		for i := 1; i < distance+1; i++ {
			var y = vbcore.MinInt(p.Location.Y+i, MapHeight-1)
			if y == MapHeight-1 {
				break
			}
			if p.Map.Matrix[y][p.Location.X].HasResident() {
				pCount++
			}
		}
	case dirWest:
		for i := 1; i < distance+1; i++ {
			var x = vbcore.MaxInt(p.Location.X-i, 0)
			if x == 0 {
				break
			}
			if p.Map.Matrix[p.Location.Y][x].HasResident() {
				pCount++
			}
		}
	}

	// find out which players need to be informed about this action and their
	// relative positions to us
	return pCount, p.Map.PInRenderArea(p.Location)
}

// Environment implements https://sdk-wiki.vikebot.com/#environment
func (p *Player) Environment() (blocktypeMatrix [][]string, ngl NotifyGroupLocated) {
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	matrix := make([][]string, RenderHeight)
	for i := range matrix {
		matrix[i] = make([]string, RenderWidth)
	}
	for y := 0; y < RenderHeight; y++ {
		for x := 0; x < RenderWidth; x++ {
			l := Location{
				Y: p.Location.Y - RenderHeight + y,
				X: p.Location.X - RenderWidth + x,
			}
			if l.IsInMap() {
				matrix[y][x] = p.Map.Matrix[l.Y][l.X].Blocktype
			} else {
				matrix[y][x] = blockEndOfMap
			}
		}
	}

	// find out which players need to be informed about this action and their
	// relative positions to us
	return matrix, p.Map.PInRenderArea(p.Location)
}

// Watch implements https://sdk-wiki.vikebot.com/#watch
func (p *Player) Watch() (playerhealthMatrix [][]int, ngl NotifyGroupLocated) {
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	matrix := make([][]int, RenderHeight)
	for i := range matrix {
		matrix[i] = make([]int, RenderWidth)
	}

	var endY, endX int
	if p.WatchDir == dirNorth || p.WatchDir == dirSouth {
		endY = HrHeight
		endX = RenderWidth
	} else {
		endY = RenderHeight
		endX = HrWidth
	}

	var loc *Location
	switch p.WatchDir {
	case dirNorth:
		loc = newLocation(p.Location.Y-HrHeight, p.Location.X-HrWidth)
	case dirEast:
		loc = newLocation(p.Location.Y-HrHeight, p.Location.X+1)
	case dirSouth:
		loc = newLocation(p.Location.Y+1, p.Location.X-HrWidth)
	case dirWest:
		loc = newLocation(p.Location.Y-HrHeight, p.Location.X-HrWidth)
	}

	for y := 0; y < endY; y++ {
		for x := 0; x < endX; x++ {
			l := loc.DeepCopy()
			l.X += x
			l.Y += y

			if l.IsInMap() {
				if p.Map.Matrix[l.Y][l.X].HasResident() {
					matrix[y][x] = p.Map.Matrix[l.Y][l.X].Resident.Health.HealthSynced()
				} else {
					matrix[y][x] = 0
				}
			} else {
				matrix[y][x] = -1
			}
		}
	}

	// find out which players need to be informed about this action and their
	// relative positions to us
	return matrix, p.Map.PInRenderArea(p.Location)
}

// PlayerHitEvent is called when a player has been hit
type PlayerHitEvent func(p *Player, healthAfterHit int, ngl NotifyGroupLocated)

// DeathEvent is called when a player has died
type DeathEvent func(p *Player, ngl NotifyGroupLocated)

// SpawnEvent is called when a player has spawned
type SpawnEvent func(p *Player, ngl NotifyGroupLocated) error

// StatsEvent is called when the stats of a player
// has changed
type StatsEvent func(p []Player)

// Attack implements https://sdk-wiki.vikebot.com/#attack
func (p *Player) Attack(onHit PlayerHitEvent, onDeath DeathEvent, onSpawn SpawnEvent, changedStats StatsEvent) (enemyHealth int, ngl NotifyGroupLocated, err error) {
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	// Get enemy location
	enemyLoc := p.Location.DeepCopy()
	enemyLoc.AddDirection(p.WatchDir)

	// Check if it's in the map
	if !enemyLoc.IsInMap() {
		return 0, nil, ErrOutOfMap
	}

	// Check if block entity even has a resident
	be := p.Map.Matrix[enemyLoc.Y][enemyLoc.X]
	if !be.HasResident() {
		return 0, nil, ErrNoEnemy
	}

	// Safe enemy pointer because we eventually delete him from the map
	enemy := be.Resident
	health := 0

	// Lock the enemies health sync to ensure we are the one who
	// enventually kills him
	enemy.Health.Lock()
	enemy.Health.TakeDamage(enemy)
	health = enemy.Health.internalValue

	// Inform all players that the enemy has been hit
	beforeRespawnNGL := enemy.Map.PInRenderArea(enemy.Location)
	onHit(enemy, health, beforeRespawnNGL)

	// Check if p killed the enemy
	if health < 1 {
		// set health to zero to avoid returning negative health values
		health = 0

		// increase kill and death counters for p and enemy respectively
		p.Kills++
		enemy.Deaths++
		enemy.Health.Unlock()

		// Notify that the stats has changed. Copy the players so the KD is
		// save for the stats. TODO: find out if neccessary
		var players = []Player{*p, *enemy}
		go func() {
			changedStats(players)
		}()

		// Notify all players that the enemy died
		onDeath(enemy, beforeRespawnNGL)

		// Try to respawn the enemy
		err = enemy.Respawn()
		if err != nil {
			return 0, nil, err
		}

		// Inform the people around the enemies new location, that he has just
		// spawned.
		afterRespawnNG := enemy.Map.PInRenderArea(enemy.Location)
		err = onSpawn(enemy, afterRespawnNG)
		if err != nil {
			return 0, nil, err
		}
	} else {
		enemy.Health.Unlock()
	}

	// find out which players need to be informed about this action and their
	// relative positions to us
	return health, p.Map.PInRenderArea(p.Location), nil
}

// Defend implements https://sdk-wiki.vikebot.com/#defend-and-undefend
func (p *Player) Defend() (ngl NotifyGroupLocated, err error) {
	if p.IsDefending {
		return nil, ErrAlreadyDef
	}
	p.IsDefending = true

	// find out which players need to be informed about this action and their
	// relative positions to us
	return p.Map.PInRenderArea(p.Location), nil
}

// Undefend implements https://sdk-wiki.vikebot.com/#defend-and-undefend
func (p *Player) Undefend() (ngl NotifyGroupLocated, err error) {
	if !p.IsDefending {
		return nil, ErrAlreadyUndef
	}
	p.IsDefending = false

	// find out which players need to be informed about this action and their
	// relative positions to us
	return p.Map.PInRenderArea(p.Location), nil
}

// Spawn places the player randomly on the map as long as the location doesn't
// already have a resident. If so Spawn will retry 100 times. If no suitable
// location is found an error is returned.
func (p *Player) Spawn() error {
	for i := 0; i < 100; i++ {
		// Randomly generate a position inside the map
		/* #nosec G404 */
		loc := Location{
			X: rand.Int() % MapWidth,
			Y: rand.Int() % MapHeight,
		}

		// Check whether there already is a player or not
		empty := !p.Map.Matrix[loc.Y][loc.X].HasResident()
		isWater := p.Map.Matrix[loc.Y][loc.X].Blocktype == blockWater

		if empty && !isWater {
			// If the field is empty and not water we place the player
			p.Map.Matrix[loc.Y][loc.X].JoinArea(p)
			p.Location = &loc
			p.Health = NewDefaultHealth()
			p.WatchDir = dirNorth
			p.IsDefending = false
			return nil
		}
	}

	return errors.New("vbge: unable to find a suitable location to place the player during spawn")
}

// SpawnSynced is like `Spawn` but locks the Map
func (p *Player) SpawnSynced() error {
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	return p.Spawn()
}

// Respawn removes the player from it's current position and add calls `Spawn`
// to place it again.
func (p *Player) Respawn() error {
	// Remove the player from the block and delete the pointer to it'
	// location
	p.Map.Matrix[p.Location.Y][p.Location.X].LeaveArea()
	p.Location = nil

	// Spwan the player again
	err := p.Spawn()
	if err != nil {
		return err
	}

	return nil
}

// RespawnSynced is like `Respawn` but locks the Map
func (p *Player) RespawnSynced() error {
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	return p.Respawn()
}
