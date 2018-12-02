package vbge

import (
	"math"
	"strconv"
	"testing"

	"github.com/vikebot/vbcore"
)

func TestPlayerRotate(t *testing.T) {
	cases := []struct {
		name  string
		from  string
		angle string
		want  string
	}{
		{"North + right = East", dirNorth, angleRight, dirEast},
		{"East + right = South", dirEast, angleRight, dirSouth},
		{"South + right = West", dirSouth, angleRight, dirWest},
		{"West + right = North", dirWest, angleRight, dirNorth},
		{"North + left = West", dirNorth, angleLeft, dirWest},
		{"East + left = North", dirEast, angleLeft, dirNorth},
		{"South + left = East", dirSouth, angleLeft, dirEast},
		{"West + left = South", dirWest, angleLeft, dirSouth},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				Map: NewMapEntity(MapWidth, MapHeight),
				Location: &Location{
					X: HrWidth,
					Y: HrHeight,
				},
				WatchDir: c.from,
				Rl:       NewOpLimitations(),
			}

			p.Rotate(c.angle)

			if p.WatchDir != c.want {
				t.Errorf("Rotate() => WatchDir = %v, want %v", p.WatchDir, c.want)
			}
		})
	}
}

type playerMoveTest struct {
	FromDir string
	ToDir   string
	FromX   int
	FromY   int
}

func newPlayerMoveTest(fromDir, toDir string, y, x int) playerMoveTest {
	return playerMoveTest{
		FromDir: fromDir,
		ToDir:   toDir,
		FromY:   y,
		FromX:   x,
	}
}

func TestPlayerMove(t *testing.T) {
	cases := []struct {
		name           string
		playerMove     playerMoveTest
		wantedLocation *Location
	}{
		{"Test01: Basic Go To North", newPlayerMoveTest(dirNorth, dirNorth, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight-1, HalfmapWidth)},
		{"Test02: Basic Go To East", newPlayerMoveTest(dirEast, dirEast, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight, HalfmapWidth+1)},
		{"Test03: Basic Go To South", newPlayerMoveTest(dirSouth, dirSouth, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight+1, HalfmapWidth)},
		{"Test04: Basic Go To West", newPlayerMoveTest(dirWest, dirWest, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight, HalfmapWidth-1)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				Map:      NewMapEntity(100, 100),
				WatchDir: c.playerMove.FromDir,
				Location: &Location{
					X: c.playerMove.FromX,
					Y: c.playerMove.FromY,
				},
				Rl: NewOpLimitations(),
			}

			p.Move(c.playerMove.ToDir)
			if p.Location.X == c.wantedLocation.X && p.Location.Y != c.wantedLocation.Y {
				t.Fail()
				t.Errorf(c.name + " wanted: X: " + strconv.Itoa(c.wantedLocation.X) + " Y: " + strconv.Itoa(c.wantedLocation.Y) + " got: X:" + strconv.Itoa(p.Location.X) + " Y: " + strconv.Itoa(p.Location.Y))
			}
		})
	}
}

func TestPlayerMoveOutOfMap(t *testing.T) {
	cases := []struct {
		name       string
		playerMove playerMoveTest
	}{
		{"Test01: Basic Go To North at top left corner", newPlayerMoveTest(dirNorth, dirNorth, 0, 0)},
		{"Test02: Basic Go To West at top lefft corner", newPlayerMoveTest(dirWest, dirWest, 0, 0)},
		{"Test03: Basic Go To North at top right corner", newPlayerMoveTest(dirNorth, dirNorth, 0, MapWidth)},
		{"Test04: Basic Go To East at top right corner", newPlayerMoveTest(dirEast, dirEast, 0, MapWidth)},
		{"Test05: Basic Go To South at bottom left corner", newPlayerMoveTest(dirSouth, dirSouth, MapHeight, 0)},
		{"Test06: Basic Go To West at bottom left corner", newPlayerMoveTest(dirWest, dirWest, MapHeight, 0)},
		{"Test07: Basic Go To South at bottom right corner", newPlayerMoveTest(dirSouth, dirSouth, MapHeight, MapWidth)},
		{"Test08: Basic Go To East at bottom right corner", newPlayerMoveTest(dirEast, dirEast, MapHeight, MapWidth)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				Map:      NewMapEntity(MapWidth, MapHeight),
				WatchDir: c.playerMove.FromDir,
				Location: &Location{
					X: c.playerMove.FromX,
					Y: c.playerMove.FromY,
				},
				Rl: NewOpLimitations(),
			}

			_, _, err := p.Move(c.playerMove.ToDir)
			if err == nil {
				t.Errorf(c.name)
			}
		})

	}
}

