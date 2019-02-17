package citygame

import (
	"io"
	"sort"

	"git.fleta.io/fleta/common/util"
)

// consts
const (
	GTileSize = 16
)

// AreaType is a area type of the target tile
type AreaType uint8

// types
const (
	EmptyAreaType       = AreaType(0)
	CommercialAreaType  = AreaType(1)
	IndustrialAreaType  = AreaType(2)
	ResidentialAreaType = AreaType(3)
	EndOfAreaType       = AreaType(4)
)

// Tile reprents a information of the target tile
type Tile struct {
	AreaType    AreaType
	Level       uint8
	BuildHeight uint32
}

// WriteTo is a serialization function
func (tile *Tile) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint8(w, uint8(tile.AreaType)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, tile.Level); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, tile.BuildHeight); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tile *Tile) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tile.AreaType = AreaType(v)
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tile.Level = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		tile.BuildHeight = v
	}
	return read, nil
}

// NewTile returns a Tile
func NewTile(AreaType AreaType, BuildHeight uint32) *Tile {
	tile := &Tile{
		AreaType:    AreaType,
		Level:       1,
		BuildHeight: BuildHeight,
	}
	return tile
}

// Resource is a status of the game data
type Resource struct {
	Balance       uint64
	PowerRemained uint32
	PowerProvided uint32
	ManRemained   uint32
	ManProvided   uint32
}

// GameData stores all data of the game
type GameData struct {
	PointBalance uint64
	PointHeight  uint32
	Tiles        []*Tile
}

// NewGameData returns a GameData
func NewGameData(TargetHeight uint32) *GameData {
	gd := &GameData{
		Tiles:        make([]*Tile, GTileSize*GTileSize),
		PointBalance: 900,
		PointHeight:  TargetHeight,
	}
	return gd
}

// Count returns the building count by a area type
func (gd *GameData) Count() map[AreaType]int {
	CountMap := map[AreaType]int{}
	for _, tile := range gd.Tiles {
		if tile != nil {
			CountMap[tile.AreaType]++
		}
	}
	return CountMap
}

// Resource returns remained resources by the target height
func (gd *GameData) Resource(TargetHeight uint32) *Resource {
	ForwardHeight := TargetHeight - gd.PointHeight
	base := &Resource{
		Balance:       4,
		PowerRemained: 5,
		PowerProvided: 5,
		ManRemained:   3,
		ManProvided:   3,
	}
	used := &Resource{}
	provide := &Resource{
		Balance:       gd.PointBalance + uint64(base.Balance/2)*uint64(ForwardHeight),
		PowerRemained: base.PowerRemained,
		ManRemained:   base.ManRemained,
	}
	for _, tile := range gd.Tiles {
		if tile != nil {
			bd := GBuildingDefine[tile.AreaType][tile.Level-1]
			used.ManRemained += bd.ManUsage
			used.PowerRemained += bd.PowerUsage

			ConstructionHeight := tile.BuildHeight + bd.BuildTime*2
			if TargetHeight < ConstructionHeight {
				if tile.Level == 1 {
					continue
				}
				bd = GBuildingDefine[tile.AreaType][tile.Level-2]
			}
			switch tile.AreaType {
			case CommercialAreaType:
				if TargetHeight > ConstructionHeight {
					if ConstructionHeight <= gd.PointHeight {
						provide.Balance += uint64(bd.Output/2) * uint64(ForwardHeight)
					} else {
						if tile.BuildHeight <= gd.PointHeight {
							provide.Balance += uint64(bd.Output/2) * uint64(ForwardHeight-(ConstructionHeight-gd.PointHeight))
						} else {
							provide.Balance += uint64(bd.Output/2) * uint64(TargetHeight-ConstructionHeight)
							if tile.Level > 1 {
								prevbd := GBuildingDefine[tile.AreaType][tile.Level-2]
								provide.Balance += uint64(prevbd.Output/2) * uint64(tile.BuildHeight-gd.PointHeight)
							}
						}
					}
				}
			case IndustrialAreaType:
				provide.PowerRemained += bd.Output
			case ResidentialAreaType:
				provide.ManRemained += bd.Output
			}
		}
	}
	res := &Resource{
		Balance:       provide.Balance,
		PowerRemained: provide.PowerRemained - used.PowerRemained,
		PowerProvided: provide.PowerRemained,
		ManRemained:   provide.ManRemained - used.ManRemained,
		ManProvided:   provide.ManRemained,
	}
	return res
}

