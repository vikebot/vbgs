package vbge

// PlayerResp is the response value of player for vbwatch
type PlayerResp struct {
	GRID          string   `json:"grid"`
	Health        int      `json:"health"`
	CharacterType string   `json:"ct"`
	WatchDir      string   `json:"watchdir"`
	Location      Location `json:"location"`
}

// EntityResp is the response value of a specific location
// with the values 'Blocktype' and 'Player'
type EntityResp struct {
	Blocktype string      `json:"bt"`
	Player    *PlayerResp `json:"p"`
}

// ViewableMapentity is nearly the same like 'MapEntity' but
// only the for a specific player visible matrix including
// other players and blocktypes
type ViewableMapentity struct {
	Height int             `json:"height"`
	Witdh  int             `json:"width"`
	Matrix [][]*EntityResp `json:"matrix"`
}

// GetViewableMapentity returns a Mapentity for a specific player
func GetViewableMapentity(width, height, userID int, game *Battle, sync bool) (viewableMapentity *ViewableMapentity, err error) {
	viewableMatrix := make([][]*EntityResp, height)
	for i := range viewableMatrix {
		viewableMatrix[i] = make([]*EntityResp, width)
	}

	gameMatrixMe := game.Map.GetMatrixSectionFromMapentity(width, height, game.Players[userID], sync)

	var me = &MapEntity{
		Height: height,
		Width:  width,
		Matrix: gameMatrixMe,
	}

	viewableMatrix = fillMatrixWithER(viewableMatrix, me, "")

	viewableMapentity = &ViewableMapentity{
		Height: height,
		Witdh:  width,
		Matrix: viewableMatrix,
	}

	return viewableMapentity, err
}

// GetNewLineMapentity returns a new mapentity with a size of 1x11 or 11x1 depends on
// moving direction
func GetNewLineMapentity(width, userID int, game *Battle, direction string) *ViewableMapentity {
	newLineMe := game.Map.GetNewLineFromMapEntity(width, game.Players[userID], direction)

	viewableMatrix := make([][]*EntityResp, len(newLineMe))
	for i := range viewableMatrix {
		viewableMatrix[i] = make([]*EntityResp, len(newLineMe[0]))
	}

	var me = &MapEntity{
		Height: len(viewableMatrix),
		Width:  len(viewableMatrix[0]),
		Matrix: newLineMe,
	}

	viewableMatrix = fillMatrixWithER(viewableMatrix, me, direction)

	return &ViewableMapentity{
		Height: len(newLineMe),
		Witdh:  len(newLineMe[0]),
		Matrix: viewableMatrix,
	}
}

// fillMatrixWithER fills an given matrix of EntityResp with the values
// from a 2D-Slice of Blockentity in an EntityResponse (ER)
func fillMatrixWithER(matrix [][]*EntityResp, me *MapEntity, direction string) [][]*EntityResp {

	var loc *Location

	if me.Height == RenderHeight && me.Width == RenderWidth {
		loc = me.Matrix[HrHeight][HrWidth].Resident.Location
	}

	for yi := 0; yi < len(matrix); yi++ {
		for xi := 0; xi < len(matrix[0]); xi++ {
			var player *PlayerResp

			if me.Matrix[yi][xi].HasResident() {
				var resident = me.Matrix[yi][xi].Resident

				if loc == nil {
					loc = &Location{}
					switch direction {
					case dirNorth:
						loc.X = resident.Location.X + (HrWidth - xi)
						loc.Y = resident.Location.Y + HrHeight
						break
					case dirEast:
						loc.X = resident.Location.X - HrWidth
						loc.Y = resident.Location.Y + (HrHeight - yi)
						break
					case dirSouth:
						loc.X = resident.Location.X + (HrWidth - xi)
						loc.Y = resident.Location.Y - HrHeight
						break
					case dirWest:
						loc.X = resident.Location.X + HrWidth
						loc.Y = resident.Location.Y + (HrHeight - yi)
						break
					}
				}

				player = &PlayerResp{
					GRID:          resident.GRenderID,
					Health:        resident.Health.HealthSynced(),
					CharacterType: resident.CharacterType,
					WatchDir:      resident.WatchDir,
					Location:      *resident.Location.RelativeFrom(loc),
				}
			}

			matrix[yi][xi] = &EntityResp{
				Blocktype: me.Matrix[yi][xi].Blocktype,
				Player:    player,
			}
		}
	}
	return matrix
}