func TestPlayerMoveWatchingWrongDirection(t *testing.T) {
	cases := []struct {
		name           string
		playerMove     playerMoveTest
		wantedLocation *Location
	}{
		{"Test01: Go To East,wrong direction", newPlayerMoveTest(dirNorth, dirEast, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight, HalfmapWidth+1)},
		{"Test02: Go To South, wrong direction", newPlayerMoveTest(dirNorth, dirSouth, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight+1, HalfmapWidth)},
		{"Test03: Go To West, wrong direction", newPlayerMoveTest(dirNorth, dirWest, HalfmapHeight, HalfmapWidth), newLocation(HalfmapHeight, HalfmapWidth-1)},
		{"Test04: Go To South, wrong direction", newPlayerMoveTest(dirNorth, dirSouth, 0, 0), newLocation(1, 0)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				Map:      NewMapEntity(MapWidth, MapHeight),
				WatchDir: c.playerMove.FromDir,
				Location: &Location{
					X: c.playerMove.FromX,
					Y: c.playerMove.FromY,
				},
				Rl: NewOpLimitations(),
			}

			p.Move(c.playerMove.ToDir)

			if p.Location.X != c.wantedLocation.X || p.Location.Y != c.wantedLocation.Y {
				t.Fail()
				t.Errorf(c.name)
			}
		})
	}
}

func TestPlayerRadar(t *testing.T) {
	cases := []struct {
		name                 string
		wantedPlayerCount    int
		mapEntityWithPlayers *MapEntity
		playerLocation       *Location
	}{
		{"Test01: One player in radarRadius", 1, newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), ""), newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test02: More than one player in radarRadius", 5, newMapEntityWithPlayers(5, newLocation(HalfmapHeight, HalfmapWidth), ""), newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test03: Playerlocation on edge of map", 2, newMapEntityWithPlayers(2, newLocation(0, 0), ""), newLocation(0, 0)},
		{"Test04: Playerlocation on edge of map", 3, newMapEntityWithPlayers(5, newLocation(0, 0), ""), newLocation(0, 0)},
		{"Test05: Playerlocation on edge of map", 1, newMapEntityWithPlayers(1, newLocation(MapHeight-1, MapWidth-1), ""), newLocation(MapHeight-1, MapWidth-1)},
		{"Test06: Playerlocation on edge of map", 4, newMapEntityWithPlayers(5, newLocation(MapHeight-1, MapWidth-1), ""), newLocation(MapHeight-1, MapWidth-1)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				UserID:   1,
				Map:      c.mapEntityWithPlayers,
				Location: c.playerLocation,
				Rl:       NewOpLimitations(),
			}

			playerCount, _, _ := p.Radar()

			if playerCount != c.wantedPlayerCount {
				t.Fail()
				t.Error(c.name + " playerCount: " + strconv.Itoa(playerCount) + " wantedPlayerCount: " + strconv.Itoa(c.wantedPlayerCount))
			}
		})
	}
}

func TestPlayerScout(t *testing.T) {
	cases := []struct {
		name                 string
		wantedPlayerCount    int
		distance             int
		dir                  string
		mapEntityWithPlayers *MapEntity
		playerLocation       *Location
	}{
		{"Test01 one player", 1, 1, dirNorth, newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirNorth), newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test02: more players", 4, 4, dirNorth, newMapEntityWithPlayers(4, newLocation(HalfmapHeight, HalfmapWidth), dirNorth), newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test03: other direction", 1, 1, dirEast, newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirEast), newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test04: whole map", 1, MapHeight, dirSouth, newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirSouth), newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test05: player on edge", 5, 5, dirEast, newMapEntityWithPlayers(5, newLocation(0, 0), dirEast), newLocation(0, 0)},
		{"Test06: dir west", 4, 10, dirWest, newMapEntityWithPlayers(4, newLocation(HalfmapHeight, HalfmapWidth), dirWest), newLocation(HalfmapHeight, HalfmapWidth)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				Map:      c.mapEntityWithPlayers,
				Location: c.playerLocation,
				WatchDir: c.dir,
				Rl:       NewOpLimitations(),
			}

			playerCount, _, _ := p.Scout(c.distance)

			if playerCount != c.wantedPlayerCount {
				t.Fail()
				t.Error(c.name + " returned playerCount: " + strconv.Itoa(playerCount) + " wanted playerCount: " + strconv.Itoa(c.wantedPlayerCount))
			}
		})
	}
}