// UpdatePoint updates the last point of the game data
func (gd *GameData) UpdatePoint(TargetHeight uint32, Balance uint64) {
	gd.PointHeight = TargetHeight
	gd.PointBalance = Balance
}

// WriteTo is a serialization function
func (gd *GameData) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint64(w, gd.PointBalance); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, gd.PointHeight); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	coords := []int{}
	tileMap := map[int]*Tile{}
	for coord, tile := range gd.Tiles {
		if tile != nil {
			coords = append(coords, coord)
			tileMap[coord] = tile
		}
	}
	sort.Ints(coords)

	if n, err := util.WriteUint32(w, uint32(len(coords))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, coord := range coords {
			if n, err := util.WriteUint32(w, uint32(coord)); err != nil {
				return wrote, err
			} else {
				wrote += n
				tile := tileMap[coord]
				if n, err := tile.WriteTo(w); err != nil {
					return wrote, err
				} else {
					wrote += n
				}
			}
		}
	}

	return wrote, nil
}

// ReadFrom is a deserialization function
func (gd *GameData) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		gd.PointBalance = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		gd.PointHeight = v
	}

	if Len, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		gd.Tiles = make([]*Tile, GTileSize*GTileSize)
		for i := 0; i < int(Len); i++ {
			if coord, n, err := util.ReadUint32(r); err != nil {
				return read, err
			} else {
				read += n
				tile := &Tile{}
				if n, err := tile.ReadFrom(r); err != nil {
					return read, err
				} else {
					read += n
					gd.Tiles[coord] = tile
				}
			}
		}
	}
	return read, nil
}

// BuildingDefine defines a building information
type BuildingDefine struct {
	CostUsage     uint64 `json:"cost_usage"`      // all
	BuildTime     uint32 `json:"build_time"`      // all
	Output        uint32 `json:"output"`          // all
	ManUsage      uint32 `json:"man_usage"`       // commerial, industrial
	PowerUsage    uint32 `json:"power_usage"`     // commerial, residential
	AccManUsage   uint32 `json:"acc_man_usage"`   // commerial, industrial
	AccPowerUsage uint32 `json:"acc_power_usage"` // commerial, residential
}

func init() {
	for _, bds := range GBuildingDefine {
		var AccManUsage uint32
		var AccPowerUsage uint32
		for _, bd := range bds {
			AccManUsage += bd.ManUsage
			AccPowerUsage += bd.PowerUsage
			bd.AccManUsage = AccManUsage
			bd.AccPowerUsage = AccPowerUsage
		}
	}
}

