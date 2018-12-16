package vbge

import (
	"sync"

	"github.com/vikebot/vbcore"
)

// MapEntity describes a single map instance.
type MapEntity struct {
	Height   int
	Width    int
	Matrix   [][]*BlockEntity
	SyncRoot sync.Mutex
}

// NewMapEntity allocates memory for a new map with the size specified by the
// `width` and `height` parameter.
func NewMapEntity(width, height int) *MapEntity {
	matrix := make([][]*BlockEntity, height)
	for i := range matrix {
		matrix[i] = make([]*BlockEntity, width)
	}

	for yi := 0; yi < height; yi++ {
		for xi := 0; xi < width; xi++ {
			matrix[yi][xi] = &BlockEntity{
				Blocktype: blockLightDirt,
			}
		}
	}

	return &MapEntity{
		Height: height,
		Width:  width,
		Matrix: matrix,
	}
}

// PInRenderArea returns all the players that are inside the render area
// around the passend location. Result is returned as a `NotifyGroup`. This
// function isn't safe for concurrent use.
func (me *MapEntity) PInRenderArea(l Location) NotifyGroup {
	startX := vbcore.MaxInt(0, l.X-HrWidth)
	endX := vbcore.MinInt(MapWidth-1, l.X+HrWidth)
	startY := vbcore.MaxInt(0, l.Y-HrHeight)
	endY := vbcore.MinInt(MapHeight-1, l.Y+HrHeight)

	inarea := []*Player{}

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if me.Matrix[y][x].HasResident() {
				inarea = append(inarea, me.Matrix[y][x].Resident)
			}
		}
	}

	return inarea
}

// PInRenderAreaCombined returns all the players that are inside the render
// area around the passend locations. The minimum rectangle containing both
// locations is calculated and searched for players. The Result is returned
// as a `NotifyGroup`. This function isn't safe for concurrent use.
func (me *MapEntity) PInRenderAreaCombined(oldL *Location, newL *Location) NotifyGroup {
	startX := vbcore.MinInt(oldL.X-HrWidth, newL.X-HrWidth)
	startX = vbcore.MaxInt(0, startX)

	endX := vbcore.MaxInt(oldL.X+HrWidth, newL.X+HrWidth)
	endX = vbcore.MinInt(MapWidth-1, endX)

	startY := vbcore.MinInt(oldL.Y-HrHeight, newL.Y-HrHeight)
	startY = vbcore.MaxInt(0, startY)

	endY := vbcore.MaxInt(oldL.Y+HrHeight, newL.Y+HrHeight)
	endY = vbcore.MinInt(MapHeight-1, endY)

	inarea := []*Player{}
	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if me.Matrix[y][x].HasResident() {
				inarea = append(inarea, me.Matrix[y][x].Resident)
			}
		}
	}

	return inarea
}

// PInMatrix returns true if any player is in the matrix of the given
// MapEntity. If a player occurs, true is returned
func (me *MapEntity) PInMatrix() bool {
	for y := 0; y < len(me.Matrix); y++ {
		for x := 0; x < len(me.Matrix[0]); x++ {
			if me.Matrix[y][x].HasResident() {
				return true
			}
		}
	}
	return false
}

// LeaveArea is calling (*BlockEntity).LeaveArea() to set a location after
// moving to nil, so the location isn't blocked anymore.
func (me *MapEntity) LeaveArea(p *Player, l Location) {
	me.Matrix[l.Y][l.X].LeaveArea()
}

// GetMatrixSectionFromMapentity returns the 'viewable' matrix of a given player
func (me *MapEntity) GetMatrixSectionFromMapentity(width, height int, p *Player, sync bool) (gameMatrix [][]*BlockEntity) {
	l := p.Location

	startX := l.X - HrWidth
	endX := l.X + HrHeight
	startY := l.Y - HrHeight
	endY := l.Y + HrWidth

	gameMatrix = make([][]*BlockEntity, height)
	for i := range gameMatrix {
		gameMatrix[i] = make([]*BlockEntity, width)
	}

	if !sync {
		gameMatrix = me.parseMeIntoNewMatrix(startX, endX, startY, endY, gameMatrix)
	} else {
		gameMatrix = me.parseMeIntoNewMatrixSynced(startX, endX, startY, endY, gameMatrix)
	}

	return gameMatrix
}

// GetNewLineFromMapEntity is returning a 11x1 dim (RenderWidth), usually called
// when a player is moving
func (me *MapEntity) GetNewLineFromMapEntity(width int, p *Player, direction string) (gameLine [][]*BlockEntity) {
	l := p.Location

	var startX, endX, startY, endY, height int

	switch direction {
	case dirNorth:
		startX = l.X - HrWidth
		endX = l.X + HrWidth
		startY = l.Y - HrHeight
		endY = startY
		height = 1
	case dirEast:
		startX = l.X + HrWidth
		endX = startX
		startY = l.Y - HrHeight
		endY = l.Y + HrHeight
		height = width
		width = 1
	case dirSouth:
		startX = l.X - HrWidth
		endX = l.X + HrWidth
		startY = l.Y + HrHeight
		endY = startY
		height = 1
	case dirWest:
		startX = l.X - HrWidth
		endX = startX
		startY = l.Y - HrHeight
		endY = l.Y + HrHeight
		height = width
		width = 1
	}

	gameLine = make([][]*BlockEntity, height)
	for i := range gameLine {
		gameLine[i] = make([]*BlockEntity, width)
	}

	gameLine = me.parseMeIntoNewMatrix(startX, endX, startY, endY, gameLine)

	return gameLine
}

// parseMeIntoNewMatrix parses the values of a mapentity, with given start and end
// values into a new Matrix of a specific size
func (me *MapEntity) parseMeIntoNewMatrix(startX, endX, startY, endY int, matrix [][]*BlockEntity) (parsedMatrix [][]*BlockEntity) {
	var x, y int

	for yi := startY; yi <= endY; yi++ {
		x = 0
		for xi := startX; xi <= endX; xi++ {
			if yi < 0 || yi >= MapHeight || xi < 0 || xi >= MapWidth {
				matrix[y][x] = &BlockEntity{
					Blocktype: blockEndOfMap,
					Resident:  nil,
				}
			} else {
				matrix[y][x] = me.Matrix[yi][xi]
			}
			x++
		}
		y++
	}

	return matrix
}

func (me *MapEntity) parseMeIntoNewMatrixSynced(startX, endX, startY, endY int, matrix [][]*BlockEntity) (parsedMatrix [][]*BlockEntity) {
	me.SyncRoot.Lock()
	defer me.SyncRoot.Unlock()

	parsedMatrix = me.parseMeIntoNewMatrix(startX, endX, startY, endY, matrix)
	return parsedMatrix
}