func TestPlayerEnvironment(t *testing.T) {
	cases := []struct {
		name     string
		location *Location
	}{
		{"Test01: items of slice", newLocation(HalfmapHeight, HalfmapWidth)},
		{"Test02: player on edge", newLocation(0, 0)},
		{"Test03: player on edge", newLocation(HalfmapHeight, HalfmapWidth)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			p := Player{
				UserID:   1,
				Map:      NewMapEntity(MapWidth, MapHeight),
				Location: c.location,
				Rl:       NewOpLimitations(),
			}

			matrix, _, _ := p.Environment()

			for y := 0; y < HrHeight; y++ {
				for x := 0; x < HrWidth; x++ {
					if matrix[y][x] == "" {
						t.Fail()
						t.Error(c.name + " BlockType: " + matrix[y][x])
					}
				}
			}
		})
	}
}

func TestPlayerWatch(t *testing.T) {
	cases := []struct {
		name              string
		watchDir          string
		location          *Location
		mapEntity         *MapEntity
		wantedHealthCount int
	}{
		{"Test01: watch North", dirNorth, newLocation(HalfmapHeight, HalfmapWidth), newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirNorth), 1},
		{"Test02: watch East", dirEast, newLocation(HalfmapHeight, HalfmapWidth), newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirEast), 1},
		{"Test03: watch South", dirSouth, newLocation(HalfmapHeight, HalfmapWidth), newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirSouth), 1},
		{"Test04: watch West", dirWest, newLocation(HalfmapHeight, HalfmapWidth), newMapEntityWithPlayers(1, newLocation(HalfmapHeight, HalfmapWidth), dirWest), 1},
		{"Test05: more players", dirNorth, newLocation(HalfmapHeight, HalfmapWidth), newMapEntityWithPlayers(10, newLocation(HalfmapHeight, HalfmapWidth), dirNorth), HrHeight},
		{"Test06: out of map north", dirNorth, newLocation(0, 0), newMapEntityWithPlayers(10, newLocation(0, 0), dirNorth), 0},
		{"Test06: out of map east", dirEast, newLocation(0, MapWidth), newMapEntityWithPlayers(10, newLocation(0, MapWidth), dirEast), 0},
		{"Test06: out of map south", dirSouth, newLocation(MapHeight, 0), newMapEntityWithPlayers(10, newLocation(MapHeight, 0), dirSouth), 0},
		{"Test06: out of map west", dirWest, newLocation(0, 0), newMapEntityWithPlayers(10, newLocation(0, 0), dirWest), 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				Map:      c.mapEntity,
				Location: c.location,
				WatchDir: c.watchDir,
				Rl:       NewOpLimitations(),
				Health:   NewDefaultHealth(),
			}

			matrix, _, _ := p.Watch()

			healthCount := 0

			for _, line := range matrix {
				for _, element := range line {
					if element == 100 {
						healthCount++
					}
				}
			}

			if healthCount != c.wantedHealthCount {
				t.Fail()
				t.Errorf(c.name + " wantedHealthCount: " + strconv.Itoa(c.wantedHealthCount) + " returnedHealthCount: " + strconv.Itoa(healthCount))
			}
		})
	}
}

