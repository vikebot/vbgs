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
		GRenderID:     vbcore.FastRandomString(32),
		WatchDir:      dirNorth,
		Health:        NewDefaultHealth(),
		Rl:            NewOpLimitations(),
		CharacterType: humanThugMale,
	}

	// Search random picture
	if rand.Int()%2 == 0 {
		p.PicLink = "male/avatar" + strconv.Itoa((rand.Int()%20)-1) + ".png"
	} else {
		p.PicLink = "female/avatar" + strconv.Itoa((rand.Int()%15)-1) + ".png"
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
func (p *Player) Rotate(angle string) NotifyGroup {
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

	return p.Map.PInRenderArea(*p.Location)
}

// Move implements https://sdk-wiki.vikebot.com/#move
func (p *Player) Move(dir string) (ng NotifyGroup, err error) {
	if p.IsDefending {
		return nil, ErrCantMoveOFDefending
	}

	oldLoc := p.Location

	locc := p.Location.DeepCopy()
	locc.AddDirection(dir)
	if !locc.IsInMap() {
		return nil, ErrNoMoveOutOfMap
	}

	joined, ng := p.Map.TryJoinAreaSynced(p, *locc)
	if !joined {
		return nil, ErrHasResident
	}
	p.Location = locc

	p.Map.LeaveArea(p, *oldLoc)

	return p.Map.PInRenderArea(*p.Location), nil
}

// Radar implements https://sdk-wiki.vikebot.com/#radar
func (p *Player) Radar() (playerCount int, ng NotifyGroup, err error) {

	pCount := 0

	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	for y := vbcore.MaxInt(p.Location.Y-radarRadius, 0); y < vbcore.MinInt(p.Location.Y+radarRadius, MapHeight); y++ {
		for x := vbcore.MaxInt(p.Location.X-radarRadius, 0); x < vbcore.MinInt(p.Location.X+radarRadius, MapWidth); x++ {
			if p.Map.Matrix[y][x].HasResident() {
				pCount++
			}
		}
	}

	return pCount, p.Map.PInRenderArea(*p.Location), nil
}

// Scout implements https://sdk-wiki.vikebot.com/#scout
func (p *Player) Scout(distance int) (playerCount int, ng NotifyGroup, err error) {
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

	return pCount, p.Map.PInRenderArea(*p.Location), nil
}

// Environment implements https://sdk-wiki.vikebot.com/#environment
func (p *Player) Environment() (blocktypeMatrix [][]string, ng NotifyGroup, err error) {
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

	return matrix, p.Map.PInRenderArea(*p.Location), nil
}

// Watch implements https://sdk-wiki.vikebot.com/#watch
func (p *Player) Watch() (playerhealthMatrix [][]int, ng NotifyGroup, err error) {
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
					matrix[y][x] = p.Map.Matrix[l.Y][l.X].Resident.Health.Value
				} else {
					matrix[y][x] = 0
				}
			} else {
				matrix[y][x] = -1
			}
		}
	}

	return matrix, p.Map.PInRenderArea(*p.Location), nil
}

// Attack implements https://sdk-wiki.vikebot.com/#attack
func (p *Player) Attack(beforeRespawn func(*Player, NotifyGroup), afterRespawn func(*Player, NotifyGroup)) (enemyHealth int, ng NotifyGroup, err error) {
	p.Map.SyncRoot.Lock()
	defer p.Map.SyncRoot.Unlock()

	l := p.Location.DeepCopy()
	l.AddDirection(p.WatchDir)

	if !l.IsInMap() {
		return 0, p.Map.PInRenderArea(*p.Location), ErrOutOfMap
	}

	be := p.Map.Matrix[l.Y][l.X]
	if !be.HasResident() {
		return 0, p.Map.PInRenderArea(*p.Location), ErrNoEnemy
	}

	// Safe enemy pointer because we eventually delete him from the map
	enemy := be.Resident
	health := 0

	// Lock the enemies health sync to ensure we are the one who
	// enventually kills him
	enemy.Health.Lock()
	enemy.Health.TakeDamage(enemy)
	health = enemy.Health.Value
	if health < 1 {
		p.Kills++
		enemy.Deaths++

		beforeRespawn(enemy, enemy.Map.PInRenderArea(*enemy.Location))
		enemy.Respawn()
		afterRespawn(enemy, enemy.Map.PInRenderArea(*enemy.Location))

		health = 0
	}
	enemy.Health.Unlock()

	return health, p.Map.PInRenderArea(*p.Location), err
}

// Defend implements https://sdk-wiki.vikebot.com/#defend-and-undefend
func (p *Player) Defend() (ng NotifyGroup, err error) {
	if p.IsDefending {
		return nil, ErrAlreadyDef
	}
	p.IsDefending = true

	return p.Map.PInRenderArea(*p.Location), nil
}

// Undefend implements https://sdk-wiki.vikebot.com/#defend-and-undefend
func (p *Player) Undefend() (ng NotifyGroup, err error) {
	if !p.IsDefending {
		return nil, ErrAlreadyUndef
	}
	p.IsDefending = false

	return p.Map.PInRenderArea(*p.Location), nil
}

// GetHealth returns the health as an int value of a player
func (p *Player) GetHealth() (health int, ng NotifyGroup) {
	return p.Health.Value, p.Map.PInRenderArea(*p.Location)
}

// Spawn places the player randomly on the map as long as the location doesn't
// already have a resident. If so Spawn will retry 20 times. If no suitable
// location is found an error is returned.
func (p *Player) Spawn() error {
	for i := 0; i < 20; i++ {
		// Randomly generate a position inside the map
		loc := Location{
			X: rand.Int() % MapWidth,
			Y: rand.Int() % MapHeight,
		}

		// Check whether there already is a player or not
		var empty bool
		if empty = !p.Map.Matrix[loc.Y][loc.X].HasResident(); empty {
			// If the field is empty we place the player
			p.Map.Matrix[loc.Y][loc.X].JoinArea(p)
			p.Location = &loc
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