// GBuildingDefine is ingame construction data
var GBuildingDefine = map[AreaType][]*BuildingDefine{
	CommercialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  1,
			BuildTime:  1,
			Output:     4,
			ManUsage:   2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  2,
			BuildTime:  2,
			Output:     10,
			ManUsage:   3,
			PowerUsage: 4,
		},
		&BuildingDefine{
			CostUsage:  3,
			BuildTime:  3,
			Output:     24,
			ManUsage:   8,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  4,
			BuildTime:  4,
			Output:     64,
			ManUsage:   40,
			PowerUsage: 30,
		},
		&BuildingDefine{
			CostUsage:  5,
			BuildTime:  5,
			Output:     160,
			ManUsage:   200,
			PowerUsage: 80,
		},
		&BuildingDefine{
			CostUsage:  6,
			BuildTime:  6,
			Output:     1600,
			ManUsage:   4000,
			PowerUsage: 1500,
		},
	},
	IndustrialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage: 1,
			BuildTime: 1,
			Output:    5,
			ManUsage:  1,
		},
		&BuildingDefine{
			CostUsage: 2,
			BuildTime: 2,
			Output:    14,
			ManUsage:  2,
		},
		&BuildingDefine{
			CostUsage: 3,
			BuildTime: 3,
			Output:    96,
			ManUsage:  8,
		},
		&BuildingDefine{
			CostUsage: 4,
			BuildTime: 4,
			Output:    390,
			ManUsage:  54,
		},
		&BuildingDefine{
			CostUsage: 5,
			BuildTime: 5,
			Output:    1440,
			ManUsage:  300,
		},
		&BuildingDefine{
			CostUsage: 6,
			BuildTime: 6,
			Output:    33000,
			ManUsage:  6100,
		},
	},
	ResidentialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  1,
			BuildTime:  1,
			Output:     3,
			PowerUsage: 2,
		},
		&BuildingDefine{
			CostUsage:  2,
			BuildTime:  2,
			Output:     10,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  3,
			BuildTime:  3,
			Output:     64,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  4,
			BuildTime:  4,
			Output:     564,
			PowerUsage: 35,
		},
		&BuildingDefine{
			CostUsage:  5,
			BuildTime:  5,
			Output:     4000,
			PowerUsage: 100,
		},
		&BuildingDefine{
			CostUsage:  6,
			BuildTime:  6,
			Output:     101000,
			PowerUsage: 1800,
		},
	},
}

/*
var GBuildingDefine = map[AreaType][]*BuildingDefine{
	CommercialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  400,
			BuildTime:  30,
			Output:     4,
			ManUsage:   2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  2400,
			BuildTime:  140,
			Output:     10,
			ManUsage:   3,
			PowerUsage: 4,
		},
		&BuildingDefine{
			CostUsage:  12000,
			BuildTime:  700,
			Output:     24,
			ManUsage:   8,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  60000,
			BuildTime:  3500,
			Output:     64,
			ManUsage:   40,
			PowerUsage: 30,
		},
		&BuildingDefine{
			CostUsage:  300000,
			BuildTime:  18000,
			Output:     160,
			ManUsage:   200,
			PowerUsage: 80,
		},
		&BuildingDefine{
			CostUsage:  6000000,
			BuildTime:  86400,
			Output:     1600,
			ManUsage:   4000,
			PowerUsage: 1500,
		},
	},
	IndustrialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage: 200,
			BuildTime: 60,
			Output:    5,
			ManUsage:  1,
		},
		&BuildingDefine{
			CostUsage: 1700,
			BuildTime: 200,
			Output:    14,
			ManUsage:  2,
		},
		&BuildingDefine{
			CostUsage: 12000,
			BuildTime: 700,
			Output:    96,
			ManUsage:  8,
		},
		&BuildingDefine{
			CostUsage: 80000,
			BuildTime: 2700,
			Output:    390,
			ManUsage:  54,
		},
		&BuildingDefine{
			CostUsage: 450000,
			BuildTime: 12000,
			Output:    1440,
			ManUsage:  300,
		},
		&BuildingDefine{
			CostUsage: 9100000,
			BuildTime: 57000,
			Output:    33000,
			ManUsage:  6100,
		},
	},
	ResidentialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  300,
			BuildTime:  45,
			Output:     3,
			PowerUsage: 2,
		},
		&BuildingDefine{
			CostUsage:  2000,
			BuildTime:  170,
			Output:     10,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  12000,
			BuildTime:  700,
			Output:     64,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  66000,
			BuildTime:  3200,
			Output:     564,
			PowerUsage: 35,
		},
		&BuildingDefine{
			CostUsage:  360000,
			BuildTime:  15000,
			Output:     4000,
			PowerUsage: 100,
		},
		&BuildingDefine{
			CostUsage:  7200000,
			BuildTime:  72000,
			Output:     101000,
			PowerUsage: 1800,
		},
	},
}
*/