func TestPlayerAttack(t *testing.T) {
	playerLoc := newLocation(HalfmapHeight, HalfmapWidth)
	cases := []struct {
		name         string
		playerLoc    *Location
		watchDir     string
		enemy        *Player
		attackCount  int
		wantedHealth int
	}{
		{"Single attack", playerLoc, dirNorth, newPlayer(100, 0, 0, HalfmapHeight-1, HalfmapWidth, false), 1, 100 - defaultDmg*1},
		{"Attack 2 times", playerLoc, dirNorth, newPlayer(100, 0, 0, HalfmapHeight-1, HalfmapWidth, false), 2, 100 - defaultDmg*2},
		{"Attack 10 times", playerLoc, dirNorth, newPlayer(100, 0, 0, HalfmapHeight-1, HalfmapWidth, false), 10, 100 - defaultDmg*10},
		{"Attack 15 times", playerLoc, dirNorth, newPlayer(150, 0, 0, HalfmapHeight-1, HalfmapWidth, false), 15, 150 - defaultDmg*15},

		{"Above Zero", playerLoc, dirNorth, newPlayer(5, 0, 0, HalfmapHeight-1, HalfmapWidth, false), 1, 0},

		{"Attack dirNorth", playerLoc, dirNorth, newPlayer(100, 0, 0, HalfmapHeight-1, HalfmapWidth, false), 1, 100 - defaultDmg*1},
		{"Attack dirEast", playerLoc, dirEast, newPlayer(100, 0, 0, HalfmapHeight, HalfmapWidth+1, false), 1, 100 - defaultDmg*1},
		{"Attack dirSouth", playerLoc, dirSouth, newPlayer(100, 0, 0, HalfmapHeight+1, HalfmapWidth, false), 1, 100 - defaultDmg*1},
		{"Attack dirWest", playerLoc, dirWest, newPlayer(100, 0, 0, HalfmapHeight, HalfmapWidth-1, false), 1, 100 - defaultDmg*1},

		{"Defend 1 Attack", playerLoc, dirNorth, newPlayer(100, 0, 0, HalfmapHeight-1, HalfmapWidth, true), 1, 100 - defaultDmg/2*1},
		{"Defend 2 Attacks ", playerLoc, dirNorth, newPlayer(100, 0, 0, HalfmapHeight-1, HalfmapWidth, true), 2, 100 - defaultDmg/2*2},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mapEntity := NewMapEntity(MapWidth, MapHeight)
			c.enemy.Map = mapEntity
			mapEntity.Matrix[c.enemy.Location.Y][c.enemy.Location.X].JoinArea(c.enemy)

			p := Player{
				Map:      mapEntity,
				Location: c.playerLoc,
				WatchDir: c.watchDir,
				Rl:       NewOpLimitations(),
			}

			// Prepare wanted health
			c.wantedHealth = vbcore.MaxInt(c.wantedHealth, 0)

			// Calculate returned health
			returnedHealth := 0
			for i := 0; i < c.attackCount; i++ {
				var err error
				returnedHealth, _, err = p.Attack(func(e *Player, ng NotifyGroup) {}, func(e *Player, ng NotifyGroup) {})
				if err != nil {
					t.Errorf("Attack() err = %v, wantedErr = false", err)
					return
				}
			}

			if returnedHealth != c.wantedHealth {
				t.Errorf("Attack() = %v, want %v", returnedHealth, c.wantedHealth)
				return
			}
		})
	}
}

func TestPlayerAttackError(t *testing.T) {
	cases := []struct {
		name        string
		location    *Location
		watchDir    string
		wantedError error
	}{
		{"Out of map (North)", newLocation(0, 0), dirNorth, ErrOutOfMap},
		{"Out of map (East)", newLocation(0, MapWidth-1), dirEast, ErrOutOfMap},
		{"Out of map (South)", newLocation(MapHeight-1, MapWidth-1), dirSouth, ErrOutOfMap},
		{"Out of map (West)", newLocation(MapHeight-1, 0), dirWest, ErrOutOfMap},

		{"No enemy (North)", newLocation(HalfmapHeight, HalfmapWidth), dirNorth, ErrNoEnemy},
		{"No enemy (East)", newLocation(HalfmapHeight, HalfmapWidth), dirEast, ErrNoEnemy},
		{"No enemy (South)", newLocation(HalfmapHeight, HalfmapWidth), dirSouth, ErrNoEnemy},
		{"No enemy (West)", newLocation(HalfmapHeight, HalfmapWidth), dirWest, ErrNoEnemy},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				UserID:   1,
				Map:      NewMapEntity(MapWidth, MapHeight),
				Location: c.location,
				WatchDir: c.watchDir,
				Rl:       NewOpLimitations(),
			}

			_, _, err := p.Attack(func(e *Player, ng NotifyGroup) {}, func(e *Player, ng NotifyGroup) {})
			if err == nil || err != c.wantedError {
				t.Errorf("Attack() = %v, want %v", err.Error(), c.wantedError.Error())
			}
		})
	}
}

