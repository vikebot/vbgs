package vbge

const (
	angleLeft  = "left"
	angleRight = "right"

	dirNorth = "north"
	dirEast  = "east"
	dirSouth = "south"
	dirWest  = "west"

	// DirNorth is the exported direction north
	DirNorth = "north"
	// DirEast is the exported direction north
	DirEast = "east"
	// DirSouth is the exported direction north
	DirSouth = "south"
	// DirWest is the exported direction north
	DirWest = "west"

	blockSwamp     = "swamp"
	blockStonetile = "stonetile"
	blockDirt      = "dirt"
	blockLightDirt = "dirt_light"
	blockGrass     = "grass"
	blockLava      = "lava"
	blockLavarock  = "lavarock"
	blockWater     = "water"
	blockEndOfMap  = "endofmap"
	blockFog       = "fog"

	humanArmoredArcherMale = "male_armored_archer"
	humanKnightMale        = "male_night"
	humanNinjaMale         = "male_ninja"
	humanThugMale          = "male_thug"
)

var (
	// MapProps are the dimensions of the complete map used for all players

	// MapWidth is the length in x-direction
	MapWidth = 11
	// MapHeight is the length in y-direction
	MapHeight = 11
	// HalfmapWidth is the half value of 'MapWidth'
	HalfmapWidth = MapWidth / 2

	// HalfmapHeight is the half value of 'MapHeight'
	HalfmapHeight = MapHeight / 2

	// RenderProps describe the area that gets rendered in vbwatch. hr stands
	// for halfRender which is the full render are minus the block we are
	// standing devided by two

	// RenderWidth is the visible area for a player in x-direction
	RenderWidth = 11
	// RenderHeight is the visible area for a player in y-direction
	RenderHeight = 11
	// HrWidth is the half value of 'RenderWidth'
	HrWidth = (RenderWidth - 1) / 2
	// HrHeight is the half value of 'RenderHeight'
	HrHeight = (RenderHeight - 1) / 2

	// MaxHealth is the default value of a player's character when he has
	// full health points
	MaxHealth = 100

	// defaultDmg is the damage which a player makes when the bot is
	// attacking somebody
	defaultDmg = 10

	// RadarRadius is the radius around a player's location we use to collect
	// counter metrics
	radarRadius = 10
)

// SetMapDimensions sets the default map dimensions (e.g. width and health)
func SetMapDimensions(width, height int) {
	MapWidth = width
	MapHeight = height
	HalfmapWidth = width / 2
	HalfmapHeight = height / 2
}

// IsAngle determines whether the `angleCandidate` is actually a valid angle
func IsAngle(angleCandidate string) bool {
	if angleCandidate == angleLeft || angleCandidate == angleRight {
		return true
	}
	return false
}

// IsDir determines whether the `dirCandidate` is actually a valid direction
func IsDir(dirCandidate string) bool {
	if dirCandidate == dirNorth || dirCandidate == dirEast || dirCandidate == dirSouth || dirCandidate == dirWest {
		return true
	}
	return false
}

// IsBlocktype determines whether the `blocktypeCandidate` is actually a valid blocktype
func IsBlocktype(blocktypeCandidate string) bool {
	btc := blocktypeCandidate
	if btc == blockSwamp || btc == blockStonetile || btc == blockDirt || btc == blockGrass || btc == blockLava || btc == blockLavarock || btc == blockWater || btc == blockEndOfMap || btc == blockFog {
		return true
	}
	return false
}

// IsDistance determines wether the `distanceCanidate` is actually a valid distance
func IsDistance(distanceCandidate int) bool {
	return distanceCandidate > 0 && distanceCandidate < 16
}