func TestPlayerDefend(t *testing.T) {
	cases := []struct {
		name         string
		isDefending  bool
		wantedResult bool
		wantedError  error
	}{
		{"Test01: Defend", false, true, nil},
		{"Test02: Already defending", true, true, ErrAlreadyDef},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				IsDefending: c.isDefending,
				Location: &Location{
					X: HalfmapWidth,
					Y: HalfmapHeight,
				},
				Map: NewMapEntity(MapWidth, MapHeight),
				Rl:  NewOpLimitations(),
			}

			_, err := p.Defend()

			if err != c.wantedError {
				t.Errorf("Defend() = %v, want %v", err, c.wantedError)
			}

			if c.wantedResult != p.IsDefending {
				t.Errorf("p.IsDefennding = %v, want %v", p.IsDefending, c.wantedResult)
			}
		})
	}
}

func TestPlayerUndefend(t *testing.T) {
	cases := []struct {
		name         string
		isDefending  bool
		wantedResult bool
		wantedError  error
	}{
		{"Test01: Undefend", true, false, nil},
		{"Test02: Already undefending", false, false, ErrAlreadyUndef},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Player{
				IsDefending: c.isDefending,
				Location: &Location{
					X: HalfmapWidth,
					Y: HalfmapHeight,
				},
				Map: NewMapEntity(MapWidth, MapHeight),
				Rl:  NewOpLimitations(),
			}

			_, err := p.Undefend()

			if err != c.wantedError {
				t.Errorf("Undefend() = %v, want %v", err, c.wantedError)
			}

			if c.wantedResult != p.IsDefending {
				t.Errorf("p.IsDefennding = %v, want %v", p.IsDefending, c.wantedResult)
			}
		})
	}
}

func newMapEntityWithPlayers(playerCount int, location *Location, dir string) *MapEntity {
	mapEntity := NewMapEntity(30, 30)

	if dir == "" {
		for i := 1; i < playerCount+1; i++ {
			if i%2 == 0 {
				p1 := Player{
					Location: &Location{
						X: int(math.Min(float64(location.X+i), float64(MapWidth-1))),
						Y: int(math.Min(float64(location.Y+i), float64(MapHeight-1))),
					},
					Health: NewDefaultHealth(),
				}
				mapEntity.Matrix[p1.Location.Y][p1.Location.X].JoinArea(&p1)
			} else {
				p2 := Player{
					Location: &Location{
						X: int(math.Max(float64(location.X-i), 0)),
						Y: int(math.Max(float64(location.Y-i), 0)),
					},
					Health: NewDefaultHealth(),
				}
				mapEntity.Matrix[p2.Location.Y][p2.Location.X].JoinArea(&p2)
			}

		}
	} else {
		switch dir {
		case dirNorth:
			for i := 1; i < playerCount+1; i++ {
				p := Player{
					Location: &Location{
						X: location.X,
						Y: int(math.Max(float64(location.Y-i), 0)),
					},
					Health: NewDefaultHealth(),
				}
				mapEntity.Matrix[p.Location.Y][p.Location.X].JoinArea(&p)
			}
		case dirEast:
			for i := 1; i < playerCount+1; i++ {
				p := Player{
					Location: &Location{
						X: int(math.Min(float64(location.X+i), float64(30-1))),
						Y: location.Y,
					},
					Health: NewDefaultHealth(),
				}
				mapEntity.Matrix[p.Location.Y][p.Location.X].JoinArea(&p)
			}
		case dirSouth:
			for i := 1; i < playerCount+1; i++ {
				p := Player{
					Location: &Location{
						X: location.X,
						Y: int(math.Min(float64(location.Y+i), float64(30-1))),
					},
					Health: NewDefaultHealth(),
				}
				mapEntity.Matrix[p.Location.Y][p.Location.X].JoinArea(&p)
			}
		case dirWest:
			for i := 1; i < playerCount+1; i++ {
				p := Player{
					Location: &Location{
						X: int(math.Max(float64(location.X-i), 0)),
						Y: location.Y,
					},
					Health: NewDefaultHealth(),
				}
				mapEntity.Matrix[p.Location.Y][p.Location.X].JoinArea(&p)
			}
		}
	}
	return mapEntity
}

func newPlayer(health, kills, deaths, y, x int, isDefending bool) *Player {
	return &Player{
		Health:      NewHealth(health),
		Kills:       kills,
		Deaths:      deaths,
		Location:    newLocation(y, x),
		IsDefending: isDefending,
		Rl:          NewOpLimitations(),
	}
}
